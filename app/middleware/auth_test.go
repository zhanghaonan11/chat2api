package middleware

import (
	"chat2api/app/conf"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupAuthTest(t *testing.T, config string) *gin.Engine {
	t.Helper()
	dir := t.TempDir()
	confDir := filepath.Join(dir, "conf")
	if err := os.MkdirAll(confDir, 0700); err != nil {
		t.Fatalf("mkdir conf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(confDir, "app.dev.yaml"), []byte(config), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Chdir(dir)
	cleanup := conf.Init(context.Background())
	t.Cleanup(func() {
		cleanup(context.Background())
	})

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(V1Auth)
	engine.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return engine
}

func TestV1AuthRejectsMissingAuthorization(t *testing.T) {
	engine := setupAuthTest(t, `
auth:
  access_tokens:
    - sk-local
`)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestV1AuthAcceptsConfiguredToken(t *testing.T) {
	engine := setupAuthTest(t, `
auth:
  access_tokens:
    - sk-local
`)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer sk-local")
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestV1AuthRejectsWhenConfiguredTokenListIsEmptyAtRuntime(t *testing.T) {
	engine := setupAuthTest(t, `
auth:
  access_tokens: []
`)
	tokens := conf.GetAuthAccessTokens()
	if len(tokens) != 1 {
		t.Fatalf("expected generated fallback auth token, got %#v", tokens)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestV1AuthAllowsDirectAccessToken(t *testing.T) {
	engine := setupAuthTest(t, `
auth:
  access_tokens:
    - sk-local
`)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer at-real-access-token")
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}
