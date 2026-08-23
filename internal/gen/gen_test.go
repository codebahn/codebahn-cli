package gen

import (
	"testing"

	"github.com/codebahn/codebahn-cli/tools"
)

func TestGroupCommandsCount(t *testing.T) {
	cmds := GroupCommands()
	groups := tools.Groups()
	if got, want := len(cmds), len(groups); got != want {
		t.Errorf("GroupCommands() = %d commands, want %d groups", got, want)
	}
}

func TestAllToolsHaveCommands(t *testing.T) {
	cmds := GroupCommands()

	subcommands := map[string]bool{}
	for _, cmd := range cmds {
		for _, sub := range cmd.Commands() {
			key := cmd.Use + " " + sub.Use
			subcommands[key] = true
		}
	}

	for _, td := range tools.All {
		key := td.Group + " " + td.CLIName
		if !subcommands[key] {
			t.Errorf("tool %s: no generated command for %q", td.Name, key)
		}
	}
}

func TestRequiredFlagsMarked(t *testing.T) {
	cmd := IssueCreateCmd()

	// owner and repo should exist but NOT be cobra-required (context-inferred)
	for _, name := range []string{"owner", "repo"} {
		f := cmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("flag %s not found", name)
			continue
		}
		if _, ok := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; ok {
			t.Errorf("flag %s should not be cobra-required (context-inferred)", name)
		}
	}

	// title should still be cobra-required
	f := cmd.Flags().Lookup("title")
	if f == nil {
		t.Fatal("flag title not found")
	}
	if _, ok := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Error("flag title should be cobra-required")
	}

	// body should exist and not be required
	f = cmd.Flags().Lookup("body")
	if f == nil {
		t.Error("flag body not found")
	}
}

func TestEmptyArgsCommand(t *testing.T) {
	cmd := UserInfoCmd()
	if cmd.Use != "info" {
		t.Errorf("Use = %q, want info", cmd.Use)
	}
	if cmd.Flags().NFlag() != 0 {
		t.Errorf("expected no flags for get_my_user_info, got %d", cmd.Flags().NFlag())
	}
}
