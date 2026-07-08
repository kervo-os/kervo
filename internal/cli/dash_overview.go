package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// dashWiring is what feeds and consumes this workspace — the adapters
// actually connected, detected from files and the ledger, never assumed.
type dashWiring struct {
	ClaudeMD, AgentsMD bool // consumer files carrying the marker block
	Hooks              bool // .claude/settings.json invokes kervo hook
	MCP                bool // .mcp.json registers the kervo server
	Inject             string
	Sources            []string // distinct observation actors seen in the ledger
}

// detectWiring checks which adapters are actually connected to dir.
func detectWiring(dir string, folder *trust.Folder) dashWiring {
	hasBlock := func(name string) bool {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		return err == nil && strings.Contains(string(raw), "<!-- kervo:begin -->")
	}
	contains := func(name, needle string) bool {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		return err == nil && strings.Contains(string(raw), needle)
	}
	w := dashWiring{
		ClaudeMD: hasBlock("CLAUDE.md"),
		AgentsMD: hasBlock("AGENTS.md"),
		Hooks:    contains(filepath.Join(".claude", "settings.json"), "kervo hook"),
		MCP:      contains(".mcp.json", `"kervo"`),
		Sources:  []string{},
	}
	if mode, err := resolveInject(dir, ""); err == nil {
		w.Inject = mode
	}
	seen := map[string]bool{}
	for _, o := range folder.Observations() {
		if !seen[o.Actor] {
			seen[o.Actor] = true
			w.Sources = append(w.Sources, o.Actor)
		}
	}
	sort.Strings(w.Sources)
	if len(w.Sources) > 6 {
		w.Sources = w.Sources[:6]
	}
	return w
}

// dashOverview is the workspace's fact skeleton, structured for the page —
// the same deterministic scan compile runs, capped for reading. Coupling
// pairs come from commit history (modules changed in the same commit):
// connections kervo can prove, not narrate.
type dashOverview struct {
	Wiring        dashWiring
	Branch        string
	Languages     []string
	Frameworks    []string
	Partial       bool
	Commands      []dashCmd
	TotalCommands int
	Commits       []dashCommit
	TotalCommits  int
	Tasks         []dashTask
	TotalTasks    int
	Modules       []dashModule
	Links         []dashLink
}

type dashCmd struct{ Run, Source string }
type dashCommit struct{ SHA, Date, Subject string }
type dashTask struct{ Loc, Text string }
type dashModule struct {
	Name  string
	Files int
}
type dashLink struct {
	A, B string
	N    int
}

// The page renders 8 rows per section and expands on click, so we ship
// enough to make expanding worth it — but not the whole 500-commit scan:
// beyond these caps the page shows a plain "+N more" that points back at
// git itself.
const (
	dashShipCommits = 100
	dashShipTasks   = 50
)

// buildOverview shapes a fact snapshot for the page.
func buildOverview(snap fact.Snapshot) *dashOverview {
	ov := &dashOverview{
		Branch: snap.Repo.Branch, Partial: snap.Partial,
		Languages:  append([]string{}, snap.Repo.Languages...),
		Frameworks: append([]string{}, snap.Repo.Frameworks...),
		Commands:   []dashCmd{}, Commits: []dashCommit{}, Tasks: []dashTask{},
		Modules: []dashModule{}, Links: []dashLink{},
		TotalCommands: len(snap.Commands), TotalCommits: len(snap.Commits), TotalTasks: len(snap.Todos),
	}
	for _, c := range snap.Commands {
		ov.Commands = append(ov.Commands, dashCmd{Run: c.Run, Source: c.Source})
	}
	for i, c := range snap.Commits {
		if i >= dashShipCommits {
			break
		}
		sha := c.SHA
		if len(sha) > 7 {
			sha = sha[:7]
		}
		ov.Commits = append(ov.Commits, dashCommit{SHA: sha, Date: c.At.UTC().Format("2006-01-02"), Subject: c.Subject})
	}
	for i, t := range snap.Todos {
		if i >= dashShipTasks {
			break
		}
		ov.Tasks = append(ov.Tasks, dashTask{Loc: fmt.Sprintf("%s:%d", t.Path, t.Line), Text: t.Text})
	}
	for _, m := range snap.Modules {
		ov.Modules = append(ov.Modules, dashModule{Name: m.Path, Files: m.Files})
	}
	current := map[string]bool{}
	for _, m := range snap.Modules {
		current[m.Path] = true
	}
	ov.Links = coupledModules(snap.Commits, current)
	return ov
}

// coupledModules counts top-level module pairs touched by the same commit.
// Commits spanning more than 6 modules are skipped as noise (mass renames,
// formatting sweeps). Deterministic: ties break lexically.
func coupledModules(commits []fact.Commit, current map[string]bool) []dashLink {
	pair := map[[2]string]int{}
	for _, c := range commits {
		mods := map[string]bool{}
		for _, f := range c.Files {
			if i := strings.IndexByte(f, '/'); i > 0 && f[0] != '.' && current[f[:i]] {
				// Dot-dirs are plumbing (.kervo travels with every commit
				// by design, .github with releases) — coupling is about
				// code modules.
				mods[f[:i]] = true
			}
		}
		if len(mods) < 2 || len(mods) > 6 {
			continue
		}
		names := make([]string, 0, len(mods))
		for m := range mods {
			names = append(names, m)
		}
		sort.Strings(names)
		for i := 0; i < len(names); i++ {
			for j := i + 1; j < len(names); j++ {
				pair[[2]string{names[i], names[j]}]++
			}
		}
	}
	out := make([]dashLink, 0, len(pair))
	for k, n := range pair {
		out = append(out, dashLink{A: k[0], B: k[1], N: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].N != out[j].N {
			return out[i].N > out[j].N
		}
		if out[i].A != out[j].A {
			return out[i].A < out[j].A
		}
		return out[i].B < out[j].B
	})
	if len(out) > 6 {
		out = out[:6]
	}
	return out
}
