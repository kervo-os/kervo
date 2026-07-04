# kervo

**Stop re-explaining your project to AI. `kervo init` once.**

kervo compiles your repository into a deterministic Context Artifact and
injects it into `CLAUDE.md` — so every AI session starts already knowing
your project. Facts are extracted deterministically; interpretations enter
only as trust-labeled proposals. *Deterministic context for
non-deterministic agents.*

```
┬┌─ ┌─┐ ┬─┐ ┬  ┬ ┌─┐
├┴┐ ├┤  ├┬┘ └┐┌┘ │ │
┴ ┴ └─┘ ┴└─  └┘  └─┘
```

## Quickstart

```bash
make build          # or: go build -o kervo ./cmd/kervo
./kervo init        # scan → .kervo/artifact.md → injected into CLAUDE.md
```

First run on a real repository takes well under a second (500-commit scan
cap, marked `Partial` when hit). The artifact covers: repository summary,
declared commands (Makefile targets, npm scripts), recent changes
(merge-commit noise excluded), open TODO/FIXME tasks, module layout — plus
slots for goal / decisions / risks / summaries.

Only the block between `<!-- kervo:begin -->` and `<!-- kervo:end -->` in
`CLAUDE.md` is ever touched. Everything you wrote by hand is preserved
byte-for-byte.

## Compilation modes

| Mode | What fills the semantic slots | Requires |
|---|---|---|
| **1 — Fact-only** (default) | Nothing — deterministic facts only. Always works. | git |
| **2 — Consumer-assisted** | Your AI session stages proposals in `.kervo/proposals.json`; `kervo compile` attaches them | an agent session |
| **3 — Dedicated backend** | Any OpenAI-compatible endpoint proposes goal/summary/risk observations | a local or remote LLM |

Mode 3 with a fully local model (nothing leaves your machine):

```bash
export KERVO_SEMANTIC_URL=http://localhost:1234/v1   # LM Studio (or Ollama :11434/v1)
export KERVO_SEMANTIC_MODEL=openai/gpt-oss-120b
kervo compile
# Artifact: .kervo/artifact.md (Mode 3 — backend:openai/gpt-oss-120b)
```

Failures never break a run: Mode 3 → Mode 2 → fact-only, each demotion
warned, the fact skeleton always produced.

Artifacts render in English by default; `--lang ko` / `--lang ja` localize
them (the choice persists per workspace).

## Capture: wire the hooks

Live capture feeds the ledger and the built-in measurement counters.
For Claude Code, add to your project's `.claude/settings.json`
(hooks run in the project directory, so `kervo` just needs to be on PATH):

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

The hook is a millisecond-budget local append — no LLM, no network, and
it never breaks your session (garbage in, exit 0 out). The committed
ledger stores **names, paths, and sizes only**: prompt and file contents
never leave your machine or enter git history.

What you get:

```bash
kervo capture -type decision -body "JWT over sessions"   # record by hand
kervo trust -id 01KWP -to verified -reason "team agreed" # judge
kervo status                                             # one-screen trust view
kervo metrics                                            # prompt sizes: with vs without artifact
```

## Why trust labels

Accumulated context rots — and wrong context is worse than none. Every
non-fact enters as a labeled proposal with provenance:

```
**[generated — backend:openai/gpt-oss-120b]**
Needs confirmation — current focus appears to be terminal input/UX
hardening… Evidence: Recent Changes 05-28..06-28.
```

States move `generated → observed → verified → stale → deprecated` — by
evidence and human confirmation, not by a decay timer. Stale entries are
listed with their exclusion reason instead of being silently dropped.

## Design guarantees

- **Deterministic skeleton** — same workspace, same language, same bytes;
  pinned by golden files in CI. No LLM in the fact path, ever.
- **Boundaries as checks** — the pure core cannot import adapters
  (`make arch-check`); data-derived text cannot impersonate structural
  markers; providers cannot self-promote past `generated`.
- **Repo-native** — all state lives in `.kervo/` and `CLAUDE.md`. Clone the
  repo, and its compiled memory moves with it. No server, no daemon, no
  database, no account.

## Status

v0. Cold-start validation (H1) passed with semantic slots enabled;
capture/experience accumulation (the Experience Store) is in active
development. Docs: PRD / RFCs / architecture live outside this repo for
now and will be published as they stabilize.

---

kervo is not a coding tool. It is a memory layer for any team that lives
on git — developers are simply the first market, because they already
store their work as commits.

Licensed under [Apache-2.0](LICENSE).
