package gate

import (
	"testing"
	"time"

	"github.com/kervo-os/kervo/internal/core/trust"
)

func TestMatch(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		// bare directory = prefix (what humans write)
		{"services/payments", "services/payments/gateway.go", true},
		{"services/payments", "services/payments.go", false},
		{"services/payments", "services/payments", true},
		{"services/pay", "services/payments/gateway.go", false},
		// ** spans segments, including zero
		{"services/payments/**", "services/payments/gateway.go", true},
		{"services/payments/**", "services/payments/deep/nested/x.go", true},
		{"services/payments/**", "services/api/x.go", false},
		{"**/testdata/**", "a/b/testdata/c.txt", true},
		{"**/testdata/**", "testdata/c.txt", true},
		// per-segment globs
		{"internal/billing/*.go", "internal/billing/ledger.go", true},
		{"internal/billing/*.go", "internal/billing/sub/ledger.go", false},
		{"*.md", "README.md", true},
		{"*.md", "docs/README.md", false},
		// junk in, false out
		{"", "a.go", false},
		{"a.go", "", false},
	}
	for _, c := range cases {
		if got := Match(c.pattern, c.path); got != c.want {
			t.Errorf("Match(%q, %q) = %v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}

func obs(id string, state trust.State, anchors ...string) trust.Observation {
	return trust.Observation{ID: id, Type: "decision", Body: id + " body", State: state, Anchors: anchors}
}

func TestConflictsGatesOnVerifiedOnly(t *testing.T) {
	all := []trust.Observation{
		obs("ver", trust.Verified, "services/payments/**"),
		obs("gen", trust.Generated, "services/payments/**"), // unsigned — must not gate
		obs("stale", trust.Stale, "services/payments/**"),   // retired — must not gate
		obs("noanchor", trust.Verified),                     // nothing anchored
	}
	got := Conflicts(all, []string{"services/payments/gateway.go", "README.md"})
	if len(got) != 1 || got[0].Obs.ID != "ver" {
		t.Fatalf("want exactly the verified anchored obs, got %+v", got)
	}
	if len(got[0].Files) != 1 || got[0].Files[0] != "services/payments/gateway.go" {
		t.Fatalf("wrong matched files: %v", got[0].Files)
	}
}

func TestConflictsNoTouchNoConflict(t *testing.T) {
	all := []trust.Observation{obs("ver", trust.Verified, "services/payments/**")}
	if got := Conflicts(all, []string{"docs/notes.md"}); len(got) != 0 {
		t.Fatalf("expected no conflicts, got %+v", got)
	}
}

func TestDrifted(t *testing.T) {
	judged := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	o := obs("dec", trust.Verified, "services/payments/**")
	o.JudgedAt = judged
	fresh := obs("fresh", trust.Verified, "services/api/**")
	fresh.JudgedAt = judged

	day := 24 * time.Hour
	var changes []Change
	for i := 1; i <= 5; i++ { // 5 post-judgment commits on payments
		changes = append(changes, Change{At: judged.Add(time.Duration(i) * day),
			Files: []string{"services/payments/gateway.go"}})
	}
	// pre-judgment churn must not count
	changes = append(changes, Change{At: judged.Add(-day), Files: []string{"services/payments/old.go"}})
	// churn elsewhere must not count for either
	changes = append(changes, Change{At: judged.Add(day), Files: []string{"docs/x.md"}})

	got := Drifted([]trust.Observation{o, fresh}, changes)
	if len(got) != 1 || got[0].Obs.ID != "dec" || got[0].Commits != 5 {
		t.Fatalf("want dec drifted with 5 commits, got %+v", got)
	}

	// one commit under threshold — silence
	few := Drifted([]trust.Observation{o}, changes[:4])
	if len(few) != 0 {
		t.Fatalf("below threshold must not drift, got %+v", few)
	}
}

func TestDead(t *testing.T) {
	all := []trust.Observation{
		obs("alive", trust.Verified, "cmd/kervo/**"),
		obs("dead", trust.Verified, "ci/deploy.sh"),
	}
	tracked := []string{"cmd/kervo/main.go", "README.md"}
	got := Dead(all, tracked)
	if len(got) != 1 || got[0].ID != "dead" {
		t.Fatalf("want the dead-anchored obs only, got %+v", got)
	}
}
