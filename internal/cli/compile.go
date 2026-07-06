package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/consumer/claudecode"
	"github.com/kervo-os/kervo/internal/adapters/semantic/consumer"
	"github.com/kervo-os/kervo/internal/adapters/semantic/openaicompat"
	"github.com/kervo-os/kervo/internal/adapters/source/files"
	"github.com/kervo-os/kervo/internal/adapters/source/gitexec"
	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/compiler"
	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/i18n"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runCompile: rescan -> skeleton -> ingest new proposals into the ledger
// (Mode 3 backend, else staged Mode 2 file — RFC-0003 §4 chain, failures
// demote with a warning) -> age-based stale sweep -> render EVERYTHING from
// the replayed trust view (PRD §7.2 treatment: Verified first, labeled
// Observed/Generated, Stale listed with reasons, Deprecated excluded from
// the artifact but preserved in the ledger).
func runCompile(args []string) error {
	fs := newFlagSet("compile")
	dir := fs.String("dir", ".", "workspace directory")
	langFlag := fs.String("lang", "", "artifact language: en, ko, ja (default: workspace setting or en)")
	injectFlag := fs.String("inject", "", "consumer-file injection: block (full artifact) or import (one @-line)")
	staleAfter := fs.Duration("stale-after", 720*time.Hour, "demote generated/observed observations older than this")
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

	snap, cursor, skeleton, err := buildSkeleton(ctx, *dir, lang)
	if err != nil {
		return err
	}
	store := jsonl.Open(*dir)

	// ── 1. Collect this run's proposals (Mode 3 → Mode 2 → none).
	mode := "Mode 1 — Fact-only"
	var proposals []artifact.Enhancement
	usedBackend := false
	if backend, berr := openaicompat.FromEnv(lang); berr != nil {
		fmt.Fprintf(os.Stderr, "kervo: Mode 3 misconfigured, falling back: %v\n", berr)
	} else if backend != nil {
		semCtx, semCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		got, perr := backend.Propose(semCtx, skeleton, snap)
		semCancel()
		if perr != nil {
			fmt.Fprintf(os.Stderr, "kervo: Mode 3 failed, falling back: %v\n", perr)
		} else {
			proposals, usedBackend = got, true
			mode = "Mode 3 — backend:" + backend.Model
		}
	}
	if !usedBackend {
		got, perr := consumer.FileProposals{Dir: *dir}.Propose(ctx, skeleton, snap)
		switch {
		case perr != nil:
			fmt.Fprintf(os.Stderr, "kervo: Mode 2 proposals invalid, fact-only: %v\n", perr)
		case len(got) > 0:
			proposals = got
			mode = "Mode 2 — staged proposals"
		}
	}

	// ── 2. Ingest proposals as Generated observations. Backends gap-fill
	// only (never spam slots that already carry live context); staged
	// proposals dedup by (slot, body).
	folder, err := replayFolder(store)
	if err != nil {
		return err
	}
	repo := filepath.Base(mustAbs(*dir))
	if n, err := ingestProposals(store, folder, proposals, repo, usedBackend); err != nil {
		return err
	} else if n > 0 {
		fmt.Fprintf(os.Stderr, "kervo: %d new observation(s) entered the ledger as generated\n", n)
	}

	// ── 3. Conservative stale sweep: age only, never semantics (PRD §7.2);
	// Verified survives — only a human may demote what a human confirmed.
	folder, err = replayFolder(store)
	if err != nil {
		return err
	}
	if err := sweepStale(store, folder, repo, *staleAfter); err != nil {
		return err
	}

	// ── 4. Render from the trust view.
	folder, err = replayFolder(store)
	if err != nil {
		return err
	}
	enh, staleNotes := renderView(folder)
	rendered := skeleton
	if len(enh) > 0 {
		rendered, err = compiler.Attach(skeleton, enh)
		if err != nil {
			fmt.Fprintf(os.Stderr, "kervo: attach failed, fact-only: %v\n", err)
			rendered = skeleton
		}
	}
	rendered, err = compiler.AttachStale(rendered, staleNotes)
	if err != nil {
		return err
	}

	injected, err := writeOutputs(ctx, *dir, rendered, cursor, lang, inject)
	if err != nil {
		return err
	}
	fmt.Printf("Artifact: .kervo/artifact.md (%s · ledger: %d live, %d stale)\n", mode, len(enh), len(staleNotes))
	fmt.Printf("Injected: %s (marker block)\n", strings.Join(injected, ", "))
	return nil
}

func ingestProposals(store *jsonl.Store, folder *trust.Folder, proposals []artifact.Enhancement, repo string, gapFillOnly bool) (int, error) {
	if len(proposals) == 0 {
		return 0, nil
	}
	liveSlot := map[string]bool{}
	seen := map[string]bool{}
	for _, o := range folder.Observations() {
		if o.State == trust.Deprecated {
			continue
		}
		slot := slotForType(o.Type)
		seen[slot+"\x00"+o.Body] = true
		if o.State != trust.Stale {
			liveSlot[slot] = true
		}
	}
	n := 0
	for _, p := range proposals {
		if gapFillOnly && liveSlot[p.Slot] {
			continue
		}
		if seen[p.Slot+"\x00"+p.Body] {
			continue
		}
		payload, err := json.Marshal(map[string]string{"body": p.Body})
		if err != nil {
			return n, err
		}
		if _, err := store.Append(context.Background(), event.Event{
			Kind:    event.KindObservation,
			Type:    typeForSlot(p.Slot),
			Repo:    repo,
			Actor:   p.Source, // provider identity — non-human, so it enters as Generated
			Source:  p.Source,
			Payload: json.RawMessage(payload),
		}); err != nil {
			return n, err
		}
		seen[p.Slot+"\x00"+p.Body] = true
		n++
	}
	return n, nil
}

func sweepStale(store *jsonl.Store, folder *trust.Folder, repo string, staleAfter time.Duration) error {
	if staleAfter <= 0 {
		return nil
	}
	now := time.Now().UTC()
	for _, o := range folder.Observations() {
		if o.State != trust.Generated && o.State != trust.Observed {
			continue
		}
		age := now.Sub(o.At)
		if age <= staleAfter {
			continue
		}
		reason := fmt.Sprintf("age %dd > %dd without reaffirmation", int(age.Hours()/24), int(staleAfter.Hours()/24))
		payload, err := json.Marshal(map[string]string{"to": string(trust.Stale), "reason": reason})
		if err != nil {
			return err
		}
		if _, err := store.Append(context.Background(), event.Event{
			Kind: event.KindTransition, Type: "transition", Repo: repo,
			Actor: "system", Source: "system", Ref: o.ID,
			Payload: json.RawMessage(payload),
		}); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "kervo: %s demoted to stale (%s)\n", shortID(o.ID), reason)
	}
	return nil
}

// renderView applies the PRD §7.2 treatment table to the folded ledger.
func renderView(folder *trust.Folder) ([]artifact.Enhancement, []compiler.StaleNote) {
	var enh []artifact.Enhancement
	var stale []compiler.StaleNote
	for _, o := range folder.Observations() {
		switch o.State {
		case trust.Verified, trust.Observed, trust.Generated:
			body := o.Body
			// Evidence travels with the claim so any reader can re-check it.
			if o.Evidence != "" {
				body += "\nEvidence: " + o.Evidence
			}
			enh = append(enh, artifact.Enhancement{
				Slot: slotForType(o.Type), Body: body,
				State: o.State, Source: o.LastActor, Conflict: o.Conflict,
			})
		case trust.Stale:
			reason := o.Reason
			if reason == "" {
				reason = "stale"
			}
			stale = append(stale, compiler.StaleNote{Body: o.Body, Reason: reason, Actor: o.LastActor})
		case trust.Deprecated:
			// excluded from the artifact; the ledger keeps the history
		}
	}
	rank := map[trust.State]int{trust.Verified: 0, trust.Observed: 1, trust.Generated: 2}
	sort.SliceStable(enh, func(i, j int) bool { return rank[enh[i].State] < rank[enh[j].State] })
	return enh, stale
}

func slotForType(t string) string {
	switch t {
	case "decision":
		return artifact.SlotDecisions
	case "risk":
		return artifact.SlotRisks
	case "goal":
		return artifact.SlotGoal
	default: // note, summary, correction, ...
		return artifact.SlotSummaries
	}
}

func typeForSlot(s string) string {
	switch s {
	case artifact.SlotDecisions:
		return "decision"
	case artifact.SlotRisks:
		return "risk"
	case artifact.SlotGoal:
		return "goal"
	default:
		return "summary"
	}
}

func mustAbs(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}

// scanFacts runs the shared fact pipeline: scan git + files, merge.
func scanFacts(ctx context.Context, dir string) (fact.Snapshot, string, error) {
	snap, cursor, err := gitexec.New().Scan(ctx, dir, "")
	if err != nil {
		return fact.Snapshot{}, "", err
	}
	fsnap, _, err := files.New().Scan(ctx, dir, "")
	if err != nil {
		return fact.Snapshot{}, "", err
	}
	return mergeSnapshots(snap, fsnap), cursor, nil
}

// buildSkeleton scans and renders the deterministic skeleton in lang.
func buildSkeleton(ctx context.Context, dir string, lang i18n.Lang) (fact.Snapshot, string, string, error) {
	snap, cursor, err := scanFacts(ctx, dir)
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	skeleton, err := compiler.BuildSkeleton(snap, lang)
	if err != nil {
		return fact.Snapshot{}, "", "", err
	}
	return snap, cursor, skeleton, nil
}

// writeOutputs stages every injection before any write (no partial state),
// then persists artifact, cursor, language, injection mode, the consumer
// files, and the RFC-0005 ignore rules (derived state never gets
// committed). Returns the consumer files that received the block.
func writeOutputs(ctx context.Context, dir, rendered, cursor string, lang i18n.Lang, inject string) ([]string, error) {
	// import mode: the marker block carries one @-line and the full
	// artifact stays in .kervo/artifact.md (decision 01KWTFTX — the
	// clean-CLAUDE.md trade: fresh clones need one `kervo compile`).
	blockContent := rendered
	if inject == injectImport {
		blockContent = importLine
	}
	injector := claudecode.Injector{}
	injPath, injContent, err := injector.Render(dir, blockContent)
	if err != nil {
		return nil, err
	}
	injected := []string{"CLAUDE.md"}
	// AGENTS.md is the second consumer surface (Codex and other AGENTS.md
	// readers get zero context from CLAUDE.md). File presence is the opt-in:
	// kervo injects into an existing AGENTS.md but never creates one.
	var agentsPath, agentsContent string
	if _, statErr := os.Stat(filepath.Join(dir, "AGENTS.md")); statErr == nil {
		agentsPath, agentsContent, err = claudecode.Injector{FileName: "AGENTS.md"}.Render(dir, blockContent)
		if err != nil {
			return nil, err
		}
		injected = append(injected, "AGENTS.md")
	}
	stateDir := filepath.Join(dir, ".kervo")
	if err := os.MkdirAll(filepath.Join(stateDir, "cache"), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "artifact.md"), []byte(rendered), 0o644); err != nil {
		return nil, err
	}
	// Incremental scan cursor lives under cache/ (RFC-0005 §2.1 layout).
	if err := os.WriteFile(filepath.Join(stateDir, "cache", "cursor"), []byte(cursor+"\n"), 0o644); err != nil {
		return nil, err
	}
	// lang and inject are workspace choices (not derivable) — committable.
	if err := os.WriteFile(filepath.Join(stateDir, "lang"), []byte(string(lang)+"\n"), 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(stateDir, "inject"), []byte(inject+"\n"), 0o644); err != nil {
		return nil, err
	}
	if err := ensureGitignore(dir); err != nil {
		return nil, err
	}
	// Machine-local, path-only, best-effort — powers `kervo dash`.
	registerWorkspace(dir)
	if err := injector.Apply(injPath, injContent); err != nil {
		return nil, err
	}
	if agentsPath != "" {
		if err := injector.Apply(agentsPath, agentsContent); err != nil {
			return nil, err
		}
	}
	return injected, nil
}
