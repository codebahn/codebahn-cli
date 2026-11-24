package gen

import (
	"context"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"

	"github.com/codebahn/codebahn-cli/client"
)

// The owner and repo come from the git remote on the server the command
// talks to, not from a fixed default server.
func TestResolveRepoContextUsesClientServer(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", "https://staging.codebahn.net/acme/widget.git"},
	} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	t.Chdir(dir)

	cmd := &cobra.Command{}
	cmd.SetContext(WithClient(context.Background(), client.New("https://staging.codebahn.net", "tok")))
	var args struct{ Owner, Repo string }
	if err := resolveRepoContext(cmd, &args); err != nil {
		t.Fatalf("resolveRepoContext() error = %v", err)
	}
	if args.Owner != "acme" || args.Repo != "widget" {
		t.Errorf("owner/repo = %s/%s, want acme/widget", args.Owner, args.Repo)
	}
}
