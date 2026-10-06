package tools

import (
	"testing"
)

func TestMCPRegistryCoversAllTools(t *testing.T) {
	covered := map[string]bool{}
	for _, mt := range MCPAll {
		for _, a := range mt.Actions {
			covered[a.ToolName] = true
		}
	}

	skip := map[string]bool{
		"convert_repo":        true, // CLI-only
		"issue_state_change":  true, // absorbed into issues update (state param)
		"update_variable":     true, // absorbed into ci_config set (upsert)
		"update_org_variable": true, // absorbed into ci_config set (upsert)
	}

	for _, td := range All {
		if skip[td.Name] {
			continue
		}
		if !covered[td.Name] {
			t.Errorf("tool %q is not covered by any MCPTool", td.Name)
		}
	}
}

func TestMCPRegistryToolNamesExist(t *testing.T) {
	allNames := map[string]bool{}
	for _, td := range All {
		allNames[td.Name] = true
	}

	for _, mt := range MCPAll {
		for _, a := range mt.Actions {
			if !allNames[a.ToolName] {
				t.Errorf("MCPTool %q action %q references unknown tool %q", mt.Name, a.Action, a.ToolName)
			}
		}
	}
}

func TestMCPRegistryCount(t *testing.T) {
	if got := len(MCPAll); got != 15 {
		t.Errorf("MCPAll has %d tools, want 15", got)
	}
}

func TestMCPRegistryUniqueNames(t *testing.T) {
	seen := map[string]bool{}
	for _, mt := range MCPAll {
		if seen[mt.Name] {
			t.Errorf("duplicate MCPTool name %q", mt.Name)
		}
		seen[mt.Name] = true
	}
}

func TestMCPActionKeys(t *testing.T) {
	for _, mt := range MCPAll {
		seen := map[string]bool{}
		for _, a := range mt.Actions {
			key := a.Key()
			if seen[key] {
				t.Errorf("MCPTool %q has duplicate action key %q", mt.Name, key)
			}
			seen[key] = true
		}
	}
}

func TestMCPByName(t *testing.T) {
	mt := MCPByName("files")
	if mt.Name != "files" {
		t.Errorf("MCPByName('files') = %q, want 'files'", mt.Name)
	}
}

func TestMCPByNamePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MCPByName with unknown name should panic")
		}
	}()
	MCPByName("nonexistent")
}
