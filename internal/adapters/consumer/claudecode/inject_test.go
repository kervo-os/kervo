package claudecode

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readClaude(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestInjectCreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := (Injector{}).Inject(context.Background(), dir, "ARTIFACT\n"); err != nil {
		t.Fatal(err)
	}
	got := readClaude(t, dir)
	want := "<!-- kervo:begin -->\nARTIFACT\n<!-- kervo:end -->\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInjectReplacesOnlyMarkerBlock(t *testing.T) {
	dir := t.TempDir()
	human := "# My rules\n\nnever touch prod\n\n<!-- kervo:begin -->\nold artifact\n<!-- kervo:end -->\n\n## More human notes\n"
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(human), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (Injector{}).Inject(context.Background(), dir, "NEW"); err != nil {
		t.Fatal(err)
	}
	got := readClaude(t, dir)
	for _, keep := range []string{"# My rules", "never touch prod", "## More human notes"} {
		if !strings.Contains(got, keep) {
			t.Errorf("human content %q lost", keep)
		}
	}
	if strings.Contains(got, "old artifact") {
		t.Error("stale artifact not replaced")
	}
	if !strings.Contains(got, "<!-- kervo:begin -->\nNEW\n<!-- kervo:end -->") {
		t.Errorf("new block malformed:\n%s", got)
	}
}

func TestInjectAppendsWhenNoMarkers(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("human only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (Injector{}).Inject(context.Background(), dir, "ART"); err != nil {
		t.Fatal(err)
	}
	got := readClaude(t, dir)
	if !strings.HasPrefix(got, "human only\n") {
		t.Errorf("human prefix lost: %q", got)
	}
	if !strings.Contains(got, "<!-- kervo:begin -->\nART\n<!-- kervo:end -->") {
		t.Errorf("block not appended: %q", got)
	}
}

func TestInjectRefusesCorruptMarkers(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("x\n<!-- kervo:begin -->\nno end marker\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (Injector{}).Inject(context.Background(), dir, "ART"); err == nil {
		t.Fatal("expected error on corrupt markers, got none")
	}
	if got := readClaude(t, dir); !strings.Contains(got, "no end marker") {
		t.Error("file was modified despite corrupt markers")
	}
}

func TestInjectIdempotent(t *testing.T) {
	dir := t.TempDir()
	inj := Injector{}
	if err := inj.Inject(context.Background(), dir, "SAME"); err != nil {
		t.Fatal(err)
	}
	first := readClaude(t, dir)
	if err := inj.Inject(context.Background(), dir, "SAME"); err != nil {
		t.Fatal(err)
	}
	if second := readClaude(t, dir); first != second {
		t.Errorf("re-inject not idempotent:\n%q\nvs\n%q", first, second)
	}
}
