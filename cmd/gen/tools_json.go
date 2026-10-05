package main

import (
	"github.com/codebahn/codebahn-cli/tools"
)

// toolsJSONParam is one parameter of a tools.json entry.
type toolsJSONParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Desc     string `json:"desc"`
}

// toolsJSONEntry is one entry in the generated tools.json tool reference.
type toolsJSONEntry struct {
	Name        string           `json:"name"`
	Group       string           `json:"group"`
	Description string           `json:"description"`
	Method      string           `json:"method"`
	Destructive bool             `json:"destructive"`
	ReadOnly    bool             `json:"read_only"`
	Params      []toolsJSONParam `json:"params"`
}

// buildToolsJSON builds the tools.json entries from tools.All, preserving
// registry order.
func buildToolsJSON() []toolsJSONEntry {
	entries := make([]toolsJSONEntry, len(tools.All))
	for i, td := range tools.All {
		entries[i] = toolsJSONEntry{
			Name:        td.Name,
			Group:       td.Group,
			Description: td.Description,
			Method:      td.Method,
			Destructive: td.Destructive,
			ReadOnly:    td.IsReadOnly(),
			Params:      []toolsJSONParam{},
		}
	}
	return entries
}
