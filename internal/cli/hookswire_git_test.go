package cli

import (
	"os"
	"os/exec"
	"path/filepath"
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
	for name, script := range map[string]string{"pre-commit": preCommitScript, "post-merge": postMergeScript} {
		p := filepath.Join(dir, ".git", "hooks", name)
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if string(raw) != script {
			t.Fatalf("%s content:\n%q", name, raw)
		}
		if fi, _ := os.Stat(p); fi.Mode()&0o111 == 0 {
			t.Fatalf("%s is not executable", name)
		}
	}
}

func TestWireGitAutoCompileMigratesOwnLegacyPostCommit(t *testing.T) {
	dir := gitRepo(t)
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	// Our exact v0.21.0 script is removed; a user-edited variant is not.
	if err := os.WriteFile(filepath.Join(hooks, "post-commit"), []byte(legacyPostCommitScript), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := wireGitAutoCompile(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(hooks, "post-commit")); !os.IsNotExist(err) {
		t.Fatal("legacy kervo post-commit should have been migrated away")
	}
}

func TestWireGitAutoCompileNeverRewritesForeignHooks(t *testing.T) {
	dir := gitRepo(t)
	foreign := "#!/bin/sh\necho mine\n"
	p := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(foreign), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := wireGitAutoCompile(dir); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(p); string(raw) != foreign {
		t.Fatalf("foreign post-commit rewritten:\n%q", raw)
	}
}
