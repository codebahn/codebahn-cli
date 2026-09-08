package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheDir_XDGOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	got := CacheDir()
	want := filepath.Join(dir, "codebahn")
	if got != want {
		t.Errorf("CacheDir() = %q, want %q", got, want)
	}
}

func TestCacheDir_Default(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")

	got := CacheDir()
	base, _ := os.UserCacheDir()
	want := filepath.Join(base, "codebahn")
	if got != want {
		t.Errorf("CacheDir() = %q, want %q", got, want)
	}
}
