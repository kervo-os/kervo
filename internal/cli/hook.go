package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/artifact"
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
	repo := filepath.Base(abs)
	store := jsonl.Open(*dir)

	// The ledger is committed to git: payloads carry SIZES AND NAMES, never
	// content (prompts and file bodies would leak into history otherwise).
	reduced, err := json.Marshal(reducePayload(payload))
	if err != nil {
		reduced = []byte("{}")
	}
	if _, err := store.Append(context.Background(), event.Event{
		Kind:    event.KindFact,
		Type:    "hook:" + hookName,
		Repo:    repo,
		Actor:   "agent:claude-code",
		Source:  "consumer:claude-code",
		Ref:     ref,
		Payload: json.RawMessage(reduced),
	}); err != nil {
		// Failing to persist is worth a warning, but never a broken session.
		fmt.Fprintf(os.Stderr, "kervo hook: append failed: %v\n", err)
	}

	// H3 counters ride the capture path (PRD §12: cannot be added after
	// the fact). Each prompt yields one deterministic metric event carrying
	// the A/B variable: was the compiled artifact present when typed?
	if hookName == "UserPromptSubmit" {
		if prompt := firstString(payload, "prompt", "user_prompt", "userPrompt"); prompt != "" {
			present, artifactBytes := artifactInjected(*dir)
			m, err := json.Marshal(map[string]any{
				"session":          firstString(payload, "session_id", "sessionId"),
				"prompt_chars":     len(prompt),
				"prompt_words":     len(strings.Fields(prompt)),
				"artifact_present": present,
				"artifact_bytes":   artifactBytes,
				"artifact_known":   true, // live observation — enters the A/B sides
			})
			if err == nil {
				if _, err := store.Append(context.Background(), event.Event{
					Kind: event.KindFact, Type: "metric:prompt", Repo: repo,
					Actor: "system", Source: "consumer:claude-code",
					Ref:     firstString(payload, "session_id", "sessionId"),
					Payload: json.RawMessage(m),
				}); err != nil {
					fmt.Fprintf(os.Stderr, "kervo hook: metric append failed: %v\n", err)
				}
			}
		}
	}
	return nil
}

// reducePayload keeps only measurement-grade fields: names, paths, sizes.
func reducePayload(payload map[string]any) map[string]any {
	out := map[string]any{}
	for _, k := range []string{"hook_event_name", "hookEventName", "tool_name", "toolName", "session_id", "sessionId"} {
		if v, ok := payload[k].(string); ok && v != "" {
			out[k] = v
		}
	}
	if prompt := firstString(payload, "prompt", "user_prompt", "userPrompt"); prompt != "" {
		out["prompt_chars"] = len(prompt)
	}
	if ti, ok := payload["tool_input"].(map[string]any); ok {
		if fp, ok := ti["file_path"].(string); ok && fp != "" {
			out["file_path"] = fp // paths are already public in git
		}
		if c, ok := ti["content"].(string); ok {
			out["content_chars"] = len(c)
		}
		if c, ok := ti["command"].(string); ok {
			out["command_chars"] = len(c) // commands may embed secrets — size only
		}
	}
	return out
}

// artifactInjected reports whether CLAUDE.md carried the kervo block at
// this moment, and how large the whole file was — the H3 A/B variable.
func artifactInjected(dir string) (bool, int) {
	raw, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		return false, 0
	}
	return strings.Contains(string(raw), artifact.MarkerBegin), len(raw)
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
