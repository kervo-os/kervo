package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// The registry is a machine-local list of workspace PATHS — nothing else —
// so `kervo dash` can survey every initialized repo. It lives outside any
// repository (~/.kervo), is never committed, and losing it costs nothing:
// the next init/compile in a workspace re-registers it.

type registryEntry struct {
	Path     string    `json:"path"`
	LastSeen time.Time `json:"lastSeen"`
}

type registryFile struct {
	Workspaces []registryEntry `json:"workspaces"`
}

// stateDir is overridable for tests (and users who dislike ~/.kervo).
func stateDir() string {
	if d := os.Getenv("KERVO_STATE_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kervo")
}

func registryPath() string {
	sd := stateDir()
	if sd == "" {
		return ""
	}
	return filepath.Join(sd, "workspaces.json")
}

func loadRegistry() registryFile {
	var reg registryFile
	p := registryPath()
	if p == "" {
		return reg
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return reg
	}
	_ = json.Unmarshal(raw, &reg)
	return reg
}

// registerWorkspace upserts dir into the registry. Best-effort by design:
// registration must never fail a compile, so errors are swallowed.
func registerWorkspace(dir string) {
	p := registryPath()
	if p == "" {
		return
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	reg := loadRegistry()
	// Prune dead entries while we are here: a path without .kervo was
	// deleted or moved, and re-registration is free on its next compile.
	kept := reg.Workspaces[:0]
	found := false
	for i := range reg.Workspaces {
		w := reg.Workspaces[i]
		if w.Path == abs {
			w.LastSeen = time.Now().UTC()
			found = true
		} else if _, err := os.Stat(filepath.Join(w.Path, ".kervo")); err != nil {
			continue
		}
		kept = append(kept, w)
	}
	reg.Workspaces = kept
	if !found {
		reg.Workspaces = append(reg.Workspaces, registryEntry{Path: abs, LastSeen: time.Now().UTC()})
	}
	sort.Slice(reg.Workspaces, func(i, j int) bool { return reg.Workspaces[i].Path < reg.Workspaces[j].Path })
	raw, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(p, raw, 0o644)
}

// registeredWorkspaces returns registry paths that still look like kervo
// workspaces (a stale path is skipped, not an error — repos move).
func registeredWorkspaces() []string {
	var out []string
	for _, w := range loadRegistry().Workspaces {
		if _, err := os.Stat(filepath.Join(w.Path, ".kervo")); err == nil {
			out = append(out, w.Path)
		}
	}
	return out
}
