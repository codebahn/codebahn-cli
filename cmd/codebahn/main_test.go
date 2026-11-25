package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/codebahn/codebahn-cli/internal/config"
)

// runRoot runs the full command tree with args and returns its error.
func runRoot(t *testing.T, args ...string) error {
	t.Helper()
	_, err := runRootIO(t, "", args...)
	return err
}

// runRootIO runs the full command tree with stdin and returns its stdout.
func runRootIO(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	cmd := newRootCmd()
	var outBuf, errBuf bytes.Buffer
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return outBuf.String(), err
}

// fakeOAuthServer answers the token exchange with a fixed token and records
// whether it was asked.
func fakeOAuthServer(t *testing.T, exchanged *bool) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login/oauth/access_token" {
			http.NotFound(w, r)
			return
		}
		*exchanged = true
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"acc","refresh_token":"ref","expires_in":3600}`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// stubBrowser replaces the browser opener with one that records the
// authorize URL and completes the redirect back to the CLI.
func stubBrowser(t *testing.T, opened *string) {
	t.Helper()
	openBrowser = func(authorizeURL string) error {
		*opened = authorizeURL
		u, err := url.Parse(authorizeURL)
		if err != nil {
			return err
		}
		q := u.Query()
		cb := q.Get("redirect_uri") + "/?code=c&state=" + url.QueryEscape(q.Get("state"))
		resp, err := http.Get(cb)
		if err != nil {
			return err
		}
		return resp.Body.Close()
	}
	t.Cleanup(func() { openBrowser = nil })
}

func TestAuthLoginURLPicksAndSavesServer(t *testing.T) {
	setupGitCredentialTest(t, config.Config{})
	var exchanged bool
	srv := fakeOAuthServer(t, &exchanged)
	var opened string
	stubBrowser(t, &opened)

	if err := runRoot(t, "auth", "login", "--url", srv.URL); err != nil {
		t.Fatalf("auth login: %v", err)
	}
	if !strings.HasPrefix(opened, srv.URL+"/login/oauth/authorize?") {
		t.Errorf("opened %q, want the authorize page on %s", opened, srv.URL)
	}
	if !exchanged {
		t.Error("token exchange did not reach the --url server")
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.URL != srv.URL || cfg.AccessToken != "acc" {
		t.Errorf("saved config = %+v, want URL %s with the new token", cfg, srv.URL)
	}
}

func TestAuthLoginDefaultsToEnvURL(t *testing.T) {
	setupGitCredentialTest(t, config.Config{})
	var exchanged bool
	srv := fakeOAuthServer(t, &exchanged)
	t.Setenv("CODEBAHN_URL", srv.URL)
	var opened string
	stubBrowser(t, &opened)

	if err := runRoot(t, "auth", "login"); err != nil {
		t.Fatalf("auth login: %v", err)
	}
	if cfg, _ := config.LoadConfig(); cfg.URL != srv.URL {
		t.Errorf("saved URL = %q, want %q", cfg.URL, srv.URL)
	}
}

func TestInstanceFlagIsGone(t *testing.T) {
	setupGitCredentialTest(t, config.Config{})

	err := runRoot(t, "auth", "status", "--instance", "https://staging.codebahn.net")
	if err == nil || !strings.Contains(err.Error(), "unknown flag: --instance") {
		t.Errorf("error = %v, want unknown flag: --instance", err)
	}
}

// A saved login talks to the server it was made on; CODEBAHN_URL naming
// another server is an error, never a silent choice.
func TestSavedLoginDisagreeingWithEnvURLFails(t *testing.T) {
	setupGitCredentialTest(t, config.Config{URL: "https://staging.codebahn.net", AccessToken: "acc"})
	t.Setenv("CODEBAHN_URL", "https://codebahn.net")

	err := runRoot(t, "auth", "status")
	if err == nil || !strings.Contains(err.Error(), "but you are logged in to https://staging.codebahn.net") {
		t.Errorf("error = %v, want the disagreement error", err)
	}
}

func TestTokenFlagUsesEnvURL(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"login":"ci"}`)
	}))
	defer srv.Close()
	setupGitCredentialTest(t, config.Config{})
	t.Setenv("CODEBAHN_URL", srv.URL)

	if err := runRoot(t, "--token", "pat", "auth", "status"); err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if gotAuth != "token pat" && gotAuth != "Bearer pat" {
		t.Errorf("Authorization = %q, want the --token value", gotAuth)
	}
}

// Commands that need no saved credentials still run when CODEBAHN_URL
// disagrees with the saved login (R19: store and erase are no-ops, and a
// foreign host gets nothing and exit 0).
func TestURLMismatchDoesNotBlockCommandsWithoutAuth(t *testing.T) {
	const mismatch = "but you are logged in to"
	tests := []struct {
		name  string
		args  []string
		stdin string
	}{
		{"store", []string{"auth", "git-credential", "store"}, "protocol=https\nhost=codebahn.net\n\n"},
		{"erase", []string{"auth", "git-credential", "erase"}, "protocol=https\nhost=codebahn.net\n\n"},
		{"get foreign host", []string{"auth", "git-credential", "get"}, "protocol=https\nhost=github.com\n\n"},
		{"get http", []string{"auth", "git-credential", "get"}, "protocol=http\nhost=codebahn.net\n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net", AccessToken: "acc"})
			t.Setenv("CODEBAHN_URL", "https://staging.codebahn.net")

			stdout, err := runRootIO(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatalf("error = %v, want nil", err)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
		})
	}

	t.Run("update", func(t *testing.T) {
		setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net", AccessToken: "acc"})
		t.Setenv("CODEBAHN_URL", "https://staging.codebahn.net")

		err := runRoot(t, "update")
		if err != nil && strings.Contains(err.Error(), mismatch) {
			t.Errorf("error = %v, want update to ignore the saved login", err)
		}
	})
}

// For either of the two disagreeing servers, get still refuses to guess.
func TestURLMismatchGetForEitherServerFails(t *testing.T) {
	for _, host := range []string{"codebahn.net", "staging.codebahn.net"} {
		t.Run(host, func(t *testing.T) {
			setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net", AccessToken: "acc"})
			t.Setenv("CODEBAHN_URL", "https://staging.codebahn.net")

			stdout, err := runRootIO(t, "protocol=https\nhost="+host+"\n\n", "auth", "git-credential", "get")
			if err == nil || !strings.Contains(err.Error(), "but you are logged in to") {
				t.Errorf("error = %v, want the disagreement error", err)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
		})
	}
}

// --token on one invocation must not leak into the next: the flag value
// lives on the command, not in package state.
func TestTokenFlagDoesNotLeakBetweenRuns(t *testing.T) {
	setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net"})

	if _, err := runRootIO(t, "protocol=https\nhost=codebahn.net\n\n", "--token", "pat", "auth", "git-credential", "get"); err != nil {
		t.Fatalf("first run: %v", err)
	}
	stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost=codebahn.net\n\n")
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty: --token from the first run leaked", stdout)
	}
}

// Shell completion and help need no credentials, so a mismatch must not
// break tab completion.
func TestURLMismatchDoesNotBlockCompletionOrHelp(t *testing.T) {
	for _, args := range [][]string{
		{"__complete", "auth", ""},
		{"completion", "bash"},
		{"help", "auth"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net", AccessToken: "acc"})
			t.Setenv("CODEBAHN_URL", "https://staging.codebahn.net")

			if err := runRoot(t, args...); err != nil {
				t.Errorf("error = %v, want nil", err)
			}
		})
	}
}

// The mismatch error says to run 'codebahn auth login --url ...'; that must
// work while CODEBAHN_URL still disagrees with the saved login.
func TestAuthLoginRunsDespiteURLMismatch(t *testing.T) {
	setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net", AccessToken: "old"})
	var exchanged bool
	srv := fakeOAuthServer(t, &exchanged)
	t.Setenv("CODEBAHN_URL", srv.URL)
	var opened string
	stubBrowser(t, &opened)

	if err := runRoot(t, "auth", "login", "--url", srv.URL); err != nil {
		t.Fatalf("auth login: %v", err)
	}
	if cfg, _ := config.LoadConfig(); cfg.URL != srv.URL || cfg.AccessToken != "acc" {
		t.Errorf("saved config = %+v, want URL %s with the new token", cfg, srv.URL)
	}
}

func TestAuthLogoutRunsDespiteURLMismatch(t *testing.T) {
	setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net", AccessToken: "old"})
	t.Setenv("CODEBAHN_URL", "https://staging.codebahn.net")

	if err := runRoot(t, "auth", "logout"); err != nil {
		t.Fatalf("auth logout: %v", err)
	}
	if cfg, _ := config.LoadConfig(); cfg.AccessToken != "" {
		t.Errorf("AccessToken = %q, want cleared", cfg.AccessToken)
	}
}

// R19: --token on get returns that token for the Codebahn host.
func TestGitCredentialGetTokenFlag(t *testing.T) {
	setupGitCredentialTest(t, config.Config{})

	stdout, err := runRootIO(t, "protocol=https\nhost=codebahn.net\n\n", "--token", "pat", "auth", "git-credential", "get")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if want := "username=oauth2\npassword=pat\n"; stdout != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}

// writeCorruptConfig isolates config and environment and writes a config
// file that exists but does not parse: token_expiry as RFC 3339 instead of
// an epoch, as seen in the field.
func writeCorruptConfig(t *testing.T) {
	t.Helper()
	setupGitCredentialTest(t, config.Config{})
	content := `{"url": "https://codebahn.net", "access_token": "acc", "token_expiry": "2026-11-24T17:00:00Z"}`
	if err := os.WriteFile(config.ConfigPath(), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// wantConfigParseError checks that err reports the corrupt config file, not
// a missing login.
func wantConfigParseError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("error = nil, want the config parse error")
	}
	if !strings.Contains(err.Error(), config.ConfigPath()) || !strings.Contains(err.Error(), "token_expiry") {
		t.Errorf("error = %q, want it to name %s and the bad field", err, config.ConfigPath())
	}
}

func TestCorruptConfigFailsCommands(t *testing.T) {
	writeCorruptConfig(t)

	wantConfigParseError(t, runRoot(t, "auth", "status"))
}

// Logging in is the way out of a corrupt config, so it must still work and
// rewrite the file, with the update check wired in as in main.
func TestAuthLoginRewritesCorruptConfig(t *testing.T) {
	writeCorruptConfig(t)
	var exchanged bool
	srv := fakeOAuthServer(t, &exchanged)
	var opened string
	stubBrowser(t, &opened)

	root := newRootCmd()
	checkUpdateInBackground(root)
	root.SetArgs([]string{"auth", "login", "--url", srv.URL})
	if err := root.Execute(); err != nil {
		t.Fatalf("auth login: %v", err)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.URL != srv.URL || cfg.AccessToken != "acc" {
		t.Errorf("saved config = %+v, want URL %s with the new token", cfg, srv.URL)
	}
}

// The update-check opt-out lives in the config; when the config cannot be
// read, the opt-out is unknown, so no check is made.
func TestUpdateCheckSkippedWhenConfigCorrupt(t *testing.T) {
	setupGitCredentialTest(t, config.Config{URL: "https://codebahn.net"})
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("CODEBAHN_NO_UPDATE_CHECK", "")
	orig := version
	version = "v1.0.0"
	t.Cleanup(func() { version = orig })

	if !wantUpdateCheck() {
		t.Fatal("wantUpdateCheck() = false with a valid config, want true")
	}
	writeCorruptConfig(t)
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if wantUpdateCheck() {
		t.Error("wantUpdateCheck() = true with a corrupt config, want false")
	}
}
