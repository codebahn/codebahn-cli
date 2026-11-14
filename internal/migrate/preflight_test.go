package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCheckGitHubRepo_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"full_name":"acme/api","private":false}`)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	err := checkGitHubRepo(context.Background(), Source{Owner: "acme", Repo: "api"}, "tok")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCheckGitHubRepo_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Not Found"}`)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	err := checkGitHubRepo(context.Background(), Source{Owner: "acme", Repo: "nope"}, "tok")
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*PreflightError)
	if !ok {
		t.Fatalf("expected PreflightError, got %T", err)
	}
	if pe.Check != "source-access" {
		t.Errorf("Check = %q, want source-access", pe.Check)
	}
}

func TestCheckGitHubRepo_NotFoundNoToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Not Found"}`)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	err := checkGitHubRepo(context.Background(), Source{Owner: "acme", Repo: "private"}, "")
	if err == nil {
		t.Fatal("expected error")
	}
	pe := err.(*PreflightError)
	if pe.Check != "source-access" {
		t.Errorf("Check = %q, want source-access", pe.Check)
	}
	if got := pe.Message; got != `acme/private not found (if private, provide --source-token or set GITHUB_TOKEN)` {
		t.Errorf("unexpected message: %s", got)
	}
}

func TestCheckGitHubRepo_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	err := checkGitHubRepo(context.Background(), Source{Owner: "acme", Repo: "api"}, "bad-token")
	if err == nil {
		t.Fatal("expected error")
	}
	pe := err.(*PreflightError)
	if pe.Check != "source-auth" {
		t.Errorf("Check = %q, want source-auth", pe.Check)
	}
}

func TestCheckGitHubRepo_RateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	err := checkGitHubRepo(context.Background(), Source{Owner: "acme", Repo: "api"}, "tok")
	if err == nil {
		t.Fatal("expected error")
	}
	pe := err.(*PreflightError)
	if pe.Check != "rate-limit" {
		t.Errorf("Check = %q, want rate-limit", pe.Check)
	}
}

func TestCheckGitHubRepo_ScopeWarning(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-OAuth-Scopes", "read:org, read:user")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"full_name":"acme/api","private":false}`)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	// Should succeed (200) but print a warning about missing repo scope.
	// We can't easily capture stderr in this test, but verify it doesn't error.
	err := checkGitHubRepo(context.Background(), Source{Owner: "acme", Repo: "api"}, "tok")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func sourceFromServer(t *testing.T, ts *httptest.Server, service, owner, repo string) Source {
	t.Helper()
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	return Source{Service: service, Scheme: u.Scheme, Host: u.Host, Owner: owner, Repo: repo}
}

func TestCheckSourceRepo_Gitea(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		token     string
		wantCheck string
		wantMsg   string
	}{
		{"ok", http.StatusOK, "tok", "", ""},
		{"not found no token", http.StatusNotFound, "", "source-access",
			"acme/private not found (if private, provide --source-token with read access)"},
		{"not found with token", http.StatusNotFound, "tok", "source-access",
			"acme/private not found or token lacks access"},
		{"unauthorized", http.StatusUnauthorized, "bad", "source-auth",
			"source token is invalid or expired"},
		{"forbidden", http.StatusForbidden, "tok", "source-auth",
			"access denied to acme/private; check token permissions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotAuth string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotAuth = r.URL.Path, r.Header.Get("Authorization")
				w.WriteHeader(tt.status)
				fmt.Fprint(w, `{}`)
			}))
			defer ts.Close()

			err := checkSourceRepo(context.Background(), sourceFromServer(t, ts, "gitea", "acme", "private"), tt.token)

			if gotPath != "/api/v1/repos/acme/private" {
				t.Errorf("path = %q", gotPath)
			}
			wantAuth := ""
			if tt.token != "" {
				wantAuth = "token " + tt.token
			}
			if gotAuth != wantAuth {
				t.Errorf("Authorization = %q, want %q", gotAuth, wantAuth)
			}
			assertPreflight(t, err, tt.wantCheck, tt.wantMsg)
		})
	}
}

func TestCheckSourceRepo_GitLab(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		token     string
		wantCheck string
		wantMsg   string
	}{
		{"ok", http.StatusOK, "tok", "", ""},
		{"not found no token", http.StatusNotFound, "", "source-access",
			"acme/private not found (if private, provide --source-token with read access)"},
		{"unauthorized", http.StatusUnauthorized, "bad", "source-auth",
			"source token is invalid or expired"},
		{"forbidden", http.StatusForbidden, "tok", "source-auth",
			"access denied to acme/private; check token permissions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotToken, gotAuth string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.EscapedPath()
				gotToken, gotAuth = r.Header.Get("PRIVATE-TOKEN"), r.Header.Get("Authorization")
				w.WriteHeader(tt.status)
				fmt.Fprint(w, `{}`)
			}))
			defer ts.Close()

			err := checkSourceRepo(context.Background(), sourceFromServer(t, ts, "gitlab", "acme", "private"), tt.token)

			if gotPath != "/api/v4/projects/acme%2Fprivate" {
				t.Errorf("path = %q", gotPath)
			}
			if gotToken != tt.token || gotAuth != "" {
				t.Errorf("PRIVATE-TOKEN = %q, Authorization = %q; want token %q only", gotToken, gotAuth, tt.token)
			}
			assertPreflight(t, err, tt.wantCheck, tt.wantMsg)
		})
	}
}

func TestPreflightSingleRepo_GiteaNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	err := PreflightSingleRepo(context.Background(), nil, sourceFromServer(t, ts, "gitea", "acme", "private"), "", "me")
	assertPreflight(t, err, "source-access",
		"acme/private not found (if private, provide --source-token with read access)")
}

func assertPreflight(t *testing.T, err error, wantCheck, wantMsg string) {
	t.Helper()
	if wantCheck == "" {
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		return
	}
	pe, ok := err.(*PreflightError)
	if !ok {
		t.Fatalf("expected PreflightError, got %T: %v", err, err)
	}
	if pe.Check != wantCheck {
		t.Errorf("Check = %q, want %q", pe.Check, wantCheck)
	}
	if pe.Message != wantMsg {
		t.Errorf("Message = %q, want %q", pe.Message, wantMsg)
	}
}
