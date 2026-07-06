package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/trust"
)

func captureFor(t *testing.T, dir, typ, body string) {
	t.Helper()
	if err := runCapture([]string{"-dir", dir, "-type", typ, "-body", body, "-actor", "agent:test"}); err != nil {
		t.Fatal(err)
	}
	// ULIDs sharing a millisecond sort by random entropy — space captures
	// out so the queue order matches capture order deterministically.
	time.Sleep(2 * time.Millisecond)
}

func reviewStates(t *testing.T, dir string) map[string]trust.State {
	t.Helper()
	folder, err := replayFolder(jsonl.Open(dir))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]trust.State{}
	for _, o := range folder.Observations() {
		out[o.Body] = o.State
	}
	return out
}

func TestReviewJudgesQueue(t *testing.T) {
	dir := t.TempDir()
	captureFor(t, dir, "decision", "first proposal")
	captureFor(t, dir, "risk", "second proposal")
	captureFor(t, dir, "summary", "third proposal")

	// verify #1 with a reason, skip #2, deprecate #3 without a reason.
	in := strings.NewReader("v\nteam agreed\n\nd\n\n")
	var out bytes.Buffer
	if err := reviewLoop(dir, "human:tester", in, &out); err != nil {
		t.Fatal(err)
	}

	states := reviewStates(t, dir)
	if states["first proposal"] != trust.Verified {
		t.Errorf("first = %s, want verified", states["first proposal"])
	}
	if states["second proposal"] != trust.Generated {
		t.Errorf("second = %s, want generated (skipped)", states["second proposal"])
	}
	if states["third proposal"] != trust.Deprecated {
		t.Errorf("third = %s, want deprecated", states["third proposal"])
	}
	for _, want := range []string{"3 awaiting judgment", "[1/3]", "first proposal", "→ verified", "Judged 2 of 3"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q in:\n%s", want, out.String())
		}
	}
}

func TestReviewQuitStopsEarly(t *testing.T) {
	dir := t.TempDir()
	captureFor(t, dir, "decision", "first proposal")
	captureFor(t, dir, "decision", "second proposal")

	var out bytes.Buffer
	if err := reviewLoop(dir, "human:tester", strings.NewReader("q\n"), &out); err != nil {
		t.Fatal(err)
	}
	states := reviewStates(t, dir)
	if states["first proposal"] != trust.Generated || states["second proposal"] != trust.Generated {
		t.Error("quit must not judge anything")
	}
	if !strings.Contains(out.String(), "Judged 0 of 2") {
		t.Errorf("missing summary in:\n%s", out.String())
	}
}

// A piped or CI stdin (immediate EOF) must end cleanly, never hang or error.
func TestReviewEOFSafe(t *testing.T) {
	dir := t.TempDir()
	captureFor(t, dir, "decision", "pending proposal")

	var out bytes.Buffer
	if err := reviewLoop(dir, "human:tester", strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Judged 0 of 1") {
		t.Errorf("missing summary in:\n%s", out.String())
	}
}

func TestReviewEmptyQueue(t *testing.T) {
	var out bytes.Buffer
	if err := reviewLoop(t.TempDir(), "human:tester", strings.NewReader(""), &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Review queue is empty") {
		t.Errorf("missing empty-queue message in:\n%s", out.String())
	}
}
