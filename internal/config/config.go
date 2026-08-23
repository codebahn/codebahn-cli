package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
	return os.WriteFile(ConfigPath(), data, 0600)
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
