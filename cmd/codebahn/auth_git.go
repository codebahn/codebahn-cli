package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/codebahn/codebahn-cli/internal/config"
	"github.com/codebahn/codebahn-cli/internal/oauth"
)

// refreshWindow is how far ahead of expiry a git-credential get call
// refreshes the access token, so git never receives a credential that
// expires mid-request.
const refreshWindow = 60 * time.Second

// gitCredentialCmd implements the git credential helper protocol
// (see git-credential(1)) so git can authenticate HTTPS requests to the
// configured Codebahn host with the CLI's stored OAuth token.
func gitCredentialCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "git-credential <get|store|erase>",
		Short:        "Git credential helper (see git-credential(1))",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "get":
				return runGitCredentialGet(cmd)
			case "store", "erase":
				// Nothing is stored or erased; the helper always re-derives the
				// credential from the CLI's own config. Drain stdin so git never
				// sees a broken pipe.
				if _, err := io.Copy(io.Discard, cmd.InOrStdin()); err != nil {
					return fmt.Errorf("reading credential input: %w", err)
				}
				return nil
			default:
				return fmt.Errorf("unknown git-credential operation %q", args[0])
			}
		},
	}
}

func runGitCredentialGet(cmd *cobra.Command) error {
	fields, err := parseCredentialInput(cmd.InOrStdin())
	if err != nil {
		return fmt.Errorf("reading credential input: %w", err)
	}

	auth := resolveAuth()
	if fields["protocol"] != "https" || !hostMatches(fields["host"], auth.URL) {
		return nil
	}
	if auth.AccessToken == "" {
		return nil
	}

	token := auth.AccessToken
	// A zero TokenExpiry means the token doesn't expire (e.g. a static
	// token from CODEBAHN_TOKEN or --token).
	if auth.TokenExpiry != 0 && tokenExpiringSoon(auth.TokenExpiry) {
		if auth.RefreshToken == "" {
			// Nothing to refresh with; fall through silently rather than
			// hand git a stale credential.
			return nil
		}
		refreshed, err := refreshTokenLocked(cmd.Context(), auth)
		if err != nil {
			return fmt.Errorf("refreshing access token: %w; run 'codebahn auth login' to re-authenticate", err)
		}
		token = refreshed
	}

	fmt.Fprintf(cmd.OutOrStdout(), "username=oauth2\npassword=%s\n", token)
	return nil
}

// tokenExpiringSoon reports whether a token with the given expiry (unix
// seconds) expires within refreshWindow.
func tokenExpiringSoon(expiry int64) bool {
	return time.Unix(expiry, 0).Sub(time.Now()) < refreshWindow
}

// refreshTokenLocked refreshes the access token under the cross-process
// config lock. After acquiring the lock it re-reads the config, since a
// concurrent git-credential invocation may have already refreshed the
// token; in that case it returns the already-refreshed token without
// contacting the server.
func refreshTokenLocked(ctx context.Context, auth config.Auth) (string, error) {
	lock, err := config.LockConfig()
	if err != nil {
		return "", fmt.Errorf("locking config: %w", err)
	}
	defer config.UnlockConfig(lock)

	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("loading config: %w", err)
	}

	if cfg.AccessToken != "" && !tokenExpiringSoon(cfg.TokenExpiry) {
		// A concurrent caller already refreshed the token.
		return cfg.AccessToken, nil
	}

	tokenResp, err := oauth.Refresh(ctx, auth.URL, cfg.RefreshToken)
	if err != nil {
		return "", err
	}

	cfg.AccessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		cfg.RefreshToken = tokenResp.RefreshToken
	}
	cfg.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Unix()
	if err := config.SaveConfig(cfg); err != nil {
		return "", fmt.Errorf("saving refreshed token: %w", err)
	}

	return cfg.AccessToken, nil
}

// parseCredentialInput reads key=value lines (the git credential helper
// protocol) until a blank line or EOF.
func parseCredentialInput(r io.Reader) (map[string]string, error) {
	fields := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		fields[key] = value
	}
	return fields, scanner.Err()
}

// hostMatches reports whether host (as sent by git, e.g. "codebahn.net")
// matches the host of the configured instance URL.
func hostMatches(host, configuredURL string) bool {
	if host == "" || configuredURL == "" {
		return false
	}
	configuredHost, err := authHost(configuredURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(configuredHost, host)
}

// authHost extracts the host (e.g. "codebahn.net") from a Codebahn instance
// URL.
func authHost(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing instance URL %q: %w", rawURL, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("instance URL %q has no host", rawURL)
	}
	return u.Host, nil
}
