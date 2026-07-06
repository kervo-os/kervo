package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/trust"
)

func webServerFor(t *testing.T, dir string) (*reviewServer, *httptest.Server) {
	t.Helper()
	store := jsonl.Open(dir)
	folder, err := replayFolder(store)
	if err != nil {
		t.Fatal(err)
	}
	var queue []trust.Observation
	for _, o := range folder.Observations() {
		if o.State == trust.Generated || o.Conflict {
			queue = append(queue, o)
		}
	}
	srv := newReviewServer(store, dir, "human:tester", queue)
	ts := httptest.NewServer(srv.handler())
	t.Cleanup(ts.Close)
	return srv, ts
}

func TestReviewWebPageAndJudge(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := captureObservation(dir, "decision", "web queue item <b>escaped</b>", "ran it locally", "agent:test"); err != nil {
		t.Fatal(err)
	}
	srv, ts := webServerFor(t, dir)

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(res.Body)
	res.Body.Close()
	for _, want := range []string{"web queue item", "evidence: ran it locally", "&lt;b&gt;escaped&lt;/b&gt;"} {
		if !strings.Contains(string(page), want) {
			t.Errorf("page missing %q", want)
		}
	}
	if strings.Contains(string(page), "<b>escaped</b>") {
		t.Error("body not HTML-escaped — data can inject markup")
	}

	folder, _ := replayFolder(jsonl.Open(dir))
	id := folder.Observations()[0].ID
	res, err = http.Post(ts.URL+"/judge", "application/json",
		strings.NewReader(`{"ID":"`+id+`","To":"verified","Reason":"looks right"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("judge status = %d", res.StatusCode)
	}
	folder, _ = replayFolder(jsonl.Open(dir))
	if o := folder.Observations()[0]; o.State != trust.Verified || o.Reason != "looks right" {
		t.Errorf("state = %s reason = %q", o.State, o.Reason)
	}

	// Double-judging the same item must 409, and judging the last item
	// closes the done channel so the command can exit.
	res, _ = http.Post(ts.URL+"/judge", "application/json",
		strings.NewReader(`{"ID":"`+id+`","To":"stale"}`))
	res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Errorf("double judge status = %d, want 409", res.StatusCode)
	}
	select {
	case <-srv.done:
	default:
		t.Error("done channel not closed after last judgment")
	}
}

func TestReviewWebQuit(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := captureObservation(dir, "risk", "left pending", "", "agent:test"); err != nil {
		t.Fatal(err)
	}
	srv, ts := webServerFor(t, dir)
	res, err := http.Post(ts.URL+"/quit", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	select {
	case <-srv.done:
	default:
		t.Error("quit did not close the done channel")
	}
	if srv.judgedCount() != 0 {
		t.Error("quit must not judge anything")
	}
}
