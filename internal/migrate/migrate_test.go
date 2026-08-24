package migrate

import "testing"

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
