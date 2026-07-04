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
)

// runCapture appends an Observation to the workspace ledger — the manual
// producer that exists regardless of the H2' Build-vs-Interop outcome.
// Everything enters as an event; trust states are derived by replay.
func runCapture(args []string) error {
	fs := newFlagSet("capture")
	dir := fs.String("dir", ".", "workspace directory")
	typ := fs.String("type", "note", "observation type: decision, note, risk, correction, goal")
	body := fs.String("body", "", "the observation text (required)")
	actor := fs.String("actor", "", `who is recording (default "human:<git user.name>")`)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*body) == "" {
		return fmt.Errorf("capture: -body is required")
	}
	payload, err := json.Marshal(map[string]string{"body": *body})
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(*dir)
	if err != nil {
		return err
	}
	id, err := jsonl.Open(*dir).Append(context.Background(), event.Event{
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
