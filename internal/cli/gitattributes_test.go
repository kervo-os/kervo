package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitattributesCreatesAndStaysIdempotent(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 2; i++ {
		if err := ensureGitattributes(dir); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".gitattributes"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(raw), unionAttr); got != 1 {
		t.Fatalf("union rule appears %d times, want exactly 1:\n%s", got, raw)
	}
}

func TestEnsureGitattributesPreservesHumanRules(t *testing.T) {
	dir := t.TempDir()
	human := "*.png binary" // no trailing newline: the append must add one
	if err := os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte(human), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureGitattributes(dir); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, ".gitattributes"))
	want := human + "\n" + unionAttr + "\n"
	if string(raw) != want {
		t.Fatalf("got:\n%q\nwant:\n%q", raw, want)
	}
}
