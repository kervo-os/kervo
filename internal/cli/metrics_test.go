package cli

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
)

// The ledger is committed to git: prompt content must NEVER be stored —
// only sizes. This is both privacy and the H3 measurement design.
func TestHookRecordsSizesNeverContent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(hookGuardEnv, "")
	writeFile(t, dir, "CLAUDE.md", "rules\n<!-- kervo:begin -->\nartifact\n<!-- kervo:end -->\n")
	secret := "SECRET-DO-NOT-PERSIST please explain the auth flow again"
	feedStdin(t, `{"hook_event_name":"UserPromptSubmit","session_id":"s1","prompt":"`+secret+`"}`)
	if err := runHook([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(dir + "/.kervo/events/" + monthFile(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "SECRET-DO-NOT-PERSIST") {
		t.Fatal("prompt content leaked into the committed ledger")
	}

	var metrics []promptMetric
	if err := jsonl.Open(dir).Replay(context.Background(), "", func(e event.Event) error {
		if e.Type == "metric:prompt" {
			var m promptMetric
			if err := json.Unmarshal(e.Payload, &m); err != nil {
				return err
			}
			metrics = append(metrics, m)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 1 {
		t.Fatalf("metric events = %d, want 1", len(metrics))
	}
	m := metrics[0]
	if m.PromptChars != len(secret) || m.PromptWords != 7 {
		t.Errorf("sizes wrong: %+v (want chars %d words 7)", m, len(secret))
	}
	if !m.ArtifactPresent || m.ArtifactBytes == 0 {
		t.Errorf("A/B variable not captured: %+v", m)
	}
	if m.Session != "s1" {
		t.Errorf("session = %q", m.Session)
	}
}

func monthFile(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir + "/.kervo/events")
	if err != nil {
		t.Fatal(err)
	}
	return entries[0].Name()
}

func TestAggregateMetricsABSides(t *testing.T) {
	mk := func(session string, chars int, present bool) event.Event {
		p, _ := json.Marshal(promptMetric{Session: session, PromptChars: chars, ArtifactPresent: present})
		return event.Event{Type: "metric:prompt", Kind: event.KindFact, Payload: p}
	}
	r := aggregateMetrics([]event.Event{
		mk("a", 1000, true), // session a: first prompt 1000, with artifact
		mk("a", 200, true),
		mk("b", 3000, false), // session b: first prompt 3000, without
		mk("b", 400, false),
		mk("b", 400, false),
	})
	with, without := r.sides()
	if with.sessions != 1 || without.sessions != 1 {
		t.Fatalf("sides = %+v / %+v", with, without)
	}
	if with.firstSum != 1000 || without.firstSum != 3000 {
		t.Errorf("first prompt sums: with=%d without=%d", with.firstSum, without.firstSum)
	}
	if with.promptSum != 2 || without.promptSum != 3 {
		t.Errorf("prompt counts: with=%d without=%d", with.promptSum, without.promptSum)
	}
	if with.charSum != 1200 || without.charSum != 3800 {
		t.Errorf("char sums: with=%d without=%d", with.charSum, without.charSum)
	}
}

// Non-prompt hooks keep only names/paths/sizes in the ledger.
func TestReducePayloadDropsBodies(t *testing.T) {
	out := reducePayload(map[string]any{
		"hook_event_name": "PostToolUse",
		"tool_name":       "Write",
		"tool_input": map[string]any{
			"file_path": "main.go",
			"content":   "package main // huge body",
			"command":   "export TOKEN=hunter2",
		},
	})
	if out["file_path"] != "main.go" {
		t.Errorf("path lost: %+v", out)
	}
	if out["content_chars"] != 25 || out["command_chars"] != 20 {
		t.Errorf("sizes wrong: %+v", out)
	}
	raw, _ := json.Marshal(out)
	for _, leak := range []string{"package main", "hunter2"} {
		if strings.Contains(string(raw), leak) {
			t.Errorf("body leaked: %s", raw)
		}
	}
}
