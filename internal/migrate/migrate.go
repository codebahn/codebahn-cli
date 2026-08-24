package migrate

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/codebahn/codebahn-cli/client"
	"github.com/codebahn/codebahn-cli/internal/output"
)

type StringSlice []string

func (s *StringSlice) String() string     { return strings.Join(*s, ",") }
func (s *StringSlice) Set(v string) error { *s = append(*s, v); return nil }
func (s *StringSlice) Type() string       { return "string" }

type MigrateRepoOptions struct {
	CloneAddr    string `json:"clone_addr"`
	RepoOwner    string `json:"repo_owner,omitempty"`
	RepoName     string `json:"repo_name"`
	Service      string `json:"service"`
	AuthToken    string `json:"auth_token,omitempty"`
	Mirror       bool   `json:"mirror"`
	Private      bool   `json:"private"`
	Wiki         bool   `json:"wiki"`
	Milestones   bool   `json:"milestones"`
	Labels       bool   `json:"labels"`
	Issues       bool   `json:"issues"`
	PullRequests bool   `json:"pull_requests"`
	Releases     bool   `json:"releases"`
	LFS          bool   `json:"lfs"`
}

func MigrateRepo(ctx context.Context, c *client.Client, opts MigrateRepoOptions) (json.RawMessage, error) {
	return c.PostRaw(ctx, "/repos/migrate", opts)
}

func ResolveDestOwner(ctx context.Context, c *client.Client, ownerFlag string) (string, error) {
	if ownerFlag != "" {
		return ownerFlag, nil
	}
	raw, err := c.GetRaw(ctx, "/user")
	if err != nil {
		return "", fmt.Errorf("could not determine destination owner: %w", err)
	}
	var user struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(raw, &user); err != nil {
		return "", fmt.Errorf("could not parse user info: %w", err)
	}
	return user.Login, nil
}

func ResolveSourceToken(tokenFlag string) string {
	if tokenFlag != "" {
		return tokenFlag
	}
	return os.Getenv("GITHUB_TOKEN")
}

func FilterRepos(repos []SourceRepo, include, exclude []string) []SourceRepo {
	var result []SourceRepo
	for _, r := range repos {
		if len(include) > 0 && !matchesAny(r.Name, include) {
			continue
		}
		if len(exclude) > 0 && matchesAny(r.Name, exclude) {
			continue
		}
		result = append(result, r)
	}
	return result
}

func matchesAny(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := path.Match(p, name); matched {
			return true
		}
	}
	return false
}

func ConfirmMigration(repos []SourceRepo, destOwner string, isMirror, yes bool) error {
	action := "Importing"
	if isMirror {
		action = "Mirroring"
	}

	fmt.Printf("Found %d repositories\n", len(repos))
	for _, r := range repos {
		fmt.Printf("  %s\n", r.FullName)
	}
	fmt.Printf("\n%s %d repositories to %s\n", action, len(repos), destOwner)

	if yes {
		return nil
	}

	if !output.IsTTY() {
		return fmt.Errorf("--yes required in non-interactive mode")
	}

	fmt.Print("Continue? [y/N] ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("cancelled")
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer != "y" && answer != "yes" {
		return fmt.Errorf("cancelled")
	}
	return nil
}

func ListSourceRepos(ctx context.Context, src Source, token string) ([]SourceRepo, error) {
	switch src.Service {
	case "github":
		return ListGitHubRepos(ctx, src.Owner, token)
	default:
		return nil, fmt.Errorf("org-level import not supported for %s; specify a full repository URL", src.Service)
	}
}
