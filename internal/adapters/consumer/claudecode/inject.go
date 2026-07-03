package claudecode

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/ports"
)

// Injector writes the rendered artifact into CLAUDE.md between the kervo
// markers. Everything outside the markers is human-owned and preserved
// byte-for-byte (ARCH-0001 §4: never overwrite the whole file).
type Injector struct {
	FileName string // defaults to CLAUDE.md
}

var _ ports.ConsumerInjector = Injector{}

func (i Injector) Inject(ctx context.Context, workspaceDir, rendered string) error {
	path, content, err := i.Render(workspaceDir, rendered)
	if err != nil {
		return err
	}
	return i.Apply(path, content)
}

// Render computes the updated consumer file without touching disk, so
// callers can validate the injection (corrupt markers, unreadable file)
// BEFORE writing any other output — no partially-applied init runs.
func (i Injector) Render(workspaceDir, rendered string) (path, content string, err error) {
	name := i.FileName
	if name == "" {
		name = "CLAUDE.md"
	}
	path = filepath.Join(workspaceDir, name)

	block := artifact.MarkerBegin + "\n" + strings.TrimRight(rendered, "\n") + "\n" + artifact.MarkerEnd

	existing, err := os.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		return path, block + "\n", nil
	case err != nil:
		return "", "", fmt.Errorf("claudecode: read %s: %w", name, err)
	}

	current := string(existing)
	begin := strings.Index(current, artifact.MarkerBegin)
	end := strings.Index(current, artifact.MarkerEnd)

	switch {
	case begin >= 0 && end > begin:
		return path, current[:begin] + block + current[end+len(artifact.MarkerEnd):], nil
	case begin < 0 && end < 0:
		sep := "\n"
		if !strings.HasSuffix(current, "\n") {
			sep = "\n\n"
		} else if !strings.HasSuffix(current, "\n\n") {
			sep = "\n"
		}
		return path, current + sep + block + "\n", nil
	default:
		// One marker without its pair: refuse to guess at human content.
		return "", "", fmt.Errorf("claudecode: %s has corrupt kervo markers — fix or remove them, then re-run", name)
	}
}

// Apply writes content produced by Render.
func (i Injector) Apply(path, content string) error {
	return writeAtomic(path, content)
}

// writeAtomic goes through a temp file + rename so a crash mid-write can
// never leave a half-written CLAUDE.md.
func writeAtomic(path, content string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".kervo-inject-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
