package compiler

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/kervo-os/kervo/internal/core/fact"
)

// The Brief opens the artifact with the lines a newcomer actually needs —
// synthesized presentation of deterministic signals, never inference. The
// bar it answers: `kervo init` alone, on anyone's repo, must brief like a
// teammate — with zero LLM in the path (Mode 1 native).

// briefWindow bounds the "recent" analysis: enough commits to see a theme,
// few enough that last quarter's work doesn't drown this week's.
const briefWindow = 25

// scopeRe extracts the conventional-commit scope: feat(search): → search.
var scopeRe = regexp.MustCompile(`^[a-z]+!?\(([^)]+)\)[:!]`)

type freq struct {
	name string
	n    int
}

func topFreq(counts map[string]int, limit int) []freq {
	out := make([]freq, 0, len(counts))
	for k, v := range counts {
		out = append(out, freq{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].n != out[j].n {
			return out[i].n > out[j].n
		}
		return out[i].name < out[j].name
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func joinFreq(fs []freq, suffix string) string {
	parts := make([]string, 0, len(fs))
	for _, f := range fs {
		parts = append(parts, fmt.Sprintf("%s%s ×%d", f.name, suffix, f.n))
	}
	return strings.Join(parts, " · ")
}

// BriefFocus counts where the recent window of commits concentrates:
// conventional-commit scopes when the repo declares them, and the
// top-level modules the commits touched. Counting, not reading minds.
func BriefFocus(s fact.Snapshot) string {
	window := s.Commits
	if len(window) > briefWindow {
		window = window[:briefWindow]
	}
	if len(window) < 3 {
		return ""
	}
	scopes := map[string]int{}
	mods := map[string]int{}
	for _, c := range window {
		if m := scopeRe.FindStringSubmatch(c.Subject); m != nil {
			scopes[m[1]]++
		}
		seen := map[string]bool{}
		for _, f := range c.Files {
			if i := strings.IndexByte(f, '/'); i > 0 && f[0] != '.' && !seen[f[:i]] {
				seen[f[:i]] = true
				mods[f[:i]]++
			}
		}
	}
	var parts []string
	if len(scopes) > 0 {
		parts = append(parts, joinFreq(topFreq(scopes, 4), ""))
	}
	if len(mods) > 0 {
		parts = append(parts, joinFreq(topFreq(mods, 3), "/"))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " — ")
}

// briefRun compresses the command list to one orientation line: compose
// services collapse to a count, the first declared build/test entries
// stay verbatim.
func briefRun(s fact.Snapshot) string {
	var head []string
	compose := 0
	for _, c := range s.Commands {
		if strings.HasPrefix(c.Run, "docker compose up") {
			compose++
			continue
		}
		if len(head) < 3 {
			head = append(head, "`"+c.Run+"`")
		}
	}
	parts := head
	if compose > 0 {
		parts = append(parts, fmt.Sprintf("`docker compose up` (%d services)", compose))
	}
	return strings.Join(parts, " · ")
}

// writeBrief renders the section; empty signals render nothing — a brief
// that pads is worse than none.
func writeBrief(b *strings.Builder, s fact.Snapshot, tr func(string) string) {
	type line struct{ label, text string }
	var lines []line

	if focus := BriefFocus(s); focus != "" {
		lines = append(lines, line{"brief.focus", focus})
	}
	if run := briefRun(s); run != "" {
		lines = append(lines, line{"brief.run", run})
	}
	if len(s.Todos) > 0 {
		paths := make([]string, 0, 3)
		for i, t := range s.Todos {
			if i >= 3 {
				break
			}
			paths = append(paths, t.Path)
		}
		lines = append(lines, line{"brief.edges",
			fmt.Sprintf("%d — %s", len(s.Todos), strings.Join(paths, ", "))})
	}
	if s.Repo.Ahead > 0 {
		lines = append(lines, line{"brief.unpushed", fmt.Sprintf("%d", s.Repo.Ahead)})
	}
	if len(lines) == 0 {
		return
	}
	b.WriteString("## " + tr("brief.title") + "\n\n")
	for _, l := range lines {
		b.WriteString("- **" + tr(l.label) + "**: " + esc(l.text) + "\n")
	}
	b.WriteString("\n")
}
