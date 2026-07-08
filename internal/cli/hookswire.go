package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// claudeHooksJSON is the exact block the README documents — one constant
// so docs, wizard, and tests can never drift apart.
const claudeHooksJSON = `{
  "hooks": {
    "UserPromptSubmit": [
      { "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ],
    "SessionStart": [
      { "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ],
    "PostToolUse": [
      { "matcher": "Edit|Write",
        "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ]
  }
}
`

// resolveHooksWiring decides whether init wires Claude Code capture hooks.
// Flag wins; an interactive init asks (default yes) when claude is among
// the consumers; non-TTY stays silent — CI keeps the legacy behavior.
// Codex has no per-repo hook system, so the question never fires for a
// codex-only choice: absent capability is stated, not faked.
func resolveHooksWiring(flagVal string, consumers []string) (bool, error) {
	claude := false
	for _, c := range consumers {
		if c == consumerClaude {
			claude = true
		}
	}
	switch flagVal {
	case "yes":
		if !claude {
			return false, fmt.Errorf("hooks: consumers do not include claude — nothing to wire")
		}
		return true, nil
	case "no":
		return false, nil
	case "":
		// fall through to the interactive default
	default:
		return false, fmt.Errorf("hooks: unsupported %q (supported: yes, no)", flagVal)
	}
	if !claude || !stdinIsTTY() {
		return false, nil
	}
	fmt.Print("Wire Claude Code hooks for automatic capture? [Y/n]: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	ans := strings.ToLower(strings.TrimSpace(line))
	return ans == "" || ans == "y" || ans == "yes", nil
}

// The two moments a workspace's facts change without a kervo command
// running: a commit and a pull. pre-commit (NOT post-commit) is
// deliberate — compiling after the commit dirties the tree with the
// refreshed digest and never converges (each fix-up commit changes
// Recent Changes again; found by dogfood within minutes of shipping
// post-commit). Compiling before and staging the consumer files makes
// every commit carry its own fresh digest and leaves the tree clean.
const preCommitScript = `#!/bin/sh
# kervo: the commit carries a fresh context artifact and its ledger
kervo compile >/dev/null 2>&1 || exit 0
for f in CLAUDE.md AGENTS.md; do [ -f "$f" ] && git add -- "$f"; done
[ -d .kervo/events ] && git add -- .kervo/events
exit 0
`

const postMergeScript = "#!/bin/sh\nkervo compile >/dev/null 2>&1 || true\n"

// legacyPostCommitScript is the v0.21.0 shape — migrated away on the
// next wire because it perpetually dirties the tree (see above). Only
// an exact match is removed; anything else is a foreign hook.
const legacyPostCommitScript = postMergeScript

// legacyPreCommitScript is the v0.21.1–v0.22.0 shape — superseded because
// leaving events unstaged forces standalone "ledger:" commits, which is
// exactly the review noise teams complain about. Exact match only.
const legacyPreCommitScript = `#!/bin/sh
# kervo: the commit carries a fresh context artifact
kervo compile >/dev/null 2>&1 || exit 0
for f in CLAUDE.md AGENTS.md; do [ -f "$f" ] && git add -- "$f"; done
exit 0
`

// wireGitAutoCompile installs post-commit and post-merge hooks. This is
// not opt-in: a memory layer for a team that stores its work as commits
// must watch commits by default — a stale artifact quietly breaks the
// product's one promise. It runs on every init AND compile (workspace
// plumbing, same rank as .gitignore/.gitattributes) because git hooks
// are machine-local: a teammate's first `kervo compile` wires their
// machine. Same three safe outcomes as wireClaudeHooks; a foreign hook
// is never rewritten — replacing our hook with your own IS the opt-out.
func wireGitAutoCompile(dir string) (string, error) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--git-path", "hooks").Output()
	if err != nil {
		return "", fmt.Errorf("autocompile: not a git repository: %v", err)
	}
	hooksDir := strings.TrimSpace(string(out))
	if !filepath.IsAbs(hooksDir) {
		hooksDir = filepath.Join(dir, hooksDir)
	}
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return "", err
	}
	// Migrate our own v0.21.0 post-commit (exact match only — anything
	// the user touched is theirs).
	if raw, err := os.ReadFile(filepath.Join(hooksDir, "post-commit")); err == nil && string(raw) == legacyPostCommitScript {
		_ = os.Remove(filepath.Join(hooksDir, "post-commit"))
	}
	var wired, kept []string
	for name, script := range map[string]string{"pre-commit": preCommitScript, "post-merge": postMergeScript} {
		p := filepath.Join(hooksDir, name)
		raw, err := os.ReadFile(p)
		switch {
		case err == nil && string(raw) == legacyPreCommitScript:
			// our own previous shape — migrate to the current script
		case err == nil && strings.Contains(string(raw), "kervo compile"):
			continue // already wired
		case err == nil:
			kept = append(kept, name) // someone else's hook — not ours to rewrite
			continue
		case err != nil && !os.IsNotExist(err):
			return "", err
		}
		if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
			return "", err
		}
		wired = append(wired, name)
	}
	sort.Strings(wired)
	sort.Strings(kept)
	switch {
	case len(kept) > 0 && len(wired) == 0:
		return "left untouched — " + strings.Join(kept, ", ") + " carry your own hooks; add `kervo compile` to them yourself", nil
	case len(kept) > 0:
		return strings.Join(wired, ", ") + " wired; " + strings.Join(kept, ", ") + " left untouched (your own hooks)", nil
	case len(wired) == 0:
		return "already wired", nil
	default:
		return "pre-commit + post-merge — every commit carries a fresh artifact, pulls refresh it too", nil
	}
}

// wireClaudeHooks connects automatic capture. Three outcomes, all safe:
// absent → created; already carries kervo → untouched; carries someone
// else's config → untouched with a pointer — rewriting a user's settings
// file is not ours to do (a JSON merge that guesses wrong breaks every
// hook they had).
func wireClaudeHooks(dir string) (string, error) {
	p := filepath.Join(dir, ".claude", "settings.json")
	raw, err := os.ReadFile(p)
	switch {
	case err == nil && strings.Contains(string(raw), "kervo hook"):
		return "already wired", nil
	case err == nil:
		return "left untouched — it has your own config; add the hooks block from the README", nil
	case !os.IsNotExist(err):
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(claudeHooksJSON), 0o644); err != nil {
		return "", err
	}
	return "created — commit it and capture fires for every teammate", nil
}
