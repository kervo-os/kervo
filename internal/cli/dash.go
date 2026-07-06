package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runDash is the fleet control tower: every registered workspace on this
// machine, one page — pending judgments, trust-state bars, last activity,
// and inline triage that writes to each repo's own ledger. Truth stays
// per-repo in git; the dashboard is a derived local lens with no state of
// its own, and it dies with the command (goal 01KWTWZX).
func runDash(args []string) error {
	fs := newFlagSet("dash")
	actor := fs.String("actor", "", `who is judging (default "human:<git user.name>")`)
	if err := fs.Parse(args); err != nil {
		return err
	}
	paths := registeredWorkspaces()
	if len(paths) == 0 {
		fmt.Println("no kervo workspaces registered on this machine — run `kervo compile` in a repo first")
		return nil
	}
	srv, err := newDashServer(paths, *actor)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String()
	fmt.Printf("kervo dash: %s — %d workspaces, %d awaiting judgment (Finish on the page ends this command)\n",
		url, len(paths), srv.pendingTotal())
	openBrowser(url)

	httpSrv := &http.Server{Handler: srv.handler()}
	go func() {
		<-srv.done
		_ = httpSrv.Shutdown(context.Background())
	}()
	if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	fmt.Printf("Judged %d. Run `kervo compile` in the affected repos to refresh their artifacts.\n", srv.judgedTotal())
	return nil
}

type dashItem struct {
	ID, ShortID, Type, State, Actor, Body string
	Evidence                              string `json:",omitempty"`
	Conflict                              bool   `json:",omitempty"`
}

type dashRepo struct {
	Path, Name, Lang string
	DisplayPath      string // Path with $HOME shortened to ~ — for humans
	Events           int
	Counts           map[string]int
	LastEvent        string // RFC3339, "" if the ledger is empty
	Items            []dashItem
}

type dashServer struct {
	mu     sync.Mutex
	actor  string
	repos  []*dashRepo
	byPath map[string]*dashRepo
	judged int
	done   chan struct{}
	once   sync.Once
}

func newDashServer(paths []string, actorFlag string) (*dashServer, error) {
	s := &dashServer{actor: actorFlag, byPath: map[string]*dashRepo{}, done: make(chan struct{})}
	for _, p := range paths {
		store := jsonl.Open(p)
		folder := trust.NewFolder()
		events := 0
		var last time.Time
		if err := store.Replay(context.Background(), "", func(e event.Event) error {
			events++
			if e.At.After(last) {
				last = e.At
			}
			folder.Add(e)
			return nil
		}); err != nil {
			return nil, fmt.Errorf("dash: replay %s: %w", p, err)
		}
		repo := &dashRepo{
			Path: p, Name: filepath.Base(p), Lang: workspaceLang(p),
			DisplayPath: displayPath(p),
			Events:      events, Counts: map[string]int{},
			// Never nil: a clear repo must marshal as [], not null — the
			// page reads .Items.length on every repo before rendering.
			Items: []dashItem{},
		}
		if !last.IsZero() {
			repo.LastEvent = last.UTC().Format(time.RFC3339)
		}
		for _, o := range folder.Observations() {
			repo.Counts[string(o.State)]++
			if o.State == trust.Generated || o.Conflict {
				repo.Items = append(repo.Items, dashItem{
					ID: o.ID, ShortID: shortID(o.ID), Type: o.Type, State: string(o.State),
					Actor: o.Actor, Body: o.Body, Evidence: o.Evidence, Conflict: o.Conflict,
				})
			}
		}
		s.repos = append(s.repos, repo)
		s.byPath[p] = repo
	}
	return s, nil
}

func displayPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || !strings.HasPrefix(p, home) {
		return p
	}
	return "~" + strings.TrimPrefix(p, home)
}

func workspaceLang(dir string) string {
	raw, err := os.ReadFile(filepath.Join(dir, ".kervo", "lang"))
	if err != nil {
		return "en"
	}
	if l := strings.TrimSpace(string(raw)); l != "" {
		return l
	}
	return "en"
}

func (s *dashServer) pendingTotal() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, r := range s.repos {
		n += len(r.Items)
	}
	return n
}

func (s *dashServer) judgedTotal() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.judged
}

func (s *dashServer) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.page)
	mux.HandleFunc("/judge", s.judge)
	mux.HandleFunc("/quit", s.quit)
	return mux
}

func (s *dashServer) page(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.mu.Lock()
	fleet, err := json.Marshal(s.repos) // json escapes <>& — safe inside <script>
	s.mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = dashTmpl.Execute(w, struct{ FleetJS template.JS }{template.JS(fleet)})
}

// judge writes the human's decision to the TARGET repo's own ledger — the
// dashboard never owns state, it only routes judgments home.
func (s *dashServer) judge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ Workspace, ID, To, Reason string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	target := trust.State(req.To)
	s.mu.Lock()
	defer s.mu.Unlock()
	repo, ok := s.byPath[req.Workspace]
	if !ok {
		http.Error(w, "unknown workspace", http.StatusConflict)
		return
	}
	idx := -1
	for i, it := range repo.Items {
		if it.ID == req.ID {
			idx = i
			break
		}
	}
	if idx < 0 {
		http.Error(w, "unknown or already judged", http.StatusConflict)
		return
	}
	it := repo.Items[idx]
	if !trust.CanTransition(trust.State(it.State), target) {
		http.Error(w, fmt.Sprintf("%s → %s is not a legal transition", it.State, target), http.StatusConflict)
		return
	}
	obs := trust.Observation{ID: it.ID, State: trust.State(it.State), Body: it.Body}
	if err := appendTransition(jsonl.Open(repo.Path), repo.Path, s.actor, obs, target, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	repo.Counts[it.State]--
	repo.Counts[req.To]++
	repo.Items = append(repo.Items[:idx], repo.Items[idx+1:]...)
	s.judged++
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *dashServer) quit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	s.once.Do(func() { close(s.done) })
}
