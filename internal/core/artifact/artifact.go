// Package artifact defines the Context Artifact domain: sections, slots,
// injection markers, and enhancement records. The artifact is a build
// output — reproducible, versioned, vendor-neutral markdown (PRD §5.3).
package artifact

import "github.com/kervo-os/kervo/internal/core/trust"

// Skeleton sections are deterministic and owned by the compiler (Mode 1).
// Enhancement slots are the only mutation points for semantic providers.
const (
	SlotGoal      = "goal"
	SlotDecisions = "decisions"
	SlotRisks     = "risks"
	SlotSummaries = "summaries"
)

// Markers delimit the injected block inside CLAUDE.md; content outside
// the markers is human-owned and must never be touched (ARCH-0001 §4).
const (
	MarkerBegin = "<!-- kervo:begin -->"
	MarkerEnd   = "<!-- kervo:end -->"
)

// ReservedPrefix opens every structural marker above. Renderers must
// neutralize it inside data-derived text (compiler esc) so workspace
// content can never impersonate a marker — a TODO whose text embedded a
// literal end marker once truncated the CLAUDE.md injection (dogfooding).
const ReservedPrefix = "<!-- kervo:"

// Slots returns every enhancement slot the skeleton renders, in canonical
// section order. The compiler and Attach share this list so the
// Skeleton/Enhancement boundary stays byte-precise (RFC-0003 §2.2).
func Slots() []string {
	return []string{SlotGoal, SlotDecisions, SlotRisks, SlotSummaries}
}

// SlotBegin/SlotEnd delimit one enhancement slot inside the artifact.
// Only content between these markers may be touched by Attach.
func SlotBegin(slot string) string { return "<!-- kervo:slot:" + slot + ":begin -->" }
func SlotEnd(slot string) string   { return "<!-- kervo:slot:" + slot + ":end -->" }

// Enhancement is a semantic proposal attached to a named slot.
// The compiler renders trust labels; it never assigns states
// (RFC-0003 §2.2 — states are set by the lifecycle, not the renderer).
type Enhancement struct {
	Slot   string
	Body   string
	State  trust.State
	Source string // provider identity, printed as provenance
}
