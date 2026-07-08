package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
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
	for _, o := range dead {
		if gha {
			fmt.Printf("::notice::anchor of %s matches no tracked file — %s — consider: kervo trust -id %s -to stale -reason \"anchored path gone\"\n",
				shortID(o.ID), claim(o.Body), shortID(o.ID))
			continue
		}
		fmt.Printf("✝ %s anchors match no tracked file — %s\n", shortID(o.ID), claim(o.Body))
		fmt.Printf("  consider: kervo trust -id %s -to stale -reason \"anchored path gone\"\n", shortID(o.ID))
	}

	for _, d := range drifted {
		if gha {
			fmt.Printf("::notice::%s was verified before %d commits landed on its anchors — %s — re-affirm (kervo trust -id %s -to verified -reason \"still true\") or retire it\n",
				shortID(d.Obs.ID), d.Commits, claim(d.Obs.Body), shortID(d.Obs.ID))
			continue
		}
		fmt.Printf("↻ %s anchored code moved in %d commits since verification — %s\n",
			shortID(d.Obs.ID), d.Commits, claim(d.Obs.Body))
		fmt.Printf("  re-affirm: kervo trust -id %s -to verified -reason \"still true\" — or retire it\n", shortID(d.Obs.ID))
	}

	fmt.Printf("check: %d changed files vs %s — %d touched, %d dead anchors, %d drifted\n",
		len(changed), *base, len(conflicts), len(dead), len(drifted))
	if *strict && len(conflicts) > 0 {
		return fmt.Errorf("check: %d verified decision(s) touched (strict mode)", len(conflicts))
	}
	return nil
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
		"--format=%x1e%aI", "--name-only", base).Output()
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
		var files []string
		for _, l := range lines[1:] {
			if l = strings.TrimSpace(l); l != "" {
				files = append(files, l)
			}
		}
		changes = append(changes, gate.Change{At: at, Files: files})
	}
	return changes
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
