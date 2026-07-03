// Package files scans README/CLAUDE.md/docs and TODO|FIXME comments.
// Perf caps: skip binaries and vendored dirs (node_modules, vendor),
// parallel walk, hard file-count ceiling -> Snapshot.Partial.
package files
