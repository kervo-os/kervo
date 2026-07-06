package jsonl

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/core/event"
)

func TestAppendAssignsULIDAndMonthlyFile(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	at := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	id, err := s.Append(context.Background(), event.Event{Kind: event.KindFact, Type: "commit", At: at})
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 26 {
		t.Errorf("ULID length = %d, want 26: %q", len(id), id)
	}
	if _, err := os.Stat(filepath.Join(dir, ".kervo", "events", "2026-07.jsonl")); err != nil {
		t.Errorf("monthly ledger file missing: %v", err)
	}
}

func TestReplayOrderAndCursor(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	base := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	var ids []string
	for i := 0; i < 3; i++ {
		id, err := s.Append(context.Background(), event.Event{
			Kind: event.KindObservation, Type: "note", At: base.Add(time.Duration(i) * time.Second),
		})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	var got []string
	if err := s.Replay(context.Background(), "", func(e event.Event) error {
		got = append(got, e.ID)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0] != ids[0] || got[2] != ids[2] {
		t.Errorf("replay order wrong: %v vs %v", got, ids)
	}

	var after []string
	if err := s.Replay(context.Background(), ids[0], func(e event.Event) error {
		after = append(after, e.ID)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(after) != 2 || after[0] != ids[1] {
		t.Errorf("fromID cursor wrong: %v", after)
	}
}

// RFC-0005 §2.2: branch merges are unions. Simulate a merged ledger file by
// interleaving two branches' appends out of order — replay must still be
// globally time-ordered and lossless.
func TestMergedLedgerReplaysAsUnion(t *testing.T) {
	dirA, dirB := t.TempDir(), t.TempDir()
	base := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	a := Open(dirA)
	b := Open(dirB)
	// branch A writes t0, t2; branch B writes t1, t3
	for i, s := range []*Store{a, b, a, b} {
		if _, err := s.Append(context.Background(), event.Event{
			Kind: event.KindObservation, Type: "note", At: base.Add(time.Duration(i) * time.Second),
			Ref: []string{"a0", "b1", "a2", "b3"}[i],
		}); err != nil {
			t.Fatal(err)
		}
	}
	// git merge = union: concatenate B's file onto A's (order scrambled on purpose).
	fileA := filepath.Join(dirA, ".kervo", "events", "2026-07.jsonl")
	fileB := filepath.Join(dirB, ".kervo", "events", "2026-07.jsonl")
	bRaw, _ := os.ReadFile(fileB)
	aRaw, _ := os.ReadFile(fileA)
	if err := os.WriteFile(fileA, append(bRaw, aRaw...), 0o644); err != nil {
		t.Fatal(err)
	}

	var refs []string
	if err := a.Replay(context.Background(), "", func(e event.Event) error {
		refs = append(refs, e.Ref)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	want := "a0,b1,a2,b3"
	got := ""
	for i, r := range refs {
		if i > 0 {
			got += ","
		}
		got += r
	}
	if got != want {
		t.Errorf("union replay = %s, want %s", got, want)
	}
}

// The ledger must be trustworthy: a corrupt line is a loud error, never a
// silent skip.
func TestCorruptLineFailsLoudly(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	if _, err := s.Append(context.Background(), event.Event{Kind: event.KindFact, Type: "x"}); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, ".kervo", "events"))
	path := filepath.Join(dir, ".kervo", "events", entries[0].Name())
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("{broken json\n")
	f.Close()
	if err := s.Replay(context.Background(), "", func(event.Event) error { return nil }); err == nil {
		t.Fatal("corrupt ledger line must fail replay loudly")
	}
}

// Same-millisecond appends must sort exactly as appended — replay folds in
// ID order, and an agent's capture → trust → capture happens well inside
// one millisecond. (Regression: random per-call entropy let a transition
// sort before the observation it referenced, silently dropping it.)
func TestULIDMonotonicWithinMillisecond(t *testing.T) {
	at := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	prev := ""
	for i := 0; i < 5000; i++ {
		id, err := newULID(at)
		if err != nil {
			t.Fatal(err)
		}
		if id <= prev {
			t.Fatalf("ULID %d not monotonic: %s after %s", i, id, prev)
		}
		prev = id
	}
}
