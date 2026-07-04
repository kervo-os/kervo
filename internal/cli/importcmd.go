package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/event"
)

// runImport is the Interop side of H2' (verdict: build-primary,
// import-supplementary). It back-fills the ledger from existing stores —
// today: Claude Code transcripts, which auto-delete after 30 days, so
// importing preserves history the hooks were not yet running for.
//
// Deterministic extraction only: prompt SIZES (retroactive H3 samples,
// flagged artifact-unknown so they never pollute the A/B sides) and a
// per-session file-operation summary. No content ever enters the ledger;
// observation mining from prompt text is semantic territory (Mode 2/3).
func runImport(args []string) error {
	if len(args) == 0 || args[0] != "claude" {
		return fmt.Errorf("import: supported sources: claude (usage: kervo import claude [-dir .] [-project <path>])")
	}
	fs := newFlagSet("import claude")
	dir := fs.String("dir", ".", "workspace directory (ledger destination)")
	project := fs.String("project", "", "project path whose transcripts to import (default: -dir)")
	from := fs.String("from", "", "transcripts directory (default: ~/.claude/projects/<sanitized project>)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	tdir := *from
	if tdir == "" {
		proj := *project
		if proj == "" {
			proj = mustAbs(*dir)
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		tdir = filepath.Join(home, ".claude", "projects", sanitizeProjectPath(proj))
	}
	matches, err := filepath.Glob(filepath.Join(tdir, "*.jsonl"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("import: no transcripts found in %s", tdir)
	}
	sort.Strings(matches)

	store := jsonl.Open(*dir)
	already := map[string]bool{}
	if err := store.Replay(context.Background(), "", func(e event.Event) error {
		if e.Type == "import:session" {
			already[e.Ref] = true
		}
		return nil
	}); err != nil {
		return err
	}

	repo := filepath.Base(mustAbs(*dir))
	imported, skipped, prompts := 0, 0, 0
	for _, path := range matches {
		session := strings.TrimSuffix(filepath.Base(path), ".jsonl")
		if already[session] {
			skipped++
			continue
		}
		n, err := importSession(store, repo, session, path)
		if err != nil {
			return fmt.Errorf("import: %s: %w", filepath.Base(path), err)
		}
		imported++
		prompts += n
	}
	fmt.Printf("imported %d session(s) (%d prompts as retroactive metrics), %d already in ledger\n",
		imported, prompts, skipped)
	return nil
}

// transcriptLine is the subset of Claude Code's (undocumented, drifting)
// transcript schema we rely on — everything else is ignored.
type transcriptLine struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	IsMeta    bool   `json:"isMeta"`
	Message   struct {
		Content any `json:"content"` // string (typed prompt) or []block
	} `json:"message"`
}

func importSession(store *jsonl.Store, repo, session, path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 8<<20)
	var started, ended, lastTS time.Time
	promptCount, fileOps := 0, 0
	filesTouched := map[string]bool{}

	for sc.Scan() {
		var l transcriptLine
		if json.Unmarshal(sc.Bytes(), &l) != nil {
			continue // foreign line shapes are expected, never fatal
		}
		if ts, err := time.Parse(time.RFC3339, l.Timestamp); err == nil {
			lastTS = ts
			if started.IsZero() {
				started = ts
			}
			ended = ts
		}
		switch c := l.Message.Content.(type) {
		case string: // a typed user prompt
			if l.Type != "user" || l.IsMeta || strings.TrimSpace(c) == "" {
				continue
			}
			promptCount++
			m, err := json.Marshal(map[string]any{
				"session":        session,
				"prompt_chars":   len(c),
				"prompt_words":   len(strings.Fields(c)),
				"artifact_known": false, // retroactive — must not enter the A/B sides
			})
			if err != nil {
				return promptCount, err
			}
			if _, err := store.Append(context.Background(), event.Event{
				Kind: event.KindFact, Type: "metric:prompt", Repo: repo,
				At: lastTS, Actor: "system", Source: "import:claude-transcript",
				Ref: session, Payload: json.RawMessage(m),
			}); err != nil {
				return promptCount, err
			}
		case []any: // assistant/tool blocks — collect file operations
			for _, item := range c {
				b, ok := item.(map[string]any)
				if !ok || b["type"] != "tool_use" {
					continue
				}
				name, _ := b["name"].(string)
				if name != "Edit" && name != "Write" {
					continue
				}
				fileOps++
				if in, ok := b["input"].(map[string]any); ok {
					if fp, ok := in["file_path"].(string); ok && fp != "" {
						filesTouched[fp] = true
					}
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return promptCount, err
	}

	files := make([]string, 0, len(filesTouched))
	for fp := range filesTouched {
		files = append(files, fp)
	}
	sort.Strings(files)
	if len(files) > 30 {
		files = files[:30]
	}
	summary, err := json.Marshal(map[string]any{
		"session": session,
		"started": started.UTC().Format(time.RFC3339),
		"ended":   ended.UTC().Format(time.RFC3339),
		"prompts": promptCount,
		"fileops": fileOps,
		"files":   files,
	})
	if err != nil {
		return promptCount, err
	}
	_, err = store.Append(context.Background(), event.Event{
		Kind: event.KindFact, Type: "import:session", Repo: repo,
		At: ended, Actor: "system", Source: "import:claude-transcript",
		Ref: session, Payload: json.RawMessage(summary),
	})
	return promptCount, err
}

// sanitizeProjectPath mirrors Claude Code's project-directory naming:
// every character outside [A-Za-z0-9-] becomes "-".
func sanitizeProjectPath(p string) string {
	var b strings.Builder
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}
