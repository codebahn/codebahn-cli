package config

import (
	"os"
	"path/filepath"
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
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".config", "codebahn")
	if got != want {
		t.Errorf("ConfigDir() = %q, want %q", got, want)
	}
}
