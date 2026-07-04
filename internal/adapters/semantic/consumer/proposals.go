package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/trust"
	"github.com/kervo-os/kervo/internal/ports"
)

// ProposalsFile is the minimal Mode-2 transport: the consumer's LLM writes
// grounded proposals here; `kervo compile` attaches them as Generated
// enhancements. Interim by design — RFC-0003 §6 leaves the transport
// (Skill/MCP/Hook) to the Phase 3 spike; a file is the smallest legal one.
// The eventual home for observations is the RFC-0005 event ledger.
const ProposalsFile = "proposals.json"

// proposal is the on-disk shape. Note there is NO state field: providers
// cannot self-promote (RFC-0003 §5) — everything enters as Generated.
type proposal struct {
	Slot   string `json:"slot"`
	Body   string `json:"body"`
	Source string `json:"source"`
}

// FileProposals implements ports.SemanticProvider by reading proposals
// staged in <dir>/.kervo/proposals.json.
type FileProposals struct {
	Dir string
}

var _ ports.SemanticProvider = FileProposals{}

// Propose returns staged proposals, or (nil, nil) when none are staged —
// absence is Mode 1, not an error (RFC-0003 §4 graceful degradation).
func (f FileProposals) Propose(ctx context.Context, skeleton string, snap fact.Snapshot) ([]artifact.Enhancement, error) {
	raw, err := os.ReadFile(filepath.Join(f.Dir, ".kervo", ProposalsFile))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("consumer: read proposals: %w", err)
	}
	var ps []proposal
	if err := json.Unmarshal(raw, &ps); err != nil {
		return nil, fmt.Errorf("consumer: proposals.json is not a JSON list: %w", err)
	}
	valid := map[string]bool{}
	for _, s := range artifact.Slots() {
		valid[s] = true
	}
	var es []artifact.Enhancement
	for i, p := range ps {
		if !valid[p.Slot] {
			return nil, fmt.Errorf("consumer: proposal %d targets unknown slot %q", i, p.Slot)
		}
		if p.Body == "" || p.Source == "" {
			return nil, fmt.Errorf("consumer: proposal %d lacks body or source", i)
		}
		es = append(es, artifact.Enhancement{
			Slot:   p.Slot,
			Body:   p.Body,
			State:  trust.Generated, // by construction — never from the file
			Source: p.Source,
		})
	}
	return es, nil
}
