package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kervo-os/kervo/internal/core/fact"
)

// ui styles the human-facing console surface only. The artifact stays plain
// deterministic markdown — style never leaks into build outputs. Color is
// dropped automatically when stdout is not a TTY, NO_COLOR is set, or
// TERM=dumb (https://no-color.org convention).
type ui struct{ color bool }

func newUI() ui {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return ui{}
	}
	fi, err := os.Stdout.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return ui{}
	}
	return ui{color: true}
}

func (u ui) paint(code, s string) string {
	if !u.color {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (u ui) bold(s string) string  { return u.paint("1", s) }
func (u ui) dim(s string) string   { return u.paint("2", s) }
func (u ui) green(s string) string { return u.paint("32", s) }
func (u ui) cyan(s string) string  { return u.paint("36", s) }

var logoLines = []string{
	"┬┌─ ┌─┐ ┬─┐ ┬  ┬ ┌─┐",
	"├┴┐ ├┤  ├┬┘ └┐┌┘ │ │",
	"┴ ┴ └─┘ ┴└─  └┘  └─┘",
}

const tagline = "deterministic context for non-deterministic agents"

func (u ui) banner(version string) string {
	var b strings.Builder
	for i, l := range logoLines {
		b.WriteString(u.cyan(l))
		if i == len(logoLines)-1 {
			b.WriteString("  " + u.dim(version))
		}
		b.WriteString("\n")
	}
	b.WriteString(u.dim(tagline) + "\n")
	return b.String()
}

const rule = "────────────────────────────────────────────────"

// renderColdStart is the `kervo init` result screen (PRD §11 Cold Start UX:
// facts asserted plainly, first impressions are the product).
func renderColdStart(u ui, s fact.Snapshot, version string, injected []string) string {
	var b strings.Builder
	b.WriteString(u.banner(version))
	b.WriteString("\n")

	hasDoc := func(name string) bool {
		for _, d := range s.Docs {
			if d.Path == name {
				return true
			}
		}
		return false
	}
	mark := func(ok bool, label string) string {
		if ok {
			return u.green("✓") + " " + label
		}
		return u.dim("– " + label)
	}
	b.WriteString(u.bold("Workspace Found") + "   " +
		mark(true, "Git") + "   " + // init fails before this screen without git
		mark(hasDoc("CLAUDE.md"), "CLAUDE.md") + "   " +
		mark(hasDoc("README.md") || hasDoc("README"), "README") + "\n")
	b.WriteString(u.dim(rule) + "\n")

	row := func(label, value string) {
		if value == "" {
			return
		}
		b.WriteString("  " + u.dim(fmt.Sprintf("%-10s", label)) + " " + value + "\n")
	}
	commits := strconv.Itoa(len(s.Commits)) + " analyzed"
	if s.Partial {
		commits += "  " + u.dim("(partial — scan capped)")
	}
	row("Commits", commits)
	row("Languages", strings.Join(s.Repo.Languages, ", "))
	row("Frameworks", strings.Join(s.Repo.Frameworks, ", "))
	row("Tasks", strconv.Itoa(len(s.Todos))+" open · "+strconv.Itoa(len(s.Modules))+" modules")
	b.WriteString(u.dim(rule) + "\n")
	row("Artifact", u.bold(".kervo/artifact.md")+"  "+u.dim("(Mode 1 — Fact-only)"))
	row("Injected", strings.Join(injected, ", ")+"  "+u.dim("(marker block)"))
	return b.String()
}
