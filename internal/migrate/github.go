package migrate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
	escaped := url.PathEscape(owner)
	repos, err := fetchGitHubRepos(ctx, fmt.Sprintf("%s/orgs/%s/repos?per_page=100", githubAPIBase, escaped), token)
	if err != nil {
		var he *httpError
		if isHTTPError(err, &he) && he.StatusCode == http.StatusNotFound {
			return fetchGitHubRepos(ctx, fmt.Sprintf("%s/users/%s/repos?per_page=100&type=all", githubAPIBase, escaped), token)
		}
		return nil, err
	}
	return repos, nil
}

func fetchGitHubRepos(ctx context.Context, startURL, token string) ([]SourceRepo, error) {
	repos, err := fetchPages[SourceRepo](ctx, fetchOpts{
		startURL:  startURL,
		token:     token,
		authStyle: "bearer",
		accept:    "application/vnd.github+json",
		label:     "GitHub",
	})
	if err != nil {
		var he *httpError
		if isHTTPError(err, &he) && he.StatusCode == http.StatusForbidden {
			if he.Headers.Get("X-RateLimit-Remaining") == "0" {
				return nil, fmt.Errorf("GitHub API rate limit exceeded; authenticate with --token or GITHUB_TOKEN")
			}
		}
		return nil, err
	}
	return repos, nil
}

func isHTTPError(err error, target **httpError) bool {
	return errors.As(err, target)
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
