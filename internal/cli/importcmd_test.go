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

func fakeTranscript(t *testing.T, dir, session string) {
	t.Helper()
	lines := []string{
		`{"type":"user","timestamp":"2026-07-03T10:00:00Z","message":{"role":"user","content":"CONFIDENTIAL-PROMPT please build the scanner"}}`,
		`{"type":"assistant","timestamp":"2026-07-03T10:01:00Z","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write","input":{"file_path":"scan.go","content":"SECRET-BODY"}},{"type":"tool_use","name":"Read","input":{"file_path":"x.go"}}]}}`,
		`{"type":"user","timestamp":"2026-07-03T10:02:00Z","message":{"role":"user","content":[{"type":"tool_result","content":"ignored"}]}}`,
		`{"type":"user","timestamp":"2026-07-03T10:03:00Z","message":{"role":"user","content":"and add tests"}}`,
		`{"type":"file-history-snapshot","noise":true}`,
	}
	writeFile(t, dir, session+".jsonl", strings.Join(lines, "\n")+"\n")
}

func TestImportClaudeTranscripts(t *testing.T) {
	work := t.TempDir()
	tdir := t.TempDir()
	fakeTranscript(t, tdir, "sess-1")

	if err := runImport([]string{"claude", "-dir", work, "-from", tdir}); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(work + "/.kervo/events/2026-07.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	for _, leak := range []string{"CONFIDENTIAL-PROMPT", "SECRET-BODY"} {
		if strings.Contains(string(raw), leak) {
			t.Fatalf("content leaked into ledger: %s", leak)
		}
	}

	var metrics []promptMetric
	var summaries int
	if err := jsonl.Open(work).Replay(context.Background(), "", func(e event.Event) error {
		switch e.Type {
		case "metric:prompt":
			var m promptMetric
			if err := json.Unmarshal(e.Payload, &m); err != nil {
				return err
			}
			metrics = append(metrics, m)
		case "import:session":
			summaries++
			if !strings.Contains(string(e.Payload), `"fileops":1`) ||
				!strings.Contains(string(e.Payload), "scan.go") {
				t.Errorf("session summary wrong: %s", e.Payload)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 2 || summaries != 1 {
		t.Fatalf("metrics=%d summaries=%d, want 2/1 (tool_result and noise lines must not count)", len(metrics), summaries)
	}
	if metrics[0].ArtifactKnown {
		t.Error("retroactive import must be artifact-unknown")
	}
	if metrics[0].PromptChars != len("CONFIDENTIAL-PROMPT please build the scanner") {
		t.Errorf("first prompt chars = %d", metrics[0].PromptChars)
	}

	// Idempotency: re-import must not duplicate the session.
	if err := runImport([]string{"claude", "-dir", work, "-from", tdir}); err != nil {
		t.Fatal(err)
	}
	count := 0
	_ = jsonl.Open(work).Replay(context.Background(), "", func(e event.Event) error {
		if e.Type == "metric:prompt" {
			count++
		}
		return nil
	})
	if count != 2 {
		t.Fatalf("re-import duplicated metrics: %d", count)
	}
}

func TestImportRejectsUnknownSource(t *testing.T) {
	if err := runImport([]string{"agentmemory"}); err == nil {
		t.Fatal("unknown source must error with supported list")
	}
}

func TestSanitizeProjectPath(t *testing.T) {
	got := sanitizeProjectPath("/Users/a_b/30_lab/c.d")
	if got != "-Users-a-b-30-lab-c-d" {
		t.Errorf("sanitize = %q", got)
	}
}
