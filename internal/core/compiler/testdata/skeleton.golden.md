<!-- kervo:artifact v1 skeleton=fact-only -->
# Context Artifact

> Machine-generated context for AI agents. Fact sections are deterministic;
> slot sections carry trust-labeled observations. Regenerate with `kervo compile`
> — do not edit by hand.

## Repository Summary

- Name: demo-api
- Branch: main
- Languages: Go, Markdown
- Frameworks: Go
- Docs: README.md, CLAUDE.md

### README.md (excerpt)

> A demo REST API used to exercise the compiler. It has two paragraphs.

## Commands

- `make build` — go build -o demo ./cmd
- `npm run test` — jest --coverage

## Recent Changes

- `aaaaaaa` 2026-07-03 add auth middleware
- `bbbbbbb` 2026-07-02 fix token refresh
- `ccccccc` 2026-07-01 initial commit

_Showing 3 of 3 analyzed commits. Scan capped — older history not analyzed (Partial)._

### Frequently Changed Files

- mw/auth.go (2)
- main.go (1)
- token.go (1)

## Open Tasks

- mw/auth.go:42 — TODO: rotate signing keys
- main.go:7 — FIXME: graceful shutdown

## Related Modules

- mw/ (3 files)
- store/ (5 files)

## Workspace Facts

- Commits analyzed: 3 (partial — caps hit)
- Open tasks (TODO/FIXME): 2
- Top-level modules: 2
- Docs captured: 2

## Possible Current Goal

<!-- kervo:slot:goal:begin -->
_No proposal yet. A confirmed goal becomes the first Verified observation._
<!-- kervo:slot:goal:end -->

## Known Decisions

<!-- kervo:slot:decisions:begin -->
_None proposed yet. Semantic providers (Mode 2/3) attach labeled observations here._
<!-- kervo:slot:decisions:end -->

## Known Risks

<!-- kervo:slot:risks:begin -->
_None proposed yet. Semantic providers (Mode 2/3) attach labeled observations here._
<!-- kervo:slot:risks:end -->

## Doc Summaries

<!-- kervo:slot:summaries:begin -->
_None proposed yet. Semantic providers (Mode 2/3) attach labeled observations here._
<!-- kervo:slot:summaries:end -->

## Deprecated / Stale Notes

_None recorded. Stale or deprecated observations are listed here with their
exclusion reason instead of being silently dropped._
