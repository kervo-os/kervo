package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
	"github.com/kervo-os/kervo/internal/core/gate"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runCheck gates a diff against the ledger: which verified, anchored
// decisions does this change touch? Advisory by default — a conflicting
// PR may be an intentional reversal, and the right response to a
// reversal is a judgment (deprecate + capture), not a red build. -strict
// exits non-zero for teams that want the gate to block.
func runCheck(args []string) error {
	fs := newFlagSet("check")
	dir := fs.String("dir", ".", "workspace directory")
	base := fs.String("base", "origin/main", "diff base ref (compared as base...HEAD)")
	strict := fs.Bool("strict", false, "exit non-zero when a verified anchored observation is touched")
	if err := fs.Parse(args); err != nil {
		return err
	}

	changed, err := gitList(*dir, "diff", "--name-only", *base+"...HEAD")
	if err != nil {
		return fmt.Errorf("check: cannot diff against %q — pass -base <ref> that exists (in CI, fetch the base branch first): %v", *base, err)
	}

	folder, err := replayFolder(jsonl.Open(*dir))
	if err != nil {
		return err
	}
	obs := folder.Observations()

	conflicts := gate.Conflicts(obs, changed)

	// Dead anchors ride along on every check: an anchor that matches no
	// tracked file is evidence its subject left the workspace — the first
	// deterministic staleness signal (the rest are age-based).
	tracked, err := gitList(*dir, "ls-files")
	if err != nil {
		tracked = nil // outside a work tree the signal is just unavailable
	}
	var dead []trust.Observation
	if tracked != nil {
		dead = gate.Dead(obs, tracked)
	}

	// Drift: reversals nobody recorded arrive as code churn under the
	// decision — when anchored paths moved ≥5 commits past the judgment,
	// ask for re-affirmation instead of trusting silence. The base ref is
	// scanned: landed churn is what erodes a decision (this diff's own
	// touches already fire the conflict warning above).
	drifted := gate.Drifted(obs, gitChanges(*dir, *base, obs))

	gha := os.Getenv("GITHUB_ACTIONS") == "true"
	for _, c := range conflicts {
		if gha {
			fmt.Printf("::warning file=%s::this change touches a verified decision — %s [%s — %s] — if it reverses the decision: kervo trust -id %s -to deprecated -reason \"<why>\" and capture the new one%s\n",
				c.Files[0], claim(c.Obs.Body), shortID(c.Obs.ID), c.Obs.LastActor, shortID(c.Obs.ID), moreFiles(c.Files))
			continue
		}
		fmt.Printf("⚠ %s %s [verified — %s]\n", shortID(c.Obs.ID), c.Obs.Type, c.Obs.LastActor)
		fmt.Printf("  %s\n", claim(c.Obs.Body))
		fmt.Printf("  touched: %s\n", strings.Join(c.Files, ", "))
		fmt.Printf("  reversal? kervo trust -id %s -to deprecated -reason \"<why>\" — then capture the new decision\n", shortID(c.Obs.ID))
	}
	// git knows about renames even when anchors don't — a dead anchor that
	// died to a refactor gets its forwarding address, not just a tombstone.
	renames := gitRenames(*dir, obs)
	for _, o := range dead {
		moved := renameHint(o.Anchors, renames)
		if gha {
			fmt.Printf("::notice::anchor of %s matches no tracked file — %s —%s consider: kervo trust -id %s -to stale -reason \"anchored path gone\"\n",
				shortID(o.ID), claim(o.Body), moved, shortID(o.ID))
			continue
		}
		fmt.Printf("✝ %s anchors match no tracked file — %s\n", shortID(o.ID), claim(o.Body))
		if moved != "" {
			fmt.Printf(" %s\n", moved)
		}
		fmt.Printf("  consider: kervo trust -id %s -to stale -reason \"anchored path gone\"\n", shortID(o.ID))
	}

	// Evidential trust, not statistical decay: the re-affirmation reason is
	// yours to write — the system never dictates one, or the signature is
	// theater.
	for _, d := range drifted {
		if gha {
			fmt.Printf("::notice::%s was verified before %d commits landed on its anchors — %s — re-affirm with your reason (kervo trust -id %s -to verified -reason \"<why it still holds>\") or retire it\n",
				shortID(d.Obs.ID), d.Commits, claim(d.Obs.Body), shortID(d.Obs.ID))
			continue
		}
		fmt.Printf("↻ %s anchored code moved since verification (%d commits, %d lines) — %s\n",
			shortID(d.Obs.ID), d.Commits, d.Lines, claim(d.Obs.Body))
		fmt.Printf("  re-affirm with your reason: kervo trust -id %s -to verified -reason \"<why it still holds>\" — or retire it\n", shortID(d.Obs.ID))
	}

	// Trust changes shipped inside the diff are review objects themselves.
	// A PR must not silently retire the decision that would have flagged
	// it — RFC-0005 §2.2: disagreement is surfaced, never hidden. This is
	// also the legitimate reversal flow (deprecate + capture beside the
	// code), so it is surfaced for the reviewer, not blocked.
	transitions := diffTransitions(*dir, *base)
	for _, tr := range transitions {
		body := ""
		if o, ok := folder.Get(tr.Ref); ok {
			body = " — " + claim(o.Body)
		}
		if gha {
			fmt.Printf("::warning::this PR itself moves %s to %s (%s)%s — the judgment ships with the code; reviewer, confirm both\n",
				shortID(tr.Ref), tr.To, tr.Actor, body)
			continue
		}
		fmt.Printf("± %s → %s by %s in this diff%s\n", shortID(tr.Ref), tr.To, tr.Actor, body)
		fmt.Printf("  the judgment ships with the code — review it like the code\n")
	}

	fmt.Printf("check: %d changed files vs %s — %d touched, %d dead anchors, %d drifted, %d trust changes in diff\n",
		len(changed), *base, len(conflicts), len(dead), len(drifted), len(transitions))
	// Strict blocks on both gate-worthy events: touching a verified
	// decision AND rewriting trust state inside the diff — surfacing that
	// strict mode ignores would be a log line, not a gate.
	if *strict && (len(conflicts) > 0 || len(transitions) > 0) {
		return fmt.Errorf("check: %d verified decision(s) touched, %d trust change(s) shipped in diff (strict mode)",
			len(conflicts), len(transitions))
	}
	return nil
}

// gitRenames collects renames since the oldest judgment among verified
// anchored observations, old path → new path.
func gitRenames(dir string, obs []trust.Observation) map[string]string {
	var oldest time.Time
	for _, o := range obs {
		if o.State == trust.Verified && len(o.Anchors) > 0 && !o.JudgedAt.IsZero() {
			if oldest.IsZero() || o.JudgedAt.Before(oldest) {
				oldest = o.JudgedAt
			}
		}
	}
	if oldest.IsZero() {
		return nil
	}
	out, err := exec.Command("git", "-C", dir, "log", "-M", "--diff-filter=R",
		"--since="+oldest.UTC().Format(time.RFC3339),
		"--format=", "--name-status").Output()
	if err != nil {
		return nil
	}
	renames := map[string]string{}
	for _, l := range strings.Split(string(out), "\n") {
		parts := strings.Split(strings.TrimSpace(l), "\t")
		if len(parts) == 3 && strings.HasPrefix(parts[0], "R") {
			renames[parts[1]] = parts[2]
		}
	}
	return renames
}

// renameHint finds where a dead anchor's files went. Glob suffixes are
// stripped to a path prefix; the first rename under it names the move.
func renameHint(anchors []string, renames map[string]string) string {
	for _, a := range anchors {
		base := a
		if i := strings.IndexAny(base, "*?["); i >= 0 {
			base = base[:i]
		}
		base = strings.TrimSuffix(strings.TrimSpace(base), "/")
		if base == "" {
			continue
		}
		for old, now := range renames {
			if old == base || strings.HasPrefix(old, base+"/") {
				return fmt.Sprintf("  moved? git says %s → %s — recapture with the new anchor, then retire this one", old, now)
			}
		}
	}
	return ""
}

// gitChanges returns the commit footprints needed for drift detection —
// everything on base after the oldest judgment among verified anchored
// observations. Empty when there is nothing to scan for.
func gitChanges(dir, base string, obs []trust.Observation) []gate.Change {
	var oldest time.Time
	for _, o := range obs {
		if o.State == trust.Verified && len(o.Anchors) > 0 && !o.JudgedAt.IsZero() {
			if oldest.IsZero() || o.JudgedAt.Before(oldest) {
				oldest = o.JudgedAt
			}
		}
	}
	if oldest.IsZero() {
		return nil
	}
	out, err := exec.Command("git", "-C", dir, "log", "--no-merges",
		"--since="+oldest.UTC().Format(time.RFC3339),
		"--format=%x1e%aI", "--numstat", base).Output()
	if err != nil {
		return nil
	}
	var changes []gate.Change
	for _, rec := range strings.Split(string(out), "\x1e") {
		lines := strings.Split(strings.TrimSpace(rec), "\n")
		if len(lines) < 2 {
			continue
		}
		at, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[0]))
		if err != nil {
			continue
		}
		c := gate.Change{At: at, Lines: map[string]int{}}
		for _, l := range lines[1:] {
			// numstat: "<added>\t<deleted>\t<path>"; "-" marks binary.
			parts := strings.SplitN(strings.TrimSpace(l), "\t", 3)
			if len(parts) != 3 || parts[2] == "" {
				continue
			}
			n := atoiSafe(parts[0]) + atoiSafe(parts[1])
			c.Files = append(c.Files, parts[2])
			c.Lines[parts[2]] = n
		}
		if len(c.Files) > 0 {
			changes = append(changes, c)
		}
	}
	return changes
}

func atoiSafe(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// claim returns the first line of a body — captures are claim-first by
// protocol, so line one is the decision.
func claim(body string) string {
	if i := strings.IndexByte(body, '\n'); i >= 0 {
		body = body[:i]
	}
	return strings.TrimSpace(body)
}

func moreFiles(files []string) string {
	if len(files) <= 1 {
		return ""
	}
	return fmt.Sprintf(" (+%d more files)", len(files)-1)
}

// diffTransition is one trust-state change carried inside the diff.
type diffTransition struct {
	Ref, To, Actor string
}

// diffTransitions parses trust transitions out of the ledger lines this
// diff ADDS. Only added lines count — history is append-only, so removed
// ledger lines never carry a judgment (a PR deleting ledger lines is its
// own kind of alarm, visible in the raw diff).
func diffTransitions(dir, base string) []diffTransition {
	out, err := exec.Command("git", "-C", dir, "diff", "--unified=0",
		base+"...HEAD", "--", ".kervo/events").Output()
	if err != nil {
		return nil
	}
	var trs []diffTransition
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		var e event.Event
		if json.Unmarshal([]byte(strings.TrimPrefix(line, "+")), &e) != nil {
			continue
		}
		if e.Kind != event.KindTransition || e.Ref == "" {
			continue
		}
		var p struct {
			To string `json:"to"`
		}
		_ = json.Unmarshal(e.Payload, &p)
		trs = append(trs, diffTransition{Ref: e.Ref, To: p.To, Actor: e.Actor})
	}
	return trs
}

// gitList runs git in dir and returns non-empty output lines.
func gitList(dir string, args ...string) ([]string, error) {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}
