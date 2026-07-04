// Package trust implements the Observation lifecycle state machine (PRD §6.3).
// No confidence numbers — states only. Humans correct, never gate.
package trust

// State of an Observation.
type State string

const (
	Generated  State = "generated"  // proposed by a semantic provider
	Observed   State = "observed"   // seen/used without contradiction
	Verified   State = "verified"   // explicitly confirmed by a human
	Stale      State = "stale"      // conservative heuristic: age or related-file change
	Deprecated State = "deprecated" // explicitly retired
)

// CanTransition encodes the allowed edges.
// Generated -> {Observed, Verified}; Observed -> Verified; any -> Stale;
// Stale -> {Verified, Deprecated}. Generated -> Verified is legal because a
// human confirmation of a fresh proposal is one act, not two (PRD §7.2:
// the Cold Start goal Confirm becomes the first Verified).
func CanTransition(from, to State) bool {
	switch from {
	case Generated:
		return to == Observed || to == Verified || to == Stale || to == Deprecated
	case Observed:
		return to == Verified || to == Stale || to == Deprecated
	case Verified:
		return to == Stale || to == Deprecated
	case Stale:
		return to == Verified || to == Deprecated
	default:
		return false
	}
}

// Weight controls artifact treatment (RFC-0002 §6 applies the same table).
// Fact is handled outside this package: always included.
func Weight(s State) string {
	switch s {
	case Verified:
		return "prefer"
	case Observed, Generated:
		return "include-labeled"
	case Stale:
		return "exclude-with-reason"
	default:
		return "exclude"
	}
}
