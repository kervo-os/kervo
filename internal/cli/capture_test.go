package cli

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
)

func TestCaptureAppendsObservation(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	if err := runCapture([]string{"-dir", dir, "-type", "decision", "-body", "JWT over sessions"}); err != nil {
		t.Fatal(err)
	}
	var got []event.Event
	if err := jsonl.Open(dir).Replay(context.Background(), "", func(e event.Event) error {
		got = append(got, e)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("events = %d, want 1", len(got))
	}
	e := got[0]
	if e.Kind != event.KindObservation || e.Type != "decision" {
		t.Errorf("kind/type = %s/%s", e.Kind, e.Type)
	}
	if !strings.HasPrefix(e.Actor, "human:") {
		t.Errorf("actor = %q, want human:<name>", e.Actor)
	}
	if !strings.Contains(string(e.Payload), "JWT over sessions") {
		t.Errorf("payload = %s", e.Payload)
	}
}

func TestCaptureRequiresBody(t *testing.T) {
	if err := runCapture([]string{"-dir", t.TempDir(), "-type", "note"}); err == nil {
		t.Fatal("expected error without -body")
	}
}

func feedStdin(t *testing.T, content string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(content); err != nil {
		t.Fatal(err)
	}
	w.Close()
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })
}

func replayAll(t *testing.T, dir string) []event.Event {
	t.Helper()
	var got []event.Event
	if err := jsonl.Open(dir).Replay(context.Background(), "", func(e event.Event) error {
		got = append(got, e)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestHookIngestsWithFieldNameVariants(t *testing.T) {
	for _, payload := range []string{
		`{"hook_event_name":"PostToolUse","tool_name":"Edit","session_id":"s1"}`,
		`{"hookEventName":"PostToolUse","toolName":"Edit","sessionId":"s1"}`,
	} {
		dir := t.TempDir()
		t.Setenv(hookGuardEnv, "")
		feedStdin(t, payload)
		if err := runHook([]string{"-dir", dir}); err != nil {
			t.Fatal(err)
		}
		got := replayAll(t, dir)
		if len(got) != 1 || got[0].Type != "hook:PostToolUse" || got[0].Ref != "Edit" {
			t.Errorf("payload %s → %+v", payload, got)
		}
	}
}

func TestHookRecursionGuard(t *testing.T) {
	// Signal 2: the payload is kervo invoking itself.
	dir := t.TempDir()
	t.Setenv(hookGuardEnv, "")
	feedStdin(t, `{"hook_event_name":"PostToolUse","tool_name":"Bash","tool_input":{"command":"kervo compile"}}`)
	if err := runHook([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	if got := replayAll(t, dir); len(got) != 0 {
		t.Errorf("kervo's own invocation captured — recursion loop: %+v", got)
	}

	// Signal 1: already inside a kervo hook (env marker).
	dir2 := t.TempDir()
	t.Setenv(hookGuardEnv, "1")
	feedStdin(t, `{"hook_event_name":"PostToolUse","tool_name":"Edit"}`)
	if err := runHook([]string{"-dir", dir2}); err != nil {
		t.Fatal(err)
	}
	if got := replayAll(t, dir2); len(got) != 0 {
		t.Errorf("env-guarded hook still captured: %+v", got)
	}
}

// A hook must never break the consumer's session — garbage in, exit 0 out.
func TestHookSwallowsGarbage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(hookGuardEnv, "")
	feedStdin(t, "not json at all")
	if err := runHook([]string{"-dir", dir}); err != nil {
		t.Fatalf("hook must swallow garbage, got: %v", err)
	}
	if got := replayAll(t, dir); len(got) != 0 {
		t.Errorf("garbage captured: %+v", got)
	}
}

// Write-back guardrail: identical live bodies are duplicates — agents must
// not be able to spam the review queue; re-capture is allowed once the
// original was judged away (deprecated = fresh claim to re-judge).
func TestCaptureDedupsLiveBodies(t *testing.T) {
	dir := t.TempDir()
	if err := runCapture([]string{"-dir", dir, "-type", "decision", "-body", "same fact", "-actor", "agent:test"}); err != nil {
		t.Fatal(err)
	}
	if err := runCapture([]string{"-dir", dir, "-type", "decision", "-body", "same fact", "-actor", "agent:test"}); err != nil {
		t.Fatalf("duplicate capture must be a no-op, got: %v", err)
	}
	folder, err := replayFolder(jsonl.Open(dir))
	if err != nil {
		t.Fatal(err)
	}
	obs := folder.Observations()
	if len(obs) != 1 {
		t.Fatalf("observations = %d, want 1 (duplicate dropped)", len(obs))
	}

	if err := runTrust([]string{"-dir", dir, "-id", obs[0].ID[:10], "-to", "deprecated", "-reason", "judged wrong", "-actor", "human:test"}); err != nil {
		t.Fatal(err)
	}
	if err := runCapture([]string{"-dir", dir, "-type", "decision", "-body", "same fact", "-actor", "agent:test"}); err != nil {
		t.Fatal(err)
	}
	folder, err = replayFolder(jsonl.Open(dir))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(folder.Observations()); got != 2 {
		t.Fatalf("observations = %d, want 2 (re-assertion after deprecation allowed)", got)
	}
}
