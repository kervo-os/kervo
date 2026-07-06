package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/trust"
)

func mcpRoundTrip(t *testing.T, dir string, lines ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	if err := serveMCP(dir, strings.NewReader(strings.Join(lines, "\n")+"\n"), &out); err != nil {
		t.Fatal(err)
	}
	var replies []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("bad reply line %q: %v", l, err)
		}
		replies = append(replies, m)
	}
	return replies
}

func toolText(t *testing.T, reply map[string]any) (string, bool) {
	t.Helper()
	res, ok := reply["result"].(map[string]any)
	if !ok {
		t.Fatalf("no result in %v", reply)
	}
	content := res["content"].([]any)[0].(map[string]any)
	return content["text"].(string), res["isError"].(bool)
}

func TestMCPLifecycleAndTools(t *testing.T) {
	dir := t.TempDir()
	replies := mcpRoundTrip(t, dir,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"kervo_capture","arguments":{"type":"decision","body":"queue works","evidence":"ran the tests","actor":"agent:test"}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"review_queue","arguments":{}}}`,
	)
	if len(replies) != 4 {
		t.Fatalf("replies = %d, want 4 (notification must not be answered)", len(replies))
	}
	init := replies[0]["result"].(map[string]any)
	if init["protocolVersion"] != "2025-06-18" {
		t.Errorf("protocolVersion not echoed: %v", init)
	}
	// notifications/initialized got no reply, so replies[1] is tools/list.
	tools := replies[1]["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 4 {
		t.Errorf("tools = %d, want 4", len(tools))
	}
	capText, isErr := toolText(t, replies[2])
	if isErr || !strings.Contains(capText, "captured") {
		t.Errorf("capture failed: %q", capText)
	}
	queueText, isErr := toolText(t, replies[3])
	if isErr || !strings.Contains(queueText, "queue works") || !strings.Contains(queueText, "ran the tests") {
		t.Errorf("queue missing capture or evidence: %q", queueText)
	}
}

func TestMCPJudgeRelaysHumanDecision(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := captureObservation(dir, "decision", "judge me", "", "agent:test"); err != nil {
		t.Fatal(err)
	}
	folder, err := replayFolder(jsonl.Open(dir))
	if err != nil {
		t.Fatal(err)
	}
	id := folder.Observations()[0].ID

	replies := mcpRoundTrip(t, dir,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"review_judge","arguments":{"id":"`+id[:10]+`","to":"verified","reason":"human said yes"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"review_judge","arguments":{"id":"`+id[:10]+`","to":"observed"}}}`,
	)
	text, isErr := toolText(t, replies[0])
	if isErr || !strings.Contains(text, "→ verified") {
		t.Errorf("judge failed: %q", text)
	}
	// verified → observed is illegal — must come back as a tool error, not
	// a silent success and not a protocol failure.
	text, isErr = toolText(t, replies[1])
	if !isErr || !strings.Contains(text, "not a legal transition") {
		t.Errorf("illegal transition not surfaced: %q", text)
	}

	folder, err = replayFolder(jsonl.Open(dir))
	if err != nil {
		t.Fatal(err)
	}
	if o := folder.Observations()[0]; o.State != trust.Verified || o.Reason != "human said yes" {
		t.Errorf("state = %s reason = %q, want verified / human said yes", o.State, o.Reason)
	}
}

func TestMCPUnknownMethodAndTool(t *testing.T) {
	dir := t.TempDir()
	replies := mcpRoundTrip(t, dir,
		`{"jsonrpc":"2.0","id":1,"method":"resources/list"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"nope","arguments":{}}}`,
	)
	if _, ok := replies[0]["error"]; !ok {
		t.Error("unknown method must return a JSON-RPC error")
	}
	if text, isErr := toolText(t, replies[1]); !isErr || !strings.Contains(text, "unknown tool") {
		t.Errorf("unknown tool not surfaced: %q", text)
	}
}
