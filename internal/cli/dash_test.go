package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/adapters/store/jsonl"
	"github.com/kervo-os/kervo/internal/core/i18n"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// TestMain sandboxes the machine-local registry for the whole package —
// tests must never write into the developer's real ~/.kervo.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "kervo-test-state-*")
	if err != nil {
		panic(err)
	}
	os.Setenv("KERVO_STATE_DIR", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestRegistryUpsertAndValidation(t *testing.T) {
	t.Setenv("KERVO_STATE_DIR", t.TempDir())

	live := t.TempDir()
	if err := os.MkdirAll(filepath.Join(live, ".kervo"), 0o755); err != nil {
		t.Fatal(err)
	}
	registerWorkspace(live)
	registerWorkspace(live) // upsert, not duplicate
	gone := filepath.Join(t.TempDir(), "moved-away")
	registerWorkspace(gone) // path without .kervo must be filtered on read

	got := registeredWorkspaces()
	abs, _ := filepath.Abs(live)
	if len(got) != 1 || got[0] != abs {
		t.Fatalf("registeredWorkspaces = %v, want just %s", got, abs)
	}
}

func TestCompileRegistersWorkspace(t *testing.T) {
	t.Setenv("KERVO_STATE_DIR", t.TempDir())
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "# demo\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "x")

	if err := runInit([]string{"-dir", dir}); err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(dir)
	for _, w := range registeredWorkspaces() {
		if w == abs {
			return
		}
	}
	t.Fatalf("init did not register %s", abs)
}

// The dash is a user surface spanning repos, so its chrome speaks the
// user's language ($LANG / -lang) while trust-state and type tokens stay
// English — they are ledger vocabulary shared with the CLI and artifact.
func TestDashKoreanChrome(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := captureObservation(dir, "decision", "한국어 제안", "", "agent:test"); err != nil {
		t.Fatal(err)
	}
	srv, err := newDashServer([]string{dir}, "human:tester", i18n.KO)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(res.Body)
	res.Body.Close()
	for _, want := range []string{"판정 대기", "워크스페이스", "승인"} {
		if !strings.Contains(string(page), want) {
			t.Errorf("Korean chrome missing %q", want)
		}
	}
}

func TestDashLangSwitchPersists(t *testing.T) {
	t.Setenv("KERVO_STATE_DIR", t.TempDir())
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "en_US.UTF-8")
	dir := t.TempDir()
	if _, _, err := captureObservation(dir, "decision", "switch me", "", "agent:test"); err != nil {
		t.Fatal(err)
	}
	srv, err := newDashServer([]string{dir}, "human:tester", i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	// Every table ships with the page, so switching is instant client-side.
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(res.Body)
	res.Body.Close()
	for _, want := range []string{"판정 대기", "判定待ち", "awaiting judgment"} {
		if !strings.Contains(string(page), want) {
			t.Errorf("bundled tables missing %q", want)
		}
	}

	res, err = http.Post(ts.URL+"/lang", "application/json", strings.NewReader(`{"Lang":"ko"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("/lang status = %d", res.StatusCode)
	}
	// The choice outlives the session: next launch resolves to it even
	// though $LANG says English.
	if l, err := dashLang(""); err != nil || l != i18n.KO {
		t.Errorf("persisted lang = %v/%v, want ko", l, err)
	}
	res, _ = http.Post(ts.URL+"/lang", "application/json", strings.NewReader(`{"Lang":"xx"}`))
	res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("unsupported lang status = %d, want 400", res.StatusCode)
	}
}

func TestDashLangDetection(t *testing.T) {
	t.Setenv("KERVO_STATE_DIR", t.TempDir())
	t.Setenv("LC_ALL", "")
	t.Setenv("LANG", "ko_KR.UTF-8")
	if l, err := dashLang(""); err != nil || l != i18n.KO {
		t.Errorf("dashLang from $LANG = %v/%v, want ko", l, err)
	}
	if l, err := dashLang("ja"); err != nil || l != i18n.JA {
		t.Errorf("dashLang flag override = %v/%v, want ja", l, err)
	}
	if _, err := dashLang("xx"); err == nil {
		t.Error("unsupported -lang must error")
	}
	t.Setenv("LANG", "fr_FR.UTF-8")
	if l, _ := dashLang(""); l != i18n.EN {
		t.Errorf("unsupported $LANG must fall back to en, got %v", l)
	}
}

func TestDashFleetAndCrossRepoJudge(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	if _, _, err := captureObservation(a, "decision", "alpha fact <b>x</b>", "ran it", "agent:test"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := captureObservation(b, "risk", "beta fact", "", "agent:test"); err != nil {
		t.Fatal(err)
	}
	srv, err := newDashServer([]string{a, b}, "human:tester", i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.handler())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(res.Body)
	res.Body.Close()
	for _, want := range []string{"alpha fact", "beta fact", "ran it"} {
		if !strings.Contains(string(page), want) {
			t.Errorf("page missing %q", want)
		}
	}
	if strings.Contains(string(page), "<b>x</b>") {
		t.Error("body not escaped inside the script payload")
	}

	// Judge repo B's item; repo A must stay untouched.
	id := srv.byPath[b].Items[0].ID
	res, err = http.Post(ts.URL+"/judge", "application/json",
		strings.NewReader(`{"Workspace":"`+b+`","ID":"`+id+`","To":"verified","Reason":"ok"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("judge status = %d", res.StatusCode)
	}
	folder, _ := replayFolder(jsonl.Open(b))
	if o := folder.Observations()[0]; o.State != trust.Verified {
		t.Errorf("repo B state = %s, want verified", o.State)
	}
	folder, _ = replayFolder(jsonl.Open(a))
	if o := folder.Observations()[0]; o.State != trust.Generated {
		t.Errorf("repo A state = %s — cross-repo judge leaked", o.State)
	}
	if srv.pendingTotal() != 1 || srv.judgedTotal() != 1 {
		t.Errorf("pending=%d judged=%d, want 1/1", srv.pendingTotal(), srv.judgedTotal())
	}

	// Regression: a repo with nothing pending must marshal Items as [],
	// never null — null killed the page script on its first .length read.
	res, err = http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	page, _ = io.ReadAll(res.Body)
	res.Body.Close()
	if strings.Contains(string(page), `"Items":null`) {
		t.Error(`fleet JSON contains "Items":null — blank-page regression`)
	}

	// A workspace not in the fleet must be rejected — the page can only
	// route judgments to repos it was launched over.
	res, _ = http.Post(ts.URL+"/judge", "application/json",
		strings.NewReader(`{"Workspace":"/tmp/evil","ID":"`+id+`","To":"verified"}`))
	res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Errorf("foreign workspace status = %d, want 409", res.StatusCode)
	}
}

// The workspace detail carries the fact skeleton — and coupling pairs are
// proven by commit history, not narrated by a model.
func TestDashOverviewAndCoupling(t *testing.T) {
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "Makefile", "build:\n\tgo build ./...\n")
	writeFile(t, dir, "api/a.go", "package api\n")
	writeFile(t, dir, "core/b.go", "package core\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "first")
	writeFile(t, dir, "api/a.go", "package api // v2\n")
	writeFile(t, dir, "core/b.go", "package core // v2\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "api+core together")

	if _, _, err := captureObservation(dir, "decision", "pending one", "", "agent:test"); err != nil {
		t.Fatal(err)
	}
	srv, err := newDashServer([]string{dir}, "human:tester", i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	ov := srv.repos[0].Overview
	if ov == nil {
		t.Fatal("overview missing for a healthy git workspace")
	}
	if ov.Branch != "main" {
		t.Errorf("branch = %q", ov.Branch)
	}
	found := false
	for _, c := range ov.Commands {
		if c.Run == "make build" {
			found = true
		}
	}
	if !found {
		t.Errorf("commands missing make build: %+v", ov.Commands)
	}
	// Both commits touched api/ and core/ together.
	if len(ov.Links) == 0 || ov.Links[0].A != "api" || ov.Links[0].B != "core" || ov.Links[0].N != 2 {
		t.Errorf("coupling = %+v, want api<->core x2", ov.Links)
	}

	// A ledger-only workspace (no git) must still serve — overview absent.
	plain := t.TempDir()
	if _, _, err := captureObservation(plain, "note", "no git here", "", "agent:test"); err != nil {
		t.Fatal(err)
	}
	srv2, err := newDashServer([]string{plain}, "human:tester", i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	if srv2.repos[0].Overview != nil {
		t.Error("overview must be nil when the scan fails")
	}
}
