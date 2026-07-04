package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
)

func artifactOf(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// The product's core loop, end to end: capture → labeled context →
// human verify → preferred context → age → demoted with a reason.
func TestTrustLoopEndToEnd(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "main.go", "package main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "initial")

	// 1. capture a decision → appears labeled [observed] in the artifact
	if err := runCapture([]string{"-dir", dir, "-type", "decision", "-body", "JWT over sessions"}); err != nil {
		t.Fatal(err)
	}
	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	art := artifactOf(t, dir)
	if !strings.Contains(art, "]**\nJWT over sessions") || !strings.Contains(art, "**[observed — human:") {
		t.Fatalf("captured decision not rendered as observed:\n%s", art)
	}

	// 2. human verifies it → label upgrades
	var obsID string
	if err := jsonl.Open(dir).Replay(context.Background(), "", func(e event.Event) error {
		if e.Kind == event.KindObservation {
			obsID = e.ID
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := runTrust([]string{"-dir", dir, "-id", obsID[:8], "-to", "verified", "-reason", "team agreed"}); err != nil {
		t.Fatal(err)
	}
	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(artifactOf(t, dir), "**[verified — human:") {
		t.Fatal("verified state not rendered")
	}

	// 3. illegal transition is refused by the state machine
	if err := runTrust([]string{"-dir", dir, "-id", obsID[:8], "-to", "observed"}); err == nil {
		t.Fatal("verified → observed must be rejected")
	}

	// 4. an old unreaffirmed observation is demoted to stale with a reason
	old, _ := json.Marshal(map[string]string{"body": "use the legacy queue"})
	if _, err := jsonl.Open(dir).Append(context.Background(), event.Event{
		Kind: event.KindObservation, Type: "decision", Repo: "x",
		Actor: "human:kim", Source: "human",
		At:      time.Now().UTC().Add(-100 * 24 * time.Hour),
		Payload: json.RawMessage(old),
	}); err != nil {
		t.Fatal(err)
	}
	if err := runCompile([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	art = artifactOf(t, dir)
	if !strings.Contains(art, "**[stale — age 100d > 30d without reaffirmation]** use the legacy queue") {
		t.Fatalf("stale demotion not listed with reason:\n%s", art)
	}
	if strings.Contains(art, "]**\nuse the legacy queue") {
		t.Fatal("stale observation still rendered in its slot")
	}
	// the verified decision must have survived the sweep
	if !strings.Contains(art, "**[verified — human:") {
		t.Fatal("verified observation was swept — only humans may demote verified")
	}
}
