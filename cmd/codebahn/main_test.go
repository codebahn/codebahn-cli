package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/codebahn/codebahn-cli/internal/config"
)

// runRoot runs the full command tree with args and returns its error.
func runRoot(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newRootCmd()
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	return cmd.Execute()
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
