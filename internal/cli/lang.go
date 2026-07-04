package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kervo-os/kervo/internal/core/i18n"
)

// resolveLang picks the artifact language: explicit flag wins, then the
// workspace's persisted choice (.kervo/lang, written on every successful
// run so re-runs stay byte-stable without repeating the flag), then English.
func resolveLang(dir, flagVal string) (i18n.Lang, error) {
	if flagVal != "" {
		return i18n.Parse(flagVal)
	}
	if raw, err := os.ReadFile(filepath.Join(dir, ".kervo", "lang")); err == nil {
		return i18n.Parse(strings.TrimSpace(string(raw)))
	}
	return i18n.EN, nil
}
