package cli

import (
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/core/fact"
)

func sampleSnap() fact.Snapshot {
	return fact.Snapshot{
		Repo: fact.RepoInfo{
			Name: "demo", Branch: "main",
			Languages:  []string{"Go", "Markdown"},
			Frameworks: []string{"Go"},
		},
		Commits: make([]fact.Commit, 3),
		Todos:   make([]fact.Todo, 2),
		Modules: make([]fact.Module, 1),
		Docs: []fact.DocSummaryInput{
			{Path: "README.md"}, {Path: "CLAUDE.md"},
		},
		Partial: true,
	}
}

func TestRenderColdStartPlain(t *testing.T) {
	out := renderColdStart(ui{}, sampleSnap(), "dev")
	for _, want := range []string{
		logoLines[0], tagline, "dev",
		"Workspace Found", "✓ Git", "✓ CLAUDE.md", "✓ README",
		"3 analyzed", "(partial — scan capped)",
		"Go, Markdown", "2 open · 1 modules",
		".kervo/artifact.md", "(Mode 1 — Fact-only)", "CLAUDE.md  (marker block)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\x1b[") {
		t.Error("ANSI escape leaked into non-color output (breaks pipes/CI)")
	}
}

func TestRenderColdStartColor(t *testing.T) {
	out := renderColdStart(ui{color: true}, sampleSnap(), "dev")
	if !strings.Contains(out, "\x1b[36m") || !strings.Contains(out, "\x1b[32m") {
		t.Error("color mode missing cyan logo / green checks")
	}
	if !strings.HasSuffix(strings.TrimSpace(strings.Split(out, "\n")[0]), "\x1b[0m") {
		t.Error("styles not reset — would bleed into the user's prompt")
	}
}

func TestRenderColdStartMissingDocs(t *testing.T) {
	s := sampleSnap()
	s.Docs = nil
	out := renderColdStart(ui{}, s, "dev")
	for _, want := range []string{"– CLAUDE.md", "– README"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing-doc marker %q not shown", want)
		}
	}
}
