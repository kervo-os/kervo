package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runTrust records a human judgment as a transition event. No review gate:
// the ledger stores immediately, people correct — never approve-to-save
// (PRD §7.2). State legality is enforced by the core state machine.
func runTrust(args []string) error {
	fs := newFlagSet("trust")
	dir := fs.String("dir", ".", "workspace directory")
	id := fs.String("id", "", "observation ID or unique prefix (required)")
	to := fs.String("to", "", "target state: observed, verified, stale, deprecated (required)")
	reason := fs.String("reason", "", "why (recorded in the ledger)")
	actor := fs.String("actor", "", `who is judging (default "human:<git user.name>")`)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" || *to == "" {
		return fmt.Errorf("trust: -id and -to are required")
	}
	target := trust.State(*to)

	store := jsonl.Open(*dir)
	folder, err := replayFolder(store)
	if err != nil {
		return err
	}
	obs, err := findByPrefix(folder, *id)
	if err != nil {
		return err
	}
	if !trust.CanTransition(obs.State, target) {
		return fmt.Errorf("trust: %s → %s is not a legal transition (current state: %s)", obs.State, target, obs.State)
	}
	if err := appendTransition(store, *dir, *actor, obs, target, *reason); err != nil {
		return err
	}
	fmt.Printf("%s: %s → %s\n", shortID(obs.ID), obs.State, target)
	return nil
}

// appendTransition records a human judgment in the ledger — shared by
// `trust` (by-ID primitive) and `review` (triage queue).
func appendTransition(store *jsonl.Store, dir, actorFlag string, obs trust.Observation, target trust.State, reason string) error {
	payload, err := json.Marshal(map[string]string{"to": string(target), "reason": reason})
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	_, err = store.Append(context.Background(), event.Event{
		Kind:    event.KindTransition,
		Type:    "transition",
		Repo:    filepath.Base(abs),
		Actor:   resolveActor(actorFlag, dir),
		Source:  "human",
		Ref:     obs.ID,
		Payload: json.RawMessage(payload),
	})
	return err
}

// runStatus is the one-screen human surface: ledger size, state counts,
// and every live observation with its provenance.
func runStatus(args []string) error {
	fs := newFlagSet("status")
	dir := fs.String("dir", ".", "workspace directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	store := jsonl.Open(*dir)
	events := 0
	folder := trust.NewFolder()
	if err := store.Replay(context.Background(), "", func(e event.Event) error {
		events++
		folder.Add(e)
		return nil
	}); err != nil {
		return err
	}
	obs := folder.Observations()
	counts := map[trust.State]int{}
	for _, o := range obs {
		counts[o.State]++
	}
	fmt.Printf("Ledger: %d events · %d observations\n", events, len(obs))
	fmt.Printf("  generated %d · observed %d · verified %d · stale %d · deprecated %d\n",
		counts[trust.Generated], counts[trust.Observed], counts[trust.Verified], counts[trust.Stale], counts[trust.Deprecated])
	pending := 0
	for _, o := range obs {
		if o.State == trust.Generated || o.Conflict {
			pending++
		}
	}
	if pending > 0 {
		fmt.Printf("  %d awaiting judgment — `kervo review`\n", pending)
	}
	fmt.Println()
	for _, o := range obs {
		mark := ""
		if o.Conflict {
			mark = "  ⚠ conflict"
		}
		fmt.Printf("%s  %-10s %-9s %-22s %s%s\n",
			shortID(o.ID), o.State, o.Type, o.LastActor, truncateBody(o.Body, 60), mark)
	}
	return nil
}

func replayFolder(store *jsonl.Store) (*trust.Folder, error) {
	folder := trust.NewFolder()
	err := store.Replay(context.Background(), "", func(e event.Event) error {
		folder.Add(e)
		return nil
	})
	return folder, err
}

func findByPrefix(folder *trust.Folder, prefix string) (trust.Observation, error) {
	var hits []trust.Observation
	for _, o := range folder.Observations() {
		if strings.HasPrefix(o.ID, strings.ToUpper(prefix)) {
			hits = append(hits, o)
		}
	}
	switch len(hits) {
	case 0:
		return trust.Observation{}, fmt.Errorf("trust: no observation matches %q (see `kervo status`)", prefix)
	case 1:
		return hits[0], nil
	default:
		return trust.Observation{}, fmt.Errorf("trust: %q is ambiguous (%d matches) — use a longer prefix", prefix, len(hits))
	}
}

func shortID(id string) string {
	if len(id) > 10 {
		return id[:10]
	}
	return id
}

func truncateBody(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
