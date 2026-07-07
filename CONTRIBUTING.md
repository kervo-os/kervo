# Contributing to kervo

Thanks for considering it. The short version: small fixes → PR directly;
anything that adds surface → open an issue first. Features here need
**evidence or a gate** — "another tool has it" is not a reason (that rule
is in this repo's own ledger; `kervo dash` renders it).

## Dev loop

```bash
make build            # go 1.23+; stdlib only — there is no step two
go test -race ./...   # includes the skeleton byte-identity golden
make arch-check       # internal/core must not import adapters
```

Changed the compiler's output on purpose? Regenerate the golden and say so
in the PR:

```bash
go test ./internal/core/compiler -run TestSkeletonByteIdentity -update
```

## Hard rules (CI enforces the first three)

- **Zero dependencies.** `go.mod` stays stdlib-only; a new dependency needs
  an exceptional, argued reason.
- **Determinism.** The skeleton never reads a clock, the network, or an
  LLM. Same input, same bytes.
- **Core purity.** `internal/core` imports no adapters, no I/O.
- **Tests ride along.** Behavior changes carry a test; bug fixes carry a
  regression test naming the incident.

## Files you don't edit by hand

`CLAUDE.md` / `AGENTS.md` marker blocks and `.kervo/` (the artifact and
the event ledger) are machine-written — kervo dogfoods itself in this
repo. Don't include hand edits or your own ledger events in a PR; the
maintainer's post-merge compile refreshes the artifact. One exception:
if your change alters compile output, the regenerated golden file belongs
in the PR (see above).

Trust states are human signatures. No PR flips an observation to
`verified` — that judgment happens in review, by a human, on the ledger.

## Licensing

Apache-2.0. Per §5 of the license, submitting a contribution means you
license it under the same terms. No CLA.
