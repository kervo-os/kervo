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

## Brief

- **Focus**: internal/ ×10
- **Run**: `make build` · `make test` · `make arch-check`

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

- `7f31a3d` 2026-07-07 community: CONTRIBUTING + SECURITY (private vuln reporting enabled)
- `8d42959` 2026-07-07 changelog: condense 0.1–0.19 + ci: least privilege, SHA-pinned actions
- `1189aef` 2026-07-07 ledger: blockchain verdict + H5 anchored registration
- `29846fc` 2026-07-07 test: drop an import the assertion rewrite left behind
- `9a2d42d` 2026-07-07 autocompile: pre-commit, not post-commit — the tree must converge
- `0ec6742` 2026-07-07 ledger: session hook events
- `6f4b69c` 2026-07-07 ledger: session hook events
- `2bc2d3c` 2026-07-07 freshness is the default — compile wires the git hooks itself
- `1c4fe44` 2026-07-07 init: the wizard keeps the digest fresh — commits and pulls auto-compile
- `fe96a81` 2026-07-07 ledger: session hook events
- `869a4f3` 2026-07-07 changelog: v0.20.0
- `d3fb51d` 2026-07-07 ledger: session hook events
- `9370eb4` 2026-07-07 ledger merges were a documented lie — now they union
- `d9a3f24` 2026-07-07 ponytail: -323 lines, zero behavior change
- `01064e6` 2026-07-06 readme: the modern OSS shape — short spine, folded depth
- `68c3f9d` 2026-07-06 readme: a proper OSS tail — contributing in, diary out
- `bd8e659` 2026-07-06 readme: the status section catches up with the product
- `76705c8` 2026-07-06 changelog: v0.19.1
- `2c493b9` 2026-07-06 protocol: govern the session start, not only the end
- `219a0f5` 2026-07-06 changelog: v0.19.0

_Showing 20 of 111 analyzed commits._

### Frequently Changed Files

- .kervo/events/2026-07.jsonl (82)
- CLAUDE.md (54)
- README.md (36)
- README.ja.md (32)
- README.ko.md (32)
- internal/core/i18n/i18n.go (16)
- CHANGELOG.md (15)
- internal/cli/compile.go (15)
- internal/cli/dash.go (14)
- internal/cli/dashpage.go (13)

## Open Tasks

_No TODO/FIXME comments found._

## Related Modules

- .github/ (2 files)
- assets/ (7 files)
- cmd/ (1 files)
- internal/ (67 files)
- packaging/ (3 files)

## Workspace Facts

- Commits analyzed: 111 (complete)
- Open tasks (TODO/FIXME): 0
- Top-level modules: 5
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

**[verified — human:refuse1993]**
H5 trigger is volume-based, not calendar-based (proposal). Agent-velocity development voids the two-week frame: today alone this repo accrued ~400 ledger events. Re-run the real-repo A/B once the adopted workspace accumulates >=10 artifact-present sessions and >=10 write-back proposals judged — at agent pace that is days, possibly hours.
Evidence: user challenge 2026-07-06: 'ai agent로 개발중인데 2주뒤에 뭐 볼게 있나'; kervo ledger: 400 events in one day

**[verified — human:refuse1993]**
Mode 3 is a bootstrap channel, not a running mate to Mode 2 (field eval on the adopted real repo, 2026-07-06). With only the artifact as input (commit messages + TODO list), a local 120b produced a rear-view goal (inferred from pushed history; the real goal lived in an unpushed branch — stale, not hallucinated) and an overreaching risk (18 TODOs exist -> 'critical logic is empty'; evidence does not support the conclusion). Session-verified Mode 2 capture measured 9.5/10 the same day. The trust gate worked as designed: both weak proposals arrived as [generated] and were quarantined pending judgment. Positioning: run Mode 3 to fill empty slots at cold start; once Mode 2 capture is live, leave KERVO_SEMANTIC_URL unset (auto Mode 1) — artifact-only inference reads history, not intent, and adds review noise.
Evidence: user-relayed eval 2026-07-06: goal C+ (rear-view), risk D (fact->overreach), 9.5s local 120b, 2 proposals; Mode 2 pilot 9.5/10 same day

**[verified — human:refuse1993]**
External producers integrate via the published intake contract, not per-tool importers (refines Phase C of goal 01KWTJGS). Any tool - graph builders, memory stores, wiki generators - stages .kervo/proposals.json entries [{slot, body, source}]; compile ingests them as generated with provenance, review gates them, and the absolute rule holds: no state field exists, so nothing self-promotes. Per-tool importers (graphify/openwiki/agentmemory) are built only when a real export file exists on this machine to test against - chasing N third-party formats is the losing game the scope rule forbids. Known gap for the first real producer: the proposal shape lacks an evidence field; add it when the first producer arrives. State nuance kept from PRD 7.2: your own session history imports may enter observed; generated-content tools always enter generated.
Evidence: consumer/proposals.go: proposal{Slot,Body,Source}, no state field, RFC-0003 5; user-relayed integration proposal 2026-07-06

**[verified — human:refuse1993]**
llm-wiki-newsroom is the first concrete producer candidate - and it needs ZERO kervo code (refines 01KWVFE1). It is agent-driven (Claude Code, CLAUDE.md-operated), so its own agent can stage .kervo/proposals.json per the published contract; no 'kervo import wiki-newsroom' importer, no format chase. Mapping when adopted: contradictions -> risks (our conflict analog), cluster overviews / entities -> summaries, timelines -> summaries; never decisions (it does not produce team decisions). Everything enters [generated - llm-wiki-newsroom]. Philosophy check: it accretes at ingest (not compile-time regeneration) and separates author from reviewer - closer to kervo than the openwiki pattern; the boundary stays: kervo absorbs judgments about its output, never its graph/wiki machinery. Adoption gate: real usage in the report-archive workflow opens the first-producer items (evidence field in proposals.json).
Evidence: README direct read 2026-07-06: 5-role newsroom, self-evolving guidelines w/ regression A/B, L2 sub-layers, /wiki-lint staleness+contradictions, WIKI_LANG=ko, MIT

**[verified — human:refuse1993]**
Phase B MVP ships as the dash knowledge view (goal 01KWTJGS phase B, refinement 01KWTKGV). The accreted wiki is a READING of the trust ledger, not a generator: verified and observed entries render in full, claim-first, evidence attached, grouped by type in stable order; stale and deprecated stay visible under retired with their reasons. Deterministic markdown export (kervo wiki -write) is deferred until someone needs the wiki outside the dash - committing derived files needs a deliberate gesture, not a default.
Evidence: dash knowledge view shipped 2026-07-06; screenshot-verified on this repo's 16 verified entries

**[verified — human:refuse1993]**
Workspace-native wiki management (plan for judgment; reframe by user 2026-07-06: kervo manages the workspace, and a wiki inside it is workspace surface to manage in place - not a corpus to copy into the ledger). Stage W1, fact scan: detect a wiki/ markdown tree and state it in the artifact and dash as facts - page counts per sub-tree, last change - declared-only (the directory structure is the declaration), deterministic, zero content copying; every session learns the wiki's existence and map for zero calls, which makes 'conclusions, not corpus' natural: the corpus is pointable. Gate: a real workspace grows a wiki (newsroom adoption). Stage W2, convention only: knowledge entries cite wiki pages as evidence anchors - already expressible, document it. Stage W3, page-level trust states (deferred, needs real design): signing a page raises the invalidation question - a signature should probably bind to a content hash, not a path - so this stage waits for demand and gets planned properly, not improvised. Boundary unchanged: wiki lint/graph stay the producer's job; kervo adds existence, scale, freshness, and the trust layer.
Evidence: user reframe 2026-07-06: '우리는 워크스페이스 기준으로 관리하는 역할. 그 워크스페이스에 위키가 있으면 그것까지 관리하는 것'

**[verified — human:refuse1993]**
The conversation IS a review surface (plan for judgment; user challenge 2026-07-06). Formalize what this project already practices: when a human explicitly affirms a specific claim in session, the agent records capture AND relays the verification in one motion - reason quotes the affirmation as evidence - and no queue round-trip happens. The queue does not disappear; it shrinks to its true purpose: knowledge no human has seen (unattended write-backs, external producers, other agents' sessions, Mode 3). Three boundaries hold: (1) affirmation is claim-scoped, not vibe-scoped - an OK covers what was actually put to the human; anything learned but not discussed stays generated; (2) provenance must say HOW it was verified ('user approved in session' + quote), so signatures stay auditable; (3) the self-signature ban is untouched - the agent relays the human's stated judgment, never its own. Conflict flow, protocol-level: a consumer that finds evidence contradicting a verified entry raises it with the human in conversation, then records the updated judgment (deprecate old with reason + capture new, or re-affirm) - conflicts become questions, not queue sediment. Implementation when judged: protocol text for the injected block and MCP tool descriptions; a capture sugar flag (one command for capture+relay) only if the two-command pattern shows friction.
Evidence: user 2026-07-06: '사람이 OK 했으니까 진행된거면 그것 자체가 검증. 과거랑 다르면 다시 물어보고 팩트를 바꾸면'; this session's own practice (capture then trust with chat-quote reasons) as the working precedent

**[verified — human:refuse1993]**
README lead adopts the signed-memory positioning (user-directed 2026-07-06, ahead of the kervo.dev landing draft): the OK you give an agent becomes the team's signed memory; any agent opening the workspace starts knowing what is true, what was decided, and what not to trust yet - and that memory grows with every session. Value sentence leads, mechanism paragraph follows; hero imperative (Stop re-explaining) stays as the hook.
Evidence: user 2026-07-06: '그 문장으로 kervo.dev 랜딩 초안 만들기전에 readme부터'

**[verified — human:refuse1993]**
kervo.dev purchased (2026-07-06, $12.20/yr). Landing launch stays deferred until the H5 full re-run numbers exist; interim plan is a redirect to the GitHub repo. The GoReleaser tap author email noreply@kervo.dev is now legitimate.
Evidence: user 2026-07-06: '일단 샀어'

**[verified — human:refuse1993]**
Hero art adopted (user-regenerated 2026-07-06 with the exact strings requested: 'kervo compile', real pipeline step names, kervo:begin/end markers). assets/hero-story.jpg leads the READMEs right after the positioning sentence; the mermaid loop moved into How it works; both dash screenshots sit in the dash section. assets/loop-poster.jpg (the full-system poster with 'Agents discover and propose. Humans judge once.') is reserved for the kervo.dev landing kit - three raster images at the top of a README is too heavy. The pixel cat is the story/illustration brand element; the K mark stays the product/favicon element.
Evidence: user-supplied regenerated images 2026-07-06 21:45/21:47; JPEG 339KB/396KB after texture-aware conversion

**[verified — human:refuse1993]**
Evidence lives in its own repo: kervo-os/experiments (h4 moved there with provenance kervo@f8f7a44; H5 lands there too). The product repo's docs/ was 60/60 experiment files - research corpora polluted the artifact's module facts and once injected fake TODOs. Three-repo structure now: kervo (product) / experiments (evidence) / kervo.dev (landing, later); READMEs link the receipts one click away. .kervoignore file removed from this repo (nothing left to exclude; the feature stays documented and tested).
Evidence: user 2026-07-06: '그렇게 하자'; docs/experiments was 516KB, 60 files = 100% of docs/

**[verified — human:refuse1993]**
History purge executed before public promotion (user-approved sequence: land parallel work, purge, release). git filter-repo rewrote all blobs across the full history - machine paths relativized, a neighbor repo's name replaced - then force-pushed with all 24+ tags; verification greps across every rev: zero occurrences. The ledger itself was KEPT: deleting it would not have removed history leaks and would have destroyed the dogfood record. Local clones require re-clone or hard reset; release assets and brew unaffected. v0.17.0 ships the Brief, consumer choice, and the privacy hardening together.
Evidence: user 2026-07-06: 'ㅇㅋ 그 순서로 가자'; post-purge grep across git rev-list --all: 0 hits for both strings

**[verified — human:refuse1993]**
Protocol gains an opening directive (proposal): the injected block should instruct consumers at the START, not only at the end - 'Begin from this artifact: state what it already answers for your task, then explore only the gaps and verify only what your task puts in question.' Field evidence: a codex session asked to analyze a freshly-initialized repo re-derived declared facts (compose topology, TODOs, pytest presence) without referencing the loaded block; exploration itself was legitimate for an audit task, but the zero-acknowledgment start wastes the paid context and hides the product from the user watching. Companion observation: the same session exposed a thin declared fact - root pytest fails without PYTHONPATH=. - which is precisely a write-back candidate; whether the session captures it at the end is a live protocol-compliance datum for a non-hooked consumer.
Evidence: user-pasted codex transcript 2026-07-06: rg/ls/README/compose/pytest exploration with no artifact reference; PYTHONPATH=. discovery, 22 pass / 23 skip

**[verified — human:refuse1993]**
ports/ stays despite zero consumers — kept by judgment, not oversight.
- An over-engineering review found all 4 port interfaces have no consuming caller (assertions only); deletion was proposed and declined: the file is the adapter contract documentation and the hexagonal boundary's physical evidence (RULE: 4 ports, ceiling 6).
- External producers never touch ports — they enter via the proposals.json file contract, so 'future openwiki integration' is not a reason to keep or grow it.
- The rest of the review landed: review -web removed (dash + terminal review cover both cases), empty sqlite placeholder removed, stdlib/shrink fixes — net -323 lines, no behavior change.
Evidence: grep: every ports.* reference outside internal/ports is a 'var _' assertion; go test -race + arch-check green after removal commit

**[verified — human:refuse1993]**
Artifact freshness is default-on, not opt-in — every init/compile wires git post-commit + post-merge auto-compile hooks.
- A memory layer for teams that store work as commits must watch commits by default; a stale artifact silently breaks the one promise.
- post-merge matters as much as post-commit: pulls are how teammates' commits arrive.
- Safety contract unchanged: create if absent, recognize our own, never rewrite a foreign hook — replacing ours with your own IS the opt-out. Hooks are machine-local, so the first compile on any machine wires that machine.
- The wizard question shipped hours earlier is deleted: freshness is plumbing (same rank as .gitignore/.gitattributes), not a preference.
Evidence: user directive 2026-07-07: '자동으로 깃 안볼꺼면 왜 쓰냐 이거?'; field case: a production repo pulled a week of commits with the artifact stale

**[verified — human:refuse1993]**
Blockchain verdict: no chain — but its two useful neighbors are adopted.
- Chain consensus is orthogonal (kervo is verifier-centric, not trustless) and conflicts with zero-dep/no-network/privacy guarantees — the history purge precedent proves mutability escape hatches are required.
- Adopted (a): H5 pre-registration hashes anchored via OpenTimestamps at registration (h5/PREREG.md + GATE.md + gate.py, .ots proofs committed) — 'the protocol predates the results' is now third-party verifiable.
- Adopted (b), deferred to W3/enterprise demand: judgment-commit signing with signatures bound to content hashes.
- H5 PREREG declares: pass = mean >=8.5/10 at <=2 tool calls on the original 5-question basis; reject <=6.5; gate counts pinned by hash; pilot's targeted captures declared as design, not leakage.
Evidence: user judgment 2026-07-07: 'ㅇㅋ 그 선으로 가자, H5 해시 앵커링 진행해'; ots stamp accepted by 4 calendars; experiments@main h5/

**[verified — human:refuse1993]**
kervo.dev landing: libraries allowed — the zero-dep guarantee binds the shipped binary and its embedded UIs, not the website. Stack: Astro 5 + Tailwind 4 static site at EXPER/kervo.dev; hosting: Vercel (project kervo.dev, prod live at kervodev.vercel.app, domain kervo.dev attached 2026-07-07); DNS pending one Cloudflare A record (76.76.21.21). Measured section carries only public H4 numbers — H5 headline still waits for the re-run. Supersedes 01KWXGDYPA (GitHub Pages detail replaced by Vercel per user direction).
Evidence: user 2026-07-07: '라이브러리 써도 되는데?' + '좋아. vercel에 올리고 내 도메인 붙일까?'; vercel inspect: alias kervodev.vercel.app, curl 200
<!-- kervo:slot:decisions:end -->

## Known Risks

<!-- kervo:slot:risks:begin -->
**[verified — human:refuse1993]**
In repos without a capture habit, Mode 1 leaves goal/decisions/risks empty and the artifact reads as a git digest only — the measured-protection value (H4) only materializes once slots are filled via session capture or Mode 2/3 (real-repo eval, 2026-07-06).

**[verified — human:refuse1993]**
merge=union was documented but never wired — fixed 2026-07-07.
- README (3 languages), event.go, and jsonl.go all claimed branch merges union the ledgers, but no .gitattributes existed and no code wrote one: the first concurrent-capture team merge would hit a hard conflict on .kervo/events/*.jsonl.
- Two-clone experiment reproduced it: default driver -> CONFLICT; with 'merge=union' -> clean union, line order != ULID order (replay sorts by ID, so folds stay correct).
- Fix: init/compile now append '.kervo/events/*.jsonl merge=union' to .gitattributes (ensureGitattributes — append-only, idempotent, human rules preserved). Committed alongside the ledger so every clone inherits it.
- Companion fact for the projection-contract work: our own single-machine ledger already has 9 line-order/ULID-order inversions in 631 events — the ledger is a merged partial order even before teams.
Evidence: git check-attr merge .kervo/events/2026-07.jsonl -> unspecified (pre-fix); scratch-repo merge experiment: CONFLICT without attr, union with attr; python scan: 631 events, 9 out-of-order boundaries
<!-- kervo:slot:risks:end -->

## Doc Summaries

<!-- kervo:slot:summaries:begin -->
**[verified — human:refuse1993]**
Phase 3 spike: JSONL ledger + capture/hook landed

**[verified — human:refuse1993]**
Real-repo eval (12-module Python monorepo, blind A/B, 2026-07-06): artifact-only answered 5.5/10 onboarding questions in 1 tool call / 21.2k tokens / 48s; exploration scored 10/10 in 19 calls / 33.7k tokens / 184s. Determinism held (identical hashes), zero hallucinations in both arms. Boundary confirmed: the artifact covers git-known facts; code internals stay exploration's job. The how-to-run-tests gap (0/2) was a declared pytest config the parser missed — fixed same day (pytestCommands).

**[verified — human:refuse1993]**
Consolidation pass (2026-07-06, no version tag by user direction): race detector in CI, i18n completeness pinned by test (both directions), atomic registry writes, README command tables verified 12/12 against the CLI in three languages, CHANGELOG.md covering v0.1.0..v0.13.0, dash_overview.go split. Dependencies: zero (stdlib-only go.mod). Known debt left on record: review -web page is a plain fallback surface (dash supersedes it visually); GitHub release notes for v0.7.0..v0.13.0 are auto-changelogs while CHANGELOG.md is now canonical.
Evidence: go test -race ./... clean; goreleaser check clean; go.mod contains no requires

**[verified — human:refuse1993]**
Write-back pilot on the adopted real repo (2026-07-06, third-party agent session): two previously unanswerable questions went 0/2 to 2/2 after targeted captures flowed capture -> ledger -> compile -> slots -> a fresh consumer; same 5-question basis moved 5.5/10 to 9.5/10 at unchanged cost (1 tool call, 48s to 21s). Trust labels reached the consumer and changed its behavior: it flagged its own answer as [generated], not human-signed. The declared-pytest scan (v0.2.1/v0.13.0) fired on the real repo. Caveat kept honest: this is a mechanism pilot on 2 targeted questions, not the pre-registered full A/B re-run with blinded judging - that still runs at the volume gate.
Evidence: user-pasted eval table 2026-07-06: Q3 0/2->2/2, Q4 0/2->2/2, 1 call, 48s->21s; consumer quote flagging [generated] vs Verified

**[observed — human:refuse1993]**
H4 run1 (n=15, agent-judged): S1+S3 primary — A(kervo)=100%, B(no-label)=90%, C(unmanaged)=80%. A-C=20%p, exactly at pass threshold; interim pass, run2 needed. Mechanism confirmed: unlabeled arms rejected TRUE facts after finding one poison (guilt-by-association); labels compartmentalized contamination. Details: EXPER/h4-kit/RESULTS-run1.md

**[observed — human:refuse1993]**
H4 final (n=30, 2 runs, agent-judged): primary S1+S3 A=100% B=90% C=85% — A-C 15%p, below the 20%p pre-registered bar: PARTIAL PASS (direction confirmed, effect size short; ceiling on S1 due to repo access). Mechanism reproduced 2/2 runs: unlabeled arms reject true facts after one poison (guilt-by-association); labels compartmentalize. Exploratory subset (code-unverifiable facts T3+T5): A=100% vs C=58% (41.7%p) — labels matter most exactly where code cannot refute. Next: confirmatory run (no-repo-access or unverifiable-fact tasks, human judging, second model). Details: EXPER/h4-kit/RESULTS-final.md

**[observed — human:refuse1993]**
H4 confirmatory run (pre-registered, n=24, no-repo-access, sonnet+haiku): composite A=91.7% B=91.7% C=62.5% — A-C=29.2%p >= 20%p bar: PASS. First real poisoning events of the program: all 3 in C-haiku (bound to dead RabbitMQ, asserted single-region, 95%-asserted disputed refund claim); same model defended in A/B. Interpretation: treatment table (stale segregation/deprecated exclusion) is the main effect, labels add anti-contagion robustness in mixed conditions; protection strongest for weaker consumers. H4 program verdict: SUPPORTED (A unbeaten across 54 responses). Remaining: human-judged replication before public claims. Details: EXPER/h4-kit/RESULTS-confirm.md
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
