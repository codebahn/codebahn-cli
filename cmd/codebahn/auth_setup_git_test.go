package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupSetupGitTest isolates git's global config and codebahn's own config
// for a single test, points the codebahn instance at instanceURL, and
// stubs executablePath so the test doesn't depend on the test binary's
// path. It returns the host-specific credential helper config key.
func setupSetupGitTest(t *testing.T, instanceURL string) (host string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("CODEBAHN_TOKEN", "")
	t.Setenv("CODEBAHN_URL", instanceURL)
	unsetEnv(t, "GIT_CONFIG_GLOBAL") // don't let the real global gitconfig leak in

	exe := filepath.Join(t.TempDir(), "codebahn")
	executablePath = func() (string, error) { return exe, nil }
	t.Cleanup(func() { executablePath = os.Executable })

	return mustHost(t, instanceURL)
}

// unsetEnv unsets an environment variable for the duration of the test,
// restoring its prior value (or absence) afterwards. Unlike t.Setenv("", ""),
// this leaves the variable genuinely absent rather than set to an empty
// string, which some programs (including git) treat differently.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	prev, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) error = %v", key, err)
	}
	t.Cleanup(func() {
		if ok {
			os.Setenv(key, prev)
		}
	})
}

func runSetupGitCmd(t *testing.T) (stdout string, err error) {
	t.Helper()
	cmd := setupGitCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	err = cmd.Execute()
	return buf.String(), err
}

// helperValues returns the configured values of the host-specific
// credential helper, in file order.
func helperValues(t *testing.T, host string) []string {
	t.Helper()
	out, err := exec.Command("git", "config", "--global", "--get-all", "credential.https://"+host+".helper").Output()
	if err != nil {
		t.Fatalf("git config --get-all: %v", err)
	}
	return strings.Split(strings.TrimRight(string(out), "\n"), "\n")
}

func TestSetupGit(t *testing.T) {
	host := setupSetupGitTest(t, "https://codebahn.net")

	if _, err := runSetupGitCmd(t); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	values := helperValues(t, host)
	if len(values) != 2 {
		t.Fatalf("helper values = %v, want 2 entries", values)
	}
	if values[0] != "" {
		t.Errorf("first helper entry = %q, want empty (resets the chain)", values[0])
	}
	if !strings.Contains(values[1], "auth git-credential") {
		t.Errorf("second helper entry = %q, want it to invoke auth git-credential", values[1])
	}
}

func TestSetupGitIdempotent(t *testing.T) {
	host := setupSetupGitTest(t, "https://codebahn.net")

	if _, err := runSetupGitCmd(t); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	first := helperValues(t, host)

	if _, err := runSetupGitCmd(t); err != nil {
		t.Fatalf("Execute() (second run) error = %v", err)
	}
	second := helperValues(t, host)

	if len(second) != 2 {
		t.Fatalf("helper values after second run = %v, want 2 entries", second)
	}
	if first[0] != second[0] || first[1] != second[1] {
		t.Errorf("helper values changed across runs: first=%v second=%v", first, second)
	}
}

func TestSetupGitWithExistingHelper(t *testing.T) {
	host := setupSetupGitTest(t, "https://codebahn.net")

	if err := exec.Command("git", "config", "--global", "credential.helper", "store").Run(); err != nil {
		t.Fatalf("pre-configuring generic credential.helper: %v", err)
	}

	if _, err := runSetupGitCmd(t); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// The generic helper is untouched...
	generic, err := exec.Command("git", "config", "--global", "--get", "credential.helper").Output()
	if err != nil {
		t.Fatalf("git config --get credential.helper: %v", err)
	}
	if strings.TrimSpace(string(generic)) != "store" {
		t.Errorf("generic credential.helper = %q, want %q", strings.TrimSpace(string(generic)), "store")
	}

	// ...but the host-specific chain resets it, so only the codebahn
	// helper runs for this host.
	values := helperValues(t, host)
	if len(values) != 2 || values[0] != "" || !strings.Contains(values[1], "auth git-credential") {
		t.Fatalf("host-specific helper chain = %v, want [\"\", codebahn helper]", values)
	}
}
