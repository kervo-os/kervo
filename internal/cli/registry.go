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
		} else if _, err := os.Stat(filepath.Join(w.Path, ".kervo")); os.IsNotExist(err) {
			// Prune ONLY what provably no longer exists. Any other stat
			// failure (permissions, sandboxing, unmounted volume) means
			// THIS process can't see it — not that it's gone; dropping it
			// here would erase another user context's registration.
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
	// Atomic replace: agent-velocity sessions compile several workspaces
	// concurrently, and a torn write would corrupt the whole registry.
	// A lost update is fine — the next compile re-registers.
	tmp, err := os.CreateTemp(filepath.Dir(p), ".workspaces-*")
	if err != nil {
		return
	}
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		os.Remove(tmp.Name())
		return
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		os.Remove(tmp.Name())
	}
}

// registeredWorkspaces returns registry paths that still look like kervo
// workspaces. Only a provable absence excludes a path — a permission or
// sandbox error is this process's problem, not the workspace's.
func registeredWorkspaces() []string {
	var out []string
	for _, w := range loadRegistry().Workspaces {
		if _, err := os.Stat(filepath.Join(w.Path, ".kervo")); !os.IsNotExist(err) {
			out = append(out, w.Path)
		}
	}
	return out
}
