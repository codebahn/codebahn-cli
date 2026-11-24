package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codebahn/codebahn-cli/client"
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

// A corrupt config is an error on stderr, not a silent "no credential".
func TestGitCredentialGetCorruptConfig(t *testing.T) {
	writeCorruptConfig(t)

	stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost=codebahn.net\n\n")
	wantConfigParseError(t, err)
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
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

func TestGitCredentialRefresh(t *testing.T) {
	t.Run("expiring soon triggers refresh", func(t *testing.T) {
		var calls int
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			if got := r.FormValue("refresh_token"); got != "refresh-tok" {
				t.Errorf("refresh_token = %q, want %q", got, "refresh-tok")
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","expires_in":3600}`)
		}))
		defer srv.Close()

		host := mustHost(t, srv.URL)
		setupGitCredentialTest(t, config.Config{
			URL:          srv.URL,
			AccessToken:  "old-token",
			RefreshToken: "refresh-tok",
			TokenExpiry:  time.Now().Add(45 * time.Second).Unix(),
		})

		stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost="+host+"\n\n")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		want := "username=oauth2\npassword=new-access\n"
		if stdout != want {
			t.Errorf("stdout = %q, want %q", stdout, want)
		}
		if calls != 1 {
			t.Errorf("server called %d times, want 1", calls)
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.AccessToken != "new-access" || cfg.RefreshToken != "new-refresh" {
			t.Errorf("saved config = %+v, want refreshed tokens", cfg)
		}
	})

	t.Run("not yet expiring returns existing token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("server contacted; token is not expiring soon, no refresh expected")
		}))
		defer srv.Close()

		host := mustHost(t, srv.URL)
		setupGitCredentialTest(t, config.Config{
			URL:          srv.URL,
			AccessToken:  "old-token",
			RefreshToken: "refresh-tok",
			TokenExpiry:  time.Now().Add(90 * time.Second).Unix(),
		})

		stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost="+host+"\n\n")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		want := "username=oauth2\npassword=old-token\n"
		if stdout != want {
			t.Errorf("stdout = %q, want %q", stdout, want)
		}
	})
}

func TestGitCredentialRefreshFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"invalid_grant","error_description":"refresh token expired"}`)
	}))
	defer srv.Close()

	host := mustHost(t, srv.URL)
	setupGitCredentialTest(t, config.Config{
		URL:          srv.URL,
		AccessToken:  "old-token",
		RefreshToken: "refresh-tok",
		TokenExpiry:  time.Now().Add(1 * time.Second).Unix(),
	})

	stdout, stderr, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost="+host+"\n\n")
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "codebahn auth login") {
		t.Errorf("stderr = %q, want to contain %q", stderr, "codebahn auth login")
	}
}

func TestGitCredentialConcurrentRefresh(t *testing.T) {
	var mu sync.Mutex
	validRefreshToken := "refresh-tok"
	calls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		mu.Lock()
		defer mu.Unlock()
		calls++
		if got := r.FormValue("refresh_token"); got != validRefreshToken {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error":"invalid_grant","error_description":"refresh token reused"}`)
			return
		}
		validRefreshToken = "refresh-tok-rotated"
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"refresh-tok-rotated","expires_in":3600}`)
	}))
	defer srv.Close()

	host := mustHost(t, srv.URL)
	setupGitCredentialTest(t, config.Config{
		URL:          srv.URL,
		AccessToken:  "old-token",
		RefreshToken: "refresh-tok",
		TokenExpiry:  time.Now().Add(1 * time.Second).Unix(),
	})

	var wg sync.WaitGroup
	results := make([]string, 2)
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost="+host+"\n\n")
			results[i] = stdout
			errs[i] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("caller %d: Execute() error = %v", i, err)
		}
	}

	want := "username=oauth2\npassword=new-access\n"
	for i, got := range results {
		if got != want {
			t.Errorf("caller %d: stdout = %q, want %q", i, got, want)
		}
	}

	mu.Lock()
	gotCalls := calls
	mu.Unlock()
	if gotCalls != 1 {
		t.Errorf("server called %d times, want 1", gotCalls)
	}
}

// TestClientRefreshReadsConfig verifies that a CLI client whose in-memory
// refresh token was rotated by a concurrent credential helper process still
// succeeds by re-reading the config file instead of sending the stale token.
func TestClientRefreshReadsConfig(t *testing.T) {
	// Server that accepts each refresh token only once (rotation).
	var mu sync.Mutex
	validRefreshToken := "refresh-tok"
	refreshCalls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/access_token") {
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			mu.Lock()
			defer mu.Unlock()
			refreshCalls++
			if got := r.FormValue("refresh_token"); got != validRefreshToken {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error":"invalid_grant","error_description":"token already used"}`)
				return
			}
			validRefreshToken = "rotated-refresh"
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"rotated-access","refresh_token":"rotated-refresh","expires_in":3600}`)
			return
		}
		// API endpoints
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"login":"test"}`)
	}))
	defer srv.Close()

	host := mustHost(t, srv.URL)
	setupGitCredentialTest(t, config.Config{
		URL:          srv.URL,
		AccessToken:  "old-access",
		RefreshToken: "refresh-tok",
		TokenExpiry:  time.Now().Add(1 * time.Second).Unix(),
	})

	// Step 1: credential helper rotates the token (simulating another process).
	stdout, _, err := runGitCredential(t, []string{"get"}, "protocol=https\nhost="+host+"\n\n")
	if err != nil {
		t.Fatalf("credential helper: %v", err)
	}
	if !strings.Contains(stdout, "rotated-access") {
		t.Fatalf("credential helper stdout = %q, want rotated-access", stdout)
	}

	// Step 2: a client created with ORIGINAL in-memory tokens (as
	// PersistentPreRunE sets up at process start) tries an API call.
	c := client.New(srv.URL, "old-access")
	c.SetOAuth(time.Now().Add(1*time.Second), lockedRefresh(srv.URL))

	// With the bug: ensureValidToken sends stale refresh-tok, server rejects it.
	// After fix: re-reads config, finds rotated-access, skips HTTP refresh.
	_, err = c.GetRaw(context.Background(), "/user")
	if err != nil {
		t.Fatalf("client API call failed (stale refresh token not handled): %v", err)
	}

	mu.Lock()
	gotCalls := refreshCalls
	mu.Unlock()
	if gotCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1 (should reuse already-refreshed token)", gotCalls)
	}
}

// TestEnsureValidTokenUpdatesState verifies that after a successful refresh
// the client uses the NEW access token for the API call (not the original) and
// that the in-memory expiry is updated so a second call does not trigger
// another refresh.
func TestEnsureValidTokenUpdatesState(t *testing.T) {
	refreshCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer refreshed-access" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message":"bad token"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"login":"test"}`)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "old-access")
	c.SetOAuth(time.Now().Add(1*time.Second), func(_ context.Context) (string, time.Time, error) {
		refreshCalls++
		return "refreshed-access", time.Now().Add(time.Hour), nil
	})

	// First call: token is expiring, triggers refresh; client must use the
	// refreshed token for the API call, not the original "old-access".
	if _, err := c.GetRaw(context.Background(), "/user"); err != nil {
		t.Fatalf("first API call failed: %v", err)
	}
	if refreshCalls != 1 {
		t.Fatalf("refresh calls = %d after first request, want 1", refreshCalls)
	}

	// Second call: the in-memory expiry was updated to ~1h, so
	// ensureValidToken should skip the refresh entirely.
	if _, err := c.GetRaw(context.Background(), "/user"); err != nil {
		t.Fatalf("second API call failed: %v", err)
	}
	if refreshCalls != 1 {
		t.Fatalf("refresh calls = %d after second request, want 1 (expiry should be updated)", refreshCalls)
	}
}

// TestAuthStatusRefreshesExpiredToken runs "auth status" through the real
// command tree and verifies that PersistentPreRunE wires up OAuth refresh so
// an expiring token is rotated before the API call.
func TestAuthStatusRefreshesExpiredToken(t *testing.T) {
	refreshCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login/oauth/access_token" {
			refreshCalls++
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			if got := r.FormValue("refresh_token"); got != "refresh-tok" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error":"invalid_grant"}`)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"new-access","refresh_token":"new-refresh","expires_in":3600}`)
			return
		}
		// API endpoints: reject the stale token, accept the refreshed one.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer new-access" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"message":"bad credentials"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"login":"test","full_name":"Test User","email":"test@example.com"}`)
	}))
	defer srv.Close()

	setupGitCredentialTest(t, config.Config{
		URL:          srv.URL,
		AccessToken:  "old-access",
		RefreshToken: "refresh-tok",
		TokenExpiry:  time.Now().Add(1 * time.Second).Unix(),
	})

	cmd := newRootCmd()
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"auth", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth status failed: %v\nstderr: %s", err, errBuf.String())
	}
	if refreshCalls != 1 {
		t.Errorf("refresh calls = %d, want 1", refreshCalls)
	}
}

// TestAuthCommandSubcommands verifies that git-credential and setup-git are
// registered as auth subcommands.
func TestAuthCommandSubcommands(t *testing.T) {
	cmd := authCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{"git-credential", "setup-git"} {
		if !names[want] {
			t.Errorf("auth command missing subcommand %q", want)
		}
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
