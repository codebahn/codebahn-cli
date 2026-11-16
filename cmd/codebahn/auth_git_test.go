package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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

func TestGitCredentialGetMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("server contacted; token is not expiring, no refresh expected")
	}))
	defer srv.Close()

	host := mustHost(t, srv.URL)
	setupGitCredentialTest(t, config.Config{
		URL:         srv.URL,
		AccessToken: "valid-token",
		TokenExpiry: time.Now().Add(time.Hour).Unix(),
	})

	stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost="+host+"\n\n")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := "username=oauth2\npassword=valid-token\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func TestGitCredentialStaticToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("server contacted; static token must not trigger a refresh")
	}))
	defer srv.Close()

	host := mustHost(t, srv.URL)
	setupGitCredentialTest(t, config.Config{URL: srv.URL})
	t.Setenv("CODEBAHN_TOKEN", "static-tok")

	stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost="+host+"\n\n")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := "username=oauth2\npassword=static-tok\n"
	if stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

func TestGitCredentialStaticTokenForeignHost(t *testing.T) {
	setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net"})
	t.Setenv("CODEBAHN_TOKEN", "static-tok")

	stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost=github.com\n\n")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}

func mustHost(t *testing.T, rawURL string) string {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", rawURL, err)
	}
	return u.Host
}
