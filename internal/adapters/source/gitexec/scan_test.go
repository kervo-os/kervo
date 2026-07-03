package gitexec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=t@t",
		"GIT_AUTHOR_DATE=2026-07-01T10:00:00+09:00",
		"GIT_COMMITTER_DATE=2026-07-01T10:00:00+09:00",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

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

// mkRepo builds a repo with 3 commits touching main.go, lib/util.go, README.md.
func mkRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	write(t, dir, "main.go", "package main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "add main")
	write(t, dir, "lib/util.go", "package lib\n")
	write(t, dir, "main.go", "package main // v2\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "add lib, touch main")
	write(t, dir, "README.md", "# demo\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "add readme")
	return dir
}

func TestScanFull(t *testing.T) {
	dir := mkRepo(t)
	snap, cursor, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if snap.Repo.Name != filepath.Base(dir) {
		t.Errorf("name = %q", snap.Repo.Name)
	}
	if snap.Repo.Branch != "main" {
		t.Errorf("branch = %q", snap.Repo.Branch)
	}
	if len(snap.Commits) != 3 {
		t.Fatalf("commits = %d, want 3", len(snap.Commits))
	}
	if snap.Commits[0].Subject != "add readme" { // newest first
		t.Errorf("newest subject = %q", snap.Commits[0].Subject)
	}
	if snap.Commits[0].At.IsZero() {
		t.Error("commit time not parsed")
	}
	if cursor != git(t, dir, "rev-parse", "HEAD") {
		t.Errorf("cursor = %q, want HEAD", cursor)
	}
	if snap.Partial {
		t.Error("unexpected Partial")
	}
	// main.go touched twice -> hottest file first.
	if len(snap.Files) == 0 || snap.Files[0].Path != "main.go" || snap.Files[0].Changes != 2 {
		t.Errorf("hot files = %+v", snap.Files)
	}
	if len(snap.Repo.Languages) == 0 || snap.Repo.Languages[0] != "Go" {
		t.Errorf("languages = %v", snap.Repo.Languages)
	}
	if len(snap.Modules) != 1 || snap.Modules[0].Path != "lib" || snap.Modules[0].Files != 1 {
		t.Errorf("modules = %+v", snap.Modules)
	}
}

func TestScanIncremental(t *testing.T) {
	dir := mkRepo(t)
	s := New()
	_, cursor, err := s.Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	snap, next, err := s.Scan(context.Background(), dir, cursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Commits) != 0 {
		t.Errorf("no-change incremental scan returned %d commits", len(snap.Commits))
	}
	if next != cursor {
		t.Errorf("cursor moved without new commits")
	}

	write(t, dir, "new.go", "package main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "fourth")
	snap, next, err = s.Scan(context.Background(), dir, cursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Commits) != 1 || snap.Commits[0].Subject != "fourth" {
		t.Errorf("incremental commits = %+v", snap.Commits)
	}
	if next == cursor {
		t.Error("cursor did not advance")
	}
}

func TestScanCapMarksPartial(t *testing.T) {
	dir := mkRepo(t)
	s := &Scanner{MaxCommits: 2}
	snap, _, err := s.Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Commits) != 2 || !snap.Partial {
		t.Errorf("commits = %d partial = %v, want 2/true", len(snap.Commits), snap.Partial)
	}
}

func TestScanBadCursorDegradesToFull(t *testing.T) {
	dir := mkRepo(t)
	snap, _, err := New().Scan(context.Background(), dir, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Commits) != 3 {
		t.Errorf("degraded scan commits = %d, want 3", len(snap.Commits))
	}
}

// Merge commits are noise on PR-driven repos (prometheus: half of Recent
// Changes was "Merge pull request #...") — the scan drops them.
func TestScanExcludesMergeCommits(t *testing.T) {
	dir := mkRepo(t)
	git(t, dir, "checkout", "-q", "-b", "feature")
	write(t, dir, "f.go", "package f\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "feature work")
	git(t, dir, "checkout", "-q", "main")
	write(t, dir, "g.go", "package g\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "main work")
	git(t, dir, "merge", "--no-ff", "-q", "-m", "merge feature into main", "feature")

	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range snap.Commits {
		if strings.HasPrefix(c.Subject, "merge feature") {
			t.Errorf("merge commit leaked into scan: %q", c.Subject)
		}
	}
	if len(snap.Commits) != 5 { // 3 from mkRepo + feature work + main work
		t.Errorf("commits = %d, want 5 non-merge", len(snap.Commits))
	}
}

func TestScanEmptyRepoStillSucceeds(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	snap, cursor, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatalf("Mode 1 must not fail on an empty repo: %v", err)
	}
	if cursor != "" || len(snap.Commits) != 0 {
		t.Errorf("empty repo: cursor=%q commits=%d", cursor, len(snap.Commits))
	}
	if snap.Repo.Branch != "main" {
		t.Errorf("branch = %q", snap.Repo.Branch)
	}
}

func TestScanNotARepoErrors(t *testing.T) {
	if _, _, err := New().Scan(context.Background(), t.TempDir(), ""); err == nil {
		t.Fatal("expected error outside a git repository")
	}
}
