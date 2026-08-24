package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/codebahn/codebahn-cli/client"
	"github.com/codebahn/codebahn-cli/internal/gen"
	"github.com/codebahn/codebahn-cli/internal/output"
)

type migrateResult struct {
	Repo   string `json:"repo"`
	Status string `json:"status,omitempty"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

func MirrorCmd() *cobra.Command {
	var (
		owner   string
		private bool
		public  bool
		token   string
		yes     bool
		include StringSlice
		exclude StringSlice
	)

	cmd := &cobra.Command{
		Use:   "mirror <source>",
		Short: "Mirror a repository from GitHub/GitLab",
		Long:  "Create a pull mirror on Codebahn that stays in sync with the source repository.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if private && public {
				return fmt.Errorf("--private and --public are mutually exclusive")
			}

			src, err := ParseSource(args[0])
			if err != nil {
				return err
			}

			c := gen.ClientFrom(cmd.Context())
			if c == nil {
				return fmt.Errorf("not logged in; run 'codebahn auth login' first")
			}

			ctx := cmd.Context()
			destOwner, err := ResolveDestOwner(ctx, c, owner)
			if err != nil {
				return err
			}

			sourceToken := ResolveSourceToken(token)
			jsonMode, _ := cmd.Flags().GetBool("json")
			p := output.NewPrinter(os.Stdout, jsonMode)

			if src.IsOrg() {
				return mirrorOrg(ctx, c, src, destOwner, sourceToken, private, public, include, exclude, yes, p)
			}
			return mirrorOne(ctx, c, src, destOwner, sourceToken, public, p)
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Destination org on Codebahn (default: your username)")
	cmd.Flags().BoolVar(&private, "private", false, "Make mirror private")
	cmd.Flags().BoolVar(&public, "public", false, "Make mirror public")
	cmd.Flags().StringVar(&token, "token", "", "Source PAT (or set GITHUB_TOKEN)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation in org mode")
	cmd.Flags().Var(&include, "include", "Glob pattern to select repos (repeatable)")
	cmd.Flags().Var(&exclude, "exclude", "Glob pattern to skip repos (repeatable)")

	return cmd
}

func mirrorOne(ctx context.Context, c *client.Client, src Source, destOwner, token string, public bool, p *output.Printer) error {
	opts := MigrateRepoOptions{
		CloneAddr: src.CloneURL,
		RepoOwner: destOwner,
		RepoName:  src.Repo,
		Service:   src.Service,
		AuthToken: token,
		Mirror:    true,
		Private:   !public,
	}

	raw, err := MigrateRepo(ctx, c, opts)
	if err != nil {
		return fmt.Errorf("mirror %s: %w", src.Repo, err)
	}

	return printResult(raw, true, p)
}

func mirrorOrg(ctx context.Context, c *client.Client, src Source, destOwner, token string, private, public bool, include, exclude []string, yes bool, p *output.Printer) error {
	if token == "" {
		return fmt.Errorf("--token or GITHUB_TOKEN required to list repos from %s/%s", src.Host, src.Owner)
	}

	repos, err := ListSourceRepos(ctx, src, token)
	if err != nil {
		return err
	}

	repos = FilterRepos(repos, include, exclude)
	if len(repos) == 0 {
		return fmt.Errorf("no repositories matched")
	}

	if err := ConfirmMigration(repos, destOwner, true, yes); err != nil {
		return err
	}

	var succeeded, failed int
	var results []migrateResult

	for _, r := range repos {
		opts := MigrateRepoOptions{
			CloneAddr: r.CloneURL,
			RepoOwner: destOwner,
			RepoName:  r.Name,
			Service:   src.Service,
			AuthToken: token,
			Mirror:    true,
			Private:   !public,
		}

		raw, err := MigrateRepo(ctx, c, opts)
		if err != nil {
			failed++
			results = append(results, migrateResult{Repo: r.FullName, Error: err.Error()})
			fmt.Printf("  %s %s: %v\n", output.Red("FAIL"), r.FullName, err)
			continue
		}

		succeeded++
		res := collectResult(raw, r.FullName)
		results = append(results, res)
		fmt.Printf("  %s %s -> %s (mirror)\n", output.Green("OK"), r.FullName, res.URL)
	}

	fmt.Printf("\n%d succeeded, %d failed\n", succeeded, failed)

	if p.IsJSON() {
		p.JSON(results)
	}

	if failed > 0 {
		return fmt.Errorf("%d repositories failed to mirror", failed)
	}
	return nil
}

func ImportCmd() *cobra.Command {
	var (
		owner        string
		private      bool
		public       bool
		token        string
		yes          bool
		noIssues     bool
		noPRs        bool
		noLabels     bool
		noMilestones bool
		noReleases   bool
		noWiki       bool
		lfs          bool
		include      StringSlice
		exclude      StringSlice
	)

	cmd := &cobra.Command{
		Use:   "import <source>",
		Short: "Import a repository from GitHub/GitLab",
		Long:  "One-time import that copies code and metadata. The Codebahn repo is independent after import.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if private && public {
				return fmt.Errorf("--private and --public are mutually exclusive")
			}

			src, err := ParseSource(args[0])
			if err != nil {
				return err
			}

			c := gen.ClientFrom(cmd.Context())
			if c == nil {
				return fmt.Errorf("not logged in; run 'codebahn auth login' first")
			}

			ctx := cmd.Context()
			destOwner, err := ResolveDestOwner(ctx, c, owner)
			if err != nil {
				return err
			}

			sourceToken := ResolveSourceToken(token)
			supportsMetadata := src.SupportsMetadata()

			meta := metadataFlags{
				Issues:     supportsMetadata && !noIssues,
				PullReqs:   supportsMetadata && !noPRs,
				Labels:     supportsMetadata && !noLabels,
				Milestones: supportsMetadata && !noMilestones,
				Releases:   supportsMetadata && !noReleases,
				Wiki:       supportsMetadata && !noWiki,
				LFS:        lfs,
			}

			if !supportsMetadata {
				fmt.Println("Source is plain git; importing code only (no issues, PRs, or other metadata).")
			}

			jsonMode, _ := cmd.Flags().GetBool("json")
			p := output.NewPrinter(os.Stdout, jsonMode)

			if src.IsOrg() {
				return importOrg(ctx, c, src, destOwner, sourceToken, public, meta, include, exclude, yes, p)
			}
			return importOne(ctx, c, src, destOwner, sourceToken, public, meta, p)
		},
	}

	cmd.Flags().StringVar(&owner, "owner", "", "Destination org on Codebahn (default: your username)")
	cmd.Flags().BoolVar(&private, "private", false, "Make imported repo private")
	cmd.Flags().BoolVar(&public, "public", false, "Make imported repo public")
	cmd.Flags().StringVar(&token, "token", "", "Source PAT (or set GITHUB_TOKEN)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation in org mode")
	cmd.Flags().BoolVar(&noIssues, "no-issues", false, "Do not import issues")
	cmd.Flags().BoolVar(&noPRs, "no-prs", false, "Do not import pull requests")
	cmd.Flags().BoolVar(&noLabels, "no-labels", false, "Do not import labels")
	cmd.Flags().BoolVar(&noMilestones, "no-milestones", false, "Do not import milestones")
	cmd.Flags().BoolVar(&noReleases, "no-releases", false, "Do not import releases")
	cmd.Flags().BoolVar(&noWiki, "no-wiki", false, "Do not import wiki")
	cmd.Flags().BoolVar(&lfs, "lfs", false, "Import LFS objects")
	cmd.Flags().Var(&include, "include", "Glob pattern to select repos (repeatable)")
	cmd.Flags().Var(&exclude, "exclude", "Glob pattern to skip repos (repeatable)")

	return cmd
}

type metadataFlags struct {
	Issues     bool
	PullReqs   bool
	Labels     bool
	Milestones bool
	Releases   bool
	Wiki       bool
	LFS        bool
}

func buildImportOpts(repoName, cloneURL, destOwner, service, token string, public bool, meta metadataFlags) MigrateRepoOptions {
	return MigrateRepoOptions{
		CloneAddr:    cloneURL,
		RepoOwner:    destOwner,
		RepoName:     repoName,
		Service:      service,
		AuthToken:    token,
		Private:      !public,
		Issues:       meta.Issues,
		PullRequests: meta.PullReqs,
		Labels:       meta.Labels,
		Milestones:   meta.Milestones,
		Releases:     meta.Releases,
		Wiki:         meta.Wiki,
		LFS:          meta.LFS,
	}
}

func importOne(ctx context.Context, c *client.Client, src Source, destOwner, token string, public bool, meta metadataFlags, p *output.Printer) error {
	opts := buildImportOpts(src.Repo, src.CloneURL, destOwner, src.Service, token, public, meta)

	raw, err := MigrateRepo(ctx, c, opts)
	if err != nil {
		return fmt.Errorf("import %s: %w", src.Repo, err)
	}

	return printResult(raw, false, p)
}

func importOrg(ctx context.Context, c *client.Client, src Source, destOwner, token string, public bool, meta metadataFlags, include, exclude []string, yes bool, p *output.Printer) error {
	if token == "" {
		return fmt.Errorf("--token or GITHUB_TOKEN required to list repos from %s/%s", src.Host, src.Owner)
	}

	repos, err := ListSourceRepos(ctx, src, token)
	if err != nil {
		return err
	}

	repos = FilterRepos(repos, include, exclude)
	if len(repos) == 0 {
		return fmt.Errorf("no repositories matched")
	}

	if err := ConfirmMigration(repos, destOwner, false, yes); err != nil {
		return err
	}

	var succeeded, failed int
	var results []migrateResult

	for _, r := range repos {
		opts := buildImportOpts(r.Name, r.CloneURL, destOwner, src.Service, token, public, meta)

		raw, err := MigrateRepo(ctx, c, opts)
		if err != nil {
			failed++
			results = append(results, migrateResult{Repo: r.FullName, Error: err.Error()})
			fmt.Printf("  %s %s: %v\n", output.Red("FAIL"), r.FullName, err)
			continue
		}

		succeeded++
		res := collectResult(raw, r.FullName)
		results = append(results, res)
		fmt.Printf("  %s %s -> %s\n", output.Green("OK"), r.FullName, res.URL)
	}

	fmt.Printf("\n%d succeeded, %d failed\n", succeeded, failed)

	if p.IsJSON() {
		p.JSON(results)
	}

	if failed > 0 {
		return fmt.Errorf("%d repositories failed to import", failed)
	}
	return nil
}

func printResult(raw json.RawMessage, isMirror bool, p *output.Printer) error {
	if p.IsJSON() {
		p.JSON(raw)
		return nil
	}

	var created struct {
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	}
	if err := json.Unmarshal(raw, &created); err != nil {
		return fmt.Errorf("could not parse response: %w", err)
	}

	mode := "import"
	if isMirror {
		mode = "mirror"
	}
	fmt.Printf("Queued: %s -> %s (%s)\n", created.FullName, created.HTMLURL, mode)
	return nil
}

func collectResult(raw []byte, fullName string) migrateResult {
	var created struct {
		HTMLURL string `json:"html_url"`
	}
	_ = json.Unmarshal(raw, &created)
	return migrateResult{Repo: fullName, Status: "queued", URL: created.HTMLURL}
}
