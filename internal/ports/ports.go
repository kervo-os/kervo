// Package ports declares the interfaces core depends on.
// Adapters implement these; core never imports adapters (ARCH-0001 §2).
//
// RULE: keep this file at 4 ports (hard ceiling 6). A new port needs a
// ports-level justification, not just a new adapter — adapters multiply,
// contracts must not.
package ports

import (
	"context"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/core/fact"
)

// SourceProvider produces deterministic facts from a workspace.
// Implementations: gitexec (git CLI via exec — NOT go-git, perf), files,
// external stores (post-H2' interop).
type SourceProvider interface {
	// Scan reads from cursor (empty = full scan) and returns a snapshot
	// plus the next cursor for incremental runs.
	Scan(ctx context.Context, dir, cursor string) (fact.Snapshot, string, error)
}

// SemanticProvider proposes enhancements. Contract per RFC-0003 §5:
// input includes skeleton + facts; output is Generated-only proposals;
// it must never rewrite skeleton sections.
type SemanticProvider interface {
	Propose(ctx context.Context, skeleton string, snap fact.Snapshot) ([]artifact.Enhancement, error)
}

// EventStore is append-only. Replay feeds views (current trust states,
// RFC-0002 retroactive index).
type EventStore interface {
	Append(ctx context.Context, e event.Event) (string, error)
	Replay(ctx context.Context, fromID string, fn func(event.Event) error) error
}

// ConsumerInjector delivers the artifact to an agent surface between
// artifact.MarkerBegin/End, preserving human-written content outside.
type ConsumerInjector interface {
	Inject(ctx context.Context, workspaceDir, rendered string) error
}
