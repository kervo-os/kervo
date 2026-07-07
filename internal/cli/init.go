package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/kervo-os/kervo/internal/core/fact"
)

// initBudget is the Mode 1 perf contract (ARCH-0001 §10.3): the LLM-free
// path must finish in 30s. Blowing it is a bug, not a config issue.
const initBudget = 30 * time.Second

// runInit: full scan -> deterministic skeleton -> .kervo/artifact.md
// -> CLAUDE.md marker injection.
// Wiring order per ARCH-0001 §9: gitexec+files -> core/fact -> core/compiler -> claudecode.
// The scan cursor is persisted so `kervo compile` stays incremental.
func runInit(args []string) error {
	fs := newFlagSet("init")
	dir := fs.String("dir", ".", "workspace directory")
	langFlag := fs.String("lang", "", "artifact language: en, ko, ja (default: en)")
	injectFlag := fs.String("inject", "", "consumer-file injection: block (full artifact) or import (one @-line)")
	consumersFlag := fs.String("consumers", "", "consumer targets: claude,codex,both,auto (default: ask on TTY, else auto)")
	hooksFlag := fs.String("hooks", "", "wire Claude Code capture hooks: yes|no (default: ask on TTY)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lang, err := resolveLang(*dir, *langFlag)
	if err != nil {
		return err
	}
	inject, err := resolveInject(*dir, *injectFlag)
	if err != nil {
		return err
	}
	consumers, err := resolveConsumersForInit(*dir, *consumersFlag)
	if err != nil {
		return err
	}
	wantHooks, err := resolveHooksWiring(*hooksFlag, consumers)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), initBudget)
	defer cancel()

	// init is always Mode 1 — the first experience must never depend on a
	// semantic provider (PRD §6: onboarding = mode order).
	snap, cursor, skeleton, err := buildSkeleton(ctx, *dir, lang)
	if err != nil {
		return err
	}
	injected, err := writeOutputs(ctx, *dir, skeleton, cursor, lang, inject, consumers)
	if err != nil {
		return err
	}
	fmt.Print(renderColdStart(newUI(), snap, Version, injected))
	if wantHooks {
		status, err := wireClaudeHooks(*dir)
		if err != nil {
			return err
		}
		fmt.Println("  " + "Hooks      .claude/settings.json — " + status)
	}
	if status, err := wireGitAutoCompile(*dir); err == nil {
		fmt.Println("  Auto       .git/hooks — " + status)
	}
	codexOnly := true
	for _, c := range injected {
		if c == consumerClaude {
			codexOnly = false
		}
	}
	if codexOnly {
		fmt.Println("  Codex      AGENTS.md carries the context and the write-back protocol; optional MCP: add kervo to ~/.codex/config.toml")
	}
	return nil
}

// mergeSnapshots overlays the files scan onto the git scan. gitexec owns
// repo identity, commits, and modules; files owns docs, todos, frameworks.
func mergeSnapshots(git, fls fact.Snapshot) fact.Snapshot {
	git.Todos = fls.Todos
	git.Docs = fls.Docs
	git.Repo.Frameworks = fls.Repo.Frameworks
	git.Commands = fls.Commands
	git.Partial = git.Partial || fls.Partial
	return git
}
