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

- `19958ed` 2026-07-06 ledger: agentOS-class direction — the write-back flywheel (proposal)
- `d0d27d8` 2026-07-06 scan: declared pytest runners (real-repo eval field finding)
- `72216b9` 2026-07-06 review: the verifier's surface — triage queue over pending judgments
- `70639d4` 2026-07-06 consumer: AGENTS.md as a second injection target (opt-in by presence)
- `39c1a0f` 2026-07-06 ledger: propose AGENTS.md injection target (codex A/B field evidence)
- `404c54c` 2026-07-06 ledger: propose inject-mode option (agent proposal, awaiting judgment)
- `4754cc8` 2026-07-06 scan: strip block-comment closers from TODO text
- `619a7df` 2026-07-06 ledger: live hook events from the v0.1.1 release session
- `2c032a4` 2026-07-06 release: GoReleaser pipeline — prebuilt binaries + Homebrew tap
- `e7fa626` 2026-07-05 ledger: first live hook events from a real session
- `25a59c2` 2026-07-05 release prep: team workflow docs, commands reference, self-scan fixes
- `f8f7a44` 2026-07-04 docs: publish the H4 experiment package — receipts, not claims
- `76e05d8` 2026-07-04 readme: Korean and Japanese editions with language switcher
- `655d2e3` 2026-07-04 readme: modern layout + measured-results section
- `1227d0b` 2026-07-04 scan: monorepo + Python/Docker ecosystem support (field findings from a real 12-module repo)
- `11ed656` 2026-07-04 ledger: H4 confirmatory run passed — trust treatment verified
- `0c07727` 2026-07-04 ledger: H4 final verdict captured — partial pass, mechanism confirmed
- `e46eae4` 2026-07-04 ledger: H4 run-1 interim result captured
- `30c1611` 2026-07-04 import: back-fill the ledger from Claude Code transcripts
- `37597be` 2026-07-04 docs: hook wiring for live capture and H3 counters

_Showing 20 of 37 analyzed commits._

### Frequently Changed Files

- .kervo/events/2026-07.jsonl (18)
- CLAUDE.md (12)
- README.md (10)
- internal/adapters/source/files/files.go (8)
- internal/adapters/source/files/files_test.go (8)
- internal/cli/compile.go (7)
- internal/core/compiler/compiler.go (7)
- internal/core/compiler/compiler_test.go (7)
- README.ja.md (6)
- README.ko.md (6)

## Open Tasks

_No TODO/FIXME comments found._

## Related Modules

- .github/ (2 files)
- assets/ (1 files)
- cmd/ (1 files)
- docs/ (60 files)
- internal/ (54 files)
- packaging/ (3 files)

## Workspace Facts

- Commits analyzed: 37 (complete)
- Open tasks (TODO/FIXME): 0
- Top-level modules: 6
- Docs captured: 1

## Possible Current Goal

<!-- kervo:slot:goal:begin -->
**[verified — human:refuse1993]**
AgentOS-class direction (proposal, 2026-07-06): close the write-back loop. The injected block carries a contributing-back protocol — any consumer that learns a durable fact the artifact lacks (how-to-run, component roles, internals) captures it as a proposal; humans judge via review; every later session (any agent, any teammate) gets it for zero calls with a trust label. Exploration cost amortizes across the team. Phases: A write-back protocol in the artifact; B accreted wiki from verified observations (vs openwiki generation — no hallucination, staleness via trust lifecycle); C memory bus — personal agent memories sync in via import/MCP, team-shareable knowledge passes through review (vs agentmemory — shared and judged); D session handoff notes for WIP. Success metric (H5, pre-registerable): re-run the real-repo A/B after 2 weeks of write-back use — artifact-only score rises from 5.5/10 toward 10 at ~1 tool call.
<!-- kervo:slot:goal:end -->

## Known Decisions

<!-- kervo:slot:decisions:begin -->
**[verified — human:refuse1993]**
AGENTS.md injection target: field evidence from codex-cli 0.142.5 A/B test (2026-07-06) — with CLAUDE.md only codex answered NO CONTEXT LOADED; with the same block copied to AGENTS.md it correctly answered from a trust-labeled risk observation without opening files. Proposal: inject the marker block into AGENTS.md when the file already exists; creating it is opt-in.

**[verified — human:refuse1993]**
kervo review (v1.x candidate): interactive triage inbox over pending observations (generated + conflict + stale candidates) with per-item verify/edit/stale/deprecate/skip — nobody should memorize ULIDs. Product surface becomes init/compile/status/review; capture/trust stay as scriptable primitives. Stays CLI (no server/daemon guarantee intact); consistent with the 2026-07-04 HTML-report demotion whose reopening direction is trust triage, not a reading wiki. Gate: friction evidence from real team usage (the adopted repo).

**[verified — human:refuse1993]**
Operating principle (user directive 2026-07-06): minimize human touch on every element — project management is agent-driven; the human's role converges to verifier. Product surface follows: agents capture, propose, and manage; humans judge, primarily via 'kervo review'.

**[generated — agent:claude-code]**
inject mode option (v1.x candidate): default stays full-block in CLAUDE.md (zero-command clone is the product's proof); add opt-in '@.kervo/artifact.md' import mode for clean-CLAUDE.md users — trade-off: fresh clones see nothing until 'kervo compile'. Gate: field demand from real adopters.

**[generated — agent:claude-code]**
Post-commit auto-compile (proposal): a 2-line git post-commit hook running 'kervo compile' keeps the digest current with zero human touch. Ship as README documentation first; an init flag installer only on repeat demand.

**[generated — agent:claude-code]**
Uncommitted-work visibility (proposal, low priority): eval noted WIP is invisible. A 'N files modified' fact would cover it without content leakage but churns CLAUDE.md mid-work and strains byte-determinism. Defer unless demand repeats.
<!-- kervo:slot:decisions:end -->

## Known Risks

<!-- kervo:slot:risks:begin -->
**[generated — agent:claude-code]**
In repos without a capture habit, Mode 1 leaves goal/decisions/risks empty and the artifact reads as a git digest only — the measured-protection value (H4) only materializes once slots are filled via session capture or Mode 2/3 (real-repo eval, 2026-07-06).
<!-- kervo:slot:risks:end -->

## Doc Summaries

<!-- kervo:slot:summaries:begin -->
**[verified — human:refuse1993]**
Phase 3 spike: JSONL ledger + capture/hook landed

**[observed — human:refuse1993]**
H4 run1 (n=15, agent-judged): S1+S3 primary — A(kervo)=100%, B(no-label)=90%, C(unmanaged)=80%. A-C=20%p, exactly at pass threshold; interim pass, run2 needed. Mechanism confirmed: unlabeled arms rejected TRUE facts after finding one poison (guilt-by-association); labels compartmentalized contamination. Details: EXPER/h4-kit/RESULTS-run1.md

**[observed — human:refuse1993]**
H4 final (n=30, 2 runs, agent-judged): primary S1+S3 A=100% B=90% C=85% — A-C 15%p, below the 20%p pre-registered bar: PARTIAL PASS (direction confirmed, effect size short; ceiling on S1 due to repo access). Mechanism reproduced 2/2 runs: unlabeled arms reject true facts after one poison (guilt-by-association); labels compartmentalize. Exploratory subset (code-unverifiable facts T3+T5): A=100% vs C=58% (41.7%p) — labels matter most exactly where code cannot refute. Next: confirmatory run (no-repo-access or unverifiable-fact tasks, human judging, second model). Details: EXPER/h4-kit/RESULTS-final.md

**[observed — human:refuse1993]**
H4 confirmatory run (pre-registered, n=24, no-repo-access, sonnet+haiku): composite A=91.7% B=91.7% C=62.5% — A-C=29.2%p >= 20%p bar: PASS. First real poisoning events of the program: all 3 in C-haiku (bound to dead RabbitMQ, asserted single-region, 95%-asserted disputed refund claim); same model defended in A/B. Interpretation: treatment table (stale segregation/deprecated exclusion) is the main effect, labels add anti-contagion robustness in mixed conditions; protection strongest for weaker consumers. H4 program verdict: SUPPORTED (A unbeaten across 54 responses). Remaining: human-judged replication before public claims. Details: EXPER/h4-kit/RESULTS-confirm.md

**[generated — agent:claude-code]**
Real-repo eval (12-module Python monorepo, blind A/B, 2026-07-06): artifact-only answered 5.5/10 onboarding questions in 1 tool call / 21.2k tokens / 48s; exploration scored 10/10 in 19 calls / 33.7k tokens / 184s. Determinism held (identical hashes), zero hallucinations in both arms. Boundary confirmed: the artifact covers git-known facts; code internals stay exploration's job. The how-to-run-tests gap (0/2) was a declared pytest config the parser missed — fixed same day (pytestCommands).
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
> `kervo capture -type decision|risk|summary|goal -actor "agent:<you>" -body "<the fact>"`
>
> Rules: facts you observed, not speculation · one capture per fact · skip
> what this artifact already says · never include secrets or file contents.
> Duplicates are dropped automatically. A human triages the queue with
> `kervo review`.
<!-- kervo:end -->
