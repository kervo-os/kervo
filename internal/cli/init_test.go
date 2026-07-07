package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=t@t",
		"GIT_AUTHOR_DATE=2026-07-01T10:00:00+09:00",
		"GIT_COMMITTER_DATE=2026-07-01T10:00:00+09:00",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInitEndToEnd(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "# demo\n\nA workspace for the init e2e test.\n")
	writeFile(t, dir, "CLAUDE.md", "# House rules\n\nalways run tests\n")
	writeFile(t, dir, "src/main.go", "package main\n// TODO: handle signals\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "initial import")

	if err := runInit([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}

	artifactPath := filepath.Join(dir, ".kervo", "artifact.md")
	art, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("artifact.md not written: %v", err)
	}
	for _, want := range []string{
		"# Context Artifact",
		"- Branch: main",
		"initial import",
		"src/main.go:2 — TODO: handle signals",
		"- src/ (1 files)",
		"A workspace for the init e2e test.",
	} {
		if !strings.Contains(string(art), want) {
			t.Errorf("artifact missing %q", want)
		}
	}

	cursor, err := os.ReadFile(filepath.Join(dir, ".kervo", "cache", "cursor"))
	if err != nil || len(strings.TrimSpace(string(cursor))) != 40 {
		t.Errorf("cursor not persisted under cache/: %q err=%v", cursor, err)
	}

	// RFC-0005 §2.4: init auto-registers derived-state ignore rules,
	// idempotently and without touching existing human rules.
	gi, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
	for _, rule := range []string{".kervo/artifact.md", ".kervo/cache/"} {
		if !strings.Contains(string(gi), rule) {
			t.Errorf(".gitignore missing %q", rule)
		}
	}
	if strings.Count(string(gi), ".kervo/artifact.md") != 1 {
		t.Error("ignore rules duplicated across runs (init ran twice above)")
	}

	claude, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# House rules", "always run tests", "<!-- kervo:begin -->", "<!-- kervo:end -->"} {
		if !strings.Contains(string(claude), want) {
			t.Errorf("CLAUDE.md missing %q", want)
		}
	}

	// AGENTS.md is opt-in by presence — init must never create it.
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("AGENTS.md was created uninvited")
	}

	// Re-running init on an unchanged workspace must be byte-idempotent —
	// the determinism contract, observed end-to-end.
	if err := runInit([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	art2, _ := os.ReadFile(artifactPath)
	if string(art) != string(art2) {
		t.Error("artifact.md differs across identical runs")
	}
	claude2, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(claude) != string(claude2) {
		t.Error("CLAUDE.md differs across identical runs")
	}
}

// Field evidence 2026-07-06 (codex A/B): AGENTS.md readers get zero context
// from CLAUDE.md. An existing AGENTS.md opts the workspace into a second
// injection target under the same marker contract.
func TestInitInjectsExistingAgentsMd(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "# demo\n")
	writeFile(t, dir, "AGENTS.md", "# Codex rules\n\nbe terse\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# Codex rules", "be terse", "<!-- kervo:begin -->", "<!-- kervo:end -->", "# Context Artifact"} {
		if !strings.Contains(string(agents), want) {
			t.Errorf("AGENTS.md missing %q", want)
		}
	}

	// Idempotency holds across both consumer files.
	if err := runInit([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	agents2, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if string(agents) != string(agents2) {
		t.Error("AGENTS.md differs across identical runs")
	}
}

func TestInitConsumersCodexCreatesAgentsOnly(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "# demo\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir, "-consumers", "codex"}); err != nil {
		t.Fatal(err)
	}
	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md not created for codex consumer: %v", err)
	}
	for _, want := range []string{"<!-- kervo:begin -->", "# Context Artifact", "<!-- kervo:end -->"} {
		if !strings.Contains(string(agents), want) {
			t.Errorf("AGENTS.md missing %q", want)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("CLAUDE.md should not be created when consumers=codex")
	}
	saved, err := os.ReadFile(filepath.Join(dir, ".kervo", "consumers"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(saved)) != "AGENTS.md" {
		t.Errorf("consumer choice not persisted: %q", saved)
	}

	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("compile ignored persisted codex-only consumer choice")
	}
}

func TestInitConsumersBothCreatesBothFiles(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "# demo\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir, "-consumers", "both"}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("%s not created: %v", name, err)
		}
		if !strings.Contains(string(raw), "# Context Artifact") {
			t.Errorf("%s missing artifact", name)
		}
	}
}

// The staging invariant covers the second target too: a corrupt AGENTS.md
// must fail the run before any write, including the CLAUDE.md injection.
func TestInitCorruptAgentsMarkersLeavesNoPartialState(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main.go", "package main\n")
	writeFile(t, dir, "AGENTS.md", "human\n<!-- kervo:begin -->\nno end marker\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir}); err == nil {
		t.Fatal("expected corrupt-marker error")
	}
	if _, err := os.Stat(filepath.Join(dir, ".kervo")); !os.IsNotExist(err) {
		t.Error(".kervo was created despite the failed injection (partial state)")
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Error("CLAUDE.md was written despite the failed AGENTS.md staging")
	}
}

// Decision 01KWTFTX: import mode trades the zero-command clone for a clean
// CLAUDE.md — one @-line in the block, the full artifact in .kervo/. The
// choice persists per workspace and flips back losslessly.
func TestInitInjectImportMode(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "# demo\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir, "-inject", "import"}); err != nil {
		t.Fatal(err)
	}
	claude, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(claude), "@.kervo/artifact.md") {
		t.Error("import mode missing the @-line")
	}
	if strings.Contains(string(claude), "# Context Artifact") {
		t.Error("import mode must not inline the artifact")
	}
	art, err := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if err != nil || !strings.Contains(string(art), "# Context Artifact") {
		t.Fatalf("full artifact must live in .kervo/artifact.md: %v", err)
	}
	mode, _ := os.ReadFile(filepath.Join(dir, ".kervo", "inject"))
	if strings.TrimSpace(string(mode)) != "import" {
		t.Errorf("inject mode not persisted: %q", mode)
	}

	// No flag → the persisted choice holds, byte-idempotently.
	if err := runInit([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	claude2, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if string(claude) != string(claude2) {
		t.Error("import mode not idempotent across runs")
	}

	// Flipping back restores the full block.
	if err := runInit([]string{"-dir", dir, "-inject", "block"}); err != nil {
		t.Fatal(err)
	}
	claude3, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(string(claude3), "# Context Artifact") {
		t.Error("block mode did not restore the inline artifact")
	}

	if err := runInit([]string{"-dir", dir, "-inject", "weird"}); err == nil {
		t.Error("unsupported inject mode must error")
	}
}

func TestInitOutsideRepoFails(t *testing.T) {
	if err := runInit([]string{"-dir", t.TempDir()}); err == nil {
		t.Fatal("expected error outside a git repository")
	}
}

// Regression: a failing injection must fail the WHOLE run before any write.
// Previously artifact.md was written first, leaving a half-applied init.
func TestInitCorruptMarkersLeavesNoPartialState(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main.go", "package main\n")
	writeFile(t, dir, "CLAUDE.md", "human\n<!-- kervo:begin -->\nno end marker\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir}); err == nil {
		t.Fatal("expected corrupt-marker error")
	}
	if _, err := os.Stat(filepath.Join(dir, ".kervo")); !os.IsNotExist(err) {
		t.Error(".kervo was created despite the failed injection (partial state)")
	}
}

// -hooks yes writes the documented settings block once; a settings file
// that carries someone else's config is never rewritten.
func TestInitWiresClaudeHooks(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "# demo\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir, "-hooks", "yes"}); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, ".claude", "settings.json")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range []string{"UserPromptSubmit", "SessionStart", "PostToolUse", "kervo hook || true"} {
		if !strings.Contains(string(raw), ev) {
			t.Errorf("settings missing %q", ev)
		}
	}
	// Idempotent: second run leaves the file byte-identical.
	if err := runInit([]string{"-dir", dir, "-hooks", "yes"}); err != nil {
		t.Fatal(err)
	}
	raw2, _ := os.ReadFile(p)
	if string(raw) != string(raw2) {
		t.Error("re-wiring rewrote an already-wired settings file")
	}

	// Someone else's config is not ours to rewrite.
	custom := t.TempDir()
	git(t, custom, "init", "-q", "-b", "main")
	writeFile(t, custom, "README.md", "# demo\n")
	writeFile(t, custom, ".claude/settings.json", `{"permissions":{"allow":["Read"]}}`)
	git(t, custom, "add", ".")
	git(t, custom, "commit", "-q", "-m", "x")
	if err := runInit([]string{"-dir", custom, "-hooks", "yes"}); err != nil {
		t.Fatal(err)
	}
	kept, _ := os.ReadFile(filepath.Join(custom, ".claude", "settings.json"))
	if string(kept) != `{"permissions":{"allow":["Read"]}}` {
		t.Error("existing user settings were modified")
	}

	// Codex-only consumers have nothing to wire — explicit yes must error.
	codex := t.TempDir()
	git(t, codex, "init", "-q", "-b", "main")
	writeFile(t, codex, "README.md", "# demo\n")
	git(t, codex, "add", ".")
	git(t, codex, "commit", "-q", "-m", "x")
	if err := runInit([]string{"-dir", codex, "-consumers", "codex", "-hooks", "yes"}); err == nil {
		t.Error("hooks=yes with codex-only consumers must error")
	}

	// Non-TTY default: no prompt, no wiring — CI behavior unchanged.
	plain := t.TempDir()
	git(t, plain, "init", "-q", "-b", "main")
	writeFile(t, plain, "README.md", "# demo\n")
	git(t, plain, "add", ".")
	git(t, plain, "commit", "-q", "-m", "x")
	if err := runInit([]string{"-dir", plain}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(plain, ".claude")); !os.IsNotExist(err) {
		t.Error(".claude created without being asked (non-TTY must stay silent)")
	}
}
