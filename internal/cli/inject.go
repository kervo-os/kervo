package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Injection modes (decision 01KWTFTX, user-verified 2026-07-06).
// block (default): the full artifact lives inside the CLAUDE.md marker
// block — a fresh clone reads context with zero commands, the product's
// proof. import: the block carries a single `@.kervo/artifact.md` line for
// clean-CLAUDE.md teams — the accepted trade-off is that fresh clones see
// nothing until `kervo compile` regenerates the (gitignored) artifact.
const (
	injectBlock  = "block"
	injectImport = "import"
)

// importLine is what Claude Code expands at load time in import mode.
const importLine = "@.kervo/artifact.md"

// resolveInject mirrors resolveLang: flag wins, then the workspace's
// persisted choice (.kervo/inject — committed, it is a team decision),
// then the block default.
func resolveInject(dir, flagVal string) (string, error) {
	v := flagVal
	if v == "" {
		raw, err := os.ReadFile(filepath.Join(dir, ".kervo", "inject"))
		if err != nil {
			return injectBlock, nil
		}
		v = strings.TrimSpace(string(raw))
	}
	switch v {
	case "", injectBlock:
		return injectBlock, nil
	case injectImport:
		return injectImport, nil
	default:
		return "", fmt.Errorf("inject: unsupported mode %q (supported: block, import)", v)
	}
}
