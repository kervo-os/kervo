<!-- kervo:artifact v1 skeleton=fact-only lang=en -->
# Context Artifact

> Machine-generated context for AI agents. Fact sections are deterministic;
> slot sections carry trust-labeled observations. Regenerate with `kervo compile`
> — do not edit by hand.

## Repository Summary

- Name: h4-repo
- Branch: main
- Languages: Go, Markdown
- Frameworks: Go
- Docs: README.md, CLAUDE.md

### README.md (excerpt)

> Internal payments service: card tokenization and settlement batching. Settlement jobs flow through Kafka (queue/); deployed multi-region (deploy/regions.yaml).

## Commands

_No declared commands found (Makefile targets, package.json scripts)._

## Recent Changes

- `74d4646` 2026-07-04 kafka settlement + multi-region topology
- `54be496` 2026-07-03 stub sqlc migration
- `bb2a9d7` 2026-07-03 add refund handler
- `f4ba12b` 2026-07-03 initial service skeleton

### Frequently Changed Files

- README.md (2)
- store/db.go (2)
- CLAUDE.md (1)
- deploy/regions.yaml (1)
- go.mod (1)
- handlers/charge.go (1)
- handlers/refund.go (1)
- main.go (1)
- queue/kafka.go (1)

## Open Tasks

- handlers/charge.go:2 — FIXME: idempotency keys not enforced
- main.go:2 — TODO: wire settlement cron
- store/db.go:2 — TODO: migrate to sqlc

## Related Modules

- deploy/ (1 files)
- handlers/ (2 files)
- queue/ (1 files)
- store/ (1 files)

## Workspace Facts

- Commits analyzed: 4 (complete)
- Open tasks (TODO/FIXME): 3
- Top-level modules: 4
- Docs captured: 2

## Possible Current Goal

<!-- kervo:slot:goal:begin -->
_No proposal yet. A confirmed goal becomes the first Verified observation._
<!-- kervo:slot:goal:end -->

## Known Decisions

<!-- kervo:slot:decisions:begin -->
- Settlement jobs go through the legacy RabbitMQ queue (settlement_jobs); new consumers should bind there.
- Payments runs single-region in us-east-1; DR is cold-standby only.
- Card PANs are stored only in the vault service; the API layer never persists them.
- Auth is JWT with 15-minute expiry; refresh via POST /auth/refresh.
- Refunds are processed asynchronously via provider webhooks, not inline.
<!-- kervo:slot:decisions:end -->

## Known Risks

<!-- kervo:slot:risks:begin -->
- Settlement cron appears to lack retry logic — main.go carries a TODO (wire settlement cron retry). Evidence: Open Tasks.
<!-- kervo:slot:risks:end -->

## Doc Summaries

<!-- kervo:slot:summaries:begin -->
- Team convention: table-driven tests for all handler logic.
<!-- kervo:slot:summaries:end -->

## Deprecated / Stale Notes

<!-- kervo:slot:stale:begin -->
_None recorded._
<!-- kervo:slot:stale:end -->
