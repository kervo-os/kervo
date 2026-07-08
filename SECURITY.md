# Security Policy

## Reporting

Use GitHub's private vulnerability reporting:
**Security tab → Report a vulnerability** on `kervo-os/kervo`.
Please don't open public issues for exploitable bugs. Best-effort
response within a few days; there is no bounty program.

## Supported

The latest tagged release.

## Scope notes

kervo is a local CLI — no server, no accounts, no network calls at
runtime (Mode 3 talks only to the endpoint you configure). The
interesting bug classes are:

- **Marker impersonation**: workspace content smuggling `<!-- kervo:`
  structural markers through the escape layer into artifacts or
  injected blocks.
- **Ledger poisoning**: crafted events that break replay, self-promote
  trust states, or corrupt `merge=union` folding.
- **Hook payload handling**: the Claude Code hook parses untrusted JSON
  from stdin; recursion-guard bypass or payload-driven writes outside
  `.kervo/` would qualify.
- **Secret leakage**: anything that copies file contents, prompts, or
  paths outside the workspace into the committed ledger (it stores
  names and sizes only, by design).

## Prompt injection through the artifact

Injecting compiled text into agent context is an injection surface, and
we treat it as one rather than deny it:

- Nothing enters the artifact anonymously — every non-fact observation
  carries an actor and a source, and machine proposals are quarantined
  as `generated` until a human judges them. Agents cannot sign their
  own claims; the transition to `verified` is a human act.
- The trust boundary is git's own. Someone with commit access could
  poison the ledger — and could equally poison your README, Makefile,
  or CLAUDE.md itself. kervo inherits the repository's existing trust
  model; it adds no new privileged channel.
- The scenario is measured, not assumed: a pre-registered blind
  experiment seeded false "decisions" into the ledger; all observed
  infections occurred in the unmanaged arm. Protocols and raw
  responses: [kervo-os/experiments](https://github.com/kervo-os/experiments).

## Supply chain

Zero runtime dependencies (stdlib-only `go.mod`). Release binaries ship
with `checksums.txt`; the npm wrapper verifies the downloaded archive
against it before executing anything; the Homebrew cask is published by
CI from the tagged commit.
