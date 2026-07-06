package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	consumerClaude = "CLAUDE.md"
	consumerAgents = "AGENTS.md"
)

func resolveConsumersForInit(dir, flagVal string) ([]string, error) {
	if flagVal != "" {
		return parseConsumers(dir, flagVal)
	}
	if saved, ok, err := readConsumers(dir); err != nil {
		return nil, err
	} else if ok {
		return saved, nil
	}
	if stdinIsTTY() {
		return promptConsumers()
	}
	return legacyConsumers(dir)
}

func resolveConsumersForCompile(dir, flagVal string) ([]string, error) {
	if flagVal != "" {
		return parseConsumers(dir, flagVal)
	}
	if saved, ok, err := readConsumers(dir); err != nil {
		return nil, err
	} else if ok {
		return saved, nil
	}
	return legacyConsumers(dir)
}

func readConsumers(dir string) ([]string, bool, error) {
	raw, err := os.ReadFile(filepath.Join(dir, ".kervo", "consumers"))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	consumers, err := parseConsumers(dir, string(raw))
	if err != nil {
		return nil, false, err
	}
	return consumers, true, nil
}

func legacyConsumers(dir string) ([]string, error) {
	consumers := []string{consumerClaude}
	if _, err := os.Stat(filepath.Join(dir, consumerAgents)); err == nil {
		consumers = append(consumers, consumerAgents)
	}
	return consumers, nil
}

func parseConsumers(dir, raw string) ([]string, error) {
	_ = dir // kept for symmetry with legacy parsing and future file-aware modes.
	t := strings.ToLower(strings.TrimSpace(raw))
	t = strings.ReplaceAll(t, "\n", ",")
	t = strings.ReplaceAll(t, " ", "")
	if t == "" || t == "auto" {
		return legacyConsumers(dir)
	}
	if t == "both" || t == "all" {
		return []string{consumerClaude, consumerAgents}, nil
	}
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	for _, part := range strings.Split(t, ",") {
		switch part {
		case "", "none":
			continue
		case "claude", "claudecode", "claude.md":
			add(consumerClaude)
		case "codex", "agents", "agents.md":
			add(consumerAgents)
		default:
			return nil, fmt.Errorf("consumers: unsupported target %q (supported: claude, codex, both, auto)", part)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("consumers: no injection targets selected")
	}
	// Keep output stable even if the user wrote "codex,claude".
	if seen[consumerClaude] && seen[consumerAgents] {
		return []string{consumerClaude, consumerAgents}, nil
	}
	return out, nil
}

func stdinIsTTY() bool {
	in, inErr := os.Stdin.Stat()
	out, outErr := os.Stdout.Stat()
	return inErr == nil && outErr == nil &&
		in.Mode()&os.ModeCharDevice != 0 &&
		out.Mode()&os.ModeCharDevice != 0
}

func promptConsumers() ([]string, error) {
	fmt.Fprintln(os.Stderr, "Which agent files should kervo inject?")
	fmt.Fprintln(os.Stderr, "  1) Claude Code  -> CLAUDE.md")
	fmt.Fprintln(os.Stderr, "  2) Codex/agents -> AGENTS.md")
	fmt.Fprintln(os.Stderr, "  3) Both         -> CLAUDE.md + AGENTS.md")
	fmt.Fprint(os.Stderr, "Select [3]: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		return nil, err
	}
	switch strings.TrimSpace(strings.ToLower(line)) {
	case "", "3", "both", "all":
		return []string{consumerClaude, consumerAgents}, nil
	case "1", "claude", "claude code", "claudecode":
		return []string{consumerClaude}, nil
	case "2", "codex", "agents":
		return []string{consumerAgents}, nil
	default:
		return parseConsumers(".", line)
	}
}
