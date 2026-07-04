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
	snap.Commands = detectCommands(dir)
	todos, truncated := s.scanTodos(ctx, dir)
	snap.Todos = todos
	snap.Partial = truncated
	return snap, "", nil
}

// maxDocs bounds the doc list on monorepos (12 modules × 2 docs adds up).
const maxDocs = 20

func (s *Scanner) readDocs(dir string) []fact.DocSummaryInput {
	var docs []fact.DocSummaryInput
	add := func(rel string) {
		if len(docs) >= maxDocs {
			return
		}
		raw, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			return
		}
		content := string(raw)
		if filepath.Base(rel) == "CLAUDE.md" {
			// Never feed our own injected block back in (feedback loop).
			content = stripMarkerBlock(content)
			// A CLAUDE.md that is nothing but our block carries zero human
			// context — counting it as a captured doc makes init change its
			// own next scan (caught by the compile==init byte test).
			if strings.TrimSpace(content) == "" {
				return
			}
		}
		docs = append(docs, fact.DocSummaryInput{
			Path:    rel,
			Content: capAtNewline(content, s.maxDocBytes()),
		})
	}
	for _, name := range docCandidates {
		add(name)
	}
	// Monorepo support: modules often carry their own context docs
	// (field evidence: a 12-module CTI repo maintained api/CLAUDE.md,
	// database/CLAUDE.md, ... by hand — the richest context on disk).
	for _, mod := range moduleDirs(dir) {
		add(filepath.Join(mod, "CLAUDE.md"))
		add(filepath.Join(mod, "README.md"))
	}
	return docs
}

// moduleDirs lists top-level module directories — the one-level monorepo
// scan surface shared by docs, manifests, and framework detection. Depth
// stays at 1 by design: recursing turns the scanner into a build system.
func moduleDirs(dir string) []string {
	entries, err := os.ReadDir(dir) // already sorted — determinism for free
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || skipDirs[name] || strings.HasPrefix(name, ".") {
			continue
		}
		out = append(out, name)
	}
	return out
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

const maxCommandsPerManifest = 12

// makeTargetRe matches simple Makefile rule lines ("build:", "arch-check:").
// Variable assignments (":="), pattern rules ("%"), dot-targets (".PHONY"),
// and multi-target lines are deliberately not matched.
var makeTargetRe = regexp.MustCompile(`^([A-Za-z0-9][\w.-]*)\s*:($|[^=])`)

// detectCommands collects runnable entry points the workspace declares.
// Evidence for the section: build/test commands are the most prevalent
// hand-written context type (72% of 401 repos, arXiv:2512.18925).
func detectCommands(dir string) []fact.Command {
	var cmds []fact.Command
	cmds = append(cmds, makefileCommands(dir)...)
	cmds = append(cmds, justfileCommands(dir)...)
	cmds = append(cmds, packageJSONCommands(dir)...)
	cmds = append(cmds, pyprojectCommands(dir)...)
	cmds = append(cmds, composeCommands(dir)...)
	return cmds
}

// pyprojectCommands reads declared script entry points from pyproject.toml —
// root first, then one level into module dirs (field evidence: a monorepo
// keeping analyzer/pyproject.toml and processor/pyproject.toml, root bare).
// Sections: [project.scripts], [tool.poetry.scripts], [tool.pdm.scripts].
// Line-based TOML subset — declared entries only, nothing inferred.
// The maxCommandsPerManifest cap is shared across all pyproject files.
func pyprojectCommands(dir string) []fact.Command {
	scriptSections := map[string]bool{
		"[project.scripts]": true, "[tool.poetry.scripts]": true, "[tool.pdm.scripts]": true,
	}
	var cmds []fact.Command
	parse := func(rel string) {
		raw, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			return
		}
		in := false
		for _, line := range strings.Split(string(raw), "\n") {
			t := strings.TrimSpace(line)
			if strings.HasPrefix(t, "[") {
				in = scriptSections[t]
				continue
			}
			if !in || t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			name, val, ok := strings.Cut(t, "=")
			if !ok {
				continue
			}
			name = strings.TrimSpace(name)
			val = strings.Trim(strings.TrimSpace(val), `"'`)
			if name == "" {
				continue
			}
			cmds = append(cmds, fact.Command{Run: name, Detail: capLine(val, 80), Source: rel})
			if len(cmds) >= maxCommandsPerManifest {
				return
			}
		}
	}
	parse("pyproject.toml")
	for _, mod := range moduleDirs(dir) {
		if len(cmds) >= maxCommandsPerManifest {
			break
		}
		parse(filepath.Join(mod, "pyproject.toml"))
	}
	return cmds
}

// justfileCommands parses just recipes (same shape as Makefile targets).
func justfileCommands(dir string) []fact.Command {
	var raw []byte
	var err error
	for _, n := range []string{"justfile", "Justfile", ".justfile"} {
		if raw, err = os.ReadFile(filepath.Join(dir, n)); err == nil {
			break
		}
	}
	if err != nil {
		return nil
	}
	lines := strings.Split(string(raw), "\n")
	var cmds []fact.Command
	for i, line := range lines {
		m := makeTargetRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		detail := ""
		if i+1 < len(lines) && (strings.HasPrefix(lines[i+1], "\t") || strings.HasPrefix(lines[i+1], "  ")) {
			detail = capLine(strings.TrimSpace(lines[i+1]), 80)
		}
		cmds = append(cmds, fact.Command{Run: "just " + m[1], Detail: detail, Source: "justfile"})
		if len(cmds) >= maxCommandsPerManifest {
			break
		}
	}
	return cmds
}

// composeCommands lists declared docker-compose services. Line-based YAML
// subset: keys indented exactly two spaces under a top-level "services:".
func composeCommands(dir string) []fact.Command {
	matches, _ := filepath.Glob(filepath.Join(dir, "docker-compose*.y*ml"))
	more, _ := filepath.Glob(filepath.Join(dir, "compose.y*ml"))
	matches = append(matches, more...)
	sort.Strings(matches)
	var cmds []fact.Command
	for _, path := range matches {
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		file := filepath.Base(path)
		flag := ""
		if file != "docker-compose.yml" && file != "compose.yml" && file != "docker-compose.yaml" && file != "compose.yaml" {
			flag = " -f " + file
		}
		in := false
		var detail string
		var pending string
		flush := func() {
			if pending != "" && len(cmds) < maxCommandsPerManifest {
				cmds = append(cmds, fact.Command{
					Run: "docker compose" + flag + " up " + pending, Detail: detail, Source: file,
				})
			}
			pending, detail = "", ""
		}
		for _, line := range strings.Split(string(raw), "\n") {
			t := strings.TrimRight(line, " \r")
			switch {
			case t == "services:":
				in = true
			case in && t != "" && !strings.HasPrefix(t, " ") && !strings.HasPrefix(t, "#"):
				in = false // next top-level key ends the services block
			case in && strings.HasPrefix(t, "  ") && !strings.HasPrefix(t, "   ") &&
				strings.HasSuffix(t, ":") && !strings.Contains(strings.TrimSpace(t), " "):
				flush()
				pending = strings.TrimSuffix(strings.TrimSpace(t), ":")
			case in && pending != "" && detail == "":
				if img, ok := strings.CutPrefix(strings.TrimSpace(t), "image:"); ok {
					// YAML: whitespace + '#' starts an inline comment.
					if i := strings.Index(img, " #"); i >= 0 {
						img = img[:i]
					}
					detail = capLine(strings.TrimSpace(img), 60)
				}
			}
		}
		flush() // trailing service (EOF or block exited)
		if len(cmds) >= maxCommandsPerManifest {
			break
		}
	}
	return cmds
}

func makefileCommands(dir string) []fact.Command {
	raw, err := os.ReadFile(filepath.Join(dir, "Makefile"))
	if err != nil {
		return nil
	}
	lines := strings.Split(string(raw), "\n")
	var cmds []fact.Command
	for i, line := range lines {
		m := makeTargetRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		detail := ""
		if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "\t") {
			detail = capLine(strings.TrimSpace(lines[i+1]), 80)
		}
		cmds = append(cmds, fact.Command{
			Run: "make " + m[1], Detail: detail, Source: "Makefile",
		})
		if len(cmds) >= maxCommandsPerManifest {
			break
		}
	}
	return cmds
}

func packageJSONCommands(dir string) []fact.Command {
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(raw, &pkg) != nil || len(pkg.Scripts) == 0 {
		return nil
	}
	names := make([]string, 0, len(pkg.Scripts))
	for n := range pkg.Scripts {
		names = append(names, n)
	}
	sort.Strings(names) // JSON maps are unordered; determinism requires sorting
	if len(names) > maxCommandsPerManifest {
		names = names[:maxCommandsPerManifest]
	}
	var cmds []fact.Command
	for _, n := range names {
		cmds = append(cmds, fact.Command{
			Run: "npm run " + n, Detail: capLine(pkg.Scripts[n], 80), Source: "package.json",
		})
	}
	return cmds
}

func capLine(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
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
	// Notable Python frameworks, from declared dependencies only. Module
	// manifests count too — monorepos often keep the root bare.
	var pyDeps []byte
	pyManifests := []string{"requirements.txt", "pyproject.toml"}
	for _, mod := range moduleDirs(dir) {
		pyManifests = append(pyManifests,
			filepath.Join(mod, "requirements.txt"), filepath.Join(mod, "pyproject.toml"))
	}
	for _, rel := range pyManifests {
		if raw, err := os.ReadFile(filepath.Join(dir, rel)); err == nil {
			set["Python"] = true
			pyDeps = append(pyDeps, raw...)
		}
	}
	if len(pyDeps) > 0 {
		low := strings.ToLower(string(pyDeps))
		for dep, label := range notablePyDeps {
			if strings.Contains(low, dep) {
				set[label] = true
			}
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
		set["Docker"] = true
	}
	if m, _ := filepath.Glob(filepath.Join(dir, "docker-compose*.y*ml")); len(m) > 0 {
		set["Docker Compose"] = true
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

var notablePyDeps = map[string]string{
	"fastapi": "FastAPI", "django": "Django", "flask": "Flask",
	"celery": "Celery", "sqlalchemy": "SQLAlchemy",
}
