package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := exec.Command("git", "-C", dir, "init", "-q").Run(); err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	return dir
}

func TestWireGitAutoCompileCreatesBothHooks(t *testing.T) {
	dir := gitRepo(t)
	for i := 0; i < 2; i++ { // second run must be a no-op, not a duplicate
		if _, err := wireGitAutoCompile(dir); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	for _, name := range []string{"post-commit", "post-merge"} {
		p := filepath.Join(dir, ".git", "hooks", name)
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if string(raw) != gitHookScript {
			t.Fatalf("%s content:\n%q", name, raw)
		}
		if fi, _ := os.Stat(p); fi.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable", name)
		}
	}
}

func TestWireGitAutoCompileNeverRewritesForeignHooks(t *testing.T) {
	dir := gitRepo(t)
	foreign := "#!/bin/sh\necho mine\n"
	p := filepath.Join(dir, ".git", "hooks", "post-commit")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(foreign), 0o755); err != nil {
		t.Fatal(err)
	}
	status, err := wireGitAutoCompile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(p); string(raw) != foreign {
		t.Fatalf("foreign post-commit rewritten:\n%q", raw)
	}
	if !strings.Contains(status, "post-merge wired") || !strings.Contains(status, "post-commit left untouched") {
		t.Fatalf("status = %q", status)
	}
}
