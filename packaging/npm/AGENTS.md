<!-- kervo:begin -->
<!-- kervo:artifact v1 skeleton=fact-only lang=en -->
# Context Artifact

> Machine-generated context for AI agents. Fact sections are deterministic;
> slot sections carry trust-labeled observations. Regenerate with `kervo compile`
> — do not edit by hand.
>
> Start from this artifact: state what it already answers for your task,
> then explore only the gaps — and verify only what your task puts in
> question.

## Repository Summary

- Name: kervo
- Branch: main
- Languages: JavaScript, Markdown
- Frameworks: Node.js
- Docs: README.md

### README.md (excerpt)

> **Stop re-explaining your project to AI. `kervo init` once.**

## Commands

_No declared commands found (Makefile targets, package.json scripts)._

## Recent Changes

- `d5e9028` 2026-07-07 init: AGENTS.md joins the workspace-found row
- `0a8da8f` 2026-07-07 readme: trust lifecycle as a state diagram — the prose arrow was a straight line, the machine is not
- `653e32a` 2026-07-07 readme: H4 result as a chart — the 29.2pp gap should not hide in a gray table
- `773f814` 2026-07-07 readme: npm joins the install paths — all three languages
- `1a2f72c` 2026-07-07 ledger: session hook events
- `184d43b` 2026-07-07 ledger: session hook events
- `8813dc6` 2026-07-07 readme: quickstart as a colored terminal card (freeze SVG of the real init styling)
- `490327b` 2026-07-07 npm: the wrapper downloads the real binary
- `3d1ffc3` 2026-07-07 community: CONTRIBUTING + SECURITY (private vuln reporting enabled)
- `a78866f` 2026-07-07 changelog: condense 0.1–0.19 + ci: least privilege, SHA-pinned actions
- `54a7646` 2026-07-07 ledger: blockchain verdict + H5 anchored registration
- `916897a` 2026-07-07 test: drop an import the assertion rewrite left behind
- `95e0a5f` 2026-07-07 autocompile: pre-commit, not post-commit — the tree must converge
- `24852c1` 2026-07-07 ledger: session hook events
- `7f4a5d1` 2026-07-07 ledger: session hook events
- `33c46f6` 2026-07-07 freshness is the default — compile wires the git hooks itself
- `e686558` 2026-07-07 init: the wizard keeps the digest fresh — commits and pulls auto-compile
- `3493ca2` 2026-07-07 ledger: session hook events
- `5ad7201` 2026-07-07 changelog: v0.20.0
- `ee473ef` 2026-07-07 ledger: session hook events

_Showing 20 of 119 analyzed commits._

### Frequently Changed Files

- README.md (40)

## Open Tasks

_No TODO/FIXME comments found._

## Related Modules

- bin/ (1 files)

## Workspace Facts

- Commits analyzed: 119 (complete)
- Open tasks (TODO/FIXME): 0
- Top-level modules: 1
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
_None proposed yet. Semantic providers (Mode 2/3) attach labeled observations here._
<!-- kervo:slot:summaries:end -->

## Deprecated / Stale Notes

<!-- kervo:slot:stale:begin -->
_None recorded. Stale or deprecated observations are listed here with their
exclusion reason instead of being silently dropped._
<!-- kervo:slot:stale:end -->

## Write-back Protocol

> For AI consumers — close the loop. If this session taught you a durable
> fact this artifact does not carry (how to run something, what a component
> does, a decision made in conversation), stage it for human judgment:
>
> `kervo capture -type decision|risk|summary|goal -actor "agent:<you>" -body "<the fact>" -evidence "<how you verified it>"`
>
> Rules: facts you observed, not speculation · conclusions, not corpus —
> what lives in a file agents can read stays there; cite it as evidence
> instead of mirroring it · start the body with a
> one-line claim, details after it · attach evidence — the command you
> ran, the doc you read — so the human can sign in one keystroke · one
> capture per fact · skip what this artifact already says · never include
> secrets or file contents. Duplicates are dropped automatically.
>
> The conversation is the review: if the human affirmed a fact in this
> session, relay their judgment with the capture (`kervo trust -to
> verified -reason "<their words>"`) — only undiscussed facts wait in
> the queue (`kervo review`). If evidence contradicts a verified entry,
> raise it with the human and record their updated judgment instead of
> re-proposing.
<!-- kervo:end -->
