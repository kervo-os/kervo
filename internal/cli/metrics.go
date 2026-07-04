package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
)

// H3 counters (PRD §12): primary KPI is prompt-token reduction with vs
// without the artifact. Tokens are estimated from chars (~4 chars/token)
// because the hook path must stay dependency-free and deterministic.
// The explanation-message RATIO needs content classification and is
// deliberately not computed here — sizes are recorded so it can be
// derived later without losing history.

type promptMetric struct {
	Session         string `json:"session"`
	PromptChars     int    `json:"prompt_chars"`
	PromptWords     int    `json:"prompt_words"`
	ArtifactPresent bool   `json:"artifact_present"`
	ArtifactBytes   int    `json:"artifact_bytes"`
}

type sessionAgg struct {
	first    int // first prompt chars (min event ID = earliest)
	prompts  int
	chars    int
	artifact bool
}

type h3Report struct {
	Sessions map[string]*sessionAgg
}

func aggregateMetrics(events []event.Event) h3Report {
	r := h3Report{Sessions: map[string]*sessionAgg{}}
	for _, e := range events { // events arrive in replay (time) order
		if e.Type != "metric:prompt" {
			continue
		}
		var m promptMetric
		if json.Unmarshal(e.Payload, &m) != nil {
			continue
		}
		key := m.Session
		if key == "" {
			key = "unknown"
		}
		s := r.Sessions[key]
		if s == nil {
			// The first prompt of the session carries the A/B variable and
			// the First Prompt Size KPI.
			s = &sessionAgg{first: m.PromptChars, artifact: m.ArtifactPresent}
			r.Sessions[key] = s
		}
		s.prompts++
		s.chars += m.PromptChars
	}
	return r
}

type h3Side struct {
	sessions, firstSum, promptSum, charSum int
}

func (r h3Report) sides() (with, without h3Side) {
	for _, s := range r.Sessions {
		side := &without
		if s.artifact {
			side = &with
		}
		side.sessions++
		side.firstSum += s.first
		side.promptSum += s.prompts
		side.charSum += s.chars
	}
	return with, without
}

func runMetrics(args []string) error {
	fs := newFlagSet("metrics")
	dir := fs.String("dir", ".", "workspace directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var events []event.Event
	if err := jsonl.Open(*dir).Replay(context.Background(), "", func(e event.Event) error {
		events = append(events, e)
		return nil
	}); err != nil {
		return err
	}
	r := aggregateMetrics(events)
	with, without := r.sides()

	fmt.Printf("H3 counters — prompt sizes by artifact presence (A/B)\n")
	fmt.Printf("Sessions measured: %d\n\n", len(r.Sessions))
	fmt.Printf("%-26s %14s %14s\n", "", "with artifact", "without")
	row := func(label string, w, wo string) { fmt.Printf("%-26s %14s %14s\n", label, w, wo) }
	row("sessions", itoa(with.sessions), itoa(without.sessions))
	row("avg first prompt (chars)", avg(with.firstSum, with.sessions), avg(without.firstSum, without.sessions))
	row("  ≈ tokens", avg(with.firstSum/4, with.sessions), avg(without.firstSum/4, without.sessions))
	row("avg prompts / session", avg(with.promptSum, with.sessions), avg(without.promptSum, without.sessions))
	row("avg chars / session", avg(with.charSum, with.sessions), avg(without.charSum, without.sessions))
	fmt.Printf("\nNote: explanation-message ratio requires content classification\n")
	fmt.Printf("(semantic) — raw sizes are recorded so it can be derived later.\n")
	return nil
}

func itoa(n int) string { return fmt.Sprintf("%d", n) }

func avg(sum, n int) string {
	if n == 0 {
		return "—"
	}
	return fmt.Sprintf("%d", sum/n)
}
