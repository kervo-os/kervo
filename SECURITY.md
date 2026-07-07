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
