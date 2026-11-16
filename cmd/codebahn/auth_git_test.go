package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/codebahn/codebahn-cli/internal/config"
)

// setupGitCredentialTest isolates config and environment for a single test
// and writes cfg as the active config file.
func setupGitCredentialTest(t *testing.T, cfg config.Config) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("CODEBAHN_TOKEN", "")
	t.Setenv("CODEBAHN_URL", "")
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
}

func runGitCredential(t *testing.T, args []string, stdin string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := gitCredentialCmd()
	var outBuf, errBuf bytes.Buffer
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestGitCredentialStoreErase(t *testing.T) {
	setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net", AccessToken: "tok"})

	for _, op := range []string{"store", "erase"} {
		t.Run(op, func(t *testing.T) {
			stdout, _, err := runGitCredential(t, []string{op}, "protocol=https\nhost=codebahn.net\n\n")
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
		})
	}
}

func TestGitCredentialGetNoMatch(t *testing.T) {
	tests := []struct {
		name  string
		cfg   config.Config
		stdin string
	}{
		{
			name:  "foreign host",
			cfg:   config.Config{URL: "https://codebahn.net", AccessToken: "tok"},
			stdin: "protocol=https\nhost=github.com\n\n",
		},
		{
			name:  "http on matching host",
			cfg:   config.Config{URL: "https://codebahn.net", AccessToken: "tok"},
			stdin: "protocol=http\nhost=codebahn.net\n\n",
		},
		{
			name:  "no auth configured",
			cfg:   config.Config{URL: "https://codebahn.net"},
			stdin: "protocol=https\nhost=codebahn.net\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupGitCredentialTest(t, tt.cfg)

			stdout, _, err := runGitCredential(t, []string{"get"}, tt.stdin)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
		})
	}
}
