package config

import (
	"encoding/json"
	"errors"
	"fmt"
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
		return Config{}, err
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
	cfg, _ := LoadConfig()
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
// CODEBAHN_URL, else the saved login's server, else DefaultURL. Without a
// token, the saved login is used with the server it was made on. If
// CODEBAHN_URL names a different server, ResolveAuth fails rather than send
// the saved credentials to one server while the user expects the other.
func ResolveAuth(token string) (Auth, error) {
	cfg, _ := LoadConfig()
	envURL := os.Getenv("CODEBAHN_URL")

	if token == "" {
		token = os.Getenv("CODEBAHN_TOKEN")
	}
	if token != "" {
		return Auth{URL: firstNonEmpty(envURL, cfg.URL, DefaultURL), AccessToken: token}, nil
	}

	if cfg.AccessToken == "" {
		return Auth{URL: firstNonEmpty(envURL, cfg.URL, DefaultURL)}, nil
	}

	savedURL := firstNonEmpty(cfg.URL, DefaultURL)
	if envURL != "" && !sameURL(envURL, savedURL) {
		return Auth{}, fmt.Errorf("CODEBAHN_URL is %s, but you are logged in to %s; "+
			"unset CODEBAHN_URL, or run 'codebahn auth login --url %s'", envURL, savedURL, envURL)
	}
	return Auth{
		URL:          savedURL,
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		TokenExpiry:  cfg.TokenExpiry,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// sameURL compares two server URLs, ignoring a trailing slash and the case
// of the scheme and host.
func sameURL(a, b string) bool {
	return strings.EqualFold(strings.TrimSuffix(a, "/"), strings.TrimSuffix(b, "/"))
}
