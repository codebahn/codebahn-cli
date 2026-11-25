package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	URL          string `json:"url"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenExpiry  int64  `json:"token_expiry,omitempty"`
	CheckUpdates *bool  `json:"check_updates,omitempty"`
}

type Auth struct {
	URL          string
	AccessToken  string
	RefreshToken string
	TokenExpiry  int64
}

func ConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "codebahn")
	}
	if runtime.GOOS == "windows" {
		dir, _ := os.UserConfigDir()
		return filepath.Join(dir, "codebahn")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "codebahn")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

func LoadConfig() (Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("config file %s is invalid: %w; "+
			"fix or delete it, or run 'codebahn auth login' to rewrite it", ConfigPath(), err)
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	// Write to a temp file in the same directory, then rename into place,
	// so a crash mid-write never leaves a truncated or partial config file.
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, ConfigPath())
}

func ClearTokens() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	cfg.AccessToken = ""
	cfg.RefreshToken = ""
	cfg.TokenExpiry = 0
	return SaveConfig(cfg)
}

// DefaultURL is the Codebahn server used when nothing names another.
const DefaultURL = "https://codebahn.net"

// ResolveAuth picks the credentials and server for a command.
//
// A token (the token argument, from --token, or else CODEBAHN_TOKEN) talks to
// CODEBAHN_URL, else DefaultURL; never to a saved login's server, so a token
// for one server is not sent to another. Without a token, the saved login is
// used with the server it was made on. If CODEBAHN_URL names a different
// server, ResolveAuth returns a *URLMismatchError rather than send the saved
// credentials to one server while the user expects the other.
func ResolveAuth(token string) (Auth, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return Auth{}, err
	}
	envURL := os.Getenv("CODEBAHN_URL")

	if token == "" {
		token = os.Getenv("CODEBAHN_TOKEN")
	}
	if token != "" {
		return Auth{URL: firstNonEmpty(envURL, DefaultURL), AccessToken: token}, nil
	}

	if cfg.AccessToken == "" {
		return Auth{URL: firstNonEmpty(envURL, cfg.URL, DefaultURL)}, nil
	}

	saved := savedAuth(cfg)
	if envURL != "" && !sameURL(envURL, saved.URL) {
		return Auth{}, &URLMismatchError{EnvURL: envURL, SavedURL: saved.URL}
	}
	return saved, nil
}

// SavedAuth returns the saved login with the server it was made on, ignoring
// any token or CODEBAHN_URL in the environment. AccessToken is empty when
// nobody is logged in.
func SavedAuth() (Auth, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return Auth{}, err
	}
	return savedAuth(cfg), nil
}

func savedAuth(cfg Config) Auth {
	return Auth{
		URL:          firstNonEmpty(cfg.URL, DefaultURL),
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		TokenExpiry:  cfg.TokenExpiry,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// URLMismatchError reports that CODEBAHN_URL names another server than the
// saved login.
type URLMismatchError struct {
	EnvURL   string
	SavedURL string
}

func (e *URLMismatchError) Error() string {
	return fmt.Sprintf("CODEBAHN_URL is %s, but you are logged in to %s; "+
		"unset CODEBAHN_URL, or run 'codebahn auth login --url %s'", e.EnvURL, e.SavedURL, e.EnvURL)
}

// sameURL reports whether two server URLs name the same server, ignoring a
// trailing slash, the case of the scheme and host, and an explicit default
// port.
func sameURL(a, b string) bool {
	return normalizeURL(a) == normalizeURL(b)
}

func normalizeURL(raw string) string {
	u, err := url.Parse(strings.TrimSuffix(raw, "/"))
	if err != nil {
		return raw
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		u.Host = u.Hostname()
	}
	return u.String()
}
