package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
)

// Hook rules (RFC-0003, hardened by the agentmemory post-mortem):
// local append ONLY — no LLM, no network, no agent spawning (their issue
// #149 was an infinite Stop-hook recursion). Must stay in the ms budget.
// A hook must also never break the consumer's session: payload problems
// are reported on stderr and swallowed with exit 0.
const hookGuardEnv = "KERVO_IN_HOOK"

// runHook ingests one consumer lifecycle event from stdin JSON and appends
// it to the ledger as a fact.
func runHook(args []string) error {
	fs := newFlagSet("hook")
	dir := fs.String("dir", ".", "workspace directory")
	if err := fs.Parse(args); err != nil {
		return err
	}

	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 256<<10))
	if err != nil || len(raw) == 0 {
		fmt.Fprintln(os.Stderr, "kervo hook: empty or unreadable stdin — ignored")
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		fmt.Fprintln(os.Stderr, "kervo hook: stdin is not JSON — ignored")
		return nil
	}

	if hookIsRecursive(payload) {
		return nil // silently drop our own echo — never loop
	}
	os.Setenv(hookGuardEnv, "1")

	// Field names drift across consumer versions (observed in the wild:
	// tool_name vs toolName) — accept known variants from day one.
	hookName := firstString(payload, "hook_event_name", "hookEventName", "event", "hook")
	if hookName == "" {
		hookName = "unknown"
	}
	ref := firstString(payload, "tool_name", "toolName", "session_id", "sessionId")

	abs, err := filepath.Abs(*dir)
	if err != nil {
		return err
	}
	if _, err := jsonl.Open(*dir).Append(context.Background(), event.Event{
		Kind:    event.KindFact,
		Type:    "hook:" + hookName,
		Repo:    filepath.Base(abs),
		Actor:   "agent:claude-code",
		Source:  "consumer:claude-code",
		Ref:     ref,
		Payload: json.RawMessage(raw),
	}); err != nil {
		// Failing to persist is worth a warning, but never a broken session.
		fmt.Fprintf(os.Stderr, "kervo hook: append failed: %v\n", err)
	}
	return nil
}

// hookIsRecursive is the single shared guard (two signals, per the
// agentmemory lesson: they built one and then inlined copies everywhere).
// Signal 1: we are already inside a kervo-initiated hook (env marker).
// Signal 2: the payload is about kervo invoking itself.
func hookIsRecursive(payload map[string]any) bool {
	if os.Getenv(hookGuardEnv) != "" {
		return true
	}
	if cmd, ok := payload["command"].(string); ok && len(cmd) >= 5 && cmd[:5] == "kervo" {
		return true
	}
	if tool, ok := payload["tool_input"].(map[string]any); ok {
		if cmd, ok := tool["command"].(string); ok && len(cmd) >= 5 && cmd[:5] == "kervo" {
			return true
		}
	}
	return false
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}
