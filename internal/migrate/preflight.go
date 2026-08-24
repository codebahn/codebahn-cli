package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/codebahn/codebahn-cli/client"
)

type PreflightError struct {
	Check   string
	Message string
}

func (e *PreflightError) Error() string {
	return fmt.Sprintf("preflight: %s", e.Message)
}

func PreflightSingleRepo(ctx context.Context, c *client.Client, src Source, sourceToken, destOwner string) error {
	if src.Service == "github" {
		if err := checkGitHubRepo(ctx, src, sourceToken); err != nil {
			return err
		}
	}

	if err := checkDestNotExists(ctx, c, destOwner, src.Repo); err != nil {
		return err
	}

	return nil
}

func checkGitHubRepo(ctx context.Context, src Source, token string) error {
	url := fmt.Sprintf("%s/repos/%s/%s", githubAPIBase, src.Owner, src.Repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("preflight: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("preflight: checking source repo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))

	switch resp.StatusCode {
	case http.StatusOK:
		// Check if private and no token
		var repo struct {
			Private bool `json:"private"`
		}
		if json.Unmarshal(body, &repo) == nil && repo.Private && token == "" {
			return &PreflightError{
				Check:   "source-auth",
				Message: fmt.Sprintf("%s/%s is private; provide --source-token or set GITHUB_TOKEN", src.Owner, src.Repo),
			}
		}

		checkGitHubTokenScopes(resp, token, src)
		return nil

	case http.StatusNotFound:
		if token == "" {
			return &PreflightError{
				Check:   "source-access",
				Message: fmt.Sprintf("%s/%s not found (if private, provide --source-token or set GITHUB_TOKEN)", src.Owner, src.Repo),
			}
		}
		return &PreflightError{
			Check:   "source-access",
			Message: fmt.Sprintf("%s/%s not found or token lacks access", src.Owner, src.Repo),
		}

	case http.StatusUnauthorized:
		return &PreflightError{
			Check:   "source-auth",
			Message: "source token is invalid or expired",
		}

	case http.StatusForbidden:
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
			return &PreflightError{
				Check:   "rate-limit",
				Message: "GitHub API rate limit exceeded; provide --source-token or wait",
			}
		}
		return &PreflightError{
			Check:   "source-auth",
			Message: fmt.Sprintf("access denied to %s/%s; check token permissions", src.Owner, src.Repo),
		}

	default:
		return &PreflightError{
			Check:   "source-access",
			Message: fmt.Sprintf("GitHub API returned HTTP %d for %s/%s", resp.StatusCode, src.Owner, src.Repo),
		}
	}
}

func checkGitHubTokenScopes(resp *http.Response, token string, src Source) {
	if token == "" {
		return
	}

	scopes := resp.Header.Get("X-OAuth-Scopes")
	if scopes == "" {
		// Fine-grained PAT or GitHub App token: no X-OAuth-Scopes header.
		// Can't check scopes; the 200 response already proves access.
		return
	}

	// Classic PAT: X-OAuth-Scopes lists granted scopes.
	// Private repos need "repo" scope.
	var repo struct {
		Private bool `json:"private"`
	}
	// We already consumed the body, but the caller checked private above.
	// For scope checking, if we got a 200 on a private repo, the token works.
	// The scope warning is only for classic PATs without "repo" that somehow
	// got access to a public repo but would fail on private ones.
	_ = repo

	scopeList := strings.Split(scopes, ", ")
	for _, s := range scopeList {
		s = strings.TrimSpace(s)
		if s == "repo" || s == "public_repo" {
			return
		}
	}

	fmt.Fprintf(
		os.Stderr,
		"warning: token scopes [%s] may be insufficient for private repos; 'repo' scope is recommended\n",
		scopes,
	)
}

func checkDestNotExists(ctx context.Context, c *client.Client, owner, repo string) error {
	_, err := c.GetRaw(ctx, fmt.Sprintf("/repos/%s/%s", owner, repo))
	if err == nil {
		return &PreflightError{
			Check:   "dest-exists",
			Message: fmt.Sprintf("%s/%s already exists on Codebahn", owner, repo),
		}
	}

	// 404 is what we want: repo doesn't exist yet
	if sce, ok := err.(*client.StatusCodeError); ok && sce.Code == http.StatusNotFound {
		return nil
	}

	// Other errors (auth issues, network): don't block the migration,
	// just skip the check
	return nil
}
