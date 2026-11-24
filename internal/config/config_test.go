package config

import (
	"errors"
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

// setupAuthTest isolates config and environment and writes cfg as the saved
// login. A zero cfg leaves no login saved.
func setupAuthTest(t *testing.T, cfg Config) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("CODEBAHN_TOKEN", "")
	t.Setenv("CODEBAHN_URL", "")
	if cfg != (Config{}) {
		if err := SaveConfig(cfg); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}
	}
}

const staging = "https://staging.codebahn.net"

func TestResolveAuth_SavedLoginUsesItsServer(t *testing.T) {
	setupAuthTest(t, Config{URL: staging, AccessToken: "a", RefreshToken: "r", TokenExpiry: 42})

	got, err := ResolveAuth("")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	want := Auth{URL: staging, AccessToken: "a", RefreshToken: "r", TokenExpiry: 42}
	if got != want {
		t.Errorf("ResolveAuth() = %+v, want %+v", got, want)
	}
}

func TestResolveAuth_SavedLoginAgreesWithEnvURL(t *testing.T) {
	setupAuthTest(t, Config{URL: staging, AccessToken: "a"})
	t.Setenv("CODEBAHN_URL", staging+"/")

	got, err := ResolveAuth("")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	if got.URL != staging || got.AccessToken != "a" {
		t.Errorf("ResolveAuth() = %+v, want the saved staging login", got)
	}
}

func TestResolveAuth_SavedLoginDisagreesWithEnvURL(t *testing.T) {
	setupAuthTest(t, Config{URL: staging, AccessToken: "a"})
	t.Setenv("CODEBAHN_URL", DefaultURL)

	_, err := ResolveAuth("")
	if err == nil {
		t.Fatal("ResolveAuth() error = nil, want a disagreement error")
	}
	want := "CODEBAHN_URL is https://codebahn.net, but you are logged in to https://staging.codebahn.net; " +
		"unset CODEBAHN_URL, or run 'codebahn auth login --url https://codebahn.net'"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestResolveAuth_TokenUsesEnvURL(t *testing.T) {
	setupAuthTest(t, Config{})
	t.Setenv("CODEBAHN_URL", staging)

	got, err := ResolveAuth("pat")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	if want := (Auth{URL: staging, AccessToken: "pat"}); got != want {
		t.Errorf("ResolveAuth() = %+v, want %+v", got, want)
	}
}

func TestResolveAuth_TokenDefaultsToProduction(t *testing.T) {
	setupAuthTest(t, Config{})

	got, err := ResolveAuth("pat")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	if want := (Auth{URL: DefaultURL, AccessToken: "pat"}); got != want {
		t.Errorf("ResolveAuth() = %+v, want %+v", got, want)
	}
}

func TestResolveAuth_EnvTokenUsesEnvURL(t *testing.T) {
	setupAuthTest(t, Config{})
	t.Setenv("CODEBAHN_TOKEN", "env-pat")
	t.Setenv("CODEBAHN_URL", staging)

	got, err := ResolveAuth("")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	if want := (Auth{URL: staging, AccessToken: "env-pat"}); got != want {
		t.Errorf("ResolveAuth() = %+v, want %+v", got, want)
	}
}

// A token does not use the saved credentials, so a saved login on another
// server is no conflict: CODEBAHN_URL names the token's server.
func TestResolveAuth_TokenIgnoresSavedLoginOnOtherServer(t *testing.T) {
	setupAuthTest(t, Config{URL: DefaultURL, AccessToken: "a"})
	t.Setenv("CODEBAHN_URL", staging)

	got, err := ResolveAuth("pat")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	if want := (Auth{URL: staging, AccessToken: "pat"}); got != want {
		t.Errorf("ResolveAuth() = %+v, want %+v", got, want)
	}
}

// A token never goes to a saved login's server by default: a production PAT
// must not reach staging because a staging login happens to be saved.
func TestResolveAuth_TokenWithoutEnvURLIgnoresSavedServer(t *testing.T) {
	setupAuthTest(t, Config{URL: staging})

	got, err := ResolveAuth("pat")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	if want := (Auth{URL: DefaultURL, AccessToken: "pat"}); got != want {
		t.Errorf("ResolveAuth() = %+v, want %+v", got, want)
	}
}

func TestResolveAuth_NotLoggedIn(t *testing.T) {
	setupAuthTest(t, Config{})

	got, err := ResolveAuth("")
	if err != nil {
		t.Fatalf("ResolveAuth() error = %v", err)
	}
	if want := (Auth{URL: DefaultURL}); got != want {
		t.Errorf("ResolveAuth() = %+v, want %+v", got, want)
	}
}

// CODEBAHN_URL spelled differently (case, default port, trailing slash)
// still names the saved login's server.
func TestResolveAuth_SavedLoginAgreesWithEquivalentEnvURL(t *testing.T) {
	for _, env := range []string{"https://codebahn.net:443", "HTTPS://CodeBahn.net/", "https://codebahn.net:443/"} {
		t.Run(env, func(t *testing.T) {
			setupAuthTest(t, Config{URL: DefaultURL, AccessToken: "a"})
			t.Setenv("CODEBAHN_URL", env)

			got, err := ResolveAuth("")
			if err != nil {
				t.Fatalf("ResolveAuth() error = %v", err)
			}
			if got.URL != DefaultURL || got.AccessToken != "a" {
				t.Errorf("ResolveAuth() = %+v, want the saved login", got)
			}
		})
	}
}

func TestResolveAuth_MismatchIsURLMismatchError(t *testing.T) {
	setupAuthTest(t, Config{URL: DefaultURL, AccessToken: "a"})
	t.Setenv("CODEBAHN_URL", staging)

	_, err := ResolveAuth("")
	var mismatch *URLMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("ResolveAuth() error = %v, want *URLMismatchError", err)
	}
	if mismatch.EnvURL != staging || mismatch.SavedURL != DefaultURL {
		t.Errorf("mismatch = %+v, want env %s, saved %s", mismatch, staging, DefaultURL)
	}
}
