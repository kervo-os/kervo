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
	id, dup, err := captureObservation(*dir, *typ, *body, *evidence, *actor)
	if err != nil {
		return err
	}
	if dup != nil {
		fmt.Printf("duplicate of %s (%s) — skipped\n", shortID(dup.ID), dup.State)
		return nil
	}
	fmt.Printf("captured %s (%s) → .kervo/events/\n", id, *typ)
	return nil
}

// backpressureCap bounds how many UNJUDGED proposals one source may pile
// up. Producers push conclusions, not corpus — and when the queue for a
// source is full, the correct next step is human judgment, not more
// proposals. Human captures enter observed and are never throttled.
const backpressureCap = 12

// captureObservation is the shared write path for CLI and MCP. A non-nil
// dup means the body already exists live and nothing was written.
// Write-back guardrail: an identical live body is a duplicate, not new
// knowledge — agents re-reading the same session must not spam the review
// queue (a duplicate is a no-op, not a failure). Stale/deprecated bodies
// may be re-captured: re-assertion is a fresh claim for the human to judge.
func captureObservation(dir, typ, body, evidence, actorFlag string) (id string, dup *trust.Observation, err error) {
	if strings.TrimSpace(body) == "" {
		return "", nil, fmt.Errorf("capture: body is required")
	}
	store := jsonl.Open(dir)
	folder, err := replayFolder(store)
	if err != nil {
		return "", nil, err
	}
	actor := resolveActor(actorFlag, dir)
	pending := 0
	for _, o := range folder.Observations() {
		if o.Body == body && o.State != trust.Stale && o.State != trust.Deprecated {
			d := o
			return "", &d, nil
		}
		if o.State == trust.Generated && o.Actor == actor {
			pending++
		}
	}
	if pending >= backpressureCap {
		return "", nil, fmt.Errorf(
			"capture: backpressure — %s already has %d proposals awaiting judgment; a human must run `kervo review` (or judge in the dash) before this source proposes more",
			actor, pending)
	}
	fields := map[string]string{"body": body}
	if strings.TrimSpace(evidence) != "" {
		fields["evidence"] = evidence
	}
	payload, err := json.Marshal(fields)
	if err != nil {
		return "", nil, err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", nil, err
	}
	id, err = store.Append(context.Background(), event.Event{
		Kind:    event.KindObservation,
		Type:    typ,
		Repo:    filepath.Base(abs),
		Actor:   actor,
		Source:  "human",
		Payload: json.RawMessage(payload),
	})
	if err != nil {
		return "", nil, err
	}
	return id, nil, nil
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
