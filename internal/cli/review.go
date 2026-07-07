package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// runReview is the verifier's surface: a triage queue over everything that
// awaits human judgment — generated proposals and ⚠ conflicts — one item
// at a time, no IDs to memorize. Division of labor by design: agents
// capture and propose, the human only judges. Every judgment is the same
// transition event `kervo trust` writes — review is sugar, the ledger
// stays the API.
func runReview(args []string) error {
	fs := newFlagSet("review")
	dir := fs.String("dir", ".", "workspace directory")
	actor := fs.String("actor", "", `who is judging (default "human:<git user.name>")`)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return reviewLoop(*dir, *actor, os.Stdin, os.Stdout)
}

// unescapeBody restores newlines an agent passed as literal backslash-n
// through shell quoting. Display-layer only — the ledger keeps the bytes
// it was given; a body legitimately containing the two characters loses
// nothing but a rare cosmetic edge.
func unescapeBody(s string) string {
	return strings.ReplaceAll(s, `\n`, "\n")
}

// reviewLoop takes reader/writer so the interaction is testable. EOF at any
// prompt ends the session cleanly — a piped or CI stdin must never hang.
func reviewLoop(dir, actorFlag string, in io.Reader, out io.Writer) error {
	store := jsonl.Open(dir)
	folder, err := replayFolder(store)
	if err != nil {
		return err
	}
	var queue []trust.Observation
	for _, o := range folder.Observations() {
		if o.State == trust.Generated || o.Conflict {
			queue = append(queue, o)
		}
	}
	if len(queue) == 0 {
		fmt.Fprintln(out, "Review queue is empty — nothing awaits judgment.")
		return nil
	}

	fmt.Fprintf(out, "%d awaiting judgment · v=verify s=stale d=deprecate Enter=skip q=quit\n", len(queue))
	reader := bufio.NewReader(in)
	judged := 0
loop:
	for i, o := range queue {
		mark := ""
		if o.Conflict {
			mark = "  ⚠ conflict"
		}
		fmt.Fprintf(out, "\n[%d/%d] %s · %s · %s · %s%s\n", i+1, len(queue), shortID(o.ID), o.Type, o.State, o.LastActor, mark)
		fmt.Fprintf(out, "  %s\n", strings.ReplaceAll(strings.TrimRight(o.Body, "\n"), "\n", "\n  "))
		if o.Evidence != "" {
			fmt.Fprintf(out, "  evidence: %s\n", o.Evidence)
		}
		fmt.Fprint(out, "> ")
		line, rerr := reader.ReadString('\n')

		var target trust.State
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "v", "verify":
			target = trust.Verified
		case "s", "stale":
			target = trust.Stale
		case "d", "deprecate", "deprecated":
			target = trust.Deprecated
		case "q", "quit":
			break loop
		}
		if target == "" {
			if rerr != nil {
				fmt.Fprintln(out)
				break loop
			}
			continue // skip
		}
		if !trust.CanTransition(o.State, target) {
			fmt.Fprintf(out, "  %s → %s is not a legal transition — skipped\n", o.State, target)
			continue
		}
		fmt.Fprint(out, "reason (optional) > ")
		reason, rrerr := reader.ReadString('\n')
		if err := appendTransition(store, dir, actorFlag, o, target, strings.TrimSpace(reason)); err != nil {
			return err
		}
		judged++
		fmt.Fprintf(out, "  %s: %s → %s\n", shortID(o.ID), o.State, target)
		if rrerr != nil {
			fmt.Fprintln(out)
			break loop
		}
	}
	fmt.Fprintf(out, "\nJudged %d of %d. `kervo compile` refreshes the artifact.\n", judged, len(queue))
	return nil
}
