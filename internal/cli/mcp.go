package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runMCP: stdio MCP server — the conversation is the frontend. Facts out
// (read_context), observations in (kervo_capture, Generated only), and the
// judging surface as chat (review_queue / review_judge): the human speaks a
// judgment, the agent relays it. review_judge records the workspace's git
// identity, exactly like the CLI — the trust boundary is process access,
// same as `kervo trust`. Zero dependencies: newline-delimited JSON-RPC 2.0.
func runMCP(args []string) error {
	fs := newFlagSet("mcp")
	dir := fs.String("dir", ".", "workspace directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return serveMCP(*dir, os.Stdin, os.Stdout)
}

type rpcRequest struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func serveMCP(dir string, in io.Reader, out io.Writer) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 64<<10), 4<<20)
	enc := json.NewEncoder(out)
	reply := func(id json.RawMessage, result any, rerr *rpcError) {
		if len(id) == 0 {
			return // notification — never answered
		}
		msg := map[string]any{"jsonrpc": "2.0", "id": json.RawMessage(id)}
		if rerr != nil {
			msg["error"] = rerr
		} else {
			msg["result"] = result
		}
		_ = enc.Encode(msg)
	}

	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue // unparseable frame — nothing to address a reply to
		}
		switch req.Method {
		case "initialize":
			var p struct {
				ProtocolVersion string `json:"protocolVersion"`
			}
			_ = json.Unmarshal(req.Params, &p)
			if p.ProtocolVersion == "" {
				p.ProtocolVersion = "2025-03-26"
			}
			reply(req.ID, map[string]any{
				"protocolVersion": p.ProtocolVersion,
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": "kervo", "version": resolveVersion()},
			}, nil)
		case "ping":
			reply(req.ID, map[string]any{}, nil)
		case "tools/list":
			reply(req.ID, map[string]any{"tools": mcpToolDefs()}, nil)
		case "tools/call":
			var p struct {
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				reply(req.ID, nil, &rpcError{Code: -32602, Message: "invalid params"})
				continue
			}
			text, callErr := mcpCall(dir, p.Name, p.Arguments)
			if callErr != nil {
				// Tool-level failures are results, not protocol errors —
				// the model should read them and adjust.
				reply(req.ID, toolResult(callErr.Error(), true), nil)
				continue
			}
			reply(req.ID, toolResult(text, false), nil)
		default:
			reply(req.ID, nil, &rpcError{Code: -32601, Message: "method not found: " + req.Method})
		}
	}
	return sc.Err()
}

func toolResult(text string, isErr bool) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": isErr,
	}
}

func mcpToolDefs() []map[string]any {
	obj := func(props map[string]any, required ...string) map[string]any {
		s := map[string]any{"type": "object", "properties": props}
		if len(required) > 0 {
			s["required"] = required
		}
		return s
	}
	str := func(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }
	return []map[string]any{
		{
			"name":        "read_context",
			"description": "Read the compiled context artifact for this workspace (facts + trust-labeled observations). Use when the workspace context was not auto-loaded.",
			"inputSchema": obj(map[string]any{}),
		},
		{
			"name":        "kervo_capture",
			"description": "Write-back: stage one durable fact this workspace's artifact lacks as a proposal for human judgment. Attach evidence (the command you ran, the doc you read). Duplicates are dropped automatically.",
			"inputSchema": obj(map[string]any{
				"type":     str("decision | risk | summary | goal | note"),
				"body":     str("the fact, one per call"),
				"evidence": str("how you verified it (optional but expected)"),
				"actor":    str(`who proposes, e.g. "agent:claude-code"`),
			}, "type", "body", "actor"),
		},
		{
			"name":        "review_queue",
			"description": "List every observation awaiting human judgment (generated proposals and conflicts). Render it for the human; do not judge on your own.",
			"inputSchema": obj(map[string]any{}),
		},
		{
			"name":        "review_judge",
			"description": "Record a judgment the human explicitly stated in conversation (verified | stale | deprecated). Only call this to relay the human's decision, never your own.",
			"inputSchema": obj(map[string]any{
				"id":     str("observation ID or unique prefix"),
				"to":     str("verified | stale | deprecated"),
				"reason": str("the human's reason (optional)"),
			}, "id", "to"),
		},
	}
}

func mcpCall(dir, name string, args json.RawMessage) (string, error) {
	switch name {
	case "read_context":
		raw, err := os.ReadFile(filepath.Join(dir, ".kervo", "artifact.md"))
		if err != nil {
			return "", fmt.Errorf("no compiled artifact — run `kervo compile` first (%v)", err)
		}
		return string(raw), nil

	case "kervo_capture":
		var a struct{ Type, Body, Evidence, Actor string }
		if err := json.Unmarshal(args, &a); err != nil {
			return "", fmt.Errorf("bad arguments: %v", err)
		}
		if a.Type == "" {
			a.Type = "note"
		}
		id, dup, err := captureObservation(dir, a.Type, a.Body, a.Evidence, a.Actor)
		if err != nil {
			return "", err
		}
		if dup != nil {
			return fmt.Sprintf("duplicate of %s (%s) — skipped", shortID(dup.ID), dup.State), nil
		}
		return fmt.Sprintf("captured %s (%s) — awaiting human judgment in `kervo review`", id, a.Type), nil

	case "review_queue":
		folder, err := replayFolder(jsonl.Open(dir))
		if err != nil {
			return "", err
		}
		type item struct {
			ID, Type, State, Actor, Body string
			Evidence                     string `json:",omitempty"`
			Conflict                     bool   `json:",omitempty"`
		}
		var queue []item
		for _, o := range folder.Observations() {
			if o.State == trust.Generated || o.Conflict {
				queue = append(queue, item{
					ID: o.ID, Type: o.Type, State: string(o.State),
					Actor: o.Actor, Body: o.Body, Evidence: o.Evidence, Conflict: o.Conflict,
				})
			}
		}
		if len(queue) == 0 {
			return "review queue is empty — nothing awaits judgment", nil
		}
		out, err := json.MarshalIndent(queue, "", "  ")
		if err != nil {
			return "", err
		}
		return string(out), nil

	case "review_judge":
		var a struct{ ID, To, Reason string }
		if err := json.Unmarshal(args, &a); err != nil {
			return "", fmt.Errorf("bad arguments: %v", err)
		}
		store := jsonl.Open(dir)
		folder, err := replayFolder(store)
		if err != nil {
			return "", err
		}
		obs, err := findByPrefix(folder, a.ID)
		if err != nil {
			return "", err
		}
		target := trust.State(a.To)
		if !trust.CanTransition(obs.State, target) {
			return "", fmt.Errorf("%s → %s is not a legal transition", obs.State, target)
		}
		if err := appendTransition(store, dir, "", obs, target, a.Reason); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s: %s → %s — run `kervo compile` to refresh the artifact", shortID(obs.ID), obs.State, target), nil

	default:
		return "", fmt.Errorf("unknown tool %q", name)
	}
}
