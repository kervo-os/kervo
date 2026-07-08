# Changelog

All notable changes, newest first. Versions are git tags; every release
ships prebuilt binaries and a Homebrew cask (`brew install kervo-os/tap/kervo`).

## v0.23.0 — 2026-07-08

Two red-team rounds, metabolized the same day. Every item below traces
to a verified decision in this repo's own ledger.

- `kervo check` grows into a full gate: **drift detection** (a verified
  decision whose anchored code moved ≥5 commits or ≥200 lines after the
  judgment is asked for re-affirmation — counting, never meaning),
  **in-diff trust transitions surfaced** (a PR cannot silently retire
  the decision that would flag it; the legitimate deprecate-beside-code
  reversal flow is surfaced for review, not blocked), **dead anchors get
  forwarding addresses** from git rename detection, and `-strict` now
  fails on shipped trust changes too.
- **Relayed judgments must quote the human**: `agent:` actors and every
  MCP `review_judge` call are rejected without a reason — one real quote
  per verified. A human at the terminal still signs in one keystroke.
- The ledger **rides your commits**: pre-commit stages events with the
  change that produced them (old hook shape migrates itself), and
  `.gitattributes` marks them `linguist-generated` so GitHub collapses
  them in PR diffs. Standalone ledger commits are obsolete.
- Judgment surfaces show what you sign: **anchors visible** in `review`
  and the dash; the dash also gains expandable lists ("+N more" is now a
  button), a 56-day commit-activity strip, and sha tooltips.
- Gauges: `kervo status` shows queue age and 24h judgment velocity;
  `compile` warns when the artifact passes ~12k tokens.
- The write-back protocol (three languages) requires consumers to carry
  the trust label when relaying artifact knowledge.
- Docs: SECURITY.md gains the injection threat model, supply-chain notes
  and the identity model ("signed" = audited human judgment at git trust
  level); FAQ.md answers the hardest objections; linked from all READMEs.

## v0.22.0 — 2026-07-08

- New: **decisions gate CI.** `kervo capture` takes a repeatable
  `-anchor <glob>` naming the paths a decision governs; `kervo check
  -base <ref>` warns any diff that touches what a verified decision
  anchors. In CI the output is GitHub annotations — PR-inline, zero bot
  code; advisory by default, `-strict` to block. The warning teaches the
  reversal loop (deprecate with a reason, capture the new decision), and
  only verified knowledge gates — unsigned proposals never block a
  build. MCP `kervo_capture` accepts anchors too.
- New: anchors whose paths vanish from the tree surface as stale
  candidates in `kervo check` — the first evidence-based invalidation
  channel beside the age timer.
- New: the `kervo init` result screen's found row shows AGENTS.md, so a
  codex user can see their consumer is wired (derived from what this
  init injected).
- npm: `@kervo-os/kervo` is now a real wrapper — first run downloads the
  matching GoReleaser binary (verified against the release's
  checksums.txt, cached per version); the npm version tracks the
  release tag.
- Internal: a replay benchmark and a 50k-event budget test pin the
  compaction tripwire (currently ~100 ms — the wall is far).

## v0.21.1 — 2026-07-07

- Auto-compile moves from post-commit to **pre-commit**: compiling after
  the commit left the refreshed digest uncommitted forever (each fix-up
  commit changed Recent Changes again — the tree never converged; caught
  by dogfood minutes after v0.21.0). Now each commit recompiles first
  and carries its own fresh artifact; the v0.21.0 post-commit hook is
  migrated away automatically (exact match only).

## v0.21.0 — 2026-07-07

- Freshness is no longer opt-in: every `init`/`compile` wires git
  `post-commit` and `post-merge` hooks that rerun `kervo compile`, so
  local commits and incoming pulls refresh the artifact by default —
  a teammate's first `compile` wires their machine. A hook you wrote
  yourself is never rewritten (replacing ours is the opt-out). Field
  demand: a production repo went stale under incoming pulls — the
  documented manual hook was never installed.

## v0.20.0 — 2026-07-07

- Fixed: `merge=union` is now actually wired. The README had claimed it
  since the team story shipped, but nothing wrote the `.gitattributes`
  rule — the first two-branch merge with concurrent captures ended in a
  hard conflict on the events file (reproduced in a two-clone
  experiment). `init`/`compile` now append
  `.kervo/events/*.jsonl merge=union`, committed so every clone inherits
  mergeable ledgers. Existing repos: run `kervo compile` once.
- Removed: `kervo review -web` (dash covers it), the empty sqlite
  placeholder package, and dead code from an over-engineering review
  (-330 lines, no behavior change).

## v0.1.0 – v0.19.1 — 2026-07-05 → 07-06 (condensed)

Nineteen pre-1.0 releases in two days. Per-version detail lives in git
(`git log v0.1.0..v0.19.1`) and the GitHub releases; highlights:

- **Core loop** — deterministic fact skeleton (golden-tested), trust
  lifecycle generated→observed→verified→stale→deprecated, committed JSONL
  event ledger, CLAUDE.md/AGENTS.md marker injection, monotonic ULIDs,
  en/ko/ja artifacts. (v0.1.0–v0.4.0)
- **Write-back** — artifacts end with the capture-back protocol and open
  with a session-start directive; evidence-attached proposals; the
  conversation-is-the-review flow with queue backpressure.
  (v0.3.0, v0.4.0, v0.16.0, v0.19.1)
- **Surfaces** — `kervo review` triage queue, `kervo mcp` (zero-dep stdio
  JSON-RPC), `kervo dash` fleet dashboard with knowledge view,
  visualizations, and switchable en/ko/ja chrome. (v0.2.0, v0.5.0–v0.15.0)
- **Facts** — the Brief (focus/run line/open edges), declared pytest
  runners, monorepo + Python/Docker scanning, history-only paths dropped,
  `.kervoignore`. (v0.2.1, v0.13.0, v0.17.0, v0.19.0)
- **Distribution & privacy** — GoReleaser pipeline (5 targets + Homebrew
  tap), workspace-relative paths only with a full history purge (NOTICE),
  assorted fixes (TODO comment-closer leak, dash null blank).
  (v0.1.1, v0.1.2, v0.6.1, v0.17.0)
