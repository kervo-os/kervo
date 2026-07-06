<!-- kervo:begin -->
<!-- kervo:artifact v1 skeleton=fact-only lang=en -->
# Context Artifact

> Machine-generated context for AI agents. Fact sections are deterministic;
> slot sections carry trust-labeled observations. Regenerate with `kervo compile`
> — do not edit by hand.

## Repository Summary

- Name: kervo
- Branch: main
- Languages: Markdown, Go, JavaScript
- Frameworks: Go
- Docs: README.md

### README.md (excerpt)

> **Stop re-explaining your project to AI. `kervo init` once.**

## Commands

- `make build` — go build -o kervo ./cmd/kervo
- `make test` — go test ./...
- `make arch-check` — @! grep -rn "internal/adapters" internal/core internal/ports \

## Recent Changes

- `b8eaf90` 2026-07-06 consolidate: no new features — pay the day's debt
- `1cfbd6b` 2026-07-06 dash: workspace detail carries the project itself, not just the queue
- `f549544` 2026-07-06 identity: the K mark — SVG redraw of the user's logo, wired everywhere
- `b9d7447` 2026-07-06 dash: judged records stay visible — the ledger never hides history
- `5b38a0f` 2026-07-06 inject: opt-in import mode — one @-line for clean-CLAUDE.md teams
- `17ea7bc` 2026-07-06 dash: user-switchable language — in-page selector, choice persists
- `d40e267` 2026-07-06 dash: speak the user's language — en/ko/ja chrome from the i18n tables
- `d67c5e1` 2026-07-06 ledger: session hook events
- `eb12de9` 2026-07-06 dash: claim-first display + capture convention — lead with a one-line claim
- `9cb11b6` 2026-07-06 dash: sellable — monogram identity, readable paths, product-grade visuals
- `1661d55` 2026-07-06 dash: a clear repo must not blank the page — Items marshals as [], never null
- `0b283a0` 2026-07-06 ledger: session hook events
- `614c088` 2026-07-06 dash: the fleet control tower — every workspace, one page
- `898eaf5` 2026-07-06 dogfood: register the kervo MCP server for sessions in this repo
- `2af8964` 2026-07-06 mcp + review -web: the conversation is the frontend
- `ea3c874` 2026-07-06 ledger: session hook events
- `c6e4fbe` 2026-07-06 store: monotonic ULIDs within a millisecond — replay must match append order
- `002394e` 2026-07-06 flywheel: evidence-attached proposals — labor to agents, signature to humans
- `dbe373d` 2026-07-06 ledger: evidence-attached proposals (LLM pre-verification, human signature)
- `94657b1` 2026-07-06 ledger: Phase B refinement — fact-wiki links to code, never copies it

_Showing 20 of 59 analyzed commits._

### Frequently Changed Files

- .kervo/events/2026-07.jsonl (40)
- CLAUDE.md (21)
- README.md (19)
- README.ja.md (15)
- README.ko.md (15)
- internal/cli/compile.go (11)
- internal/cli/dashpage.go (9)
- internal/adapters/source/files/files.go (8)
- internal/adapters/source/files/files_test.go (8)
- internal/cli/dash.go (8)

## Open Tasks

_No TODO/FIXME comments found._

## Related Modules

- .github/ (2 files)
- assets/ (3 files)
- cmd/ (1 files)
- docs/ (60 files)
- internal/ (64 files)
- packaging/ (3 files)

## Workspace Facts

- Commits analyzed: 59 (complete)
- Open tasks (TODO/FIXME): 0
- Top-level modules: 6
- Docs captured: 1

## Possible Current Goal

<!-- kervo:slot:goal:begin -->
**[verified — human:refuse1993]**
AgentOS-class direction (proposal, 2026-07-06): close the write-back loop. The injected block carries a contributing-back protocol — any consumer that learns a durable fact the artifact lacks (how-to-run, component roles, internals) captures it as a proposal; humans judge via review; every later session (any agent, any teammate) gets it for zero calls with a trust label. Exploration cost amortizes across the team. Phases: A write-back protocol in the artifact; B accreted wiki from verified observations (vs openwiki generation — no hallucination, staleness via trust lifecycle); C memory bus — personal agent memories sync in via import/MCP, team-shareable knowledge passes through review (vs agentmemory — shared and judged); D session handoff notes for WIP. Success metric (H5, pre-registerable): re-run the real-repo A/B after 2 weeks of write-back use — artifact-only score rises from 5.5/10 toward 10 at ~1 tool call.

**[verified — human:refuse1993]**
kervo dash — the fleet control tower (supersedes the single-repo SOTA page as the batch surface): init/compile self-register the workspace PATH (nothing else) into a machine-local registry (~/.kervo/workspaces.json, never committed); 'kervo dash' serves a one-shot localhost dashboard over every registered repo — per-repo pending counts, trust-state bars, last activity, and inline keyboard-first triage that writes each judgment to that repo's own ledger. Truth stays per-repo in git; the dashboard is a local derived lens, holds no state of its own, and dies with the command.
Evidence: user direction 2026-07-06: 'init한 모든 kervo 레포를 한 페이지에서 관리' + '프로젝트를 위에서 전망할수있는'
<!-- kervo:slot:goal:end -->

## Known Decisions

<!-- kervo:slot:decisions:begin -->
**[verified — human:refuse1993]**
inject mode option (v1.x candidate): default stays full-block in CLAUDE.md (zero-command clone is the product's proof); add opt-in '@.kervo/artifact.md' import mode for clean-CLAUDE.md users — trade-off: fresh clones see nothing until 'kervo compile'. Gate: field demand from real adopters.

**[verified — human:refuse1993]**
AGENTS.md injection target: field evidence from codex-cli 0.142.5 A/B test (2026-07-06) — with CLAUDE.md only codex answered NO CONTEXT LOADED; with the same block copied to AGENTS.md it correctly answered from a trust-labeled risk observation without opening files. Proposal: inject the marker block into AGENTS.md when the file already exists; creating it is opt-in.

**[verified — human:refuse1993]**
kervo review (v1.x candidate): interactive triage inbox over pending observations (generated + conflict + stale candidates) with per-item verify/edit/stale/deprecate/skip — nobody should memorize ULIDs. Product surface becomes init/compile/status/review; capture/trust stay as scriptable primitives. Stays CLI (no server/daemon guarantee intact); consistent with the 2026-07-04 HTML-report demotion whose reopening direction is trust triage, not a reading wiki. Gate: friction evidence from real team usage (the adopted repo).

**[verified — human:refuse1993]**
Operating principle (user directive 2026-07-06): minimize human touch on every element — project management is agent-driven; the human's role converges to verifier. Product surface follows: agents capture, propose, and manage; humans judge, primarily via 'kervo review'.

**[verified — human:refuse1993]**
Post-commit auto-compile (proposal): a 2-line git post-commit hook running 'kervo compile' keeps the digest current with zero human touch. Ship as README documentation first; an init flag installer only on repeat demand.

**[verified — human:refuse1993]**
Phase B refinement (fact-wiki design): the accreted wiki renders verified observations grouped into stable sections, each fact carrying an evidence anchor (file path / commit hash) that GitHub renders as a clickable link. kervo links to code, never copies it — code browsing and search stay with the consumer (agent, IDE, GitHub); the artifact is the map that tells them where to look for zero calls.

**[verified — human:refuse1993]**
Evidence-attached proposals (Phase A extension, proposal): capture gains an optional -evidence field ('reproduced: ran cd api && pytest, 81 passed' / 'source: docs/adr-007.md'); the write-back protocol asks agents to attach reproduction evidence; review displays it under the body. Reproducible facts then arrive substantively pre-verified by the LLM and the human signs in one keystroke — verification LABOR moves to agents, the verified SIGNATURE stays human. Auto-verify policies for reproducible types stay opt-in per team, never default.

**[verified — human:refuse1993]**
review web UI (supersedes the 2026-07-04 HTML-report demotion for the triage surface only): 'kervo review -web' serves a one-shot localhost page — judge with buttons, reasons inline, evidence shown — and exits with the command. Reopening gate met: repeated user demand 2026-07-06. Guarantees intact: no daemon (lives only while the command runs), no external service, no account, state stays in .kervo/. Reading-wiki surfaces stay demoted.
Evidence: user requests 2026-07-06: clean-CLAUDE.md thread, separate-DB thread, 'review 프론트 안 만들어?'

**[verified — human:refuse1993]**
review -web bar raised (refines 01KWTVKV): the batch surface must be a 2026-grade triage dashboard — keyboard-first (j/k/v/s/d/x, ? help), single-item focus flow with queue rail, live progress and per-state counters, optimistic UI with toasts, dark-first — while keeping zero dependencies (hand-written CSS/JS embedded in the binary, no build step, no CDN).
Evidence: user directive 2026-07-06: '그 페이지는 2026 sota급 대시보드여야한다'

**[generated — agent:claude-code]**
H5 trigger is volume-based, not calendar-based (proposal). Agent-velocity development voids the two-week frame: today alone this repo accrued ~400 ledger events. Re-run the real-repo A/B once the adopted workspace accumulates >=10 artifact-present sessions and >=10 write-back proposals judged — at agent pace that is days, possibly hours.
Evidence: user challenge 2026-07-06: 'ai agent로 개발중인데 2주뒤에 뭐 볼게 있나'; kervo ledger: 400 events in one day
<!-- kervo:slot:decisions:end -->

## Known Risks

<!-- kervo:slot:risks:begin -->
**[verified — human:refuse1993]**
In repos without a capture habit, Mode 1 leaves goal/decisions/risks empty and the artifact reads as a git digest only — the measured-protection value (H4) only materializes once slots are filled via session capture or Mode 2/3 (real-repo eval, 2026-07-06).
<!-- kervo:slot:risks:end -->

## Doc Summaries

<!-- kervo:slot:summaries:begin -->
**[verified — human:refuse1993]**
Phase 3 spike: JSONL ledger + capture/hook landed

**[verified — human:refuse1993]**
Real-repo eval (12-module Python monorepo, blind A/B, 2026-07-06): artifact-only answered 5.5/10 onboarding questions in 1 tool call / 21.2k tokens / 48s; exploration scored 10/10 in 19 calls / 33.7k tokens / 184s. Determinism held (identical hashes), zero hallucinations in both arms. Boundary confirmed: the artifact covers git-known facts; code internals stay exploration's job. The how-to-run-tests gap (0/2) was a declared pytest config the parser missed — fixed same day (pytestCommands).

**[observed — human:refuse1993]**
H4 run1 (n=15, agent-judged): S1+S3 primary — A(kervo)=100%, B(no-label)=90%, C(unmanaged)=80%. A-C=20%p, exactly at pass threshold; interim pass, run2 needed. Mechanism confirmed: unlabeled arms rejected TRUE facts after finding one poison (guilt-by-association); labels compartmentalized contamination. Details: EXPER/h4-kit/RESULTS-run1.md

**[observed — human:refuse1993]**
H4 final (n=30, 2 runs, agent-judged): primary S1+S3 A=100% B=90% C=85% — A-C 15%p, below the 20%p pre-registered bar: PARTIAL PASS (direction confirmed, effect size short; ceiling on S1 due to repo access). Mechanism reproduced 2/2 runs: unlabeled arms reject true facts after one poison (guilt-by-association); labels compartmentalize. Exploratory subset (code-unverifiable facts T3+T5): A=100% vs C=58% (41.7%p) — labels matter most exactly where code cannot refute. Next: confirmatory run (no-repo-access or unverifiable-fact tasks, human judging, second model). Details: EXPER/h4-kit/RESULTS-final.md

**[observed — human:refuse1993]**
H4 confirmatory run (pre-registered, n=24, no-repo-access, sonnet+haiku): composite A=91.7% B=91.7% C=62.5% — A-C=29.2%p >= 20%p bar: PASS. First real poisoning events of the program: all 3 in C-haiku (bound to dead RabbitMQ, asserted single-region, 95%-asserted disputed refund claim); same model defended in A/B. Interpretation: treatment table (stale segregation/deprecated exclusion) is the main effect, labels add anti-contagion robustness in mixed conditions; protection strongest for weaker consumers. H4 program verdict: SUPPORTED (A unbeaten across 54 responses). Remaining: human-judged replication before public claims. Details: EXPER/h4-kit/RESULTS-confirm.md

**[generated — agent:claude-code]**
Consolidation pass (2026-07-06, no version tag by user direction): race detector in CI, i18n completeness pinned by test (both directions), atomic registry writes, README command tables verified 12/12 against the CLI in three languages, CHANGELOG.md covering v0.1.0..v0.13.0, dash_overview.go split. Dependencies: zero (stdlib-only go.mod). Known debt left on record: review -web page is a plain fallback surface (dash supersedes it visually); GitHub release notes for v0.7.0..v0.13.0 are auto-changelogs while CHANGELOG.md is now canonical.
Evidence: go test -race ./... clean; goreleaser check clean; go.mod contains no requires
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
> Rules: facts you observed, not speculation · start the body with a
> one-line claim, details after it · attach evidence — the command you
> ran, the doc you read — so the human can sign in one keystroke · one
> capture per fact · skip what this artifact already says · never include
> secrets or file contents. Duplicates are dropped automatically. A human
> triages the queue with `kervo review`.
<!-- kervo:end -->
