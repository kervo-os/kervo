package openaicompat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/i18n"
	"github.com/kervo-os/kervo/internal/core/trust"
)

// fake returns an OpenAI-compatible server that captures the request and
// replies with content wrapped the way sloppy backends wrap it.
func fake(t *testing.T, status int, content string, gotReq *map[string]any, gotAuth *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		*gotAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, gotReq)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": content}}},
		})
	}))
}

func TestProposeParsesWrappedJSON(t *testing.T) {
	var req map[string]any
	var auth string
	srv := fake(t, 200, "Sure! Here are the proposals:\n```json\n"+
		`[{"slot":"goal","body":"Ship X. Evidence: Recent Changes."},{"slot":"risks","body":"Y is fragile."}]`+
		"\n```", &req, &auth)
	defer srv.Close()

	p := &Provider{BaseURL: srv.URL, Model: "test-model", APIKey: "sk-test", Lang: i18n.KO}
	es, err := p.Propose(context.Background(), "## Repository Summary\nfacts here", fact.Snapshot{})
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 2 {
		t.Fatalf("proposals = %d, want 2", len(es))
	}
	if es[0].State != trust.Generated || es[0].Source != "backend:test-model" {
		t.Errorf("labeling wrong: %+v", es[0])
	}
	if auth != "Bearer sk-test" {
		t.Errorf("auth header = %q", auth)
	}
	msgs := req["messages"].([]any)
	sys := msgs[0].(map[string]any)["content"].(string)
	if !strings.Contains(sys, "Korean") {
		t.Error("system prompt does not carry the artifact language")
	}
	user := msgs[1].(map[string]any)["content"].(string)
	if !strings.Contains(user, "Repository Summary") {
		t.Error("skeleton facts not sent to backend")
	}
	if req["temperature"].(float64) != 0 {
		t.Error("temperature not pinned to 0")
	}
}

func TestProposeRejectsUnknownSlotAndErrors(t *testing.T) {
	var req map[string]any
	var auth string
	srv := fake(t, 200, `[{"slot":"repository-summary","body":"rewrite!"}]`, &req, &auth)
	defer srv.Close()
	p := &Provider{BaseURL: srv.URL, Model: "m"}
	if _, err := p.Propose(context.Background(), "s", fact.Snapshot{}); err == nil {
		t.Fatal("skeleton-section slot must be rejected")
	}

	srv2 := fake(t, 500, "boom", &req, &auth)
	defer srv2.Close()
	p2 := &Provider{BaseURL: srv2.URL, Model: "m"}
	if _, err := p2.Propose(context.Background(), "s", fact.Snapshot{}); err == nil {
		t.Fatal("HTTP 500 must be an error (compile demotes, not us)")
	}
}

func TestFromEnv(t *testing.T) {
	t.Setenv(EnvURL, "")
	if p, err := FromEnv(i18n.EN); p != nil || err != nil {
		t.Errorf("unset URL: want (nil, nil), got (%v, %v)", p, err)
	}
	t.Setenv(EnvURL, "http://localhost:11434/v1/")
	t.Setenv(EnvModel, "")
	if _, err := FromEnv(i18n.EN); err == nil {
		t.Error("URL without model must be a config error")
	}
	t.Setenv(EnvModel, "llama3.2")
	p, err := FromEnv(i18n.KO)
	if err != nil || p == nil {
		t.Fatalf("valid env: %v %v", p, err)
	}
	if p.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("trailing slash not trimmed: %q", p.BaseURL)
	}
}
