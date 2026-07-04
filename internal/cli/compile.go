package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/consumer/claudecode"
	"github.com/kervo-os/kervo/internal/adapters/semantic/consumer"
	"github.com/kervo-os/kervo/internal/adapters/semantic/openaicompat"
	"github.com/kervo-os/kervo/internal/adapters/source/files"
	"github.com/kervo-os/kervo/internal/adapters/source/gitexec"
	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/compiler"
	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/i18n"
)

// runCompile: rescan -> deterministic skeleton -> attach staged Enhancement
// proposals (Mode 2, file transport) -> artifact + CLAUDE.md injection.
// Degradation is the RFC-0003 §4 contract: any semantic failure demotes to
// the fact-only skeleton with a warning — never a failed run.
// (Event-store replay and Mode 3 backends join here later; the cursor is
// refreshed for future incremental scans.)
func runCompile(args []string) error {
	fs := newFlagSet("compile")
	dir := fs.String("dir", ".", "workspace directory")
	langFlag := fs.String("lang", "", "artifact language: en, ko, ja (default: workspace setting or en)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lang, err := resolveLang(*dir, *langFlag)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), initBudget)
	defer cancel()

	snap, cursor, skeleton, err := buildSkeleton(ctx, *dir, lang)
	if err != nil {
		return err
	}

	// RFC-0003 §4 fallback order: Mode 3 (backend) → Mode 2 (staged
	// proposals) → Mode 1 (fact-only). A mode failing is a demotion with a
	// warning, never a failed run.
	rendered := skeleton
	mode := "Mode 1 — Fact-only"
	var enh []artifact.Enhancement

	if backend, berr := openaicompat.FromEnv(lang); berr != nil {
		fmt.Fprintf(os.Stderr, "kervo: Mode 3 misconfigured, falling back: %v\n", berr)
	} else if backend != nil {
		// The LLM is the real bottleneck — it gets its own budget, off the
		// 30s Mode-1 contract.
		semCtx, semCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		got, perr := backend.Propose(semCtx, skeleton, snap)
		semCancel()
		if perr != nil {
			fmt.Fprintf(os.Stderr, "kervo: Mode 3 failed, falling back: %v\n", perr)
		} else {
			enh = got
			mode = "Mode 3 — backend:" + backend.Model
		}
	}

	if enh == nil {
		got, perr := consumer.FileProposals{Dir: *dir}.Propose(ctx, skeleton, snap)
		switch {
		case perr != nil:
			fmt.Fprintf(os.Stderr, "kervo: Mode 2 proposals invalid, fact-only: %v\n", perr)
		case len(got) > 0:
			enh = got
			mode = fmt.Sprintf("Mode 2 — %d proposals attached (generated)", len(got))
		}
	}

	if len(enh) > 0 {
		attached, aerr := compiler.Attach(skeleton, enh)
		if aerr != nil {
			fmt.Fprintf(os.Stderr, "kervo: attach failed, fact-only: %v\n", aerr)
			rendered, mode = skeleton, "Mode 1 — Fact-only"
		} else {
			rendered = attached
		}
	}

	if err := writeOutputs(ctx, *dir, rendered, cursor, lang); err != nil {
		return err
	}
	fmt.Printf("Artifact: .kervo/artifact.md (%s)\n", mode)
	fmt.Println("Injected: CLAUDE.md (marker block)")
	return nil
}

// buildSkeleton runs the shared fact pipeline: scan git + files, merge,
// render the deterministic skeleton in lang.
func buildSkeleton(ctx context.Context, dir string, lang i18n.Lang) (fact.Snapshot, string, string, error) {
	snap, cursor, err := gitexec.New().Scan(ctx, dir, "")
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	fsnap, _, err := files.New().Scan(ctx, dir, "")
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	snap = mergeSnapshots(snap, fsnap)
	skeleton, err := compiler.BuildSkeleton(snap, lang)
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	return snap, cursor, skeleton, nil
}

// writeOutputs stages the injection before any write (no partial state),
// then persists artifact, cursor, language, the consumer file, and the
// RFC-0005 ignore rules (derived state never gets committed).
func writeOutputs(ctx context.Context, dir, rendered, cursor string, lang i18n.Lang) error {
	injector := claudecode.Injector{}
	injPath, injContent, err := injector.Render(dir, rendered)
	if err != nil {
		return err
	}
	stateDir := filepath.Join(dir, ".kervo")
	if err := os.MkdirAll(filepath.Join(stateDir, "cache"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "artifact.md"), []byte(rendered), 0o644); err != nil {
		return err
	}
	// Incremental scan cursor lives under cache/ (RFC-0005 §2.1 layout).
	if err := os.WriteFile(filepath.Join(stateDir, "cache", "cursor"), []byte(cursor+"\n"), 0o644); err != nil {
		return err
	}
	// lang is a workspace choice (not derivable) — stays committable.
	if err := os.WriteFile(filepath.Join(stateDir, "lang"), []byte(string(lang)+"\n"), 0o644); err != nil {
		return err
	}
	if err := ensureGitignore(dir); err != nil {
		return err
	}
	return injector.Apply(injPath, injContent)
}
