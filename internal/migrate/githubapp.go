package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/codebahn/codebahn-cli/client"
	"github.com/codebahn/codebahn-cli/internal/output"
)

type GitHubAppConfig struct {
	Enabled  bool   `json:"enabled"`
	ClientID string `json:"client_id"`
	AppSlug  string `json:"app_slug"`
}

type Connection struct {
	InstallationID int64  `json:"installation_id"`
	AccountLogin   string `json:"account_login"`
	AccountType    string `json:"account_type"`
	Error          string `json:"error,omitempty"`
	InstallURL     string `json:"install_url,omitempty"`
}

type GitHubAppImportRequest struct {
	InstallationID int64             `json:"installation_id"`
	Repos          []ghAppImportRepo `json:"repos"`
	Owner          string            `json:"owner,omitempty"`
	Mirror         bool              `json:"mirror"`
	Issues         bool              `json:"issues"`
	PullRequests   bool              `json:"pull_requests"`
	Labels         bool              `json:"labels"`
	Milestones     bool              `json:"milestones"`
	Releases       bool              `json:"releases"`
}

type ghAppImportRepo struct {
	FullName string `json:"full_name"`
}

type ghAppImportResponse struct {
	Results []migrateResult `json:"results"`
}

func ghHeaders(ghToken string) map[string]string {
	return map[string]string{"X-GitHub-Token": ghToken}
}

func FetchGitHubAppConfig(ctx context.Context, c *client.Client) (*GitHubAppConfig, error) {
	raw, err := c.GetRaw(ctx, "/github-app/config")
	if err != nil {
		return nil, fmt.Errorf("fetching GitHub App config: %w", err)
	}
	var cfg GitHubAppConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parsing GitHub App config: %w", err)
	}
	return &cfg, nil
}

func ConnectGitHubApp(ctx context.Context, c *client.Client, ghToken, account string) (*Connection, error) {
	path := "/github-app/connect"
	if account != "" {
		path += "?account=" + url.QueryEscape(account)
	}
	raw, err := c.DoRawWithHeaders(ctx, http.MethodPost, path, nil, ghHeaders(ghToken))
	if err != nil {
		return nil, fmt.Errorf("connecting GitHub App: %w", err)
	}
	var conn Connection
	if err := json.Unmarshal(raw, &conn); err != nil {
		return nil, fmt.Errorf("parsing connection response: %w", err)
	}
	return &conn, nil
}

func ListGitHubAppRepos(ctx context.Context, c *client.Client, installationID int64, ghToken string) ([]SourceRepo, error) {
	path := fmt.Sprintf("/github-app/repos?installation_id=%d", installationID)
	raw, err := c.DoRawWithHeaders(ctx, http.MethodGet, path, nil, ghHeaders(ghToken))
	if err != nil {
		return nil, fmt.Errorf("listing GitHub App repos: %w", err)
	}

	var resp struct {
		Repos []struct {
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			Private     bool   `json:"private"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parsing repos response: %w", err)
	}

	repos := make([]SourceRepo, 0, len(resp.Repos))
	for _, r := range resp.Repos {
		name := r.FullName
		if idx := len(r.FullName) - 1; idx > 0 {
			for i := range r.FullName {
				if r.FullName[i] == '/' {
					name = r.FullName[i+1:]
					break
				}
			}
		}
		repos = append(repos, SourceRepo{
			Name:        name,
			FullName:    r.FullName,
			CloneURL:    fmt.Sprintf("https://github.com/%s.git", r.FullName),
			Description: r.Description,
			Private:     r.Private,
		})
	}
	return repos, nil
}

func ImportViaGitHubApp(ctx context.Context, c *client.Client, ghToken string, req GitHubAppImportRequest) ([]migrateResult, error) {
	raw, err := c.DoRawWithHeaders(ctx, http.MethodPost, "/github-app/import", req, ghHeaders(ghToken))
	if err != nil {
		return nil, fmt.Errorf("importing via GitHub App: %w", err)
	}
	var resp ghAppImportResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parsing import response: %w", err)
	}
	return resp.Results, nil
}

func authenticateGitHubApp(ctx context.Context, c *client.Client, account string) (ghToken string, conn *Connection, err error) {
	cfg, err := FetchGitHubAppConfig(ctx, c)
	if err != nil {
		return "", nil, err
	}
	if !cfg.Enabled {
		return "", nil, fmt.Errorf("GitHub App not configured on this instance; use --source-token instead")
	}

	code, err := RequestDeviceCode(ctx, cfg.ClientID)
	if err != nil {
		return "", nil, err
	}

	fmt.Printf("\nEnter the code at %s\n", code.VerificationURI)
	fmt.Printf("Code: %s\n\n", output.Bold(code.UserCode))

	_ = openBrowser(code.VerificationURI)

	fmt.Print("Waiting for authorization...")
	token, err := PollForToken(ctx, cfg.ClientID, code.DeviceCode, code.Interval)
	if err != nil {
		fmt.Println()
		return "", nil, err
	}
	fmt.Println(" done")

	result, err := ConnectGitHubApp(ctx, c, token, account)
	if err != nil {
		return "", nil, err
	}

	if result.Error == "not_installed" {
		return "", nil, fmt.Errorf("GitHub App not installed. Install it here:\n  %s\nThen re-run this command.", result.InstallURL)
	}
	if result.Error != "" {
		return "", nil, fmt.Errorf("GitHub App connection failed: %s", result.Error)
	}

	fmt.Printf("Connected as %s (%s)\n\n", output.Green(result.AccountLogin), result.AccountType)
	return token, result, nil
}
