// Package cli defines the command surface: init, compile, capture, hook, mcp.
// Zero external deps by design; swap to cobra later only if flags outgrow flag pkg.
package cli

import (
	"flag"
	"fmt"
)

type command struct {
	name, summary string
	run           func(args []string) error
}

func commands() []command {
	return []command{
		{"init", "Scan workspace, build Cold Start artifact (Mode 1, 30s budget)", runInit},
		{"compile", "Incremental scan + recompile artifact (fallback: RFC-0003 §4)", runCompile},
		{"capture", "Record an Observation into the event ledger", runCapture},
		{"trust", "Judge an observation: -id <prefix> -to verified|stale|deprecated", runTrust},
		{"review", "Triage queue: judge generated proposals and ⚠ conflicts one by one", runReview},
		{"check", "Gate a diff: which verified anchored decisions does this change touch?", runCheck},
		{"dash", "Fleet dashboard: every registered workspace on one page (one-shot, 127.0.0.1)", runDash},
		{"status", "One-screen view: ledger size, trust states, observations", runStatus},
		{"metrics", "H3 counters: prompt sizes with vs without the artifact", runMetrics},
		{"import", "Back-fill the ledger from existing stores: claude (transcripts, sizes only)", runImport},
		{"hook", "Entry point invoked by Consumer hooks (stdin JSON)", runHook},
		{"mcp", "Serve stdio MCP (Facts out, Observations in)", runMCP},
		{"version", "Print version", runVersion},
	}
}

// Run dispatches to a subcommand. Kept trivial on purpose.
func Run(args []string) error {
	if len(args) == 0 {
		usage()
		return nil
	}
	for _, c := range commands() {
		if c.name == args[0] {
			return c.run(args[1:])
		}
	}
	usage()
	return fmt.Errorf("unknown command %q", args[0])
}

func usage() {
	fmt.Println("kervo — workspace context compiler")
	fmt.Println("\nCommands:")
	for _, c := range commands() {
		fmt.Printf("  %-8s %s\n", c.name, c.summary)
	}
}

// newFlagSet is shared boilerplate for subcommand flags.
func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	return fs
}
