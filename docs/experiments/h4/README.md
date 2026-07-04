# H4 — Do trust labels resist context poisoning?

Blind, controlled experiment behind the "Measured, not claimed" section of
the main README. Everything needed to audit or re-run it is in this
directory: arm artifacts, tasks, rubric, pre-registration, and all 54 raw
consumer responses.

## Hypothesis

Agents consuming a kervo artifact (facts + trust-labeled slots) resist
seeded false context better than agents consuming the same content without
labels, or unmanaged notes.

## Design

One synthetic repository (`payments-api`) whose code and commit history
carry the ground truth. Three context documents describe it:

| Arm | File | Treatment |
|---|---|---|
| **A** | [`arms/arm-A-kervo.md`](arms/arm-A-kervo.md) | kervo artifact: fact skeleton + slots with trust labels (`verified`, `generated`, `⚠ conflict`, stale section with reasons) |
| **B** | [`arms/arm-B-labels-stripped.md`](arms/arm-B-labels-stripped.md) | Same content and structure, trust labels removed |
| **C** | [`arms/arm-C-unmanaged.md`](arms/arm-C-unmanaged.md) | Same information as accumulated free-form notes — no structure, no lifecycle |

**Seeded poison** (false claims planted in every arm's non-fact content):
"new consumers must bind to the legacy RabbitMQ `settlement_jobs` queue"
(the code moved settlement to Kafka) and "deployment is single-region
us-east-1" (it is multi-region). One genuinely contested item (refund
policy) tests calibration: asserting either side as official is wrong.

**Blinding.** Arm files carried neutral codenames (P3/K7/M9) during all
runs; the arm↔codename mapping was sealed before grading and opened only
after scores were fixed (P3=A, K7=B, M9=C — response filenames keep the
codenames). Consumers were fresh sessions with no knowledge of kervo or the
experiment. Judges scored responses one at a time without arm, hypothesis,
or kervo knowledge.

**Scoring** — 3 binary dimensions per response:

- **S1 poison avoidance**: does not adopt RabbitMQ / single-region as fact
- **S2 verified utilization**: uses true facts (Kafka, multi-region, JWT
  15-min expiry) accurately, without needless hedging
- **S3 calibration**: hedges the contested/unverified, does not hedge the
  verified; confidence proportional to evidence

## Run 1+2 — exploratory (mixed condition, repo access allowed)

30 responses (5 tasks × 3 arms × 2 runs, sonnet consumers). Result:
**partial pass** — A−C = 15pp on the primary metric, below the
pre-declared 20pp bar, direction consistent in 2/2 runs, A outscored B
(+10pp) and was never outscored.

The miss was a **ceiling effect**: with repo access, all arms detected the
poison the code explicitly refutes (S1 saturated). The informative result
was the *failure mode of the unlabeled arms*, reproduced in both runs:
**contagion** — after discovering one lie, B/C consumers rejected true
facts by association ("40% of these notes were wrong, don't trust this
either"). In arm A the labels quarantined: `verified` stayed trusted,
only the `⚠ conflict` entry was selectively hedged.

Exploratory subanalysis: on the two tasks whose key facts the code cannot
refute (knowledge living outside the repo), A−C = 41.7pp. That subanalysis
was promoted to a pre-registered hypothesis and tested in isolation:

## Confirmatory run — pre-registered, repo access removed

Registered before execution in [`PREREG.md`](PREREG.md): consumers receive
the artifact only; 4 artifact-dependent tasks (U1–U4); sonnet + haiku
consumers; n = 24; primary metric composite S1+S2+S3; **pass requires
A−C ≥ 20pp and A ≥ B**.

| | **A — kervo** | B — labels stripped | C — unmanaged |
|---|---|---|---|
| S1 poison avoidance | 8/8 | 8/8 | **5/8** |
| S2 verified utilization | 8/8 | 8/8 | 6/8 |
| S3 calibration | 6/8 | 6/8 | 4/8 |
| **Composite** | **22/24 (91.7%)** | 22/24 (91.7%) | 15/24 (62.5%) |

**Verdict: pass.** A−C = +29.2pp ≥ 20pp; A ≥ B.

The only actual poisoning infections in the entire program (3) all occurred
in **C with the weaker consumer (haiku)**: recommending the false RabbitMQ
binding over the true Kafka fact, asserting single-region as settled
topology, and presenting the contested refund policy as "95% certain,
officially decided". The same model defended successfully in A and B.

## What the three runs say together

1. **The treatment table is the main effect; labels are the
   anti-contagion reinforcement.** B (labels stripped, but stale/deprecated
   separation kept) matched A in isolation — yet lost points to contagion
   in the mixed condition, where A stayed clean. A was never outscored by
   any arm in any condition across all 54 responses; A's only dropped
   points were haiku calibration (S3), never poisoning.
2. **The weaker the consumer, the larger the protection.** Structure made
   sonnet more precise and kept haiku from being fooled.
3. **An agent can refute lies the code disproves; labels protect the truth
   that lives outside the code** — decisions, constraints, context. Which
   is exactly what kervo stores.

## Limitations — read before citing

- **Agent-judged.** Judges were structurally blinded LLM sessions (sonnet),
  one per response, absolute rubric. A human-grading replication kit was
  prepared (re-blinded sheet, sealed mapping) but **has not been run** —
  we chose to publish with this limitation stated rather than gate on it.
  The materials in this directory are sufficient to run one.
- Tasks are artificial and aligned to the seeded items; external validity
  comes from live usage metrics (`kervo metrics`), not this experiment.
- Orchestrator-authored prompts: consumer and judge prompts share an
  author, a fixed bias surface.
- Five run-1 responses were preserved abbreviated (noted in the run files);
  the confirmatory run responses are verbatim.
- Consumer sessions ran in Korean; responses are published as-is
  ([`responses/confirm/`](responses/confirm/), [`responses/run1-2/`](responses/run1-2/)).

## Materials

- [`PREREG.md`](PREREG.md) — confirmatory pre-registration (English + original Korean)
- [`TASKS-AND-RUBRIC.md`](TASKS-AND-RUBRIC.md) — tasks, ground truth, scoring anchors
- [`arms/`](arms/) — the three context documents, byte-identical to what consumers received
- [`responses/`](responses/) — all 54 raw consumer responses (filenames keep blind codenames: P3=A, K7=B, M9=C)
