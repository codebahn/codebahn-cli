package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListGitHubRepos(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"name":"api","full_name":"acme/api","clone_url":"https://github.com/acme/api.git","description":"API","private":false}]`)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	repos, err := ListGitHubRepos(context.Background(), "acme", "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "api" {
		t.Errorf("Name = %q, want %q", repos[0].Name, "api")
	}
	if repos[0].FullName != "acme/api" {
		t.Errorf("FullName = %q, want %q", repos[0].FullName, "acme/api")
	}
}

func TestListGitHubReposFallbackToUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/orgs/simon/repos" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"message":"Not Found"}`)
			return
		}
		if r.URL.Path == "/users/simon/repos" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `[{"name":"dotfiles","full_name":"simon/dotfiles","clone_url":"https://github.com/simon/dotfiles.git"}]`)
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	repos, err := ListGitHubRepos(context.Background(), "simon", "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0].Name != "dotfiles" {
		t.Errorf("expected dotfiles, got %v", repos)
	}
}

func TestListGitHubReposPagination(t *testing.T) {
	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "2" {
			w.Header().Set("Link", fmt.Sprintf(`<%s%s?page=2>; rel="next"`, serverURL, r.URL.Path))
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `[{"name":"repo1","full_name":"acme/repo1","clone_url":"https://github.com/acme/repo1.git"}]`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"name":"repo2","full_name":"acme/repo2","clone_url":"https://github.com/acme/repo2.git"}]`)
	}))
	defer ts.Close()
	serverURL = ts.URL

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	repos, err := ListGitHubRepos(context.Background(), "acme", "tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
}

func TestListGitHubReposRateLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprint(w, `{"message":"rate limit exceeded"}`)
	}))
	defer ts.Close()

	orig := githubAPIBase
	SetGitHubAPIBase(ts.URL)
	defer SetGitHubAPIBase(orig)

	_, err := ListGitHubRepos(context.Background(), "acme", "tok")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	if got := err.Error(); got != "GitHub API rate limit exceeded; authenticate with --token or GITHUB_TOKEN" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestParseNextLink(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{`<https://api.github.com/orgs/acme/repos?page=2>; rel="next", <https://api.github.com/orgs/acme/repos?page=5>; rel="last"`, "https://api.github.com/orgs/acme/repos?page=2"},
		{`<https://api.github.com/orgs/acme/repos?page=5>; rel="last"`, ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := parseNextLink(tt.header); got != tt.want {
			t.Errorf("parseNextLink(%q) = %q, want %q", tt.header, got, tt.want)
		}
	}
}
