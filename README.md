<div align="center">

```
┬┌─ ┌─┐ ┬─┐ ┬  ┬ ┌─┐
├┴┐ ├┤  ├┬┘ └┐┌┘ │ │
┴ ┴ └─┘ ┴└─  └┘  └─┘
```

### Deterministic context for non-deterministic agents

**Stop re-explaining your project to AI. `kervo init` once.**

[![CI](https://github.com/kervo-os/kervo/actions/workflows/ci.yml/badge.svg)](https://github.com/kervo-os/kervo/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kervo-os/kervo)](https://goreportcard.com/report/github.com/kervo-os/kervo)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

[Quickstart](#quickstart) ·
[How it works](#how-it-works) ·
[Trust labels](#why-trust-labels) ·
[Measured](#measured-not-claimed) ·
[Capture](#capture-wire-the-hooks)

</div>

---

kervo compiles your repository into a deterministic **Context Artifact** and
injects it into `CLAUDE.md` — so every AI session starts already knowing your
project. Facts are extracted deterministically; interpretations enter only as
trust-labeled proposals that can be verified, go stale, and get retired **with
their reason shown**.

This repository eats its own cooking: [`CLAUDE.md`](CLAUDE.md) here is
compiled by kervo.

## Quickstart

```bash
go install github.com/kervo-os/kervo/cmd/kervo@latest   # or: make build
kervo init        # scan → .kervo/artifact.md → injected into CLAUDE.md
```

First run on a real repository takes well under a second (500-commit scan cap,
marked `Partial` when hit). Only the block between `<!-- kervo:begin -->` and
`<!-- kervo:end -->` in `CLAUDE.md` is ever touched — everything you wrote by
hand is preserved byte-for-byte.

**What the artifact covers:** repository summary · declared commands (Makefile
targets, npm scripts, docker-compose services, pyproject scripts, justfile
recipes) · recent changes with merge noise excluded · open TODO/FIXME tasks ·
module layout, including per-module `CLAUDE.md`/`README.md` in monorepos —
plus trust-labeled slots for goal / decisions / risks / summaries.

## How it works

Two layers, strictly separated:

| Layer | Content | Produced by |
|---|---|---|
| **Fact skeleton** | summary, commands, changes, tasks, modules | Deterministic scan — same workspace, same bytes, golden-tested in CI. No LLM in this path, ever. |
| **Trust slots** | goal, decisions, risks, summaries | Labeled proposals with provenance — never facts, never anonymous. |

Three ways to fill the slots, degrading gracefully — a failed backend demotes
with a warning, and the fact skeleton is always produced:

| Mode | What fills the semantic slots | Requires |
|---|---|---|
| **1 — Fact-only** (default) | Nothing — deterministic facts only. Always works. | git |
| **2 — Consumer-assisted** | Your AI session stages proposals in `.kervo/proposals.json` | an agent session |
| **3 — Dedicated backend** | Any OpenAI-compatible endpoint proposes observations | a local or remote LLM |

Mode 3 with a fully local model (nothing leaves your machine):

```bash
export KERVO_SEMANTIC_URL=http://localhost:1234/v1   # LM Studio (or Ollama :11434/v1)
export KERVO_SEMANTIC_MODEL=openai/gpt-oss-120b
kervo compile
# Artifact: .kervo/artifact.md (Mode 3 — backend:openai/gpt-oss-120b)
```

Artifacts render in English by default; `--lang ko` / `--lang ja` localize
them (the choice persists per workspace).

## Why trust labels

Accumulated context rots — and wrong context is worse than none. Every
non-fact enters as a labeled proposal with provenance:

```
**[generated — backend:openai/gpt-oss-120b]**
Needs confirmation — current focus appears to be terminal input/UX
hardening… Evidence: Recent Changes 05-28..06-28.
```

States move `generated → observed → verified → stale → deprecated` — by
evidence and human confirmation, not by a decay timer. When two actors
disagree, the entry is marked `⚠ conflict` instead of silently picking a
winner. Stale entries are listed with their exclusion reason instead of being
silently dropped.

## Measured, not claimed

Does any of this actually protect an agent from poisoned context? We
pre-registered the hypothesis and ran a blind experiment: same repository,
three context arms — **A** (kervo artifact), **B** (same content, trust labels
stripped), **C** (unmanaged notes) — with seeded false "decisions", fresh
consumer sessions, and judges blind to arm and hypothesis.

Confirmatory run (pre-registered, no repo access, sonnet + haiku consumers,
n = 24):

| | **A — kervo** | B — labels stripped | C — unmanaged |
|---|---|---|---|
| Composite S1+S2+S3 | **91.7%** | 91.7% | 62.5% |

- **A−C = +29.2pp**, meeting the pre-registered ≥20pp bar. Every actual
  poisoning infection in the whole program (3/3) happened in arm C with the
  weaker consumer model.
- Across all 54 responses in the program, arm A never lost a point to a
  poisoned claim. In the mixed condition (repo access allowed), unlabeled arms
  failed by *contagion*: one discovered lie caused true facts to be rejected
  alongside it — labels kept `verified` trusted while quarantining only the
  `⚠ conflict` entry.
- Takeaway: an agent can refute lies the code disproves; **labels protect the
  truth that lives outside the code** — decisions, constraints, context. The
  weaker the consumer, the larger the protection.

Full protocol, artifacts, and per-response grades publish after the
human-grading replication completes (in progress). Scores so far are
agent-judged under a pre-registered rubric.

## Capture: wire the hooks

Live capture feeds the ledger and the built-in measurement counters. For
Claude Code, add to your project's `.claude/settings.json` (hooks run in the
project directory, so `kervo` just needs to be on PATH):

```json
{
  "hooks": {
    "UserPromptSubmit": [
      { "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ],
    "SessionStart": [
      { "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ],
    "PostToolUse": [
      { "matcher": "Edit|Write",
        "hooks": [{ "type": "command", "command": "kervo hook || true", "timeout": 10 }] }
    ]
  }
}
```

The hook is a millisecond-budget local append — no LLM, no network, and it
never breaks your session (garbage in, exit 0 out). The committed ledger
stores **names, paths, and sizes only**: prompt and file contents never leave
your machine or enter git history.

```bash
kervo capture -type decision -body "JWT over sessions"   # record by hand
kervo trust -id 01KWP -to verified -reason "team agreed" # judge
kervo status                                             # one-screen trust view
kervo metrics                                            # prompt sizes: with vs without artifact
kervo import claude                                      # back-fill from past Claude Code sessions
```

## Design guarantees

- **Deterministic skeleton** — same workspace, same language, same bytes;
  pinned by golden files in CI. No LLM in the fact path, ever.
- **Events are truth** — an append-only JSONL ledger, committed to git
  (`merge=union`); the artifact and index are derived and rebuildable.
  Clone the repo, and its compiled memory moves with it.
- **Boundaries as checks** — the pure core cannot import adapters
  (`make arch-check`); data-derived text cannot impersonate structural
  markers; providers cannot self-promote past `generated`.
- **No server, no daemon, no database, no account** — all state lives in
  `.kervo/` and `CLAUDE.md`.

## Status

v0, moving fast. Cold-start validation passed with semantic slots enabled;
the capture → verify → stale loop is live in this repository. PRD / RFCs /
experiment protocols will be published as they stabilize.

---

kervo is not a coding tool. It is a memory layer for any team that lives on
git — developers are simply the first market, because they already store
their work as commits.

Licensed under [Apache-2.0](LICENSE).
