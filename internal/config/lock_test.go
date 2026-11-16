package config

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"
)

func TestConfigAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cfg := Config{
				URL:          "https://codebahn.net",
				AccessToken:  "access",
				RefreshToken: "refresh",
				TokenExpiry:  int64(n),
			}
			if err := SaveConfig(cfg); err != nil {
				t.Errorf("SaveConfig() error = %v", err)
			}
		}(i)
	}
	wg.Wait()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("file content is not valid JSON: %v\ncontent: %s", err, data)
	}
	if cfg.URL == "" || cfg.AccessToken == "" || cfg.RefreshToken == "" {
		t.Errorf("Config = %+v, missing fields", cfg)
	}
}

func TestConfigLockExclusion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	first, err := LockConfig()
	if err != nil {
		t.Fatalf("first LockConfig() error = %v", err)
	}

	acquired := make(chan struct{})
	go func() {
		second, err := LockConfig()
		if err != nil {
			t.Errorf("second LockConfig() error = %v", err)
			return
		}
		close(acquired)
		_ = UnlockConfig(second)
	}()

	select {
	case <-acquired:
		t.Fatal("second LockConfig() acquired the lock before the first was released")
	case <-time.After(200 * time.Millisecond):
	}

	if err := UnlockConfig(first); err != nil {
		t.Fatalf("UnlockConfig() error = %v", err)
	}

	select {
	case <-acquired:
	case <-time.After(2 * time.Second):
		t.Fatal("second LockConfig() did not acquire the lock after the first was released")
	}
}
