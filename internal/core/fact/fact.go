// Package fact defines the deterministic scan result model.
// Adapters produce these structs; core never touches I/O.
package fact

import "time"

// Snapshot is everything Mode 1 needs to build a skeleton.
// (Revised once during gitexec implementation, as budgeted by PRD risk #12:
// Modules added for the deterministic Related Modules section.)
type Snapshot struct {
	Repo    RepoInfo
	Commits []Commit
	Files   []ChangedFile
	Modules []Module
	Todos   []Todo
	Docs    []DocSummaryInput
	Partial bool // true if scan caps were hit (ARCH perf budget)
}

// Module is a top-level directory of the repository with its tracked-file
// count — the deterministic input for Related Modules (RFC-0003 §2.1).
type Module struct {
	Path  string
	Files int
}

type RepoInfo struct {
	Name       string
	Languages  []string
	Frameworks []string // from manifests (package.json, go.mod, ...)
	Branch     string
}

type Commit struct {
	SHA     string
	At      time.Time
	Subject string
	Files   []string
}

type ChangedFile struct {
	Path    string
	Changes int
}

type Todo struct {
	Path string
	Line int
	Text string
}

// DocSummaryInput is raw doc content handed to the compiler.
// Summarizing it with an LLM is an Enhancement, never part of the skeleton.
type DocSummaryInput struct {
	Path    string
	Content string
}
