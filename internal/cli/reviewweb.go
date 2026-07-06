package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runReviewWeb is the same triage queue as the terminal loop, served as a
// one-shot localhost page. It is NOT a daemon: the server lives exactly as
// long as the command, binds 127.0.0.1 only, and all state stays in
// .kervo/ — the design guarantees hold (decision 01KWTVKV).
func runReviewWeb(dir, actorFlag string) error {
	store := jsonl.Open(dir)
	folder, err := replayFolder(store)
	if err != nil {
		return err
	}
	var queue []trust.Observation
	for _, o := range folder.Observations() {
		if o.State == trust.Generated || o.Conflict {
			queue = append(queue, o)
		}
	}
	if len(queue) == 0 {
		fmt.Println("Review queue is empty — nothing awaits judgment.")
		return nil
	}

	srv := newReviewServer(store, dir, actorFlag, queue)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String()
	fmt.Printf("Review UI: %s — %d awaiting judgment (the page's Finish button ends this command)\n", url, len(queue))
	openBrowser(url)

	httpSrv := &http.Server{Handler: srv.handler()}
	go func() {
		<-srv.done
		_ = httpSrv.Shutdown(context.Background())
	}()
	if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	fmt.Printf("Judged %d of %d. `kervo compile` refreshes the artifact.\n", srv.judgedCount(), len(queue))
	return nil
}

type webItem struct {
	trust.Observation
	ShortID string
	Judged  bool
}

type reviewServer struct {
	mu    sync.Mutex
	store *jsonl.Store
	dir   string
	actor string
	items []*webItem
	byID  map[string]*webItem
	done  chan struct{}
	once  sync.Once
}

func newReviewServer(store *jsonl.Store, dir, actorFlag string, queue []trust.Observation) *reviewServer {
	s := &reviewServer{
		store: store, dir: dir, actor: actorFlag,
		byID: map[string]*webItem{}, done: make(chan struct{}),
	}
	for _, o := range queue {
		it := &webItem{Observation: o, ShortID: shortID(o.ID)}
		s.items = append(s.items, it)
		s.byID[o.ID] = it
	}
	return s
}

func (s *reviewServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.page)
	mux.HandleFunc("/judge", s.judge)
	mux.HandleFunc("/quit", s.quit)
	return mux
}

func (s *reviewServer) judgedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, it := range s.items {
		if it.Judged {
			n++
		}
	}
	return n
}

func (s *reviewServer) page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pageTmpl.Execute(w, struct{ Items []*webItem }{s.items})
}

// judge appends the same transition event the CLI writes — the page is
// sugar over the primitive, exactly like the terminal loop.
func (s *reviewServer) judge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ ID, To, Reason string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	target := trust.State(req.To)
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.byID[req.ID]
	if !ok || it.Judged {
		http.Error(w, "unknown or already judged", http.StatusConflict)
		return
	}
	if !trust.CanTransition(it.State, target) {
		http.Error(w, fmt.Sprintf("%s → %s is not a legal transition", it.State, target), http.StatusConflict)
		return
	}
	if err := appendTransition(s.store, s.dir, s.actor, it.Observation, target, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	it.Judged = true
	all := true
	for _, x := range s.items {
		if !x.Judged {
			all = false
			break
		}
	}
	if all {
		s.once.Do(func() { close(s.done) })
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *reviewServer) quit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	s.once.Do(func() { close(s.done) })
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return
	}
	_ = exec.Command(cmd, url).Start()
}

var pageTmpl = template.Must(template.New("review").Parse(`<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>kervo review</title>
<style>
:root{--bg:#fff;--fg:#1a1a1a;--muted:#6a6a6a;--card:#f6f6f4;--line:#e2e2de;--accent:#0a7d4f}
@media(prefers-color-scheme:dark){:root{--bg:#111;--fg:#e8e8e6;--muted:#9a9a96;--card:#1c1c1a;--line:#2e2e2a;--accent:#3fb37f}}
body{margin:0 auto;max-width:44rem;padding:2rem 1rem;background:var(--bg);color:var(--fg);font:15px/1.55 ui-sans-serif,system-ui,sans-serif}
h1{font-size:1.15rem;margin:0 0 1.2rem}
.item{background:var(--card);border:1px solid var(--line);border-radius:8px;padding:.9rem 1rem;margin-bottom:.8rem}
.item.judged,.item.skipped{opacity:.45}
.meta{font-size:.78rem;color:var(--muted);margin-bottom:.4rem}
.conflict{color:#c0392b;font-weight:600}
.body{white-space:pre-wrap;margin-bottom:.4rem}
.evidence{font-size:.82rem;color:var(--muted);border-left:3px solid var(--line);padding-left:.6rem;margin-bottom:.5rem}
.row{display:flex;gap:.4rem;flex-wrap:wrap;align-items:center}
input{flex:1;min-width:10rem;background:var(--bg);color:var(--fg);border:1px solid var(--line);border-radius:6px;padding:.35rem .5rem;font:inherit;font-size:.85rem}
button{border:1px solid var(--line);background:var(--bg);color:var(--fg);border-radius:6px;padding:.35rem .7rem;font:inherit;font-size:.85rem;cursor:pointer}
button.v{border-color:var(--accent);color:var(--accent);font-weight:600}
button:disabled{cursor:default;opacity:.5}
.verdict{font-size:.8rem;color:var(--accent);font-weight:600}
#finish{margin-top:1rem;width:100%;padding:.6rem}
</style></head><body>
<h1>kervo review — <span id="count">{{len .Items}}</span> awaiting judgment</h1>
{{range .Items}}
<div class="item" id="item-{{.ID}}">
  <div class="meta">{{.ShortID}} · {{.Type}} · {{.State}} · {{.Actor}}{{if .Conflict}} · <span class="conflict">⚠ conflict</span>{{end}}</div>
  <div class="body">{{.Body}}</div>
  {{if .Evidence}}<div class="evidence">evidence: {{.Evidence}}</div>{{end}}
  <div class="row">
    <input id="reason-{{.ID}}" placeholder="reason (optional)">
    <button class="v" onclick="judge('{{.ID}}','verified')">verify</button>
    <button onclick="judge('{{.ID}}','stale')">stale</button>
    <button onclick="judge('{{.ID}}','deprecated')">deprecate</button>
    <button onclick="skip('{{.ID}}')">skip</button>
    <span class="verdict" id="verdict-{{.ID}}"></span>
  </div>
</div>
{{end}}
<button id="finish" onclick="finish()">Finish — end the session</button>
<script>
let left = document.querySelectorAll('.item').length;
async function judge(id, to){
  const reason = document.getElementById('reason-'+id).value;
  const r = await fetch('/judge',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({ID:id,To:to,Reason:reason})});
  if(!r.ok){ alert(await r.text()); return; }
  settle(id, '→ '+to);
}
function skip(id){ settle(id, 'skipped'); }
function settle(id, verdict){
  const el = document.getElementById('item-'+id);
  el.classList.add('judged');
  el.querySelectorAll('button,input').forEach(x=>x.disabled=true);
  document.getElementById('verdict-'+id).textContent = verdict;
  left--; document.getElementById('count').textContent = left;
  if(left===0) done();
}
function finish(){ fetch('/quit',{method:'POST'}).finally(done); }
function done(){ document.body.innerHTML='<h1>Done — run <code>kervo compile</code>, then close this tab.</h1>'; }
</script>
</body></html>`))
