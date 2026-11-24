package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"

	"github.com/codebahn/codebahn-cli/client"
	"github.com/codebahn/codebahn-cli/internal/browser"
	"github.com/codebahn/codebahn-cli/internal/config"
	"github.com/codebahn/codebahn-cli/internal/gen"
	"github.com/codebahn/codebahn-cli/internal/migrate"
	"github.com/codebahn/codebahn-cli/internal/oauth"
	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/internal/update"
)

var (
	version = "dev"
	noColor bool

	openBrowser = browser.Open
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "codebahn",
		Short:   "Codebahn CLI",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if noColor {
				output.SetNoColor(true)
			}

			if cmd.Annotations[noClientAnnotation] != "" || isCobraBuiltin(cmd) {
				return nil
			}

			auth, err := resolveAuth(cmd)
			if err != nil {
				return err
			}
			if auth.AccessToken == "" {
				return nil
			}

			c := client.New(auth.URL, auth.AccessToken)
			if auth.RefreshToken != "" {
				c.SetOAuth(time.Unix(auth.TokenExpiry, 0), lockedRefresh(auth.URL))
			}

			cmd.SetContext(gen.WithClient(cmd.Context(), c))
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().String("token", "",
		"Personal access token (or set CODEBAHN_TOKEN); talks to CODEBAHN_URL, default "+config.DefaultURL)
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().Bool("json", false, "Output raw JSON")

	for _, cmd := range gen.GroupCommands() {
		if cmd.Name() == "repo" {
			cmd.AddCommand(migrate.MirrorCmd())
			cmd.AddCommand(migrate.ImportCmd())
			cmd.AddCommand(migrate.ListRemoteCmd())
		}
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(authCmd())
	rootCmd.AddCommand(updateCmd())

	return rootCmd
}

func main() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	rootCmd := newRootCmd()

	notice := checkUpdateInBackground(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		output.Errorf("%v", err)
		os.Exit(1)
	}

	if msg := notice(); msg != "" {
		fmt.Fprint(os.Stderr, msg)
	}
}

// noClientAnnotation marks commands that need no API client from the root
// pre-run: they use no credentials, or resolve them themselves.
const noClientAnnotation = "codebahn/no-client"

var noClient = map[string]string{noClientAnnotation: "true"}

// isCobraBuiltin reports whether cmd is, or sits under, one of cobra's own
// top-level commands (help, completion and the hidden shell-completion
// requests). Cobra adds them at Execute time, so they cannot carry the
// noClient annotation, and they need no credentials.
func isCobraBuiltin(cmd *cobra.Command) bool {
	top := cmd
	for top.HasParent() && top.Parent().HasParent() {
		top = top.Parent()
	}
	switch top.Name() {
	case "help", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return true
	}
	return false
}

// resolveAuth picks the credentials and server for a command: its --token
// flag if given, otherwise as config.ResolveAuth describes.
func resolveAuth(cmd *cobra.Command) (config.Auth, error) {
	token := ""
	if f := cmd.Flag("token"); f != nil {
		token = f.Value.String()
	}
	return config.ResolveAuth(token)
}

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long: `Manage authentication.

Log in with 'codebahn auth login'; --url picks the server (default
CODEBAHN_URL, else https://codebahn.net) and is saved with the login. Every
other command uses the saved login's server.

Without a saved login, pass --token or set CODEBAHN_TOKEN, and set
CODEBAHN_URL to the server (default https://codebahn.net). If CODEBAHN_URL
names another server than the saved login, commands fail instead of guessing.`,
	}
	cmd.AddCommand(authLoginCmd())
	cmd.AddCommand(authLogoutCmd())
	cmd.AddCommand(authStatusCmd())
	cmd.AddCommand(gitCredentialCmd())
	cmd.AddCommand(setupGitCmd())
	return cmd
}

func authLoginCmd() *cobra.Command {
	var loginURL string
	cmd := &cobra.Command{
		Use:         "login",
		Annotations: noClient,
		Short:       "Authenticate via browser (OAuth2 + PKCE)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			tokenResp, err := oauth.Login(cmd.Context(), loginURL, openBrowser)
			if err != nil {
				return err
			}

			cfg := config.Config{
				URL:          loginURL,
				AccessToken:  tokenResp.AccessToken,
				RefreshToken: tokenResp.RefreshToken,
				TokenExpiry:  time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Unix(),
			}
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Printf("Logged in to %s.\n", loginURL)
			return nil
		},
	}
	defaultURL := os.Getenv("CODEBAHN_URL")
	if defaultURL == "" {
		defaultURL = config.DefaultURL
	}
	cmd.Flags().StringVar(&loginURL, "url", defaultURL, "Codebahn server to log in to; saved with the login and used by every other command")
	return cmd
}

func authLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "logout",
		Annotations: noClient,
		Short:       "Clear saved tokens",
		RunE: func(*cobra.Command, []string) error {
			if err := config.ClearTokens(); err != nil {
				return err
			}
			migrate.ClearGitHubSession()
			fmt.Println("Logged out.")
			return nil
		},
	}
}

func authStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Verify connection",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c := gen.ClientFrom(cmd.Context())
			if c == nil {
				return fmt.Errorf("not logged in; run 'codebahn auth login' first")
			}

			raw, err := c.GetRaw(cmd.Context(), "/user")
			if err != nil {
				return fmt.Errorf("verifying connection: %w", err)
			}

			var user struct {
				Login    string `json:"login"`
				FullName string `json:"full_name"`
				Email    string `json:"email"`
				IsAdmin  bool   `json:"is_admin"`
			}
			if err := json.Unmarshal(raw, &user); err != nil {
				return fmt.Errorf("parsing user info: %w", err)
			}

			fmt.Printf("Logged in to %s as %s", c.BaseURL(), user.Login)
			if user.FullName != "" {
				fmt.Printf(" (%s)", user.FullName)
			}
			fmt.Println()
			return nil
		},
	}
}

func checkUpdateInBackground(rootCmd *cobra.Command) func() string {
	ch := make(chan string, 1)

	origPreRun := rootCmd.PersistentPreRunE
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := origPreRun(cmd, args); err != nil {
			return err
		}

		if cmd.Name() == "update" {
			return nil
		}

		cfg, _ := config.LoadConfig()
		if !update.ShouldCheck(version, cfg.CheckUpdates) {
			return nil
		}

		go func() {
			rel, err := update.CheckLatest(version)
			if err != nil {
				ch <- ""
				return
			}
			update.RecordCheck()
			if rel.Newer {
				var buf bytes.Buffer
				update.PrintUpdateNotice(&buf, version, rel.Version)
				ch <- buf.String()
			} else {
				ch <- ""
			}
		}()

		return nil
	}

	return func() string {
		select {
		case msg := <-ch:
			return msg
		default:
			return ""
		}
	}
}
