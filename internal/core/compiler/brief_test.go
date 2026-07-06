package compiler

import (
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/core/fact"
)

// History-only paths must not appear in the Focus line: a repo extracted
// from a parent directory carries the old prefix in every pre-split
// commit, and the brief describes the repo that exists.
func TestBriefFocusIgnoresHistoricalModules(t *testing.T) {
	at := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	snap := fact.Snapshot{
		Modules: []fact.Module{{Path: "services", Files: 10}, {Path: "packages", Files: 5}},
		Commits: []fact.Commit{
			{Subject: "feat(search): federated hunt", At: at, Files: []string{"services/a.py", "oldroot/x.py"}},
			{Subject: "feat(search): semantic sim", At: at, Files: []string{"services/b.py", "packages/c.py"}},
			{Subject: "fix(ui): panel", At: at, Files: []string{"oldroot/y.ts"}},
			{Subject: "chore: deps", At: at, Files: []string{"services/d.py"}},
		},
	}
	got := BriefFocus(snap)
	for _, want := range []string{"search ×2", "ui ×1", "services/ ×3", "packages/ ×1"} {
		if !contains(got, want) {
			t.Errorf("focus %q missing %q", got, want)
		}
	}
	if contains(got, "oldroot") {
		t.Errorf("focus %q counts a module that no longer exists", got)
	}
}

func TestBriefFocusNeedsSignal(t *testing.T) {
	if got := BriefFocus(fact.Snapshot{}); got != "" {
		t.Errorf("empty snapshot must yield no focus, got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}
