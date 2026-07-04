package openaicompat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kervo-os/kervo/internal/core/artifact"
	"github.com/kervo-os/kervo/internal/core/fact"
	"github.com/kervo-os/kervo/internal/core/i18n"
	"github.com/kervo-os/kervo/internal/core/trust"
	"github.com/kervo-os/kervo/internal/ports"
)

// Env configuration (RFC-0003 §3.2: Mode 3 holds only user-provided backend
// credentials — a local Ollama needs none).
const (
	EnvURL   = "KERVO_SEMANTIC_URL"   // e.g. http://localhost:11434/v1
	EnvModel = "KERVO_SEMANTIC_MODEL" // e.g. llama3.2, gpt-4o-mini
	EnvKey   = "KERVO_SEMANTIC_KEY"   // optional (Bearer)
)

const (
	defaultTimeout  = 120 * time.Second // the LLM is the real bottleneck — off the 30s Mode-1 path
	maxSkeletonSend = 12000             // chars of skeleton context sent to the backend
	maxBodyRunes    = 800               // per-proposal cap; slots are context, not essays
	maxProposals    = 4
)

// Provider implements ports.SemanticProvider against any OpenAI-compatible
// chat-completions endpoint (Ollama, LM Studio, OpenRouter, OpenAI, ...).
type Provider struct {
	BaseURL string
	Model   string
	APIKey  string
	Lang    i18n.Lang
	Client  *http.Client
	Timeout time.Duration
}

var _ ports.SemanticProvider = (*Provider)(nil)

// FromEnv returns a configured provider, or (nil, nil) when Mode 3 is not
// configured — absence is not an error (RFC-0003 §4 fallback order).
func FromEnv(lang i18n.Lang) (*Provider, error) {
	url := strings.TrimRight(os.Getenv(EnvURL), "/")
	if url == "" {
		return nil, nil
	}
	model := os.Getenv(EnvModel)
	if model == "" {
		return nil, fmt.Errorf("openaicompat: %s is set but %s is empty", EnvURL, EnvModel)
	}
	return &Provider{
		BaseURL: url,
		Model:   model,
		APIKey:  os.Getenv(EnvKey),
		Lang:    lang,
	}, nil
}

// Propose asks the backend for slot proposals grounded in the skeleton.
// Contract per RFC-0003 §5: output is Generated-only, provenance-labeled,
// and can never rewrite skeleton sections (Attach enforces the boundary).
func (p *Provider) Propose(ctx context.Context, skeleton string, snap fact.Snapshot) ([]artifact.Enhancement, error) {
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(skeleton) > maxSkeletonSend {
		skeleton = skeleton[:maxSkeletonSend]
	}
	reqBody, err := json.Marshal(map[string]any{
		"model":       p.Model,
		"temperature": 0,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt(p.Lang)},
			{"role": "user", "content": "FACT CONTEXT:\n\n" + skeleton},
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	client := p.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openaicompat: backend call failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openaicompat: backend returned %s: %s", resp.Status, truncate(string(raw), 200))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Choices) == 0 {
		return nil, fmt.Errorf("openaicompat: unexpected response shape: %s", truncate(string(raw), 200))
	}
	return p.parseProposals(out.Choices[0].Message.Content)
}

// systemPrompt encodes the grounding discipline (borrowed as concepts, not
// code, from field-tested tools): claims must cite the provided facts, the
// goal is a proposal needing confirmation, output is strict JSON.
func systemPrompt(lang i18n.Lang) string {
	langName := map[i18n.Lang]string{i18n.EN: "English", i18n.KO: "Korean", i18n.JA: "Japanese"}[lang]
	if langName == "" {
		langName = "English"
	}
	return `You analyze a repository's FACT CONTEXT (deterministically extracted) and propose observations for an AI-agent context document.

Rules:
1. Use ONLY the provided facts. Do not invent files, features, or intent.
2. Every claim must cite its evidence from the facts (e.g. "Evidence: Recent Changes 07-01..07-03").
3. The goal is a PROPOSAL: phrase it as needing confirmation, never as established truth.
4. Write the body text in ` + langName + `.
5. Output ONLY a JSON array, no prose, no code fences:
   [{"slot":"goal","body":"..."},{"slot":"summaries","body":"..."},{"slot":"risks","body":"..."}]
   Allowed slots: goal, summaries, risks, decisions. 1 to 4 items. Each body under 500 characters.`
}

// parseProposals tolerates prose/code-fence wrapping around the JSON array —
// backends differ in how strictly they honor "JSON only".
func (p *Provider) parseProposals(content string) ([]artifact.Enhancement, error) {
	i := strings.Index(content, "[")
	j := strings.LastIndex(content, "]")
	if i < 0 || j <= i {
		return nil, fmt.Errorf("openaicompat: no JSON array in backend output: %s", truncate(content, 200))
	}
	var items []struct {
		Slot string `json:"slot"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal([]byte(content[i:j+1]), &items); err != nil {
		return nil, fmt.Errorf("openaicompat: proposal JSON invalid: %w", err)
	}
	if len(items) > maxProposals {
		items = items[:maxProposals]
	}
	valid := map[string]bool{}
	for _, s := range artifact.Slots() {
		valid[s] = true
	}
	var es []artifact.Enhancement
	for n, it := range items {
		if !valid[it.Slot] {
			return nil, fmt.Errorf("openaicompat: proposal %d targets unknown slot %q", n, it.Slot)
		}
		body := strings.TrimSpace(it.Body)
		if body == "" {
			return nil, fmt.Errorf("openaicompat: proposal %d has empty body", n)
		}
		es = append(es, artifact.Enhancement{
			Slot:   it.Slot,
			Body:   truncate(body, maxBodyRunes),
			State:  trust.Generated, // by construction — backends cannot self-promote
			Source: "backend:" + p.Model,
		})
	}
	return es, nil
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
