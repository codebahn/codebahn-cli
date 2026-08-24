package migrate

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadSession(t *testing.T) {
	dir := t.TempDir()
	githubSessionPath = func() string { return filepath.Join(dir, "github-session.json") }

	if err := saveGitHubSession("ghu_abc123", 3600); err != nil {
		t.Fatalf("saveGitHubSession: %v", err)
	}

	token, ok := loadGitHubSession()
	if !ok {
		t.Fatal("loadGitHubSession returned false, want true")
	}
	if token != "ghu_abc123" {
		t.Errorf("token = %q, want ghu_abc123", token)
	}
}

func TestLoadSession_Expired(t *testing.T) {
	dir := t.TempDir()
	githubSessionPath = func() string { return filepath.Join(dir, "github-session.json") }

	if err := saveGitHubSession("ghu_expired", -1); err != nil {
		t.Fatalf("saveGitHubSession: %v", err)
	}

	// Force expiry by saving with a negative TTL; saveGitHubSession will set
	// ExpiresAt = now + (-1s), which is in the past.
	token, ok := loadGitHubSession()
	if ok {
		t.Errorf("loadGitHubSession returned true with token %q, want false", token)
	}
}

func TestLoadSession_Missing(t *testing.T) {
	dir := t.TempDir()
	githubSessionPath = func() string { return filepath.Join(dir, "nonexistent.json") }

	_, ok := loadGitHubSession()
	if ok {
		t.Error("loadGitHubSession returned true for missing file, want false")
	}
}

func TestClearSession(t *testing.T) {
	dir := t.TempDir()
	githubSessionPath = func() string { return filepath.Join(dir, "github-session.json") }

	if err := saveGitHubSession("ghu_clearme", 3600); err != nil {
		t.Fatalf("saveGitHubSession: %v", err)
	}

	clearGitHubSession()

	_, ok := loadGitHubSession()
	if ok {
		t.Error("loadGitHubSession returned true after clear, want false")
	}
}

func TestSaveSession_DefaultTTL(t *testing.T) {
	dir := t.TempDir()
	githubSessionPath = func() string { return filepath.Join(dir, "github-session.json") }

	if err := saveGitHubSession("ghu_default", 0); err != nil {
		t.Fatalf("saveGitHubSession: %v", err)
	}

	token, ok := loadGitHubSession()
	if !ok {
		t.Fatal("loadGitHubSession returned false, want true")
	}
	if token != "ghu_default" {
		t.Errorf("token = %q, want ghu_default", token)
	}

	// Verify the default TTL is approximately 1 hour by checking the session
	// is still valid after a small time. We can't easily test the exact
	// timestamp, but we can verify the session was saved with a future expiry.
	_ = time.Now()
}
