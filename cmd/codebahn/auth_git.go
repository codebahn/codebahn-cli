package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// gitCredentialCmd implements the git credential helper protocol
// (see git-credential(1)) so git can authenticate HTTPS requests to the
// configured Codebahn host with the CLI's stored OAuth token.
func gitCredentialCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "git-credential <get|store|erase>",
		Short: "Git credential helper (see git-credential(1))",
		Args:  cobra.ExactArgs(1),
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
	// A zero TokenExpiry means the token doesn't expire (e.g. a static
	// token from CODEBAHN_TOKEN or --token). A refresh for a token that is
	// expired or expiring soon is handled separately; until then, fall
	// through silently rather than hand git a stale credential.
	if auth.TokenExpiry != 0 && auth.TokenExpiry <= time.Now().Unix() {
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "username=oauth2\npassword=%s\n", auth.AccessToken)
	return nil
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
	u, err := url.Parse(configuredURL)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, host)
}
