package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"

	"github.com/codebahn/codebahn-cli/client"
	"github.com/codebahn/codebahn-cli/internal/config"
	"github.com/codebahn/codebahn-cli/internal/gen"
	"github.com/codebahn/codebahn-cli/internal/oauth"
	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/internal/update"
)

var (
	version  = "dev"
	instance string
	token    string
	noColor  bool
)

func main() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	rootCmd := &cobra.Command{
		Use:     "codebahn",
		Short:   "Codebahn CLI",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if noColor {
				output.SetNoColor(true)
			}

			if cmd.Name() == "login" || cmd.Name() == "logout" {
				return nil
			}

			auth := resolveAuth()
			if auth.AccessToken == "" {
				return nil
			}

			c := client.New(auth.URL, auth.AccessToken)
			if auth.RefreshToken != "" {
				c.SetOAuth(auth.RefreshToken, oauth.ClientID, time.Unix(auth.TokenExpiry, 0), func(access, refresh string, expiry time.Time) {
					cfg, _ := config.LoadConfig()
					cfg.AccessToken = access
					cfg.RefreshToken = refresh
					cfg.TokenExpiry = expiry.Unix()
					_ = config.SaveConfig(cfg)
				})
			}

			cmd.SetContext(gen.WithClient(cmd.Context(), c))
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().StringVar(&instance, "instance", "https://codebahn.net", "Codebahn instance URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Personal access token")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().Bool("json", false, "Output raw JSON")

	for _, cmd := range gen.GroupCommands() {
		rootCmd.AddCommand(cmd)
	}
	rootCmd.AddCommand(authCmd())
	rootCmd.AddCommand(updateCmd())

	notice := checkUpdateInBackground(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		output.Errorf("%v", err)
		os.Exit(1)
	}

	if msg := notice(); msg != "" {
		fmt.Fprint(os.Stderr, msg)
	}
}

func resolveAuth() config.Auth {
	if token != "" {
		url := instance
		return config.Auth{
			URL:         url,
			AccessToken: token,
		}
	}

	auth := config.ResolveAuth()
	if auth.URL == "" {
		auth.URL = instance
	}
	return auth
}

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}
	cmd.AddCommand(authLoginCmd())
	cmd.AddCommand(authLogoutCmd())
	cmd.AddCommand(authStatusCmd())
	return cmd
}

func authLoginCmd() *cobra.Command {
	var loginURL string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate via browser (OAuth2 + PKCE)",
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

			fmt.Println("Logged in successfully.")
			return nil
		},
	}
	cmd.Flags().StringVar(&loginURL, "url", "https://codebahn.net", "Codebahn instance URL")
	return cmd
}

func authLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear saved tokens",
		RunE: func(*cobra.Command, []string) error {
			if err := config.ClearTokens(); err != nil {
				return err
			}
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

			auth := config.ResolveAuth()
			fmt.Printf("Logged in to %s as %s", auth.URL, user.Login)
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

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return nil
	}
}
