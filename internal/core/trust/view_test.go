package trust

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/core/event"
)

func obsEvent(id, actor, body string) event.Event {
	p, _ := json.Marshal(map[string]string{"body": body})
	return event.Event{ID: id, Kind: event.KindObservation, Type: "decision",
		At: time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC), Actor: actor, Payload: p}
}

func transEvent(id, ref, actor string, to State, reason string) event.Event {
	p, _ := json.Marshal(map[string]string{"to": string(to), "reason": reason})
	return event.Event{ID: id, Kind: event.KindTransition, Type: "transition",
		Actor: actor, Ref: ref, Payload: p}
}

func TestInitialStateByActor(t *testing.T) {
	f := NewFolder()
	f.Add(obsEvent("01A", "human:kim", "we chose JWT"))
	f.Add(obsEvent("01B", "backend:llama", "goal guess"))
	obs := f.Observations()
	if obs[0].State != Observed {
		t.Errorf("human observation starts %s, want observed", obs[0].State)
	}
	if obs[1].State != Generated {
		t.Errorf("machine proposal starts %s, want generated", obs[1].State)
	}
}

func TestLatestTransitionWinsAndReasonTracked(t *testing.T) {
	f := NewFolder()
	f.Add(obsEvent("01A", "human:kim", "we chose JWT"))
	f.Add(transEvent("01B", "01A", "human:kim", Verified, "confirmed in review"))
	f.Add(transEvent("01C", "01A", "system", Stale, "age 40d"))
	o, _ := f.Get("01A")
	if o.State != Stale || o.Reason != "age 40d" || o.LastActor != "system" {
		t.Errorf("latest transition not applied: %+v", o)
	}
}

// RFC-0005 §2.2: disagreement between actors is surfaced, never hidden.
func TestConflictWhenActorsDisagree(t *testing.T) {
	f := NewFolder()
	f.Add(obsEvent("01A", "human:kim", "we chose JWT"))
	f.Add(transEvent("01B", "01A", "human:kim", Verified, ""))
	f.Add(transEvent("01C", "01A", "human:lee", Deprecated, "superseded by sessions"))
	o, _ := f.Get("01A")
	if !o.Conflict {
		t.Error("kim=verified vs lee=deprecated must flag conflict")
	}
	if o.State != Deprecated {
		t.Errorf("latest wins: state = %s, want deprecated", o.State)
	}
}

// A transition whose observation is unknown (compacted away on another
// branch) must never be fatal.
func TestUnknownRefIgnored(t *testing.T) {
	f := NewFolder()
	f.Add(transEvent("01B", "GONE", "human:kim", Verified, ""))
	if len(f.Observations()) != 0 {
		t.Error("orphan transition created an observation")
	}
}
