package trust

import (
	"encoding/json"
	"time"

	"github.com/kervo-os/kervo/internal/core/event"
)

// Observation is the replay-derived current view of one observation.
// Nothing here is stored — state is always folded from the ledger
// (RFC-0005: events are truth, views are derived).
type Observation struct {
	ID        string
	Type      string // capture type: decision, risk, goal, note, summary, ...
	Body      string
	Evidence  string // how the proposer verified it — "" if none attached
	At        time.Time
	Actor     string // who proposed it
	Source    string
	State     State
	LastActor string // who put it in its current state
	Reason    string // reason of the last transition, "" if none
	// Conflict is set when different actors' latest judgments disagree
	// (RFC-0005 §2.2: latest wins, disagreement is surfaced, resolution
	// is v2 multi-party territory — v1 only refuses to hide it).
	Conflict bool
}

type transitionPayload struct {
	To     State  `json:"to"`
	Reason string `json:"reason"`
}

type observationPayload struct {
	Body     string `json:"body"`
	Evidence string `json:"evidence"`
}

// Folder folds ledger events into current observation views.
// Feed events in replay (ID) order.
type Folder struct {
	byID  map[string]*Observation
	order []string
	// latest transition target per observation per actor, for conflict detection
	actorLatest map[string]map[string]State
}

func NewFolder() *Folder {
	return &Folder{byID: map[string]*Observation{}, actorLatest: map[string]map[string]State{}}
}

// Add folds one event. Facts are ignored here (they are always true and
// never lifecycle-managed); unknown refs are ignored (a transition may
// arrive from a branch whose observation was compacted — never fatal).
func (f *Folder) Add(e event.Event) {
	switch e.Kind {
	case event.KindObservation:
		var p observationPayload
		_ = json.Unmarshal(e.Payload, &p)
		o := &Observation{
			ID: e.ID, Type: e.Type, Body: p.Body, Evidence: p.Evidence, At: e.At,
			Actor: e.Actor, Source: e.Source,
			State: initialState(e.Actor), LastActor: e.Actor,
		}
		f.byID[e.ID] = o
		f.order = append(f.order, e.ID)
	case event.KindTransition:
		o, ok := f.byID[e.Ref]
		if !ok {
			return
		}
		var p transitionPayload
		if json.Unmarshal(e.Payload, &p) != nil || p.To == "" {
			return
		}
		// Latest wins unconditionally during fold: the ledger may contain
		// edges that were legal against a state later reordered by merge.
		o.State = p.To
		o.Reason = p.Reason
		o.LastActor = e.Actor
		if f.actorLatest[e.Ref] == nil {
			f.actorLatest[e.Ref] = map[string]State{}
		}
		f.actorLatest[e.Ref][e.Actor] = p.To
		o.Conflict = disagree(f.actorLatest[e.Ref])
	}
}

// initialState: a human recording their own observation has observed it;
// machine proposals start as Generated (PRD §7.2 — imports enter Observed,
// provider output enters Generated).
func initialState(actor string) State {
	if len(actor) >= 6 && actor[:6] == "human:" {
		return Observed
	}
	return Generated
}

func disagree(latest map[string]State) bool {
	var first State
	seen := false
	for _, s := range latest {
		if !seen {
			first, seen = s, true
			continue
		}
		if s != first {
			return true
		}
	}
	return false
}

// Observations returns views in ledger order (stable, ULID-time-ordered).
func (f *Folder) Observations() []Observation {
	out := make([]Observation, 0, len(f.order))
	for _, id := range f.order {
		out = append(out, *f.byID[id])
	}
	return out
}

// Get returns the view for a full observation ID.
func (f *Folder) Get(id string) (Observation, bool) {
	o, ok := f.byID[id]
	if !ok {
		return Observation{}, false
	}
	return *o, true
}
