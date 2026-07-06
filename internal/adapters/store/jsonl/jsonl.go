// Package jsonl implements the event ledger — the Source of Truth per
// RFC-0005: one event per line in .kervo/events/YYYY-MM.jsonl, append-only,
// ULID-keyed, committed to git so branch merges are unions. SQLite (the
// sqlite package) is only a rebuildable index over these files.
package jsonl

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/ports"
)

// Store appends to and replays the workspace ledger.
type Store struct {
	dir string // <workspace>/.kervo/events
}

var _ ports.EventStore = (*Store)(nil)

func Open(workspaceDir string) *Store {
	return &Store{dir: filepath.Join(workspaceDir, ".kervo", "events")}
}

// Append writes e as one JSON line to the current month's file. Missing
// ID/At are assigned here (ULID / now) — callers may preset them for
// deterministic tests. O_APPEND keeps concurrent hook writes line-atomic
// for our record sizes.
func (s *Store) Append(ctx context.Context, e event.Event) (string, error) {
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	if e.ID == "" {
		id, err := newULID(e.At)
		if err != nil {
			return "", err
		}
		e.ID = id
	}
	line, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", err
	}
	name := filepath.Join(s.dir, e.At.UTC().Format("2006-01")+".jsonl")
	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return "", err
	}
	return e.ID, nil
}

// Replay streams events in ID (= time) order, starting after fromID
// (empty = from the beginning). Events merged in from other branches are
// interleaved correctly because ULIDs sort by creation time.
func (s *Store) Replay(ctx context.Context, fromID string, fn func(event.Event) error) error {
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return nil // no ledger yet — empty replay, not an error
	}
	if err != nil {
		return err
	}
	var files []string
	for _, en := range entries {
		if !en.IsDir() && strings.HasSuffix(en.Name(), ".jsonl") {
			files = append(files, en.Name())
		}
	}
	sort.Strings(files) // YYYY-MM sorts chronologically

	var all []event.Event
	for _, name := range files {
		f, err := os.Open(filepath.Join(s.dir, name))
		if err != nil {
			return err
		}
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64<<10), 4<<20)
		lineNo := 0
		for sc.Scan() {
			lineNo++
			raw := strings.TrimSpace(sc.Text())
			if raw == "" {
				continue
			}
			var e event.Event
			if err := json.Unmarshal([]byte(raw), &e); err != nil {
				f.Close()
				// A broken line means a corrupted ledger — surface loudly,
				// never skip silently (the ledger must be trustworthy).
				return fmt.Errorf("jsonl: %s line %d corrupt: %w", name, lineNo, err)
			}
			all = append(all, e)
		}
		f.Close()
		if err := sc.Err(); err != nil {
			return err
		}
	}
	// Within one file appends are ordered, but merged branches interleave —
	// sort by ID (ULID: time-prefixed, lexicographically time-ordered).
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	for _, e := range all {
		if fromID != "" && e.ID <= fromID {
			continue
		}
		if err := fn(e); err != nil {
			return err
		}
	}
	return nil
}

// crockford is the ULID alphabet (no I, L, O, U).
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// ulidMu guards the monotonic state below.
var (
	ulidMu      sync.Mutex
	ulidLastMS  uint64
	ulidLastRnd [10]byte
)

// newULID builds a spec-shaped ULID: 48-bit millisecond timestamp +
// 80 random bits, Crockford base32, 26 chars, lexicographically sortable.
// Within one process the entropy is monotonic inside a millisecond
// (previous entropy + 1, not a re-roll): replay folds in ID order, so a
// rapid same-process sequence — an agent's capture → trust → capture —
// must sort exactly as appended or a transition can fold before the
// observation it references and silently vanish.
func newULID(t time.Time) (string, error) {
	var b [16]byte
	ms := uint64(t.UnixMilli())
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	ulidMu.Lock()
	if ms == ulidLastMS {
		for i := len(ulidLastRnd) - 1; i >= 0; i-- {
			ulidLastRnd[i]++
			if ulidLastRnd[i] != 0 {
				break
			}
		}
	} else {
		if _, err := rand.Read(ulidLastRnd[:]); err != nil {
			ulidMu.Unlock()
			return "", err
		}
		ulidLastMS = ms
	}
	copy(b[6:], ulidLastRnd[:])
	ulidMu.Unlock()
	// 16 bytes = 128 bits → 26 base32 chars (130 bits, top 2 bits zero).
	dst := make([]byte, 26)
	var acc uint32
	bits := 0
	pos := 25
	for i := 15; i >= 0; i-- {
		acc |= uint32(b[i]) << bits
		bits += 8
		for bits >= 5 && pos >= 0 {
			dst[pos] = crockford[acc&31]
			acc >>= 5
			bits -= 5
			pos--
		}
	}
	for pos >= 0 {
		dst[pos] = crockford[acc&31]
		acc >>= 5
		pos--
	}
	return string(dst), nil
}
