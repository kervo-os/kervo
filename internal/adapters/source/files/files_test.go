package files

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/core/fact"
)

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanDocsAndTodos(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "README.md", "# Demo\n\nA demo project.\n")
	write(t, dir, "CLAUDE.md", "human notes\n<!-- kervo:begin -->\ninjected artifact\n<!-- kervo:end -->\nmore human notes\n")
	write(t, dir, "src/main.go", "package main\n// TODO: wire the thing\nfunc main() {}\n// FIXME broken edge case\n")
	write(t, dir, "node_modules/x/index.js", "// TODO: vendored, must be skipped\n")
	write(t, dir, ".hidden/notes.txt", "TODO: hidden dir, must be skipped\n")

	snap, cursor, err := New().Scan(context.Background(), dir, "ignored")
	if err != nil {
		t.Fatal(err)
	}
	if cursor != "" {
		t.Errorf("files scan must be stateless, cursor = %q", cursor)
	}

	byPath := map[string]string{}
	for _, d := range snap.Docs {
		byPath[d.Path] = d.Content
	}
	if !strings.Contains(byPath["README.md"], "A demo project.") {
		t.Errorf("README not captured: %q", byPath["README.md"])
	}
	claude := byPath["CLAUDE.md"]
	if strings.Contains(claude, "injected artifact") {
		t.Error("CLAUDE.md marker block was not stripped (feedback loop)")
	}
	if !strings.Contains(claude, "human notes") || !strings.Contains(claude, "more human notes") {
		t.Errorf("human-owned CLAUDE.md content lost: %q", claude)
	}

	if len(snap.Todos) != 2 {
		t.Fatalf("todos = %+v, want 2", snap.Todos)
	}
	want := fact.Todo{Path: filepath.Join("src", "main.go"), Line: 2, Text: "TODO: wire the thing"}
	if snap.Todos[0] != want {
		t.Errorf("todo[0] = %+v, want %+v", snap.Todos[0], want)
	}
	if !strings.HasPrefix(snap.Todos[1].Text, "FIXME:") {
		t.Errorf("todo[1] = %+v", snap.Todos[1])
	}
}

func TestScanTodoCapSetsPartial(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.go", "// TODO: one\n// TODO: two\n// TODO: three\n")
	s := New()
	s.MaxTodos = 2
	snap, _, err := s.Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Todos) != 2 || !snap.Partial {
		t.Errorf("todos = %d partial = %v, want 2/true", len(snap.Todos), snap.Partial)
	}
}

func TestDetectFrameworks(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module demo\n")
	write(t, dir, "package.json", `{"dependencies":{"react":"^19"},"devDependencies":{"typescript":"^5"}}`)
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(snap.Repo.Frameworks, ",")
	for _, want := range []string{"Go", "Node.js", "React", "TypeScript"} {
		if !strings.Contains(got, want) {
			t.Errorf("frameworks = %v, missing %s", snap.Repo.Frameworks, want)
		}
	}
}

// Regression: TODOs inside the injected marker block are artifact echoes,
// not workspace facts. Counting them makes every init invent new tasks.
func TestTodosInsideMarkerBlockIgnored(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "CLAUDE.md", "human\n<!-- kervo:begin -->\n- x.go:1 — TODO: echoed from artifact\n<!-- kervo:end -->\n// TODO: real one after block\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Todos) != 1 {
		t.Fatalf("todos = %+v, want only the one outside the block", snap.Todos)
	}
	if snap.Todos[0].Line != 5 {
		t.Errorf("line = %d, want 5 (masking must preserve line numbers)", snap.Todos[0].Line)
	}
}

func TestBinaryFilesSkipped(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "bin.dat", "TODO\x00binary")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Todos) != 0 {
		t.Errorf("binary file scanned: %+v", snap.Todos)
	}
}
