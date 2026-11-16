package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
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

func ResolveAuth() Auth {
	cfg, _ := LoadConfig()

	envURL := os.Getenv("CODEBAHN_URL")
	envToken := os.Getenv("CODEBAHN_TOKEN")

	if envToken != "" {
		url := envURL
		if url == "" {
			url = cfg.URL
		}
		return Auth{
			URL:         url,
			AccessToken: envToken,
		}
	}

	url := envURL
	if url == "" {
		url = cfg.URL
	}

	return Auth{
		URL:          url,
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		TokenExpiry:  cfg.TokenExpiry,
	}
}
