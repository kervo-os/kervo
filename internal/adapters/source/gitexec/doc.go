// Package gitexec implements ports.SourceProvider by shelling out to the
// git CLI (`git log --format=...`, `git diff --stat`).
//
// Decision record: go-git was rejected — pure-Go packfile parsing is too
// slow on large repos and threatens the 30s init budget. git CLI parses
// 10k commits in <1s. Caps: last N commits (default 500), respect
// .gitignore, mark fact.Snapshot.Partial when limits hit.
package gitexec
