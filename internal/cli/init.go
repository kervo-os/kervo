package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/consumer/claudecode"
	"github.com/kervo-os/kervo/internal/adapters/source/files"
	"github.com/kervo-os/kervo/internal/adapters/source/gitexec"
	"github.com/kervo-os/kervo/internal/core/compiler"
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
	if err := fs.Parse(args); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), initBudget)
	defer cancel()

	snap, cursor, err := gitexec.New().Scan(ctx, *dir, "")
	if err != nil {
		return err
	}
	fsnap, _, err := files.New().Scan(ctx, *dir, "")
	if err != nil {
		return err
	}
	snap = mergeSnapshots(snap, fsnap)

	skeleton, err := compiler.BuildSkeleton(snap)
	if err != nil {
		return err
	}

	// Stage the injection before any write: a corrupt CLAUDE.md must fail
	// the whole run, not leave a half-applied .kervo/ behind.
	injector := claudecode.Injector{}
	injPath, injContent, err := injector.Render(*dir, skeleton)
	if err != nil {
		return err
	}

	stateDir := filepath.Join(*dir, ".kervo")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "artifact.md"), []byte(skeleton), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "cursor"), []byte(cursor+"\n"), 0o644); err != nil {
		return err
	}

	if err := injector.Apply(injPath, injContent); err != nil {
		return err
	}

	fmt.Print(renderColdStart(newUI(), snap, Version))
	return nil
}

// mergeSnapshots overlays the files scan onto the git scan. gitexec owns
// repo identity, commits, and modules; files owns docs, todos, frameworks.
func mergeSnapshots(git, fls fact.Snapshot) fact.Snapshot {
	git.Todos = fls.Todos
	git.Docs = fls.Docs
	git.Repo.Frameworks = fls.Repo.Frameworks
	git.Partial = git.Partial || fls.Partial
	return git
}
