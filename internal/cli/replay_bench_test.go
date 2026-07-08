package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
)

// writeBenchLedger writes n synthetic events as one month file. IDs are
// zero-padded 26-char strings, so they sort exactly like ULIDs; every
// fifth event is a verify transition on the observation before it, so the
// fold path is exercised, not just the scan.
func writeBenchLedger(tb testing.TB, dir string, n int) {
	tb.Helper()
	evDir := filepath.Join(dir, ".kervo", "events")
	if err := os.MkdirAll(evDir, 0o755); err != nil {
		tb.Fatal(err)
	}
	f, err := os.Create(filepath.Join(evDir, "2026-01.jsonl"))
	if err != nil {
		tb.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("01BENCH%019d", i)
		var e event.Event
		if i%5 == 4 {
			e = event.Event{
				ID: id, Kind: event.KindTransition, Type: "decision", At: at,
				Repo: "bench", Actor: "human:bench", Source: "human",
				Ref:     fmt.Sprintf("01BENCH%019d", i-1),
				Payload: json.RawMessage(`{"to":"verified","reason":"bench"}`),
			}
		} else {
			e = event.Event{
				ID: id, Kind: event.KindObservation, Type: "decision", At: at,
				Repo: "bench", Actor: "agent:bench", Source: "consumer:bench",
				Payload: json.RawMessage(fmt.Sprintf(`{"body":"decision %d","evidence":"bench"}`, i)),
			}
		}
		line, err := json.Marshal(e)
		if err != nil {
			tb.Fatal(err)
		}
		w.Write(line)
		w.WriteByte('\n')
	}
	if err := w.Flush(); err != nil {
		tb.Fatal(err)
	}
}

// BenchmarkReplayFold pins the events-vs-replay-time curve (ledger entry
// 01KWZHFV29): replay is linear in events by design; this is where we
// watch the slope.
func BenchmarkReplayFold(b *testing.B) {
	for _, n := range []int{1_000, 10_000, 50_000} {
		b.Run(fmt.Sprintf("events=%d", n), func(b *testing.B) {
			dir := b.TempDir()
			writeBenchLedger(b, dir, n)
			store := jsonl.Open(dir)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := replayFolder(store); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestReplayBudget is the compaction tripwire: when a 50k-event replay
// blows this budget, the answer is the fold-cache keyed by content-hash
// of closed monthly files — not raising the budget. The bound is generous
// because CI boxes are slow and -race multiplies cost; the point is the
// order of magnitude.
func TestReplayBudget(t *testing.T) {
	dir := t.TempDir()
	writeBenchLedger(t, dir, 50_000)
	store := jsonl.Open(dir)
	start := time.Now()
	folder, err := replayFolder(store)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if got := len(folder.Observations()); got != 40_000 {
		t.Fatalf("fold dropped events: %d observations, want 40000", got)
	}
	if budget := 10 * time.Second; elapsed > budget {
		t.Fatalf("50k-event replay took %v (budget %v) — time to build the fold cache", elapsed, budget)
	}
	t.Logf("50k-event replay+fold: %v", elapsed)
}
