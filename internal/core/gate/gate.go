// Package gate matches changed workspace paths against the anchors of
// verified observations — the deterministic core of `kervo check`. No
// git, no I/O: callers bring the changed-file list, gate brings the
// verdict. Only verified knowledge gates a diff; unsigned proposals must
// never block a build.
package gate

import (
	"path"
	"strings"

	"github.com/kervo-os/kervo/internal/core/trust"
)

// Match reports whether a workspace-relative file path matches an anchor
// pattern. Each pattern segment matches per path.Match; "**" matches any
// number of segments, including zero. A pattern with no glob
// metacharacters also matches as a directory prefix — humans write
// "services/payments", and that should cover every file under it.
func Match(pattern, p string) bool {
	pattern = strings.TrimSuffix(strings.TrimSpace(pattern), "/")
	if pattern == "" || p == "" {
		return false
	}
	if !strings.ContainsAny(pattern, "*?[") {
		return p == pattern || strings.HasPrefix(p, pattern+"/")
	}
	return matchSegs(strings.Split(pattern, "/"), strings.Split(p, "/"))
}

func matchSegs(pat, segs []string) bool {
	if len(pat) == 0 {
		return len(segs) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(segs); i++ {
			if matchSegs(pat[1:], segs[i:]) {
				return true
			}
		}
		return false
	}
	if len(segs) == 0 {
		return false
	}
	ok, err := path.Match(pat[0], segs[0])
	if err != nil || !ok {
		return false
	}
	return matchSegs(pat[1:], segs[1:])
}

// Conflict is one verified, anchored observation whose anchors intersect
// the changed set.
type Conflict struct {
	Obs   trust.Observation
	Files []string // matched changed files, in input order
}

// Conflicts gates a changed-file list against verified anchored
// observations. Order of the result follows ledger order of obs — stable
// for identical inputs.
func Conflicts(obs []trust.Observation, changed []string) []Conflict {
	var out []Conflict
	for _, o := range gated(obs) {
		var hit []string
		for _, f := range changed {
			for _, a := range o.Anchors {
				if Match(a, f) {
					hit = append(hit, f)
					break
				}
			}
		}
		if len(hit) > 0 {
			out = append(out, Conflict{Obs: o, Files: hit})
		}
	}
	return out
}

// Dead returns verified anchored observations none of whose anchors match
// any tracked file — the anchored subject is gone from the workspace,
// which is a deterministic stale signal for a human to judge (the age
// timer is no longer the only staleness channel).
func Dead(obs []trust.Observation, tracked []string) []trust.Observation {
	var out []trust.Observation
	for _, o := range gated(obs) {
		alive := false
		for _, a := range o.Anchors {
			for _, f := range tracked {
				if Match(a, f) {
					alive = true
					break
				}
			}
			if alive {
				break
			}
		}
		if !alive {
			out = append(out, o)
		}
	}
	return out
}

func gated(obs []trust.Observation) []trust.Observation {
	var out []trust.Observation
	for _, o := range obs {
		if o.State == trust.Verified && len(o.Anchors) > 0 {
			out = append(out, o)
		}
	}
	return out
}
