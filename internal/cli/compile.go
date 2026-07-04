package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kervo-os/kervo/internal/adapters/consumer/claudecode"
	"github.com/kervo-os/kervo/internal/adapters/semantic/consumer"
	"github.com/kervo-os/kervo/internal/adapters/source/files"
	"github.com/kervo-os/kervo/internal/adapters/source/gitexec"
	"github.com/kervo-os/kervo/internal/core/compiler"
	"github.com/kervo-os/kervo/internal/core/fact"
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
	if err := fs.Parse(args); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), initBudget)
	defer cancel()

	snap, cursor, skeleton, err := buildSkeleton(ctx, *dir)
	if err != nil {
		return err
	}

	rendered := skeleton
	mode := "Mode 1 — Fact-only"
	enh, err := consumer.FileProposals{Dir: *dir}.Propose(ctx, skeleton, snap)
	switch {
	case err != nil:
		fmt.Fprintf(os.Stderr, "kervo: semantic degraded to fact-only: %v\n", err)
	case len(enh) > 0:
		attached, aerr := compiler.Attach(skeleton, enh)
		if aerr != nil {
			fmt.Fprintf(os.Stderr, "kervo: semantic degraded to fact-only: %v\n", aerr)
		} else {
			rendered = attached
			mode = fmt.Sprintf("Mode 2 — %d proposals attached (generated)", len(enh))
		}
	}

	if err := writeOutputs(ctx, *dir, rendered, cursor); err != nil {
		return err
	}
	fmt.Printf("Artifact: .kervo/artifact.md (%s)\n", mode)
	fmt.Println("Injected: CLAUDE.md (marker block)")
	return nil
}

// buildSkeleton runs the shared fact pipeline: scan git + files, merge,
// render the deterministic skeleton.
func buildSkeleton(ctx context.Context, dir string) (fact.Snapshot, string, string, error) {
	snap, cursor, err := gitexec.New().Scan(ctx, dir, "")
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	fsnap, _, err := files.New().Scan(ctx, dir, "")
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	snap = mergeSnapshots(snap, fsnap)
	skeleton, err := compiler.BuildSkeleton(snap)
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	return snap, cursor, skeleton, nil
}

// writeOutputs stages the injection before any write (no partial state),
// then persists artifact, cursor, and the consumer file.
func writeOutputs(ctx context.Context, dir, rendered, cursor string) error {
	injector := claudecode.Injector{}
	injPath, injContent, err := injector.Render(dir, rendered)
	if err != nil {
		return err
	}
	stateDir := filepath.Join(dir, ".kervo")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "artifact.md"), []byte(rendered), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "cursor"), []byte(cursor+"\n"), 0o644); err != nil {
		return err
	}
	return injector.Apply(injPath, injContent)
}
