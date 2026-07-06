// Package compiler is the process that turns a fact snapshot plus
// experience into a Context Artifact.
//
// HARD BOUNDARY (RFC-0003 §2, enforced by `make arch-check` + review):
// this package must never import LLM clients or adapters. Skeleton output
// is byte-identical for identical input — guarded by tests. The language
// is part of the input: per-language golden files pin determinism.
package compiler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/i18n"
)

// Display caps. Snapshot keeps more; the artifact stays prompt-sized.
// Constants are part of the byte-identity contract — changing them changes
// the golden files.
const (
	maxRecentCommits = 20
	maxOpenTasks     = 30
	maxHotFiles      = 10
	maxExcerptLen    = 300
)

// slotTitleKeys maps each enhancement slot to its heading's i18n key.
var slotTitleKeys = map[string]string{
	artifact.SlotGoal:      "slot.goal",
	artifact.SlotDecisions: "slot.decisions",
	artifact.SlotRisks:     "slot.risks",
	artifact.SlotSummaries: "slot.summaries",
}

// placeholderFor is the empty-slot content. Detach restores it, which is
// what makes the round-trip invariant byte-exact.
func placeholderFor(slot string, lang i18n.Lang) string {
	if slot == artifact.SlotGoal {
		return i18n.T(lang, "ph.goal")
	}
	return i18n.T(lang, "ph.generic")
}

// langRe extracts the language tag from the artifact header so Detach can
// restore the right placeholders without being told the language.
var langRe = regexp.MustCompile(`<!-- kervo:artifact v1 skeleton=fact-only lang=([a-z]{2}) -->`)

// BuildSkeleton renders the deterministic Mode-1 artifact in lang.
// Input -> template -> markdown. No LLM, no clock beyond snapshot data.
func BuildSkeleton(s fact.Snapshot, lang i18n.Lang) (string, error) {
	tr := func(key string) string { return i18n.T(lang, key) }
	var b strings.Builder

	b.WriteString("<!-- kervo:artifact v1 skeleton=fact-only lang=" + string(lang) + " -->\n")
	b.WriteString("# Context Artifact\n\n")
	b.WriteString(tr("hdr.quote") + "\n\n")

	writeBrief(&b, s, tr)
	writeRepoSummary(&b, s, tr)
	writeCommands(&b, s, tr)
	writeRecentChanges(&b, s, tr)
	writeOpenTasks(&b, s, tr)
	writeRelatedModules(&b, s, tr)
	writeWorkspaceFacts(&b, s, tr)

	for _, slot := range artifact.Slots() {
		b.WriteString("## " + tr(slotTitleKeys[slot]) + "\n\n")
		b.WriteString(artifact.SlotBegin(slot) + "\n")
		b.WriteString(placeholderFor(slot, lang) + "\n")
		b.WriteString(artifact.SlotEnd(slot) + "\n\n")
	}

	b.WriteString("## " + tr("sec.stale") + "\n\n")
	b.WriteString(artifact.SlotBegin(staleSlot) + "\n")
	b.WriteString(tr("stale.empty") + "\n")
	b.WriteString(artifact.SlotEnd(staleSlot) + "\n\n")

	// The write-back flywheel: the artifact instructs its consumers to
	// return durable facts it lacks, so exploration cost is paid once and
	// amortized across every later session (goal 01KWTJGS).
	b.WriteString("## " + tr("sec.writeback") + "\n\n")
	b.WriteString(tr("writeback.body") + "\n")

	return b.String(), nil
}

// staleSlot is compiler-owned: providers cannot target it (it is not in
// slotTitleKeys), only the trust view fills it via AttachStale.
const staleSlot = "stale"

// StaleNote is one excluded observation, shown with its exclusion reason
// instead of being silently dropped (PRD §7.2 treatment table).
type StaleNote struct {
	Body   string
	Reason string
	Actor  string // who demoted it (often "system")
}

// AttachStale fills the Deprecated/Stale section. Empty notes = no-op
// (the placeholder stays, keeping Detach round-trips byte-exact).
func AttachStale(doc string, notes []StaleNote) (string, error) {
	if len(notes) == 0 {
		return doc, nil
	}
	var lines []string
	for _, n := range notes {
		lines = append(lines, "- **[stale — "+esc(n.Reason)+"]** "+esc(n.Body)+" _("+esc(n.Actor)+")_")
	}
	return replaceSlot(doc, staleSlot, strings.Join(lines, "\n"))
}

// esc neutralizes the reserved marker prefix inside snapshot-derived text.
// Every data write site below MUST go through it: workspace content that
// impersonates a marker corrupts slot replacement and CLAUDE.md injection.
func esc(s string) string {
	return strings.ReplaceAll(s, artifact.ReservedPrefix, `<!-- kervo\:`)
}

func writeRepoSummary(b *strings.Builder, s fact.Snapshot, tr func(string) string) {
	b.WriteString("## " + tr("sec.summary") + "\n\n")
	b.WriteString("- " + tr("lbl.name") + ": " + orDash(esc(s.Repo.Name)) + "\n")
	b.WriteString("- " + tr("lbl.branch") + ": " + orDash(esc(s.Repo.Branch)) + "\n")
	b.WriteString("- " + tr("lbl.languages") + ": " + orDash(esc(strings.Join(s.Repo.Languages, ", "))) + "\n")
	b.WriteString("- " + tr("lbl.frameworks") + ": " + orDash(esc(strings.Join(s.Repo.Frameworks, ", "))) + "\n")
	var docNames []string
	for _, d := range s.Docs {
		docNames = append(docNames, d.Path)
	}
	// Monorepos can carry a doc per module; keep the summary line scannable.
	const maxDocsListed = 12
	docsLine := strings.Join(docNames, ", ")
	if len(docNames) > maxDocsListed {
		docsLine = strings.Join(docNames[:maxDocsListed], ", ") +
			" (+" + strconv.Itoa(len(docNames)-maxDocsListed) + ")"
	}
	b.WriteString("- " + tr("lbl.docs") + ": " + orDash(esc(docsLine)) + "\n\n")

	for _, d := range s.Docs {
		if d.Path != "README.md" && d.Path != "README" {
			continue
		}
		if ex := firstParagraph(d.Content); ex != "" {
			b.WriteString("### " + esc(d.Path) + " " + tr("excerpt.suffix") + "\n\n")
			b.WriteString("> " + esc(ex) + "\n\n")
		}
		break
	}
}

// writeCommands surfaces workspace-declared entry points. Evidence for the
// section: Environment/commands is the most prevalent hand-written context
// type (72% of 401 repos, arXiv:2512.18925) — the thing users re-explain most.
func writeCommands(b *strings.Builder, s fact.Snapshot, tr func(string) string) {
	b.WriteString("## " + tr("sec.commands") + "\n\n")
	if len(s.Commands) == 0 {
		b.WriteString(tr("commands.empty") + "\n\n")
		return
	}
	for _, c := range s.Commands {
		line := "- `" + esc(c.Run) + "`"
		if c.Detail != "" {
			line += " — " + esc(c.Detail)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
}

func writeRecentChanges(b *strings.Builder, s fact.Snapshot, tr func(string) string) {
	b.WriteString("## " + tr("sec.recent") + "\n\n")
	if len(s.Commits) == 0 {
		b.WriteString(tr("recent.empty") + "\n\n")
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
		note := "\n_" + fmt.Sprintf(tr("recent.showing"), shown, len(s.Commits))
		if s.Partial {
			note += tr("recent.capped")
		}
		b.WriteString(note + "_\n")
	}
	b.WriteString("\n")

	if len(s.Files) > 0 {
		b.WriteString("### " + tr("sec.hotfiles") + "\n\n")
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

func writeOpenTasks(b *strings.Builder, s fact.Snapshot, tr func(string) string) {
	b.WriteString("## " + tr("sec.tasks") + "\n\n")
	if len(s.Todos) == 0 {
		b.WriteString(tr("tasks.empty") + "\n\n")
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
		b.WriteString("\n_" + fmt.Sprintf(tr("tasks.showing"), shown, len(s.Todos)) + "_\n")
	}
	b.WriteString("\n")
}

func writeRelatedModules(b *strings.Builder, s fact.Snapshot, tr func(string) string) {
	b.WriteString("## " + tr("sec.modules") + "\n\n")
	if len(s.Modules) == 0 {
		b.WriteString(tr("modules.empty") + "\n\n")
		return
	}
	for _, m := range s.Modules {
		b.WriteString(fmt.Sprintf(tr("modules.line"), esc(m.Path), m.Files) + "\n")
	}
	b.WriteString("\n")
}

func writeWorkspaceFacts(b *strings.Builder, s fact.Snapshot, tr func(string) string) {
	b.WriteString("## " + tr("sec.facts") + "\n\n")
	completeness := tr("facts.complete")
	if s.Partial {
		completeness = tr("facts.partial")
	}
	b.WriteString("- " + tr("facts.commits") + ": " + strconv.Itoa(len(s.Commits)) + " (" + completeness + ")\n")
	b.WriteString("- " + tr("facts.tasks") + ": " + strconv.Itoa(len(s.Todos)) + "\n")
	b.WriteString("- " + tr("facts.modules") + ": " + strconv.Itoa(len(s.Modules)) + "\n")
	b.WriteString("- " + tr("facts.docs") + ": " + strconv.Itoa(len(s.Docs)) + "\n\n")
}

// Attach inserts enhancements into their slots without touching skeleton
// sections. Removing all enhancements must yield the skeleton unchanged
// (RFC-0003 §2.2 invariant — covered by a round-trip test).
// Trust-state labels are protocol tokens and are never localized.
func Attach(skeleton string, es []artifact.Enhancement) (string, error) {
	bySlot := map[string][]artifact.Enhancement{}
	for _, e := range es {
		if _, ok := slotTitleKeys[e.Slot]; !ok {
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
			label := "**[" + string(e.State) + " — " + e.Source + "]**"
			if e.Conflict {
				label += " ⚠ conflict"
			}
			parts = append(parts, label+"\n"+strings.TrimSpace(e.Body))
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
// The language is read from the artifact's own header tag.
func Detach(rendered string) (string, error) {
	lang := i18n.EN
	if m := langRe.FindStringSubmatch(rendered); m != nil {
		parsed, err := i18n.Parse(m[1])
		if err != nil {
			return "", fmt.Errorf("compiler: artifact declares %v", err)
		}
		lang = parsed
	}
	out := rendered
	for _, slot := range artifact.Slots() {
		var err error
		out, err = replaceSlot(out, slot, placeholderFor(slot, lang))
		if err != nil {
			return "", err
		}
	}
	// The compiler-owned stale region round-trips too.
	out, err := replaceSlot(out, staleSlot, i18n.T(lang, "stale.empty"))
	if err != nil {
		return "", err
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
	codeFenceRe      = regexp.MustCompile("(?s)```.*?(```|\\z)")
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
	// Fenced code is never prose — a README opening with an ASCII-art logo
	// block otherwise becomes the "excerpt" (found by self-scan).
	content = codeFenceRe.ReplaceAllString(content, "\n\n")
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
