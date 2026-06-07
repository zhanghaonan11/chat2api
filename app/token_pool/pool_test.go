package token_pool

import (
	"sync"
	"testing"
	"time"
)

func TestAccessTokenPoolRoundRobinStartsFromFirstUsableToken(t *testing.T) {
	pool := newAccessTokenPool()
	expiresAt := time.Now().Add(time.Hour).Unix()
	pool.ReplaceAccessTokens([]*AccessToken{
		{Token: "Bearer token-1", ExpiresAt: expiresAt},
		{Token: "Bearer token-2", ExpiresAt: expiresAt},
	})

	first := pool.GetAccessToken()
	if first == nil || first.Token != "Bearer token-1" {
		t.Fatalf("expected first token, got %#v", first)
	}
	second := pool.GetAccessToken()
	if second == nil || second.Token != "Bearer token-2" {
		t.Fatalf("expected second token, got %#v", second)
	}
}

func TestAccessTokenPoolSkipsCoolingAndExpiredTokens(t *testing.T) {
	pool := newAccessTokenPool()
	now := time.Now()
	pool.ReplaceAccessTokens([]*AccessToken{
		{Token: "Bearer cooling", ExpiresAt: now.Add(time.Hour).Unix(), CanUseAt: now.Add(time.Hour).Unix()},
		{Token: "Bearer expired", ExpiresAt: now.Add(-time.Hour).Unix()},
		{Token: "Bearer usable", ExpiresAt: now.Add(time.Hour).Unix()},
	})

	token := pool.GetAccessToken()
	if token == nil || token.Token != "Bearer usable" {
		t.Fatalf("expected usable token, got %#v", token)
	}
	if got := pool.CanUseSize(); got != 1 {
		t.Fatalf("expected one usable token, got %d", got)
	}
}

func TestAccessTokenPoolCopiesTokensOnReadAndWrite(t *testing.T) {
	pool := newAccessTokenPool()
	source := &AccessToken{Token: "Bearer token", ExpiresAt: time.Now().Add(time.Hour).Unix()}
	pool.AddAccessToken(source)
	source.Token = "Bearer changed"

	token := pool.GetAccessToken()
	if token == nil || token.Token != "Bearer token" {
		t.Fatalf("expected pool to keep write-time copy, got %#v", token)
	}
	token.Token = "Bearer mutated-read"
	token = pool.GetAccessToken()
	if token == nil || token.Token != "Bearer token" {
		t.Fatalf("expected returned token to be read copy, got %#v", token)
	}
}

func TestAccessTokenPoolConcurrentAccess(t *testing.T) {
	pool := newAccessTokenPool()
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				pool.ReplaceAccessTokens([]*AccessToken{
					{Token: "Bearer token", ExpiresAt: time.Now().Add(time.Hour).Unix()},
				})
			}
		}()
	}
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = pool.Size()
				_ = pool.CanUseSize()
				_ = pool.GetAccessToken()
				pool.SetCanUseAt("Bearer token", 0)
			}
		}()
	}
	wg.Wait()
}
