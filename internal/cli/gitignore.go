package cli

import (
	"os"
	"path/filepath"
	"strings"
)

// derivedIgnores are the RFC-0005 §2.4 rules: derived state is never
// committed — the truth is .kervo/events/ (committed), everything
// rebuildable is ignored. Registered automatically by init/compile.
var derivedIgnores = []string{
	".kervo/artifact.md",
	".kervo/cache/",
}

const ignoreHeader = "# kervo derived state (RFC-0005: events are truth, artifacts are derived)"

// ensureGitignore appends missing kervo rules to the workspace .gitignore.
// Append-only and idempotent: human content is never rewritten, and a
// second run adds nothing.
func ensureGitignore(dir string) error {
	path := filepath.Join(dir, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	present := map[string]bool{}
	for _, line := range strings.Split(string(existing), "\n") {
		present[strings.TrimSpace(line)] = true
	}
	var missing []string
	for _, rule := range derivedIgnores {
		if !present[rule] {
			missing = append(missing, rule)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	var b strings.Builder
	b.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		b.WriteString("\n")
	}
	if !present[ignoreHeader] {
		b.WriteString(ignoreHeader + "\n")
	}
	b.WriteString(strings.Join(missing, "\n") + "\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
