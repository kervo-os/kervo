package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runCapture appends an Observation to the workspace ledger — the manual
// producer that exists regardless of the H2' Build-vs-Interop outcome.
// Everything enters as an event; trust states are derived by replay.
func runCapture(args []string) error {
	fs := newFlagSet("capture")
	dir := fs.String("dir", ".", "workspace directory")
	typ := fs.String("type", "note", "observation type: decision, note, risk, correction, goal")
	body := fs.String("body", "", "the observation text (required)")
	evidence := fs.String("evidence", "", "how you verified it — command run, doc read (optional)")
	actor := fs.String("actor", "", `who is recording (default "human:<git user.name>")`)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*body) == "" {
		return fmt.Errorf("capture: -body is required")
	}
	store := jsonl.Open(*dir)
	// Write-back guardrail: an identical live body is a duplicate, not new
	// knowledge — agents re-reading the same session must not spam the
	// review queue. Exit 0: a duplicate is a no-op, not an agent failure.
	// Stale/deprecated bodies may be re-captured (re-assertion is a fresh
	// claim for the human to judge).
	folder, err := replayFolder(store)
	if err != nil {
		return err
	}
	for _, o := range folder.Observations() {
		if o.Body == *body && o.State != trust.Stale && o.State != trust.Deprecated {
			fmt.Printf("duplicate of %s (%s) — skipped\n", shortID(o.ID), o.State)
			return nil
		}
	}
	fields := map[string]string{"body": *body}
	if strings.TrimSpace(*evidence) != "" {
		fields["evidence"] = *evidence
	}
	payload, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(*dir)
	if err != nil {
		return err
	}
	id, err := store.Append(context.Background(), event.Event{
		Kind:    event.KindObservation,
		Type:    *typ,
		Repo:    filepath.Base(abs),
		Actor:   resolveActor(*actor, *dir),
		Source:  "human",
		Payload: json.RawMessage(payload),
	})
	if err != nil {
		return err
	}
	fmt.Printf("captured %s (%s) → .kervo/events/\n", id, *typ)
	return nil
}

// resolveActor defaults to the git identity — the RFC-0005 vocabulary:
// "human:<name>", "agent:<consumer>", "system".
func resolveActor(flagVal, dir string) string {
	if flagVal != "" {
		return flagVal
	}
	out, err := exec.Command("git", "-C", dir, "config", "user.name").Output()
	if name := strings.TrimSpace(string(out)); err == nil && name != "" {
		return "human:" + name
	}
	return "human:unknown"
}
