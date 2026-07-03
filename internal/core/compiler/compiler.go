// Package compiler is the process that turns a fact snapshot plus
// experience into a Context Artifact.
//
// HARD BOUNDARY (RFC-0003 §2, enforced by `make arch-check` + review):
// this package must never import LLM clients or adapters. Skeleton output
// is byte-identical for identical input — guarded by tests.
package compiler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/fact"
)

// Display caps. Snapshot keeps more; the artifact stays prompt-sized.
// Constants are part of the byte-identity contract — changing them changes
// the golden file.
const (
	maxRecentCommits = 20
	maxOpenTasks     = 30
	maxHotFiles      = 10
	maxExcerptLen    = 300
)

// slotTitles maps each enhancement slot to its section heading.
var slotTitles = map[string]string{
	artifact.SlotGoal:      "Possible Current Goal",
	artifact.SlotDecisions: "Known Decisions",
	artifact.SlotRisks:     "Known Risks",
	artifact.SlotSummaries: "Doc Summaries",
}

// placeholderFor is the empty-slot content. Detach restores it, which is
// what makes the round-trip invariant byte-exact.
func placeholderFor(slot string) string {
	switch slot {
	case artifact.SlotGoal:
		return "_No proposal yet. A confirmed goal becomes the first Verified observation._"
	default:
		return "_None proposed yet. Semantic providers (Mode 2/3) attach labeled observations here._"
	}
}

// BuildSkeleton renders the deterministic Mode-1 artifact.
// Input -> template -> markdown. No LLM, no clock beyond snapshot data.
func BuildSkeleton(s fact.Snapshot) (string, error) {
	var b strings.Builder

	b.WriteString("<!-- kervo:artifact v1 skeleton=fact-only -->\n")
	b.WriteString("# Context Artifact\n\n")
	b.WriteString("> Machine-generated context for AI agents. Fact sections are deterministic;\n")
	b.WriteString("> slot sections carry trust-labeled observations. Regenerate with `kervo compile`\n")
	b.WriteString("> — do not edit by hand.\n\n")

	writeRepoSummary(&b, s)
	writeRecentChanges(&b, s)
	writeOpenTasks(&b, s)
	writeRelatedModules(&b, s)
	writeWorkspaceFacts(&b, s)

	for _, slot := range artifact.Slots() {
		b.WriteString("## " + slotTitles[slot] + "\n\n")
		b.WriteString(artifact.SlotBegin(slot) + "\n")
		b.WriteString(placeholderFor(slot) + "\n")
		b.WriteString(artifact.SlotEnd(slot) + "\n\n")
	}

	b.WriteString("## Deprecated / Stale Notes\n\n")
	b.WriteString("_None recorded. Stale or deprecated observations are listed here with their\nexclusion reason instead of being silently dropped._\n")

	return b.String(), nil
}

// esc neutralizes the reserved marker prefix inside snapshot-derived text.
// Every data write site below MUST go through it: workspace content that
// impersonates a marker corrupts slot replacement and CLAUDE.md injection.
func esc(s string) string {
	return strings.ReplaceAll(s, artifact.ReservedPrefix, `<!-- kervo\:`)
}

func writeRepoSummary(b *strings.Builder, s fact.Snapshot) {
	b.WriteString("## Repository Summary\n\n")
	b.WriteString("- Name: " + orDash(esc(s.Repo.Name)) + "\n")
	b.WriteString("- Branch: " + orDash(esc(s.Repo.Branch)) + "\n")
	b.WriteString("- Languages: " + orDash(esc(strings.Join(s.Repo.Languages, ", "))) + "\n")
	b.WriteString("- Frameworks: " + orDash(esc(strings.Join(s.Repo.Frameworks, ", "))) + "\n")
	var docNames []string
	for _, d := range s.Docs {
		docNames = append(docNames, d.Path)
	}
	b.WriteString("- Docs: " + orDash(esc(strings.Join(docNames, ", "))) + "\n\n")

	for _, d := range s.Docs {
		if d.Path != "README.md" && d.Path != "README" {
			continue
		}
		if ex := firstParagraph(d.Content); ex != "" {
			b.WriteString("### " + esc(d.Path) + " (excerpt)\n\n")
			b.WriteString("> " + esc(ex) + "\n\n")
		}
		break
	}
}

func writeRecentChanges(b *strings.Builder, s fact.Snapshot) {
	b.WriteString("## Recent Changes\n\n")
	if len(s.Commits) == 0 {
		b.WriteString("_No commits found._\n\n")
		return
	}
	shown := len(s.Commits)
	if shown > maxRecentCommits {
		shown = maxRecentCommits
	}
	for _, c := range s.Commits[:shown] {
		b.WriteString("- `" + shortSHA(c.SHA) + "` " + c.At.UTC().Format("2006-01-02") + " " + esc(c.Subject) + "\n")
	}
	if shown < len(s.Commits) || s.Partial {
		note := "\n_Showing " + strconv.Itoa(shown) + " of " + strconv.Itoa(len(s.Commits)) + " analyzed commits."
		if s.Partial {
			note += " Scan capped — older history not analyzed (Partial)."
		}
		b.WriteString(note + "_\n")
	}
	b.WriteString("\n")

	if len(s.Files) > 0 {
		b.WriteString("### Frequently Changed Files\n\n")
		n := len(s.Files)
		if n > maxHotFiles {
			n = maxHotFiles
		}
		for _, f := range s.Files[:n] {
			b.WriteString("- " + esc(f.Path) + " (" + strconv.Itoa(f.Changes) + ")\n")
		}
		b.WriteString("\n")
	}
}

func writeOpenTasks(b *strings.Builder, s fact.Snapshot) {
	b.WriteString("## Open Tasks\n\n")
	if len(s.Todos) == 0 {
		b.WriteString("_No TODO/FIXME comments found._\n\n")
		return
	}
	shown := len(s.Todos)
	if shown > maxOpenTasks {
		shown = maxOpenTasks
	}
	for _, td := range s.Todos[:shown] {
		b.WriteString("- " + esc(td.Path) + ":" + strconv.Itoa(td.Line) + " — " + esc(td.Text) + "\n")
	}
	if shown < len(s.Todos) {
		b.WriteString("\n_Showing " + strconv.Itoa(shown) + " of " + strconv.Itoa(len(s.Todos)) + " open tasks._\n")
	}
	b.WriteString("\n")
}

func writeRelatedModules(b *strings.Builder, s fact.Snapshot) {
	b.WriteString("## Related Modules\n\n")
	if len(s.Modules) == 0 {
		b.WriteString("_No top-level modules (flat repository)._\n\n")
		return
	}
	for _, m := range s.Modules {
		b.WriteString("- " + esc(m.Path) + "/ (" + strconv.Itoa(m.Files) + " files)\n")
	}
	b.WriteString("\n")
}

func writeWorkspaceFacts(b *strings.Builder, s fact.Snapshot) {
	b.WriteString("## Workspace Facts\n\n")
	completeness := "complete"
	if s.Partial {
		completeness = "partial — caps hit"
	}
	b.WriteString("- Commits analyzed: " + strconv.Itoa(len(s.Commits)) + " (" + completeness + ")\n")
	b.WriteString("- Open tasks (TODO/FIXME): " + strconv.Itoa(len(s.Todos)) + "\n")
	b.WriteString("- Top-level modules: " + strconv.Itoa(len(s.Modules)) + "\n")
	b.WriteString("- Docs captured: " + strconv.Itoa(len(s.Docs)) + "\n\n")
}

// Attach inserts enhancements into their slots without touching skeleton
// sections. Removing all enhancements must yield the skeleton unchanged
// (RFC-0003 §2.2 invariant — covered by a round-trip test).
func Attach(skeleton string, es []artifact.Enhancement) (string, error) {
	bySlot := map[string][]artifact.Enhancement{}
	for _, e := range es {
		if _, ok := slotTitles[e.Slot]; !ok {
			return "", fmt.Errorf("compiler: unknown slot %q", e.Slot)
		}
		// An unlabeled enhancement is a §2.3 boundary violation, not a warning.
		if e.State == "" || e.Source == "" {
			return "", fmt.Errorf("compiler: enhancement for slot %q lacks state or source label", e.Slot)
		}
		if strings.TrimSpace(e.Body) == "" {
			return "", fmt.Errorf("compiler: empty enhancement body for slot %q", e.Slot)
		}
		bySlot[e.Slot] = append(bySlot[e.Slot], e)
	}

	out := skeleton
	for _, slot := range artifact.Slots() {
		entries, ok := bySlot[slot]
		if !ok {
			continue
		}
		var parts []string
		for _, e := range entries {
			parts = append(parts, "**["+string(e.State)+" — "+e.Source+"]**\n"+strings.TrimSpace(e.Body))
		}
		var err error
		out, err = replaceSlot(out, slot, strings.Join(parts, "\n\n"))
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

// Detach restores every slot to its placeholder: Detach(Attach(s, es)) == s.
func Detach(rendered string) (string, error) {
	out := rendered
	for _, slot := range artifact.Slots() {
		var err error
		out, err = replaceSlot(out, slot, placeholderFor(slot))
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

func replaceSlot(doc, slot, content string) (string, error) {
	begin, end := artifact.SlotBegin(slot), artifact.SlotEnd(slot)
	i := strings.Index(doc, begin)
	j := strings.Index(doc, end)
	if i < 0 || j < i {
		return "", fmt.Errorf("compiler: slot %q markers missing or corrupt", slot)
	}
	return doc[:i+len(begin)] + "\n" + content + "\n" + doc[j:], nil
}

var (
	paragraphSplitRe = regexp.MustCompile(`\n\s*\n`)
	htmlTagRe        = regexp.MustCompile(`<[^>]*>`)
	mdImageRe        = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	mdLinkRe         = regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`) // keep link text
)

// minExcerptLen filters decoration paragraphs: a stripped logo block or
// badge row leaves almost no text, real prose does.
const minExcerptLen = 20

// firstParagraph extracts the first prose paragraph, working per paragraph
// and stripping HTML tags / markdown images / badge wrappers before judging
// it. Line-based skipping was not enough: prometheus's README opens with a
// multi-line <p> block whose closing line ("…guides.</p>") masqueraded as
// prose. Purely extractive — deterministic (summarizing is Semantic).
func firstParagraph(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	for _, block := range paragraphSplitRe.Split(content, -1) {
		text := strings.Join(strings.Fields(block), " ") // collapse whitespace
		if text == "" || strings.HasPrefix(text, "#") {
			continue // heading or empty
		}
		text = htmlTagRe.ReplaceAllString(text, " ")
		text = mdImageRe.ReplaceAllString(text, " ")
		text = mdLinkRe.ReplaceAllString(text, "$1")
		text = strings.Join(strings.Fields(text), " ")
		if len([]rune(text)) >= minExcerptLen {
			return capRunes(text, maxExcerptLen)
		}
	}
	return ""
}

func capRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
