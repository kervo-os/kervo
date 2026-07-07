# Changelog

All notable changes, newest first. Versions are git tags; every release
ships prebuilt binaries and a Homebrew cask (`brew install kervo-os/tap/kervo`).

## Unreleased

- `kervo init` wires git auto-compile: a third wizard question installs
  `post-commit` and `post-merge` hooks (`-autocompile yes|no` in
  scripts), so local commits and incoming pulls refresh the artifact
  without anyone remembering to run `compile`. Foreign hooks are never
  rewritten. Field demand: a production repo went stale under incoming
  pulls — the documented manual hook was never installed.

## v0.20.0 — 2026-07-07

- Fixed: `merge=union` is now actually wired. The README had claimed it
  since the team story shipped, but nothing wrote the `.gitattributes`
  rule — the first two-branch merge with concurrent captures ended in a
  hard conflict on the events file (reproduced in a two-clone
  experiment). `init`/`compile` now append
  `.kervo/events/*.jsonl merge=union`, committed so every clone inherits
  mergeable ledgers. Existing repos: run `kervo compile` once.
- Removed: `kervo review -web`. The dash judges every repo with the same
  ledger primitive and a better surface; the terminal `kervo review`
  covers the no-browser case. One judging page, not two.
- Removed: the empty sqlite index placeholder package and the
  `.kervo/index.db` gitignore rule — no code ever created that file.
  Existing `.gitignore` lines are harmless and can be deleted.
- Internal: dead-code and duplication cleanup from an over-engineering
  review (-330 lines, no behavior change).

## v0.19.1 — 2026-07-06

- The artifact opens with a session-start directive: start here, state
  what it already answers, explore only the gaps.

## v0.19.0 — 2026-07-06

- Facts describe the repo that exists: Focus, coupling, and hot files
  drop paths that only history knows (repos extracted from a parent
  directory carry the old prefix in pre-split commits).
- Observation bodies read like documents: the dash renders a minimal
  markdown subset (XSS-impossible by construction), literal backslash-n
  from shell-quoted captures displays as real newlines, and the protocol
  now asks for markdown bodies.

## v0.18.0 — 2026-07-06

- Interactive init completes the wizard: after the consumer question it
  offers to wire Claude Code capture hooks (`-hooks yes|no` for scripts) —
  created once, never rewriting an existing settings file. Codex-only
  choices get a stated fact instead of a fake: no per-repo hook system
  exists there; AGENTS.md carries the write-back protocol regardless.

## v0.17.0 — 2026-07-06

- The Brief: artifacts open with deterministic orientation — where recent
  commits concentrate (scopes + modules), one run line, open edges,
  unpushed count. Zero LLM; empty signals render nothing. The init screen
  gains the Focus row.
- `-consumers claude|codex|both|auto`: pick injection targets; interactive
  init asks once and the choice persists in `.kervo/consumers`.
- Privacy: hook and import store workspace-relative paths only (outside
  the workspace → basename); the committed ledger was scrubbed and the
  full git history rewritten to carry no machine paths or neighbor-repo
  names. NOTICE added.

## v0.16.0 — 2026-07-06

- The conversation is the review: the write-back protocol and MCP tools now
  instruct agents to relay in-session affirmations with the capture, quoting
  the human's words — the queue holds only what no human has seen. Conflicts
  with verified entries become questions to the human, not re-proposals.
- Queue guardrails: "conclusions, not corpus" joins the protocol, and
  capture applies backpressure — a source with 12 unjudged proposals must
  seek judgment before proposing more (humans are never throttled).

## v0.15.0 — 2026-07-06

- The knowledge view: workspace detail renders every verified and observed
  entry in full — claim first, evidence attached, grouped by type — and
  retired entries keep their reasons. The accreted wiki, phase B MVP:
  judged knowledge is the wiki; nothing is generated.

## v0.14.0 — 2026-07-06

- Dash visualization layer: 28-day activity sparklines, trust-state donut,
  commit-proven coupling ring, and a CONNECTED panel showing which adapters
  actually feed and consume each workspace. All inline SVG, zero deps.
- README overhaul in three languages: the loop as a diagram, dashboard
  screenshots, and the real-repo measurements (write-back 5.5→9.5,
  Mode 3 grading, labels reaching consumers).

## v0.13.1 — 2026-07-06

- Judgment semantics in the dash: a hint line under the actions and a
  "what a judgment does" help section — verify is a reversible signature,
  deprecate records why, skip stays harmlessly unverified.

- Consolidation pass: CI runs the race detector; i18n tables are pinned
  complete by test; the workspace registry writes atomically; command
  tables in all three READMEs match the CLI exactly; this changelog.

## v0.13.0 — 2026-07-06

- `kervo dash` workspace detail gains a project overview: branch, languages,
  declared commands, recent changes, open tasks, modules — the same
  deterministic scan `compile` runs, capped for reading.
- Module coupling from commit history ("change together") — connections
  proven by commits, not narrated by a model. Dot-dirs excluded.

## v0.12.0 — 2026-07-06

- The K mark: `assets/logo.svg` (full lockup, background-agnostic via a
  transparent ring punch) and `assets/mark.svg` (small tile, legible at
  16px). README heroes, dash header, and the dash favicon use it.

## v0.11.0 — 2026-07-06

- Dash judged-records rail: every past judgment stays visible with its
  state and reason — the page hides no history.

## v0.10.0 — 2026-07-06

- `-inject import`: opt-in one-line `@.kervo/artifact.md` block for
  clean-CLAUDE.md teams; choice persists in `.kervo/inject` (committed).
  Full block stays the default.
- READMEs document the git post-commit auto-compile hook.

## v0.9.0 — 2026-07-06

- Dash language is user-switchable in the page; the choice persists in
  `~/.kervo/ui-lang` and outranks `$LANG` on the next launch.

## v0.8.0 — 2026-07-06

- Dash chrome speaks en/ko/ja from the shared i18n tables ($LANG detection,
  `-lang` override). Trust-state and type tokens stay English everywhere —
  they are ledger vocabulary.

## v0.7.0 — 2026-07-06

- Dash redesign: per-repo monogram, headline pending count, home-shortened
  paths, per-state legend, activity pulse, deep links (`#N`).
- Claim-first display: the write-back protocol asks for a one-line claim
  first; the triage card renders it as the headline.

## v0.6.1 — 2026-07-06

- Fix: a workspace with nothing pending marshaled `Items` as `null` and
  blanked the dash. Now always `[]`, with a regression test.

## v0.6.0 — 2026-07-06

- `kervo dash`: one-shot 127.0.0.1 dashboard over every registered
  workspace — pending judgments, trust-state bars, inline keyboard triage
  routed to each repo's own ledger. `compile` self-registers workspace
  paths (path only) in `~/.kervo/workspaces.json`.

## v0.5.0 — 2026-07-06

- `kervo mcp` is real: zero-dependency stdio JSON-RPC server with
  `read_context`, `kervo_capture`, `review_queue`, `review_judge` —
  judging happens in the chat, the agent only relays the human's decision.
- `kervo review -web`: one-shot local page for batch triage.

## v0.4.0 — 2026-07-06

- Evidence-attached proposals: `capture -evidence` rides the ledger, shows
  in review, renders with the claim. Verification labor moves to agents;
  the verified signature stays human.
- Fix: ULIDs are monotonic within a millisecond, so replay order matches
  append order for agent-speed capture→trust→capture sequences.

## v0.3.0 — 2026-07-06

- The write-back protocol: every artifact ends with instructions for AI
  consumers to capture durable facts it lacks as proposals. Duplicate
  live bodies are dropped; re-assertion after stale/deprecated is allowed.

## v0.2.1 — 2026-07-06

- Declared pytest runners (`pytest.ini` / `[tool.pytest.ini_options]`)
  surface as commands — found by a real-repo eval.

## v0.2.0 — 2026-07-06

- `kervo review`: triage queue over generated proposals and ⚠ conflicts,
  no IDs to memorize. Judgments are the same transition events `trust`
  writes.
- AGENTS.md injection target (opt-in by file presence) — Codex and other
  AGENTS.md readers get the same block.

## v0.1.2 — 2026-07-06

- Fix: block-comment closers (`-->`, `*/`) no longer leak into TODO text.

## v0.1.1 — 2026-07-06

- GoReleaser pipeline: prebuilt binaries for five targets and the
  Homebrew tap.

## v0.1.0 — 2026-07-05

- First public release: deterministic fact skeleton (golden-tested),
  trust-labeled slots with the generated→observed→verified→stale→deprecated
  lifecycle, committed JSONL event ledger (`merge=union`), CLAUDE.md marker
  injection, capture/trust/status/metrics/import/hook, en/ko/ja artifacts,
  monorepo + Python/Docker ecosystem scanning, `.kervoignore`.
