package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// Mode 3 outranks Mode 2 when configured; on backend failure compile
// falls through to the staged proposals (RFC-0003 §4 chain, e2e).
func TestCompileMode3ChainAndFallback(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main.go", "package main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "initial")
	writeFile(t, dir, ".kervo/proposals.json", `[{"slot":"goal","body":"Mode 2 fallback goal.","source":"consumer:test"}]`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{
				"content": `[{"slot":"goal","body":"Mode 3 backend goal. Evidence: commits."}]`,
			}}},
		})
	}))
	defer srv.Close()

	t.Setenv("KERVO_SEMANTIC_URL", srv.URL)
	t.Setenv("KERVO_SEMANTIC_MODEL", "test-model")
	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	art, _ := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if !strings.Contains(string(art), "backend:test-model") || !strings.Contains(string(art), "Mode 3 backend goal.") {
		t.Errorf("Mode 3 result missing:\n%s", art)
	}
	if strings.Contains(string(art), "Mode 2 fallback goal.") {
		t.Error("Mode 2 applied even though Mode 3 succeeded")
	}

	// Kill the backend: same run must demote to the staged Mode 2 proposals.
	srv.Close()
	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatalf("backend death must demote, not fail: %v", err)
	}
	art, _ = os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if !strings.Contains(string(art), "Mode 2 fallback goal.") {
		t.Error("fallback to Mode 2 proposals did not happen")
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
