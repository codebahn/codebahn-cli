package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigDir_XDGOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	got := ConfigDir()
	want := filepath.Join(dir, "codebahn")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	got := ConfigDir()
	var want string
	if runtime.GOOS == "windows" {
		dir, _ := os.UserConfigDir()
		want = filepath.Join(dir, "codebahn")
	} else {
		home, _ := os.UserHomeDir()
		want = filepath.Join(home, ".config", "codebahn")
	}
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}
