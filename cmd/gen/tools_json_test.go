package main

import (
	"encoding/json"
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

func TestToolsJSONFields(t *testing.T) {
	entries := buildToolsJSON()

	wantKeys := map[string]bool{
		"name": true, "group": true, "description": true, "method": true,
		"destructive": true, "read_only": true, "params": true,
	}

	for i, entry := range entries {
		td := tools.All[i]

		raw, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("marshal entries[%d]: %v", i, err)
		}

		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal entries[%d]: %v", i, err)
		}

		if len(m) != len(wantKeys) {
			t.Errorf("entries[%d] (%s) has %d keys, want %d: %v", i, td.Name, len(m), len(wantKeys), m)
		}
		for k := range m {
			if !wantKeys[k] {
				t.Errorf("entries[%d] (%s) has unexpected key %q", i, td.Name, k)
			}
		}

		if entry.Group != td.Group {
			t.Errorf("entries[%d].Group = %q, want %q", i, entry.Group, td.Group)
		}
		if entry.Description != td.Description {
			t.Errorf("entries[%d].Description = %q, want %q (verbatim, not backtick-escaped)", i, entry.Description, td.Description)
		}
		if entry.Method != td.Method {
			t.Errorf("entries[%d].Method = %q, want %q", i, entry.Method, td.Method)
		}
		if entry.Destructive != td.Destructive {
			t.Errorf("entries[%d].Destructive = %v, want %v", i, entry.Destructive, td.Destructive)
		}
		if entry.ReadOnly != td.IsReadOnly() {
			t.Errorf("entries[%d].ReadOnly = %v, want %v", i, entry.ReadOnly, td.IsReadOnly())
		}
		if entry.Params == nil {
			t.Errorf("entries[%d] (%s) Params is nil, want non-null array", i, td.Name)
		}
	}
}
