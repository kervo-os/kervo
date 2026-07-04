// Package event defines the append-only event model (Source of Truth).
// Rule: events are never mutated or deleted. State is derived by replay.
package event

import "time"

// Kind separates the two experience sources (PRD §6.2).
type Kind string

const (
	KindFact        Kind = "fact"        // machine-generated, always true
	KindObservation Kind = "observation" // proposed by AI/human, lifecycle-managed
	KindTransition  Kind = "transition"  // trust state change for an observation
)

// Event is the single append-only record type.
// Persistence per RFC-0005: newline JSON in .kervo/events/YYYY-MM.jsonl
// (committed — branch merges are unions), SQLite only as a local index.
type Event struct {
	ID   string // ULID, assigned by store — global uniqueness makes merge=union safe
	Kind Kind
	Type string // e.g. "commit", "decision", "todo", "goal"
	At   time.Time
	Repo string // repository identity within the workspace (RFC-0004 §2:
	// a Workspace holds 1+ repos; v1 implements one, the
	// model assumes many — set even when it seems redundant)
	Actor string // who recorded/judged — multi-party identity (RFC-0005 §3;
	// same logic as Repo: costs nothing now, is a migration later)
	Source  string // "git", "files", "consumer:claude-code", "human", ...
	Ref     string // subject id: commit sha, file path, observation id
	Payload []byte // JSON body; schema per Type
}
