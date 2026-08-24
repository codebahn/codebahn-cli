package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
