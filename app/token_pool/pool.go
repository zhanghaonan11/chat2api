package token_pool

import (
	"chat2api/app/common"
	"sync"
)

var (
	instance *AccessTokenPool
	once     sync.Once
)

type AccessTokenPool struct {
	lock         sync.RWMutex
	AccessTokens []*AccessToken
	index        int
}

type AccessToken struct {
	Token     string `yaml:"token,omitempty"`
	ExpiresAt int64  `yaml:"expires_at,omitempty"`
	Proxy     string `yaml:"proxy,omitempty"`
	CanUseAt  int64  `yaml:"-"`
}

func newAccessTokenPool() *AccessTokenPool {
	return &AccessTokenPool{
		AccessTokens: make([]*AccessToken, 0),
		index:        -1,
	}
}

func GetAccessTokenPool() *AccessTokenPool {
	once.Do(func() {
		instance = newAccessTokenPool()
	})
	return instance
}

func (a *AccessTokenPool) AddAccessToken(accessToken *AccessToken) {
	if accessToken == nil {
		return
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	tokenCopy := *accessToken
	a.AccessTokens = append(a.AccessTokens, &tokenCopy)
}

func (a *AccessTokenPool) Reset() {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.AccessTokens = make([]*AccessToken, 0)
	a.index = -1
}

func (a *AccessTokenPool) AppendAccessTokens(accessTokens []*AccessToken) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.AccessTokens = append(a.AccessTokens, cloneAccessTokens(accessTokens)...)
}

func (a *AccessTokenPool) ReplaceAccessTokens(accessTokens []*AccessToken) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.AccessTokens = cloneAccessTokens(accessTokens)
	a.index = -1
}

func (a *AccessTokenPool) Size() int {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return len(a.AccessTokens)
}

func (a *AccessTokenPool) IsEmpty() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return len(a.AccessTokens) == 0
}

func (a *AccessTokenPool) CanUseSize() int {
	a.lock.RLock()
	defer a.lock.RUnlock()
	var count int
	now := common.GetTimestampSecond(0)
	for _, v := range a.AccessTokens {
		if v != nil && v.CanUseAt <= now && v.ExpiresAt > now {
			count++
		}
	}
	return count
}

func (a *AccessTokenPool) GetToken() string {
	accessToken := a.GetAccessToken()
	if accessToken == nil {
		return ""
	}
	return accessToken.Token
}

func (a *AccessTokenPool) GetAccessToken() *AccessToken {
	a.lock.Lock()
	defer a.lock.Unlock()
	if len(a.AccessTokens) == 0 {
		return nil
	}
	now := common.GetTimestampSecond(0)
	for range a.AccessTokens {
		a.index = (a.index + 1) % len(a.AccessTokens)
		accessToken := a.AccessTokens[a.index]
		if accessToken == nil {
			continue
		}
		if accessToken.CanUseAt <= now && accessToken.ExpiresAt > now {
			tokenCopy := *accessToken
			return &tokenCopy
		}
	}
	return nil
}

func (a *AccessTokenPool) SetCanUseAt(token string, canUseAt int64) {
	a.lock.Lock()
	defer a.lock.Unlock()
	for _, v := range a.AccessTokens {
		if v != nil && v.Token == token {
			v.CanUseAt = canUseAt
			break
		}
	}
}

func cloneAccessTokens(accessTokens []*AccessToken) []*AccessToken {
	clone := make([]*AccessToken, 0, len(accessTokens))
	for _, accessToken := range accessTokens {
		if accessToken == nil {
			continue
		}
		tokenCopy := *accessToken
		clone = append(clone, &tokenCopy)
	}
	return clone
}
