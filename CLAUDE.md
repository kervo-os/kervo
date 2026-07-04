<!-- kervo:begin -->
<!-- kervo:artifact v1 skeleton=fact-only lang=en -->
# Context Artifact

> Machine-generated context for AI agents. Fact sections are deterministic;
> slot sections carry trust-labeled observations. Regenerate with `kervo compile`
> — do not edit by hand.

## Repository Summary

- Name: kervo
- Branch: main
- Languages: Go, Markdown, JavaScript
- Frameworks: Go
- Docs: README.md

### README.md (excerpt)

> **Stop re-explaining your project to AI. `kervo init` once.**

## Commands

- `make build` — go build -o kervo ./cmd/kervo
- `make test` — go test ./...
- `make arch-check` — @! grep -rn "internal/adapters" internal/core internal/ports \

## Recent Changes

- `1ed8e65` 2026-07-04 capture: event ledger + manual capture + consumer hook (H2' spike, build side)
- `990755f` 2026-07-04 event: adopt Actor value convention from planning-session skeleton
- `307b330` 2026-07-04 storage: adopt RFC-0005 layout — events are truth, artifacts are derived
- `cafedbe` 2026-07-04 packaging: npm name is @kervo-os/kervo (unscoped blocked by typosquat filter)
- `b61de3e` 2026-07-04 packaging: npm wrapper v0.0.1 (name reservation + future npx channel)
- `92d18b3` 2026-07-04 openaicompat: tolerate trailing commas (field finding from gpt-oss-120b)
- `ec25e11` 2026-07-04 prepare for public: Apache-2.0 license + real README
- `0ab4acf` 2026-07-04 semantic: Mode 3 backend via any OpenAI-compatible endpoint
- `2546883` 2026-07-04 i18n: artifact language setting — en default, ko/ja supported
- `134ae46` 2026-07-04 compile: minimal Mode 2 semantic path (file-transport proposals)
- `ce0bad5` 2026-07-03 ci: enforce the release gates on every push
- `d92f04c` 2026-07-03 track the compiled artifact (PRD §8 primitive team sharing)
- `32e0bf1` 2026-07-03 skeleton: add Commands section (workspace-declared entry points)
- `72ca9b4` 2026-07-03 fix scanner noise and marker impersonation found by dogfooding
- `e6a0d98` 2026-07-03 kervo v0: deterministic context compiler — Phase 1 complete

### Frequently Changed Files

- internal/cli/compile.go (5)
- .gitignore (4)
- .kervo/artifact.md (4)
- CLAUDE.md (4)
- internal/adapters/source/files/files.go (4)
- internal/adapters/source/files/files_test.go (4)
- internal/cli/init.go (4)
- internal/core/compiler/compiler.go (4)
- internal/core/compiler/compiler_test.go (4)
- internal/core/event/event.go (4)

## Open Tasks

_No TODO/FIXME comments found._

## Related Modules

- .github/ (1 files)
- assets/ (1 files)
- cmd/ (1 files)
- internal/ (44 files)
- packaging/ (3 files)

## Workspace Facts

- Commits analyzed: 15 (complete)
- Open tasks (TODO/FIXME): 0
- Top-level modules: 5
- Docs captured: 1

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
**[verified — human:refuse1993]**
Phase 3 spike: JSONL ledger + capture/hook landed
<!-- kervo:slot:summaries:end -->

## Deprecated / Stale Notes

<!-- kervo:slot:stale:begin -->
_None recorded. Stale or deprecated observations are listed here with their
exclusion reason instead of being silently dropped._
<!-- kervo:slot:stale:end -->
<!-- kervo:end -->
