package main

import (
	"testing"

	"github.com/codebahn/codebahn-cli/tools"
)

func TestToolsJSONCount(t *testing.T) {
	entries := buildToolsJSON()

	if len(entries) != len(tools.All) {
		t.Fatalf("got %d entries, want %d", len(entries), len(tools.All))
	}

	for i, td := range tools.All {
		if entries[i].Name != td.Name {
			t.Errorf("entries[%d].Name = %q, want %q", i, entries[i].Name, td.Name)
		}
	}
}
