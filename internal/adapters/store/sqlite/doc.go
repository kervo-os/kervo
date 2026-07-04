// Package sqlite implements the local read index over the event ledger.
//
// Role per RFC-0005 §2.3: the truth is .kervo/events/*.jsonl (append-only,
// ULID-keyed, committed to git — merges are unions); this package is only a
// derived index (.kervo/index.db, gitignored), rebuildable at any time via
// `kervo reindex`. On corruption or conflict: delete and rebuild.
// Append = JSONL write + index update behind the unchanged ports.EventStore.
// No UPDATE/DELETE statements may exist in this package by rule.
package sqlite
