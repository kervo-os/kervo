// Package sqlite implements ports.EventStore on a single .kervo/events.db
// file (WAL mode). Append-only table; no UPDATE/DELETE statements exist in
// this package by rule. Driver choice (modernc.org/sqlite for CGO-free
// builds vs mattn/go-sqlite3) decided at Issue #1 step 4.
package sqlite
