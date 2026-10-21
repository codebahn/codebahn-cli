package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFilterRepos(t *testing.T) {
	repos := []SourceRepo{
		{Name: "api"},
		{Name: "web"},
		{Name: "old-api.archive"},
		{Name: "docs"},
		{Name: "infra.archive"},
	}

	t.Run("no filters", func(t *testing.T) {
		result := FilterRepos(repos, nil, nil)
		if len(result) != 5 {
			t.Errorf("expected 5, got %d", len(result))
		}
	})

	t.Run("exclude archive", func(t *testing.T) {
		result := FilterRepos(repos, nil, []string{"*.archive"})
		if len(result) != 3 {
			t.Errorf("expected 3, got %d", len(result))
		}
		for _, r := range result {
			if r.Name == "old-api.archive" || r.Name == "infra.archive" {
				t.Errorf("should have excluded %s", r.Name)
			}
		}
	})

	t.Run("include api and web", func(t *testing.T) {
		result := FilterRepos(repos, []string{"api", "web"}, nil)
		if len(result) != 2 {
			t.Errorf("expected 2, got %d", len(result))
		}
	})

	t.Run("include with glob", func(t *testing.T) {
		result := FilterRepos(repos, []string{"*api*"}, nil)
		if len(result) != 2 {
			t.Errorf("expected 2 (api, old-api.archive), got %d", len(result))
		}
	})

	t.Run("include and exclude", func(t *testing.T) {
		result := FilterRepos(repos, []string{"*api*"}, []string{"*.archive"})
		if len(result) != 1 {
			t.Errorf("expected 1 (api only), got %d", len(result))
		}
		if len(result) > 0 && result[0].Name != "api" {
			t.Errorf("expected api, got %s", result[0].Name)
		}
	})
}

func TestResolveSourceToken(t *testing.T) {
	t.Run("flag takes precedence", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "env-token")
		got := ResolveSourceToken("flag-token")
		if got != "flag-token" {
			t.Errorf("got %q, want %q", got, "flag-token")
		}
	})

	t.Run("falls back to env", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "env-token")
		got := ResolveSourceToken("")
		if got != "env-token" {
			t.Errorf("got %q, want %q", got, "env-token")
		}
	})

	t.Run("empty when neither set", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "")
		got := ResolveSourceToken("")
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}

func TestListGiteaOrgRepos(t *testing.T) {
	t.Run("paginates until a short page", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/orgs/acme/repos" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			page := r.URL.Query().Get("page")
			w.Header().Set("Content-Type", "application/json")
			switch page {
			case "1":
				repos := make([]string, 0, 50)
				for i := 0; i < 50; i++ {
					repos = append(repos, fmt.Sprintf(`{"name":"repo%d","full_name":"acme/repo%d","clone_url":"https://example.com/acme/repo%d.git"}`, i, i, i))
				}
				_, _ = fmt.Fprintf(w, `[%s]`, joinJSON(repos))
			case "2":
				_, _ = fmt.Fprint(w, `[{"name":"repo50","full_name":"acme/repo50","clone_url":"https://example.com/acme/repo50.git"}]`)
			default:
				_, _ = fmt.Fprint(w, `[]`)
			}
		}))
		defer ts.Close()

		scheme, host := splitTestURL(t, ts.URL)
		src := Source{Scheme: scheme, Host: host, Owner: "acme"}
		repos, err := listGiteaOrgRepos(context.Background(), src, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repos) != 51 {
			t.Fatalf("expected 51 repos, got %d", len(repos))
		}
		if repos[0].Name != "repo0" || repos[50].Name != "repo50" {
			t.Errorf("unexpected repo order: first=%q last=%q", repos[0].Name, repos[50].Name)
		}
	})

	t.Run("empty org returns empty slice", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `[]`)
		}))
		defer ts.Close()

		scheme, host := splitTestURL(t, ts.URL)
		src := Source{Scheme: scheme, Host: host, Owner: "empty-org"}
		repos, err := listGiteaOrgRepos(context.Background(), src, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repos) != 0 {
			t.Fatalf("expected empty slice, got %d repos", len(repos))
		}
	})
}

func joinJSON(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ","
		}
		result += item
	}
	return result
}
