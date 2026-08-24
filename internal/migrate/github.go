package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var githubAPIBase = "https://api.github.com"

func SetGitHubAPIBase(url string) { githubAPIBase = url }

type SourceRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	CloneURL    string `json:"clone_url"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

func ListGitHubRepos(ctx context.Context, owner, token string) ([]SourceRepo, error) {
	repos, err := fetchGitHubRepos(ctx, fmt.Sprintf("%s/orgs/%s/repos?per_page=100", githubAPIBase, owner), token)
	if err != nil {
		var httpErr *ghHTTPError
		if isGHHTTPError(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			return fetchGitHubRepos(ctx, fmt.Sprintf("%s/users/%s/repos?per_page=100&type=all", githubAPIBase, owner), token)
		}
		return nil, err
	}
	return repos, nil
}

func fetchGitHubRepos(ctx context.Context, startURL, token string) ([]SourceRepo, error) {
	var all []SourceRepo
	nextURL := startURL

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		req.Header.Set("Accept", "application/vnd.github+json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("GitHub API request: %w", err)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading GitHub response: %w", readErr)
		}

		if resp.StatusCode == http.StatusForbidden {
			if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
				return nil, fmt.Errorf("GitHub API rate limit exceeded; authenticate with --token or GITHUB_TOKEN")
			}
		}

		if resp.StatusCode != http.StatusOK {
			return nil, &ghHTTPError{StatusCode: resp.StatusCode, Body: string(body)}
		}

		var page []ghRepo
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("decoding GitHub response: %w", err)
		}

		for _, r := range page {
			all = append(all, SourceRepo(r))
		}

		nextURL = parseNextLink(resp.Header.Get("Link"))
	}

	return all, nil
}

type ghRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	CloneURL    string `json:"clone_url"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

type ghHTTPError struct {
	StatusCode int
	Body       string
}

func (e *ghHTTPError) Error() string {
	return fmt.Sprintf("GitHub API: HTTP %d: %s", e.StatusCode, e.Body)
}

func isGHHTTPError(err error, target **ghHTTPError) bool {
	if e, ok := err.(*ghHTTPError); ok {
		*target = e
		return true
	}
	return false
}

var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func parseNextLink(header string) string {
	if header == "" {
		return ""
	}
	for _, part := range strings.Split(header, ",") {
		if m := linkNextRe.FindStringSubmatch(part); m != nil {
			return m[1]
		}
	}
	return ""
}
