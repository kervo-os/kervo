package gitexec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/ports"
)

// DefaultMaxCommits is the scan cap from the perf budget (ARCH-0001 §10.3):
// last 500 commits, Partial marked beyond that.
const DefaultMaxCommits = 500

const (
	maxChangedFiles = 50 // aggregated hot-file list kept in the snapshot
	maxLanguages    = 5
)

// Scanner shells out to the git CLI. Zero state; safe for reuse.
type Scanner struct {
	MaxCommits int
}

var _ ports.SourceProvider = (*Scanner)(nil)

func New() *Scanner { return &Scanner{MaxCommits: DefaultMaxCommits} }

// Scan reads the repository at dir. cursor is the last-seen HEAD SHA from a
// previous scan (empty = full scan); the returned cursor is the current HEAD.
// An unknown/garbage cursor degrades to a full scan — never an error
// (Mode 1 must not fail while a git repo exists, RFC-0003 §4).
func (s *Scanner) Scan(ctx context.Context, dir, cursor string) (fact.Snapshot, string, error) {
	var snap fact.Snapshot

	top, err := s.run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return snap, "", fmt.Errorf("gitexec: not a git repository: %s", dir)
	}
	snap.Repo.Name = filepath.Base(strings.TrimSpace(top))

	head, err := s.run(ctx, dir, "rev-parse", "HEAD")
	if err != nil {
		// Repo exists but has no commits yet. Fact-only still succeeds.
		snap.Repo.Branch = s.symbolicBranch(ctx, dir)
		return snap, "", nil
	}
	next := strings.TrimSpace(head)

	if branch, err := s.run(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		snap.Repo.Branch = strings.TrimSpace(branch)
	}

	if cursor != "" && !s.commitExists(ctx, dir, cursor) {
		cursor = "" // history rewritten or foreign cursor: degrade to full scan
	}

	cap := s.MaxCommits
	if cap <= 0 {
		cap = DefaultMaxCommits
	}
	commits, partial, err := s.log(ctx, dir, cursor, cap)
	if err != nil {
		return snap, "", err
	}
	snap.Commits = commits
	snap.Partial = partial
	snap.Files = aggregateFiles(commits)

	tracked, err := s.lsFiles(ctx, dir)
	if err == nil {
		snap.Repo.Languages = detectLanguages(tracked)
		snap.Modules = detectModules(tracked)
	}

	return snap, next, nil
}

// log returns commits newest-first, capped. partial reports a hit cap.
// Merge commits are excluded: on PR-driven repos they double the list
// while carrying no information the merged commits don't (prometheus:
// half of Recent Changes was "Merge pull request #...").
func (s *Scanner) log(ctx context.Context, dir, cursor string, cap int) ([]fact.Commit, bool, error) {
	args := []string{
		"log",
		"--no-merges",
		"--max-count=" + strconv.Itoa(cap+1), // +1 probe detects overflow
		"--format=%x1e%H%x1f%aI%x1f%s",
		"--name-only",
	}
	if cursor != "" {
		args = append(args, cursor+"..HEAD")
	}
	out, err := s.run(ctx, dir, args...)
	if err != nil {
		return nil, false, fmt.Errorf("gitexec: git log: %w", err)
	}

	var commits []fact.Commit
	for _, rec := range strings.Split(out, "\x1e") {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		lines := strings.Split(rec, "\n")
		fields := strings.Split(lines[0], "\x1f")
		if len(fields) != 3 {
			continue
		}
		at, err := time.Parse(time.RFC3339, fields[1])
		if err != nil {
			return nil, false, fmt.Errorf("gitexec: bad author date %q: %w", fields[1], err)
		}
		c := fact.Commit{SHA: fields[0], At: at, Subject: fields[2]}
		for _, l := range lines[1:] {
			if l = strings.TrimSpace(l); l != "" {
				c.Files = append(c.Files, l)
			}
		}
		commits = append(commits, c)
	}

	partial := false
	if len(commits) > cap {
		commits = commits[:cap]
		partial = true
	}
	return commits, partial, nil
}

func (s *Scanner) lsFiles(ctx context.Context, dir string) ([]string, error) {
	out, err := s.run(ctx, dir, "ls-files", "-z")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

func (s *Scanner) commitExists(ctx context.Context, dir, sha string) bool {
	_, err := s.run(ctx, dir, "cat-file", "-e", sha+"^{commit}")
	return err == nil
}

func (s *Scanner) symbolicBranch(ctx context.Context, dir string) string {
	out, err := s.run(ctx, dir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func (s *Scanner) run(ctx context.Context, dir string, args ...string) (string, error) {
	full := append([]string{"-C", dir, "-c", "core.quotepath=false"}, args...)
	cmd := exec.CommandContext(ctx, "git", full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %v: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// aggregateFiles counts how often each path appears across the scanned
// commits — a deterministic "hot files" view for Recent Changes context.
func aggregateFiles(commits []fact.Commit) []fact.ChangedFile {
	counts := map[string]int{}
	for _, c := range commits {
		for _, f := range c.Files {
			counts[f]++
		}
	}
	files := make([]fact.ChangedFile, 0, len(counts))
	for p, n := range counts {
		files = append(files, fact.ChangedFile{Path: p, Changes: n})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Changes != files[j].Changes {
			return files[i].Changes > files[j].Changes
		}
		return files[i].Path < files[j].Path
	})
	if len(files) > maxChangedFiles {
		files = files[:maxChangedFiles]
	}
	return files
}

// vendoredDirs are excluded from language/module statistics (perf budget:
// ".gitignore/바이너리/vendored 제외" — ls-files already honors .gitignore).
var vendoredDirs = map[string]bool{
	"vendor": true, "node_modules": true, "third_party": true,
	"dist": true, "build": true, ".kervo": true,
}

var extLanguages = map[string]string{
	".go": "Go", ".ts": "TypeScript", ".tsx": "TypeScript",
	".js": "JavaScript", ".jsx": "JavaScript", ".mjs": "JavaScript",
	".py": "Python", ".rb": "Ruby", ".rs": "Rust", ".java": "Java",
	".kt": "Kotlin", ".swift": "Swift", ".c": "C", ".cpp": "C++",
	".cc": "C++", ".hpp": "C++", ".cs": "C#", ".php": "PHP",
	".sh": "Shell", ".zsh": "Shell", ".bash": "Shell", ".sql": "SQL",
	".html": "HTML", ".css": "CSS", ".scss": "CSS", ".md": "Markdown",
}

func detectLanguages(tracked []string) []string {
	counts := map[string]int{}
	for _, p := range tracked {
		if isVendored(p) {
			continue
		}
		if lang, ok := extLanguages[strings.ToLower(filepath.Ext(p))]; ok {
			counts[lang]++
		}
	}
	langs := make([]string, 0, len(counts))
	for l := range counts {
		langs = append(langs, l)
	}
	sort.Slice(langs, func(i, j int) bool {
		if counts[langs[i]] != counts[langs[j]] {
			return counts[langs[i]] > counts[langs[j]]
		}
		return langs[i] < langs[j]
	})
	if len(langs) > maxLanguages {
		langs = langs[:maxLanguages]
	}
	return langs
}

func detectModules(tracked []string) []fact.Module {
	counts := map[string]int{}
	for _, p := range tracked {
		if isVendored(p) {
			continue
		}
		top, _, found := strings.Cut(p, "/")
		if !found {
			continue // root-level file, not a module
		}
		counts[top]++
	}
	mods := make([]fact.Module, 0, len(counts))
	for p, n := range counts {
		mods = append(mods, fact.Module{Path: p, Files: n})
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].Path < mods[j].Path })
	return mods
}

func isVendored(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if vendoredDirs[seg] {
			return true
		}
	}
	return false
}
