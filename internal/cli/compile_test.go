package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompileAttachesProposals(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main.go", "package main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "initial")

	writeFile(t, dir, ".kervo/proposals.json", `[
	  {"slot":"goal","body":"Ship the payments MVP. Evidence: recent commits.","source":"consumer:claude-code"},
	  {"slot":"risks","body":"Auth flow untested.","source":"consumer:claude-code"}
	]`)

	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	art, err := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"**[generated — consumer:claude-code]**\nShip the payments MVP.",
		"Auth flow untested.",
		"# Context Artifact", // skeleton intact
	} {
		if !strings.Contains(string(art), want) {
			t.Errorf("artifact missing %q", want)
		}
	}
	claude, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(string(claude), "Ship the payments MVP.") {
		t.Error("enhancements not injected into CLAUDE.md")
	}
}

// RFC-0003 §4: semantic failure demotes to fact-only — never a failed run.
func TestCompileDegradesOnBadProposals(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main.go", "package main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "initial")
	writeFile(t, dir, ".kervo/proposals.json", `{"not":"a list"}`)

	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatalf("bad proposals must degrade, not fail: %v", err)
	}
	art, _ := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if !strings.Contains(string(art), "_No proposal yet.") {
		t.Error("degraded artifact is not the fact-only skeleton")
	}
}

// Without proposals, compile output is byte-identical to init output.
func TestCompileWithoutProposalsEqualsInit(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main.go", "package main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "initial")

	if err := runInit([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	fromInit, _ := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	fromCompile, _ := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if string(fromInit) != string(fromCompile) {
		t.Error("compile without proposals diverged from init output")
	}
}
