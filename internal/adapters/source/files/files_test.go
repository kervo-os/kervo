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

func TestDetectCommands(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "Makefile", ".PHONY: build test\nVAR := x\n\nbuild:\n\tgo build -o app ./cmd\n\ntest: build\n\tgo test ./...\n\n%.o: %.c\n\tcc -c $<\n")
	write(t, dir, "package.json", `{"scripts":{"test":"jest --coverage","build":"tsc -p ."}}`)
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	var runs []string
	for _, c := range snap.Commands {
		runs = append(runs, c.Run)
	}
	// Makefile targets in file order, then package.json scripts sorted.
	want := []string{"make build", "make test", "npm run build", "npm run test"}
	if strings.Join(runs, ",") != strings.Join(want, ",") {
		t.Fatalf("commands = %v, want %v", runs, want)
	}
	if snap.Commands[0].Detail != "go build -o app ./cmd" {
		t.Errorf("make build detail = %q", snap.Commands[0].Detail)
	}
	if snap.Commands[3].Detail != "jest --coverage" {
		t.Errorf("npm test detail = %q", snap.Commands[3].Detail)
	}
	for _, c := range snap.Commands {
		if strings.Contains(c.Run, ".PHONY") || strings.Contains(c.Run, "VAR") || strings.Contains(c.Run, "%") {
			t.Errorf("non-command leaked: %+v", c)
		}
	}
}

// Regression (self-scan dogfooding): bare-word TODO matching produced 23
// false tasks from 0 real ones — string literals, regex sources, and prose
// all counted. Only comment-marker-prefixed TODOs are tasks.
func TestTodoRequiresCommentMarker(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "code.go", strings.Join([]string{
		`var s = "TODO: inside a string literal"`,     // no
		`var re = "\\b(TODO|FIXME)\\b"`,               // no
		`// this parser reads TODO comments from src`, // no: mid-comment prose
		`// TODO: real task`,                          // yes
		`x := 1 // FIXME(alice): trailing comment`,    // yes
		`# TODO hash style`,                           // yes
		`// TODO/FIXME lines are collected here`,      // no: wrapped compound prose
	}, "\n")+"\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Todos) != 3 {
		t.Fatalf("todos = %+v, want 3", snap.Todos)
	}
	wants := []string{"TODO: real task", "FIXME: trailing comment", "TODO: hash style"}
	for i, w := range wants {
		if snap.Todos[i].Text != w {
			t.Errorf("todo[%d] = %q, want %q", i, snap.Todos[i].Text, w)
		}
	}
}

func TestPytestDeclaredRunner(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pyproject.toml", "[project]\nname = \"x\"\n\n[tool.pytest.ini_options]\ntestpaths = [\"tests\"]\n")
	write(t, dir, "worker/pytest.ini", "[pytest]\naddopts = -q\n")
	write(t, dir, "plain/pyproject.toml", "[project]\nname = \"plain\"\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, c := range snap.Commands {
		got[c.Run] = c.Source
	}
	if got["pytest"] != "pyproject.toml" {
		t.Errorf("root pytest source = %q, want pyproject.toml", got["pytest"])
	}
	if got["cd worker && pytest"] != "worker/pytest.ini" {
		t.Errorf("module pytest source = %q, want worker/pytest.ini", got["cd worker && pytest"])
	}
	if _, ok := got["cd plain && pytest"]; ok {
		t.Error("plain/ has no pytest declaration — nothing must be inferred")
	}
}

func TestTodoStripsBlockCommentClosers(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "report.md", "<!-- TODO: confirm IOC list with the vendor -->\n")
	write(t, dir, "style.css", "/* FIXME: contrast ratio */\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	wants := []string{"TODO: confirm IOC list with the vendor", "FIXME: contrast ratio"}
	if len(snap.Todos) != len(wants) {
		t.Fatalf("todos = %+v, want %d", snap.Todos, len(wants))
	}
	for i, w := range wants {
		if snap.Todos[i].Text != w {
			t.Errorf("todo[%d] = %q, want %q", i, snap.Todos[i].Text, w)
		}
	}
}

func TestTestFilesAndTestdataSkipped(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "scan_test.go", "// TODO: fixture, not a task\n")
	write(t, dir, "testdata/golden.md", "<!-- TODO: fixture too -->\n")
	write(t, dir, "real.go", "// TODO: the only real one\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Todos) != 1 || snap.Todos[0].Path != "real.go" {
		t.Fatalf("todos = %+v, want only real.go", snap.Todos)
	}
}

// Regression: a CLAUDE.md that is only our injected block (created by init
// in a repo that had none) must not count as a captured doc — otherwise
// init changes its own next scan.
func TestMarkerOnlyClaudeMdNotADoc(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "CLAUDE.md", "<!-- kervo:begin -->\nartifact body\n<!-- kervo:end -->\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Docs) != 0 {
		t.Errorf("marker-only CLAUDE.md captured as doc: %+v", snap.Docs)
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

// Field evidence (12-module CTI monorepo): Commands was empty because the
// repo declared everything via compose + pyproject, and four hand-written
// module CLAUDE.md files were invisible. Declared-only, still.
func TestComposeCommands(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "docker-compose.yml", strings.Join([]string{
		"version: \"3.9\"",
		"services:",
		"  api:",
		"    image: acme/api:latest  # inline comment must not leak",
		"    ports:",
		"      - \"8000:8000\"",
		"  worker:",
		"    build: ./worker",
		"  # commented: not a service",
		"volumes:",
		"  data:",
		"", // trailing newline
	}, "\n"))
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	var runs []string
	for _, c := range snap.Commands {
		runs = append(runs, c.Run)
	}
	want := []string{"docker compose up api", "docker compose up worker"}
	if strings.Join(runs, ",") != strings.Join(want, ",") {
		t.Fatalf("commands = %v, want %v", runs, want)
	}
	if snap.Commands[0].Detail != "acme/api:latest" {
		t.Errorf("api detail = %q, want image ref", snap.Commands[0].Detail)
	}
	// "data:" under volumes: must not leak — the services block ended.
	for _, c := range snap.Commands {
		if strings.Contains(c.Run, "data") || strings.Contains(c.Run, "commented") {
			t.Errorf("non-service leaked: %+v", c)
		}
	}
}

func TestComposeVariantFileGetsFlag(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "docker-compose.dev.yml", "services:\n  db:\n    image: postgres:16\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Commands) != 1 || snap.Commands[0].Run != "docker compose -f docker-compose.dev.yml up db" {
		t.Fatalf("commands = %+v, want -f flag for non-default file", snap.Commands)
	}
}

func TestPyprojectScripts(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pyproject.toml", strings.Join([]string{
		"[project]",
		`name = "demo"`,
		`dependencies = ["fastapi>=0.115", "celery"]`,
		"",
		"[project.scripts]",
		`ingest = "acme.ingest:main"`,
		"# a comment",
		`serve = "acme.api:run"`,
		"",
		"[tool.poetry.scripts]",
		`report = "acme.report:cli"`,
		"",
		"[tool.black]",
		"line-length = 100",
	}, "\n"))
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	var runs []string
	for _, c := range snap.Commands {
		runs = append(runs, c.Run)
	}
	want := []string{"ingest", "serve", "report"}
	if strings.Join(runs, ",") != strings.Join(want, ",") {
		t.Fatalf("commands = %v, want %v (file order, script sections only)", runs, want)
	}
	if snap.Commands[0].Detail != "acme.ingest:main" {
		t.Errorf("ingest detail = %q", snap.Commands[0].Detail)
	}
	// [tool.black] entries must not count as scripts.
	for _, c := range snap.Commands {
		if strings.Contains(c.Run, "line-length") {
			t.Errorf("non-script section leaked: %+v", c)
		}
	}
	// Python frameworks come from declared deps, not imports.
	got := strings.Join(snap.Repo.Frameworks, ",")
	for _, wantFw := range []string{"Python", "FastAPI", "Celery"} {
		if !strings.Contains(got, wantFw) {
			t.Errorf("frameworks = %v, missing %s", snap.Repo.Frameworks, wantFw)
		}
	}
}

func TestJustfileCommands(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "justfile", "set shell := [\"bash\", \"-c\"]\n\nbuild:\n  go build ./...\n\ntest: build\n  go test ./...\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	var runs []string
	for _, c := range snap.Commands {
		runs = append(runs, c.Run)
	}
	want := []string{"just build", "just test"}
	if strings.Join(runs, ",") != strings.Join(want, ",") {
		t.Fatalf("commands = %v, want %v", runs, want)
	}
	if snap.Commands[0].Detail != "go build ./..." {
		t.Errorf("build detail = %q", snap.Commands[0].Detail)
	}
}

func TestDockerFrameworkDetection(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "Dockerfile", "FROM golang:1.26\n")
	write(t, dir, "docker-compose.yml", "services:\n  app:\n    build: .\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(snap.Repo.Frameworks, ",")
	for _, want := range []string{"Docker", "Docker Compose"} {
		if !strings.Contains(got, want) {
			t.Errorf("frameworks = %v, missing %s", snap.Repo.Frameworks, want)
		}
	}
}

// Monorepo: top-level module dirs often carry their own CLAUDE.md/README.md
// (field evidence: 4 hand-written module CLAUDE.md files ignored on a real
// 12-module repo). Root docs come first; module docs follow in lexical order.
func TestModuleDocsCaptured(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "README.md", "# Root\n\nRoot readme.\n")
	write(t, dir, "api/CLAUDE.md", "API module context.\n")
	write(t, dir, "api/README.md", "API readme.\n")
	write(t, dir, "database/CLAUDE.md", "DB module context.\n")
	write(t, dir, "node_modules/pkg/README.md", "vendored, skipped\n")
	write(t, dir, ".hidden/CLAUDE.md", "hidden, skipped\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, d := range snap.Docs {
		paths = append(paths, filepath.ToSlash(d.Path))
	}
	want := []string{"README.md", "api/CLAUDE.md", "api/README.md", "database/CLAUDE.md"}
	if strings.Join(paths, ",") != strings.Join(want, ",") {
		t.Fatalf("docs = %v, want %v", paths, want)
	}
}

// Module CLAUDE.md files get the same marker-strip treatment as the root one.
func TestModuleClaudeMdMarkerStripped(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "api/CLAUDE.md", "<!-- kervo:begin -->\nartifact body\n<!-- kervo:end -->\n")
	write(t, dir, "db/CLAUDE.md", "human notes\n<!-- kervo:begin -->\ninjected\n<!-- kervo:end -->\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Docs) != 1 {
		t.Fatalf("docs = %+v, want only db/CLAUDE.md (marker-only skipped)", snap.Docs)
	}
	if strings.Contains(snap.Docs[0].Content, "injected") {
		t.Error("module CLAUDE.md marker block not stripped")
	}
}

func TestDocsCapOnManyModules(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "README.md", "# Root\n")
	for _, m := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"} {
		write(t, dir, filepath.Join(m, "CLAUDE.md"), "module "+m+"\n")
		write(t, dir, filepath.Join(m, "README.md"), "readme "+m+"\n")
	}
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Docs) != maxDocs {
		t.Fatalf("docs = %d, want capped at %d", len(snap.Docs), maxDocs)
	}
}

// Monorepo manifests: root can be bare while modules declare everything
// (field evidence: analyzer/pyproject.toml + processor/pyproject.toml,
// no root pyproject at all). One level deep, vendored dirs excluded.
func TestModulePyprojectScripts(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "analyzer/pyproject.toml", "[project]\nname = \"analyzer\"\ndependencies = [\"celery>=5\"]\n\n[project.scripts]\nanalyze = \"analyzer.cli:main\"\n")
	write(t, dir, "processor/pyproject.toml", "[tool.poetry.scripts]\nprocess = \"processor.run:cli\"\n")
	write(t, dir, "node_modules/x/pyproject.toml", "[project.scripts]\nvendored = \"x:y\"\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	var runs, sources []string
	for _, c := range snap.Commands {
		runs = append(runs, c.Run)
		sources = append(sources, filepath.ToSlash(c.Source))
	}
	if strings.Join(runs, ",") != "analyze,process" {
		t.Fatalf("commands = %v, want [analyze process] (lexical module order)", runs)
	}
	if sources[0] != "analyzer/pyproject.toml" || sources[1] != "processor/pyproject.toml" {
		t.Errorf("sources = %v, want module-relative manifest paths", sources)
	}
	// Frameworks must surface from module manifests despite a bare root.
	got := strings.Join(snap.Repo.Frameworks, ",")
	for _, want := range []string{"Python", "Celery"} {
		if !strings.Contains(got, want) {
			t.Errorf("frameworks = %v, missing %s", snap.Repo.Frameworks, want)
		}
	}
}

// .kervoignore excludes archival material from the TODO scan (found by
// self-scan: published experiment transcripts quoted TODO comments that
// showed up as open tasks).
func TestKervoignoreExcludesTodoScan(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".kervoignore", "# archive\ndocs/experiments/\n")
	write(t, dir, "docs/experiments/resp.md", "// TODO: quoted from an experiment response\n")
	write(t, dir, "docs/guide.md", "<!-- TODO: real doc task -->\n")
	write(t, dir, "main.go", "// TODO: real code task\n")
	snap, _, err := New().Scan(context.Background(), dir, "")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, td := range snap.Todos {
		paths = append(paths, filepath.ToSlash(td.Path))
	}
	if strings.Join(paths, ",") != "docs/guide.md,main.go" {
		t.Fatalf("todos = %v, want ignored dir excluded but siblings kept", paths)
	}
}
