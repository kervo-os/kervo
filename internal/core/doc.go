// Package core contains the pure domain model of Kervo OS.
//
// The system domain is Workspace Experience.
// Workspace Experience appears in two forms:
//   - Facts: deterministic records derived from source-controlled evidence.
//   - Observations: interpreted knowledge governed by the trust lifecycle.
//
// Events are the append-only source of truth.
// Facts and observations are recorded through events, but events are not
// subordinate to experience.
//
// Core packages must not import I/O, network, filesystem, database,
// process, or vendor SDK packages. (enforced: `make arch-check`)
package core
