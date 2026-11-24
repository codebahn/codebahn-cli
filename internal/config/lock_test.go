package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// TestConfigAtomicWrite verifies that SaveConfig's temp-file-then-rename
// strategy means a concurrent reader never sees a truncated or partial
// config file. The test runs writers and readers concurrently: every read
// must yield either valid JSON or ErrNotExist (before the first write
// lands). With a non-atomic os.WriteFile a reader can catch the file
// mid-write and get invalid JSON or an incomplete read.
func TestConfigAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	const writers = 10
	const readers = 10
	const rounds = 50

	stop := make(chan struct{})
	var readErrors []string
	var mu sync.Mutex

	// Concurrent readers: keep reading the config file and check every
	// read returns valid JSON (or file-not-found before the first write).
	var rg sync.WaitGroup
	for i := 0; i < readers; i++ {
		rg.Add(1)
		go func() {
			defer rg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				data, err := os.ReadFile(ConfigPath())
				if err != nil {
					continue // file not yet created
				}
				if len(data) == 0 {
					continue
				}
				var cfg Config
				if err := json.Unmarshal(data, &cfg); err != nil {
					mu.Lock()
					readErrors = append(readErrors, fmt.Sprintf("invalid JSON (%v): %s", err, data))
					mu.Unlock()
					return
				}
				if cfg.URL == "" || cfg.AccessToken == "" {
					mu.Lock()
					readErrors = append(readErrors, fmt.Sprintf("incomplete config: %+v", cfg))
					mu.Unlock()
					return
				}
			}
		}()
	}

	// Concurrent writers.
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				cfg := Config{
					URL:          "https://codebahn.net",
					AccessToken:  fmt.Sprintf("access-%d-%d", n, r),
					RefreshToken: "refresh",
					TokenExpiry:  int64(n*rounds + r),
				}
				if err := SaveConfig(cfg); err != nil {
					t.Errorf("SaveConfig() error = %v", err)
				}
			}
		}(i)
	}
	wg.Wait()
	close(stop)
	rg.Wait()

	mu.Lock()
	defer mu.Unlock()
	for _, e := range readErrors {
		t.Errorf("reader saw corrupt file: %s", e)
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
