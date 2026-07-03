<!-- kervo:begin -->
<!-- kervo:artifact v1 skeleton=fact-only -->
# Context Artifact

> Machine-generated context for AI agents. Fact sections are deterministic;
> slot sections carry trust-labeled observations. Regenerate with `kervo compile`
> — do not edit by hand.

## Repository Summary

- Name: kervo
- Branch: main
- Languages: Go, Markdown
- Frameworks: Go
- Docs: README.md, CLAUDE.md

### README.md (excerpt)

> Workspace context compiler. See `docs/` for PRD / RFC / ARCH.

## Commands

- `make build` — go build -o kervo ./cmd/kervo
- `make test` — go test ./...
- `make arch-check` — @! grep -rn "internal/adapters" internal/core internal/ports \

## Recent Changes

- `32e0bf1` 2026-07-03 skeleton: add Commands section (workspace-declared entry points)
- `72ca9b4` 2026-07-03 fix scanner noise and marker impersonation found by dogfooding
- `e6a0d98` 2026-07-03 kervo v0: deterministic context compiler — Phase 1 complete

### Frequently Changed Files

- internal/adapters/source/files/files.go (3)
- internal/adapters/source/files/files_test.go (3)
- internal/core/compiler/compiler.go (3)
- internal/core/compiler/compiler_test.go (3)
- internal/adapters/consumer/claudecode/inject.go (2)
- internal/adapters/consumer/claudecode/inject_test.go (2)
- internal/adapters/source/gitexec/scan.go (2)
- internal/adapters/source/gitexec/scan_test.go (2)
- internal/cli/init.go (2)
- internal/core/artifact/artifact.go (2)

## Open Tasks

_No TODO/FIXME comments found._

## Related Modules

- assets/ (1 files)
- cmd/ (1 files)
- internal/ (33 files)

## Workspace Facts

- Commits analyzed: 3 (complete)
- Open tasks (TODO/FIXME): 0
- Top-level modules: 3
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
<!-- kervo:end -->
