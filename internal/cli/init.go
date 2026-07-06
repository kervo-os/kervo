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
// (Fact events into the store follow once the sqlite adapter lands; the
// scan cursor is persisted now so `kervo compile` can be incremental.)
func runInit(args []string) error {
	fs := newFlagSet("init")
	dir := fs.String("dir", ".", "workspace directory")
	langFlag := fs.String("lang", "", "artifact language: en, ko, ja (default: en)")
	injectFlag := fs.String("inject", "", "consumer-file injection: block (full artifact) or import (one @-line)")
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
	ctx, cancel := context.WithTimeout(context.Background(), initBudget)
	defer cancel()

	// init is always Mode 1 — the first experience must never depend on a
	// semantic provider (PRD §6: onboarding = mode order).
	snap, cursor, skeleton, err := buildSkeleton(ctx, *dir, lang)
	if err != nil {
		return err
	}
	injected, err := writeOutputs(ctx, *dir, skeleton, cursor, lang, inject)
	if err != nil {
		return err
	}
	fmt.Print(renderColdStart(newUI(), snap, Version, injected))
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
