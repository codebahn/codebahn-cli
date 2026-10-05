package main

import (
	"github.com/codebahn/codebahn-cli/tools"
)

// toolsJSONEntry is one entry in the generated tools.json tool reference.
type toolsJSONEntry struct {
	Name string `json:"name"`
}

// buildToolsJSON builds the tools.json entries from tools.All, preserving
// registry order.
func buildToolsJSON() []toolsJSONEntry {
	entries := make([]toolsJSONEntry, len(tools.All))
	for i, td := range tools.All {
		entries[i] = toolsJSONEntry{Name: td.Name}
	}
	return entries
}
