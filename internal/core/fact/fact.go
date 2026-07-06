// Package fact defines the deterministic scan result model.
// Adapters produce these structs; core never touches I/O.
package fact

import "time"

// Snapshot is everything Mode 1 needs to build a skeleton.
// (Revision log — PRD risk #12 budgeted one: r1 Modules for Related
// Modules; r2 Commands, evidence-backed by arXiv:2512.18925 — build/test
// commands are the single most prevalent hand-written context type, 72%
// of 401 studied repos.)
type Snapshot struct {
	Repo     RepoInfo
	Commits  []Commit
	Files    []ChangedFile
	Modules  []Module
	Commands []Command
	Todos    []Todo
	Docs     []DocSummaryInput
	Partial  bool // true if scan caps were hit (ARCH perf budget)
}

// Module is a top-level directory of the repository with its tracked-file
// count — the deterministic input for Related Modules (RFC-0003 §2.1).
type Module struct {
	Path  string
	Files int
}

// Command is a runnable entry point the workspace itself declares
// (Makefile target, package.json script). Declared-only — never inferred:
// synthesizing commands would be Semantic territory.
type Command struct {
	Run    string // what to type: "make build", "npm run test"
	Detail string // underlying recipe/script line, may be empty
	Source string // manifest of origin: "Makefile", "package.json"
}

type RepoInfo struct {
	Name       string
	Languages  []string
	Frameworks []string // from manifests (package.json, go.mod, ...)
	Branch     string
	Ahead      int // commits ahead of upstream; 0 if none or no upstream
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
