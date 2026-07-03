package compiler

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/trust"
)

var update = flag.Bool("update", false, "rewrite golden files")

// fixture exercises every skeleton section: caps, partial flag, docs,
// todos, modules. Times are fixed — the compiler must not read a clock.
func fixture() fact.Snapshot {
	at := func(day int) time.Time {
		return time.Date(2026, 7, day, 10, 0, 0, 0, time.UTC)
	}
	return fact.Snapshot{
		Repo: fact.RepoInfo{
			Name:       "demo-api",
			Branch:     "main",
			Languages:  []string{"Go", "Markdown"},
			Frameworks: []string{"Go"},
		},
		Commits: []fact.Commit{
			{SHA: "aaaaaaa1111111111111111111111111111111111", At: at(3), Subject: "add auth middleware", Files: []string{"mw/auth.go"}},
			{SHA: "bbbbbbb2222222222222222222222222222222222", At: at(2), Subject: "fix token refresh", Files: []string{"mw/auth.go", "token.go"}},
			{SHA: "ccccccc3333333333333333333333333333333333", At: at(1), Subject: "initial commit", Files: []string{"main.go"}},
		},
		Files: []fact.ChangedFile{
			{Path: "mw/auth.go", Changes: 2},
			{Path: "main.go", Changes: 1},
			{Path: "token.go", Changes: 1},
		},
		Modules: []fact.Module{
			{Path: "mw", Files: 3},
			{Path: "store", Files: 5},
		},
		Commands: []fact.Command{
			{Run: "make build", Detail: "go build -o demo ./cmd", Source: "Makefile"},
			{Run: "npm run test", Detail: "jest --coverage", Source: "package.json"},
		},
		Todos: []fact.Todo{
			{Path: "mw/auth.go", Line: 42, Text: "TODO: rotate signing keys"},
			{Path: "main.go", Line: 7, Text: "FIXME: graceful shutdown"},
		},
		Docs: []fact.DocSummaryInput{
			{Path: "README.md", Content: "# demo-api\n\nA demo REST API used to exercise the compiler.\nIt has two paragraphs.\n\nSecond paragraph is not excerpted.\n"},
			{Path: "CLAUDE.md", Content: "human notes\n"},
		},
		Partial: true,
	}
}

// TestSkeletonByteIdentity is the release gate from ARCH-0001: identical
// input must produce a byte-identical skeleton, pinned by a golden file.
func TestSkeletonByteIdentity(t *testing.T) {
	got, err := BuildSkeleton(fixture())
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "skeleton.golden.md")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("missing golden file (run with -update): %v", err)
	}
	if got != string(want) {
		t.Errorf("skeleton diverged from golden file\n--- got ---\n%s", got)
	}

	again, _ := BuildSkeleton(fixture())
	if got != again {
		t.Error("two runs over identical input differ — nondeterminism in BuildSkeleton")
	}
}

func TestAttachDetachRoundTrip(t *testing.T) {
	skel, err := BuildSkeleton(fixture())
	if err != nil {
		t.Fatal(err)
	}
	es := []artifact.Enhancement{
		{Slot: artifact.SlotGoal, Body: "Ship the auth middleware", State: trust.Generated, Source: "consumer:claude-code"},
		{Slot: artifact.SlotDecisions, Body: "JWT over sessions", State: trust.Generated, Source: "consumer:claude-code"},
		{Slot: artifact.SlotDecisions, Body: "SQLite for v1", State: trust.Generated, Source: "backend:ollama"},
	}
	rendered, err := Attach(skel, es)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "**[generated — consumer:claude-code]**\nShip the auth middleware") {
		t.Error("goal enhancement not rendered with state+source label")
	}
	if !strings.Contains(rendered, "JWT over sessions") || !strings.Contains(rendered, "SQLite for v1") {
		t.Error("multiple enhancements in one slot not all rendered")
	}
	// Skeleton sections must be untouched outside the slot regions.
	if !strings.Contains(rendered, "## Repository Summary") || !strings.Contains(rendered, "- `aaaaaaa` 2026-07-03 add auth middleware") {
		t.Error("Attach modified skeleton sections")
	}

	back, err := Detach(rendered)
	if err != nil {
		t.Fatal(err)
	}
	if back != skel {
		t.Error("Detach(Attach(skeleton)) != skeleton — round-trip invariant broken (RFC-0003 §2.2)")
	}
}

func TestAttachRejectsBoundaryViolations(t *testing.T) {
	skel, _ := BuildSkeleton(fixture())
	cases := []struct {
		name string
		e    artifact.Enhancement
	}{
		{"unknown slot", artifact.Enhancement{Slot: "repository-summary", Body: "x", State: trust.Generated, Source: "s"}},
		{"missing state label", artifact.Enhancement{Slot: artifact.SlotRisks, Body: "x", Source: "s"}},
		{"missing source", artifact.Enhancement{Slot: artifact.SlotRisks, Body: "x", State: trust.Generated}},
		{"empty body", artifact.Enhancement{Slot: artifact.SlotRisks, Body: "  ", State: trust.Generated, Source: "s"}},
	}
	for _, c := range cases {
		if _, err := Attach(skel, []artifact.Enhancement{c.e}); err == nil {
			t.Errorf("%s: expected error, got none", c.name)
		}
	}
}

// Regression (prometheus validation): READMEs that open with a centered
// HTML logo block or badge rows rendered as tag soup in the excerpt.
func TestExcerptSkipsHTMLAndBadges(t *testing.T) {
	s := fixture()
	s.Docs[0].Content = "<h1 align=\"center\"><img src=\"logo.svg\"></h1>\n" +
		"<p align=\"center\">tagline</p>\n\n" +
		"[![Build](badge.svg)](ci)\n![screenshot](s.png)\n\n" +
		"Real prose starts here.\nAnd continues.\n\nSecond paragraph.\n"
	out, err := BuildSkeleton(s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "> Real prose starts here. And continues.") {
		t.Errorf("excerpt did not reach the first prose paragraph:\n%s", out)
	}
	if strings.Contains(out, "<h1") || strings.Contains(out, "[![") {
		t.Error("HTML/badge lines leaked into the excerpt")
	}
}

// Regression (self-scan dogfooding): a TODO whose text contained a literal
// end marker impersonated the artifact's structure and truncated CLAUDE.md
// injection. Data-derived text must never yield a reserved marker.
func TestSnapshotDataCannotImpersonateMarkers(t *testing.T) {
	s := fixture()
	s.Commits[0].Subject = "evil <!-- kervo:begin --> subject"
	s.Todos[0].Text = "TODO: echoed <!-- kervo:end --> marker"
	s.Todos[1].Text = "TODO: fake slot <!-- kervo:slot:goal:end --> here"
	out, err := BuildSkeleton(s)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, artifact.MarkerBegin) || strings.Contains(out, artifact.MarkerEnd) {
		t.Error("data-derived begin/end marker survived into the skeleton")
	}
	// Exactly the template's own structural markers: header + 4 slot pairs.
	if got := strings.Count(out, artifact.ReservedPrefix); got != 9 {
		t.Errorf("reserved-prefix count = %d, want 9 (impostor escaped none?)", got)
	}
	// Slot machinery must still work on the escaped skeleton.
	es := []artifact.Enhancement{{Slot: artifact.SlotGoal, Body: "g", State: trust.Generated, Source: "s"}}
	rendered, err := Attach(out, es)
	if err != nil {
		t.Fatal(err)
	}
	back, err := Detach(rendered)
	if err != nil {
		t.Fatal(err)
	}
	if back != out {
		t.Error("round-trip broke on escaped skeleton")
	}
}

func TestEmptySnapshotStillBuilds(t *testing.T) {
	out, err := BuildSkeleton(fact.Snapshot{})
	if err != nil {
		t.Fatalf("Mode 1 must not fail on an empty snapshot: %v", err)
	}
	for _, want := range []string{"# Context Artifact", "_No commits found._", "_No TODO/FIXME comments found._"} {
		if !strings.Contains(out, want) {
			t.Errorf("empty skeleton missing %q", want)
		}
	}
}
