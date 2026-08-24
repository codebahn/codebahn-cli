package migrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const defaultSessionTTL = 1 * time.Hour

type githubSession struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

var githubSessionPath = func() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "codebahn", "github-session.json")
}

func loadGitHubSession() (string, bool) {
	data, err := os.ReadFile(githubSessionPath())
	if err != nil {
		return "", false
	}
	var s githubSession
	if err := json.Unmarshal(data, &s); err != nil {
		return "", false
	}
	if time.Now().After(s.ExpiresAt) {
		return "", false
	}
	return s.AccessToken, true
}

func saveGitHubSession(token string, expiresIn int) error {
	ttl := defaultSessionTTL
	if expiresIn != 0 {
		ttl = time.Duration(expiresIn) * time.Second
	}
	s := githubSession{
		AccessToken: token,
		ExpiresAt:   time.Now().Add(ttl),
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	path := githubSessionPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func clearGitHubSession() {
	_ = os.Remove(githubSessionPath())
}

func ClearGitHubSession() {
	clearGitHubSession()
}
