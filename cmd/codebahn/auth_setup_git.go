package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// executablePath returns the path to the running codebahn binary; it is a
// var so tests can substitute a fixed path instead of the test binary's.
var executablePath = os.Executable

// setupGitCmd configures git to use "codebahn auth git-credential" as the
// credential helper for the configured Codebahn host.
func setupGitCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "setup-git",
		Short:        "Configure git to use Codebahn's credential helper",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSetupGit(cmd)
		},
	}
}

func runSetupGit(cmd *cobra.Command) error {
	auth, err := resolveAuth()
	if err != nil {
		return err
	}
	host, err := authHost(auth.URL)
	if err != nil {
		return fmt.Errorf("determining Codebahn host: %w", err)
	}

	exe, err := executablePath()
	if err != nil {
		return fmt.Errorf("locating codebahn binary: %w", err)
	}

	key := fmt.Sprintf("credential.https://%s.helper", host)
	value := fmt.Sprintf("!%s auth git-credential", shellQuote(exe))

	// Reset the helper chain for this host first: an empty helper value
	// clears every helper configured ahead of it for a matching URL,
	// including a system-level credential.helper. --replace-all (rather
	// than a plain set) makes this idempotent, since after the first run
	// the key already has two values and a plain set refuses to collapse
	// a multi-valued key.
	if err := runGitConfig("--global", "--replace-all", key, ""); err != nil {
		return err
	}
	if err := runGitConfig("--global", "--add", key, value); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", key, value)
	return nil
}

func runGitConfig(args ...string) error {
	gitCmd := exec.Command("git", append([]string{"config"}, args...)...)
	var stderr bytes.Buffer
	gitCmd.Stderr = &stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git config %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// shellQuote wraps s in single quotes for safe use as a single shell word
// in a git credential helper command line, escaping any embedded single
// quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
