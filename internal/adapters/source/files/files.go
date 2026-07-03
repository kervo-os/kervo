package files

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/ports"
)

const (
	DefaultMaxTodos    = 200
	DefaultMaxDocBytes = 16 << 10 // per doc, cut at a newline
	maxScanFileBytes   = 512 << 10
	maxTodoTextLen     = 160
)

// docCandidates are read in this order; content becomes DocSummaryInput
// (raw input for the compiler — LLM summarizing is an Enhancement, never
// part of the skeleton).
var docCandidates = []string{"README.md", "README", "CLAUDE.md", "TODO.md"}

// Scanner reads workspace files directly (no git). It complements gitexec:
// docs, TODO/FIXME comments, and manifest-derived frameworks.
type Scanner struct {
	MaxTodos    int
	MaxDocBytes int
}

var _ ports.SourceProvider = (*Scanner)(nil)

func New() *Scanner {
	return &Scanner{MaxTodos: DefaultMaxTodos, MaxDocBytes: DefaultMaxDocBytes}
}

// Scan is stateless: the cursor is ignored and returned empty (file scanning
// is cheap enough to redo; incrementality lives in gitexec).
func (s *Scanner) Scan(ctx context.Context, dir, _ string) (fact.Snapshot, string, error) {
	var snap fact.Snapshot
	snap.Docs = s.readDocs(dir)
	snap.Repo.Frameworks = detectFrameworks(dir)
	todos, truncated := s.scanTodos(ctx, dir)
	snap.Todos = todos
	snap.Partial = truncated
	return snap, "", nil
}

func (s *Scanner) readDocs(dir string) []fact.DocSummaryInput {
	var docs []fact.DocSummaryInput
	for _, name := range docCandidates {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		content := string(raw)
		if name == "CLAUDE.md" {
			// Never feed our own injected block back in (feedback loop).
			content = stripMarkerBlock(content)
		}
		docs = append(docs, fact.DocSummaryInput{
			Path:    name,
			Content: capAtNewline(content, s.maxDocBytes()),
		})
	}
	return docs
}

// stripMarkerBlock removes the kervo-owned region of CLAUDE.md, keeping
// only human-written content (ARCH-0001 §4: content outside markers is
// human-owned — and it is the only part that is a source).
func stripMarkerBlock(content string) string {
	begin := strings.Index(content, artifact.MarkerBegin)
	end := strings.Index(content, artifact.MarkerEnd)
	if begin < 0 || end < begin {
		return content
	}
	return content[:begin] + content[end+len(artifact.MarkerEnd):]
}

func capAtNewline(s string, max int) string {
	if len(s) <= max {
		return s
	}
	cut := strings.LastIndexByte(s[:max], '\n')
	if cut <= 0 {
		cut = max
	}
	return s[:cut] + "\n[truncated]\n"
}

func (s *Scanner) maxDocBytes() int {
	if s.MaxDocBytes > 0 {
		return s.MaxDocBytes
	}
	return DefaultMaxDocBytes
}

// todoRe accepts TODO/FIXME only immediately after a comment marker
// (`// TODO: x`, `# FIXME y`, `* TODO(alice): z`). A bare word match drowned
// real tasks in noise on self-scan: string literals, regexes, and prose
// like "parses TODO comments" all counted as open tasks.
// Groups: 1=token, 2=colon (if any), 3=rest.
var todoRe = regexp.MustCompile(`(?:^|\s)(?://+|#+|/\*+|\*+|--+|<!--|;+)\s*(TODO|FIXME)\b(?:\([^)]*\))?(:)?[ \t]*(.*)`)

// skipDirs are never descended into. Hidden dirs are skipped separately.
// testdata holds fixtures — their TODOs are test material, not open tasks.
var skipDirs = map[string]bool{
	"node_modules": true, "vendor": true, "third_party": true,
	"dist": true, "build": true, "target": true,
	"venv": true, "__pycache__": true, "testdata": true,
}

// scanTodos walks the tree in deterministic (lexical) order collecting
// TODO/FIXME lines. Returns truncated=true when the cap was hit.
func (s *Scanner) scanTodos(ctx context.Context, dir string) ([]fact.Todo, bool) {
	max := s.MaxTodos
	if max <= 0 {
		max = DefaultMaxTodos
	}
	var todos []fact.Todo
	truncated := false

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable entries are skipped, never fatal (Mode 1)
		}
		if ctx.Err() != nil {
			return filepath.SkipAll
		}
		name := d.Name()
		if d.IsDir() {
			if path == dir {
				return nil
			}
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if len(todos) >= max {
			truncated = true
			return filepath.SkipAll
		}
		// Test fixtures fabricate TODOs on purpose; they are not open tasks.
		if strings.HasSuffix(name, "_test.go") {
			return nil
		}
		if info, err := d.Info(); err != nil || info.Size() > maxScanFileBytes {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		found, hitCap := scanFileTodos(path, rel, max-len(todos))
		todos = append(todos, found...)
		if hitCap {
			truncated = true
			return filepath.SkipAll
		}
		return nil
	})
	return todos, truncated
}

func scanFileTodos(path, rel string, budget int) ([]fact.Todo, bool) {
	raw, err := os.ReadFile(path)
	if err != nil || isBinary(raw) {
		return nil, false
	}
	// TODOs inside our own injected block are echoes of the artifact, not
	// workspace facts — scanning them creates a feedback loop where every
	// init invents new tasks (caught by the init idempotency e2e test).
	raw = maskMarkerBlock(raw)
	var todos []fact.Todo
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	line := 0
	for sc.Scan() {
		line++
		m := todoRe.FindStringSubmatch(sc.Text())
		if m == nil {
			continue
		}
		// "// TODO/FIXME handling" is wrapped prose, not a task: a colon-less
		// token flowing straight into "/" or "|" is compound wording.
		if m[2] == "" && (strings.HasPrefix(m[3], "/") || strings.HasPrefix(m[3], "|")) {
			continue
		}
		text := strings.TrimSpace(m[1] + ": " + strings.TrimSpace(m[3]))
		if len(text) > maxTodoTextLen {
			text = text[:maxTodoTextLen] + "…"
		}
		todos = append(todos, fact.Todo{Path: rel, Line: line, Text: text})
		if len(todos) >= budget {
			return todos, true
		}
	}
	return todos, false
}

// maskMarkerBlock blanks the kervo-owned region while keeping newlines,
// so line numbers of real findings stay accurate.
func maskMarkerBlock(raw []byte) []byte {
	begin := bytes.Index(raw, []byte(artifact.MarkerBegin))
	end := bytes.Index(raw, []byte(artifact.MarkerEnd))
	if begin < 0 || end < begin {
		return raw
	}
	masked := make([]byte, len(raw))
	copy(masked, raw)
	for i := begin; i < end+len(artifact.MarkerEnd) && i < len(masked); i++ {
		if masked[i] != '\n' {
			masked[i] = ' '
		}
	}
	return masked
}

func isBinary(raw []byte) bool {
	probe := raw
	if len(probe) > 8000 {
		probe = probe[:8000]
	}
	return bytes.IndexByte(probe, 0) >= 0
}

// manifestEcosystems maps manifest files to a framework/ecosystem label.
// Deterministic detection only; anything smarter is Semantic territory.
var manifestEcosystems = map[string]string{
	"go.mod": "Go", "Cargo.toml": "Rust", "Gemfile": "Ruby",
	"pyproject.toml": "Python", "requirements.txt": "Python",
	"pom.xml": "Java (Maven)", "build.gradle": "Java (Gradle)",
}

// notableNodeDeps surface well-known frameworks from package.json.
var notableNodeDeps = map[string]string{
	"react": "React", "next": "Next.js", "vue": "Vue",
	"svelte": "Svelte", "express": "Express", "@nestjs/core": "NestJS",
	"typescript": "TypeScript",
}

func detectFrameworks(dir string) []string {
	set := map[string]bool{}
	for manifest, label := range manifestEcosystems {
		if _, err := os.Stat(filepath.Join(dir, manifest)); err == nil {
			set[label] = true
		}
	}
	if raw, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		set["Node.js"] = true
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(raw, &pkg) == nil {
			for dep, label := range notableNodeDeps {
				if _, ok := pkg.Dependencies[dep]; ok {
					set[label] = true
				} else if _, ok := pkg.DevDependencies[dep]; ok {
					set[label] = true
				}
			}
		}
	}
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for l := range set {
		out = append(out, l)
	}
	sort.Strings(out)
	return out
}
