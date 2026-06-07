package conf

import (
	"chat2api/app/token_pool"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetConfigForTest(t *testing.T) {
	t.Helper()
	setApp(defaultApp())
	token_pool.GetAccessTokenPool().ReplaceAccessTokens(nil)
	t.Cleanup(func() {
		setApp(defaultApp())
		token_pool.GetAccessTokenPool().ReplaceAccessTokens(nil)
	})
}

func TestLoadFallbackConfigGeneratesRuntimeAuthToken(t *testing.T) {
	resetConfigForTest(t)

	if err := loadFallbackConfig(context.Background()); err != nil {
		t.Fatalf("load fallback config: %v", err)
	}

	tokens := GetAuthAccessTokens()
	if len(tokens) != 1 || !strings.HasPrefix(tokens[0], "sk-") {
		t.Fatalf("expected one generated sk token, got %#v", tokens)
	}
	if got := token_pool.GetAccessTokenPool().Size(); got != 0 {
		t.Fatalf("expected empty account pool, got %d", got)
	}
}

func TestLoadConfigNormalizesAuthAndRebuildsAccountPool(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "app.dev.yaml")
	data := []byte(`
bind: 127.0.0.1
port: 3041
auth:
  access_tokens:
    - " Bearer sk-local "
    - ""
proxy: http://127.0.0.1:7890
chatgpt_base_url: https://example.com/
chatgpts:
  - access_token: " Bearer upstream-1 "
    proxy: http://127.0.0.1:7891
  - access_token: ""
  - access_token: upstream-2
`)

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := loadConfig(context.Background(), path, data, true); err != nil {
		t.Fatalf("load config: %v", err)
	}

	cfg := GetApp()
	if cfg.Bind != "127.0.0.1" || cfg.Port != 3041 {
		t.Fatalf("unexpected bind/port: %#v", cfg)
	}
	if got := GetAuthAccessTokens(); len(got) != 1 || got[0] != "sk-local" {
		t.Fatalf("expected normalized auth token, got %#v", got)
	}
	pool := token_pool.GetAccessTokenPool()
	if got := pool.Size(); got != 2 {
		t.Fatalf("expected two upstream tokens, got %d", got)
	}
	first := pool.GetAccessToken()
	if first == nil || first.Token != "Bearer upstream-1" || first.Proxy != "http://127.0.0.1:7891" {
		t.Fatalf("unexpected first upstream token: %#v", first)
	}
	second := pool.GetAccessToken()
	if second == nil || second.Token != "Bearer upstream-2" {
		t.Fatalf("unexpected second upstream token: %#v", second)
	}
}

func TestGetAppReturnsSnapshotCopy(t *testing.T) {
	resetConfigForTest(t)
	setApp(app{
		Bind: "127.0.0.1",
		Auth: auth{AccessTokens: []string{"sk-original"}},
		ChatGPTs: []chatgpt{{
			AccessToken: "upstream-original",
		}},
	})

	cfg := GetApp()
	cfg.Auth.AccessTokens[0] = "sk-mutated"
	cfg.ChatGPTs[0].AccessToken = "upstream-mutated"

	got := GetApp()
	if got.Auth.AccessTokens[0] != "sk-original" {
		t.Fatalf("auth token snapshot was mutated: %#v", got.Auth.AccessTokens)
	}
	if got.ChatGPTs[0].AccessToken != "upstream-original" {
		t.Fatalf("chatgpt snapshot was mutated: %#v", got.ChatGPTs)
	}
}

func TestLoadConfigGeneratesAndPersistsAuthToken(t *testing.T) {
	resetConfigForTest(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "app.dev.yaml")
	data := []byte("auth:\n  access_tokens: []\n")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := loadConfig(context.Background(), path, data, true); err != nil {
		t.Fatalf("load config: %v", err)
	}

	tokens := GetAuthAccessTokens()
	if len(tokens) != 1 || !strings.HasPrefix(tokens[0], "sk-") {
		t.Fatalf("expected generated auth token, got %#v", tokens)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	if !strings.Contains(string(updated), tokens[0]) {
		t.Fatalf("expected generated token to be persisted, config=%q token=%q", string(updated), tokens[0])
	}
}
