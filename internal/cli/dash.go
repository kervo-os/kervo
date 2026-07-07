package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/core/i18n"
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
	langFlag := fs.String("lang", "", "dash language: en, ko, ja (default: $LANG)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lang, err := dashLang(*langFlag)
	if err != nil {
		return err
	}
	paths := registeredWorkspaces()
	if len(paths) == 0 {
		fmt.Println("no kervo workspaces registered on this machine — run `kervo compile` in a repo first")
		return nil
	}
	srv, err := newDashServer(paths, *actor, lang)
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
	Reason                                string `json:",omitempty"`
	Conflict                              bool   `json:",omitempty"`
}

type dashRepo struct {
	Path, Name, Lang string
	DisplayPath      string // Path with $HOME shortened to ~ — for humans
	Events           int
	Counts           map[string]int
	LastEvent        string        // RFC3339, "" if the ledger is empty
	Activity         []int         // ledger events per day, oldest→today (28 days)
	Items            []dashItem    // awaiting judgment
	History          []dashItem    // already judged — newest first, reasons shown
	Overview         *dashOverview `json:",omitempty"` // nil if the fact scan failed
}

const activityDays = 28

type dashServer struct {
	mu     sync.Mutex
	actor  string
	lang   i18n.Lang
	repos  []*dashRepo
	byPath map[string]*dashRepo
	judged int
	done   chan struct{}
	once   sync.Once
}

// dashLang: the dash is a user surface spanning many repos, so its language
// is the USER's — flag, then the persisted in-page choice (~/.kervo/ui-lang),
// then $LC_ALL/$LANG — never any one workspace's .kervo/lang.
func dashLang(flagVal string) (i18n.Lang, error) {
	if flagVal != "" {
		return i18n.Parse(flagVal)
	}
	if sd := stateDir(); sd != "" {
		if raw, err := os.ReadFile(filepath.Join(sd, "ui-lang")); err == nil {
			if l, err := i18n.Parse(strings.TrimSpace(string(raw))); err == nil {
				return l, nil
			}
		}
	}
	for _, env := range []string{"LC_ALL", "LANG"} {
		if v := os.Getenv(env); len(v) >= 2 {
			if l, err := i18n.Parse(v[:2]); err == nil {
				return l, nil
			}
		}
	}
	return i18n.EN, nil
}

// dashKeys lists every UI string the page needs; trust-state and type
// tokens are deliberately NOT localized — they are ledger vocabulary and
// must read identically in the CLI, the artifact, and the page.
var dashKeys = []string{
	"workspaces", "totals", "localnote", "finish", "keys", "awaiting", "clear",
	"events", "emptyledger", "queue", "records", "pos", "cleared", "reasonph",
	"verify", "stale", "deprecate", "skip", "back", "donetitle", "donenote",
	"evidence", "justnow", "minago", "hourago", "dayago",
	"overview", "links", "more", "partialscan", "connected", "knowledge", "retired",
	"jhint", "jtitle", "jv", "js", "jd", "jx",
	"helptitle", "hopen", "hmove", "hjudge", "hskip", "hreason", "hback",
}

func dashStrings(l i18n.Lang) map[string]string {
	out := make(map[string]string, len(dashKeys)+4)
	for _, k := range dashKeys {
		out[k] = i18n.T(l, "dash."+k)
	}
	// Overview section titles reuse the artifact's vocabulary — the page
	// and CLAUDE.md must name the same things the same way.
	for k, ak := range map[string]string{
		"commands": "sec.commands", "recent": "sec.recent",
		"tasks": "sec.tasks", "modules": "sec.modules",
	} {
		out[k] = i18n.T(l, ak)
	}
	return out
}

func newDashServer(paths []string, actorFlag string, lang i18n.Lang) (*dashServer, error) {
	s := &dashServer{actor: actorFlag, lang: lang, byPath: map[string]*dashRepo{}, done: make(chan struct{})}
	for _, p := range paths {
		store := jsonl.Open(p)
		folder := trust.NewFolder()
		events := 0
		var last time.Time
		today := time.Now().UTC().Truncate(24 * time.Hour)
		activity := make([]int, activityDays)
		if err := store.Replay(context.Background(), "", func(e event.Event) error {
			events++
			if e.At.After(last) {
				last = e.At
			}
			if d := int(today.Sub(e.At.UTC().Truncate(24*time.Hour)).Hours() / 24); d >= 0 && d < activityDays {
				activity[activityDays-1-d]++
			}
			folder.Add(e)
			return nil
		}); err != nil {
			// One unreadable workspace must not blank the whole fleet —
			// skip it; it stays registered and returns when readable.
			fmt.Fprintf(os.Stderr, "kervo dash: skipping %s: %v\n", p, err)
			continue
		}
		repo := &dashRepo{
			Path: p, Name: filepath.Base(p), Lang: workspaceLang(p),
			DisplayPath: displayPath(p),
			Events:      events, Counts: map[string]int{},
			// Never nil: a clear repo must marshal as [], not null — the
			// page reads .Items.length on every repo before rendering.
			Items: []dashItem{}, History: []dashItem{},
			Activity: activity,
		}
		if !last.IsZero() {
			repo.LastEvent = last.UTC().Format(time.RFC3339)
		}
		// Best-effort: a workspace whose scan fails (broken git, moved
		// tree) still shows its ledger — the overview is just absent.
		if snap, _, err := scanFacts(context.Background(), p); err == nil {
			repo.Overview = buildOverview(snap)
			repo.Overview.Wiring = detectWiring(p, folder)
		}
		for _, o := range folder.Observations() {
			repo.Counts[string(o.State)]++
			it := dashItem{
				ID: o.ID, ShortID: shortID(o.ID), Type: o.Type, State: string(o.State),
				Actor: o.Actor, Body: unescapeBody(o.Body), Evidence: unescapeBody(o.Evidence), Reason: o.Reason,
				Conflict: o.Conflict,
			}
			if o.State == trust.Generated || o.Conflict {
				repo.Items = append(repo.Items, it)
			} else {
				// Judged records stay visible with their reasons — the
				// ledger never hides history, so neither does the page.
				repo.History = append(repo.History, it)
			}
		}
		slices.Reverse(repo.History) // newest judgments first
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
	mux.HandleFunc("/lang", s.setLang)
	mux.HandleFunc("/quit", s.quit)
	return mux
}

// setLang persists the in-page language choice machine-locally, so the
// next `kervo dash` opens in it without a flag.
func (s *dashServer) setLang(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct{ Lang string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	l, err := i18n.Parse(req.Lang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.lang = l
	s.mu.Unlock()
	if sd := stateDir(); sd != "" {
		if err := os.MkdirAll(sd, 0o755); err == nil {
			_ = os.WriteFile(filepath.Join(sd, "ui-lang"), []byte(string(l)+"\n"), 0o644)
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
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
	// All three tables ship with the page: switching languages is instant
	// and offline; only the persistence round-trips.
	all := map[string]map[string]string{}
	for _, l := range i18n.Supported() {
		all[string(l)] = dashStrings(l)
	}
	strs, err := json.Marshal(all)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = dashTmpl.Execute(w, struct {
		FleetJS, TTJS template.JS
		Lang          string
	}{template.JS(fleet), template.JS(strs), string(s.lang)})
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
	it.State, it.Reason = req.To, req.Reason
	repo.History = append([]dashItem{it}, repo.History...)
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
