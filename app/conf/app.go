package conf

import "sync"

type app struct {
	LogLevel       string    `yaml:"log_level"`
	LogPath        string    `yaml:"log_path"`
	LogFile        string    `yaml:"log_file"`
	Bind           string    `yaml:"bind"`
	Port           uint16    `yaml:"port"`
	Auth           auth      `yaml:"auth"`
	Proxy          string    `yaml:"proxy"`
	ChatGPTBaseUrl string    `yaml:"chatgpt_base_url"`
	ChatGPTs       []chatgpt `yaml:"chatgpts"`
}

func (a app) TextAccessTokens() []string {
	tokens := make([]string, 0, len(a.ChatGPTs))
	for _, account := range a.ChatGPTs {
		if account.AccessToken != "" {
			tokens = append(tokens, account.AccessToken)
		}
	}
	return tokens
}

type auth struct {
	AccessTokens []string `yaml:"access_tokens"`
}

type chatgpt struct {
	IdToken      string `yaml:"id_token"`
	AccessToken  string `yaml:"access_token"`
	RefreshToken string `yaml:"refresh_token"`
	AccountId    string `yaml:"account_id"`
	LastRefresh  string `yaml:"last_refresh"`
	Email        string `yaml:"email"`
	Type         string `yaml:"type"`
	Expired      string `yaml:"expired"`
	Proxy        string `yaml:"proxy"`
}

var appState = struct {
	sync.RWMutex
	value app
}{
	value: defaultApp(),
}

var App = defaultApp()

func GetApp() app {
	appState.RLock()
	defer appState.RUnlock()
	return cloneApp(appState.value)
}

func setApp(value app) {
	appState.Lock()
	defer appState.Unlock()
	appState.value = cloneApp(value)
	App = cloneApp(value)
}

func GetAuthAccessTokens() []string {
	cfg := GetApp()
	return append([]string(nil), cfg.Auth.AccessTokens...)
}

func cloneApp(value app) app {
	value.Auth.AccessTokens = append([]string(nil), value.Auth.AccessTokens...)
	value.ChatGPTs = append([]chatgpt(nil), value.ChatGPTs...)
	return value
}
