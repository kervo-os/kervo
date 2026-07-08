# FAQ

Honest answers to the questions skeptics ask first.

### Isn't this just a `## Decisions` section in CLAUDE.md?

That is where everyone starts, and it works — until it rots. A markdown
section has no way to show that an entry went stale, who asserted it, or
why it was retired; an agent reads a reversed decision as current fact.
kervo's machinery is the minimum needed to make staleness *visible*:
states with reasons, provenance on every entry, retirement that leaves a
tombstone instead of a silent deletion. If your one file never rots, you
don't need kervo.

### Isn't this ADRs with extra steps?

The lineage is real — the trust lifecycle descends from Nygard's status
field. Three things are new: the writer and the signer are separated
(agents draft, humans judge — an agent can never sign its own claim);
the record is compiled into agent context on every commit instead of
waiting to be read; and decisions can carry path anchors, so a PR that
touches what a verified decision governs gets warned in CI. ADRs are
documents you read; this is state that acts.

### What happens when nobody judges the queue?

Three backstops. A source that piles up 12 unjudged proposals is blocked
from proposing more — production halts before the queue rots. Facts a
human affirmed in session skip the queue entirely; only unattended
knowledge waits. And unjudged entries stay quarantined as `generated`,
which consumers are told to hedge — measured behavior, not hope. A team
that never judges still gets the deterministic fact skeleton; it just
leaves the trust layer's value on the table.

### Won't the ledger pollute my git history?

It rides your commits instead of making its own: the pre-commit hook
stages the events file with the change that produced it, and
`.gitattributes` marks it `linguist-generated`, so GitHub collapses it
in PR diffs. Events live in their own files — `git blame` on code is
untouched. Hook capture is opt-in; without it, events appear only when
someone captures deliberately.

### Isn't injecting text into agent context a prompt-injection surface?

Yes, and we measured it instead of hand-waving: seeded false "decisions"
in a pre-registered blind experiment infected the unmanaged arm and not
the labeled ones (all three real infections happened without labels).
Structurally: nothing enters anonymously, machine proposals are
quarantined until a human signs, and the trust boundary is git's own —
someone with commit access could already poison your README. See
[SECURITY.md](SECURITY.md).

### What if Claude/OpenAI ship team memory tomorrow?

Then agents get better memory and kervo still holds the part a vendor
can't: your team's signed record, in your repository, in plain text any
agent can read — not in one vendor's account. The bet is on ownership,
not features. If every vendor someday agrees on a git-native, signed,
portable memory format, this project retires happily — that's the world
it argues for.

### "Signed memory" — where's the signature?

The signature is a recorded human judgment with provenance, at git's
trust level — the same level as your commit authorship, no more, no
less. Anyone who can forge a commit author can forge an actor; teams
that sign commits get judgment commits signed for free. What the word
buys you today is auditability (who judged what, when, why — replayable)
plus two hard rules: agents can never promote claims to `verified`, and
an agent relaying your judgment must quote your words or the ledger
rejects it. Cryptographic content-hash binding is v2 territory, for
teams whose threat model needs signatures stronger than their commits.

### Who governs the governor? The design doc is written by the same agent that builds the features.

True, and v1 doesn't pretend otherwise: this is a single-human team, so
every check ultimately routes through one person plus their agent. The
honest controls are external ones — the ledger is public and replayable,
experiments are pre-registered with anchored hashes, and the replication
kit invites anyone to check the claims. Multi-party trust (judgments
that require more than one human) is explicitly v2's boundary, gated on
the first team that needs it.

### Is the evidence any good?

It is honest more than it is heavy: pre-registered thresholds (anchored
via OpenTimestamps before results existed), blinded judging, raw
responses public, and the misses reported alongside the passes. It is
also agent-graded with a small n — the paper says "a strong
pre-registered signal, not statistical proof" in exactly those words.
The human-grading replication kit is public in
[kervo-os/experiments](https://github.com/kervo-os/experiments);
running it and proving us wrong would be a contribution.
