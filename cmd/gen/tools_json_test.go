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

// TestToolsJSONReadOnly verifies that read_only is computed from IsReadOnly(),
// not the raw ReadOnly struct field. get_issue_by_index has Method "GET" and
// an unset (false) ReadOnly field, so the two diverge: using the raw field
// would wrongly report read_only: false.
func TestToolsJSONReadOnly(t *testing.T) {
	entries := buildToolsJSON()

	var found bool
	for i, td := range tools.All {
		if td.Name != "get_issue_by_index" {
			continue
		}
		found = true

		if td.Method != "GET" || td.ReadOnly {
			t.Fatalf("fixture assumption broken: get_issue_by_index Method=%q ReadOnly=%v", td.Method, td.ReadOnly)
		}
		if !td.IsReadOnly() {
			t.Fatalf("fixture assumption broken: get_issue_by_index IsReadOnly() = false, want true")
		}

		if entries[i].ReadOnly != true {
			t.Errorf("entries[%d] (get_issue_by_index) ReadOnly = %v, want true (from IsReadOnly(), not raw ReadOnly field)", i, entries[i].ReadOnly)
		}
	}
	if !found {
		t.Fatal("get_issue_by_index not found in tools.All")
	}
}

// TestToolsJSONParams verifies param shape (exactly name, type, required,
// desc), struct declaration order (not schema.For's alphabetical order), and
// the schema.For type vocabulary (string/number/boolean).
func TestToolsJSONParams(t *testing.T) {
	entries := buildToolsJSON()

	var params []toolsJSONParam
	for i, td := range tools.All {
		if td.Name == "list_repo_issues" {
			params = entries[i].Params
			break
		}
	}
	if params == nil {
		t.Fatal("list_repo_issues not found in tools.All")
	}

	wantParamKeys := map[string]bool{"name": true, "type": true, "required": true, "desc": true}
	for _, p := range params {
		raw, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("marshal param %q: %v", p.Name, err)
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatalf("unmarshal param %q: %v", p.Name, err)
		}
		if len(m) != len(wantParamKeys) {
			t.Errorf("param %q has %d keys, want %d: %v", p.Name, len(m), len(wantParamKeys), m)
		}
		for k := range m {
			if !wantParamKeys[k] {
				t.Errorf("param %q has unexpected key %q", p.Name, k)
			}
		}
	}

	// ListRepoIssuesArgs field order: owner, repo, state, type, milestones,
	// labels, page, limit. Mixed string/int types and required/optional.
	want := []toolsJSONParam{
		{Name: "owner", Type: "string", Required: true, Desc: "Repository owner"},
		{Name: "repo", Type: "string", Required: true, Desc: "Repository name"},
		{Name: "state", Type: "string", Required: false, Desc: "State (open|closed|all)"},
		{Name: "type", Type: "string", Required: false, Desc: "Type (issues|pulls)"},
		{Name: "milestones", Type: "string", Required: false, Desc: "Milestone names/IDs (comma-separated)"},
		{Name: "labels", Type: "string", Required: false, Desc: "Labels (comma-separated)"},
		{Name: "page", Type: "number", Required: false, Desc: "Page number (1-based)"},
		{Name: "limit", Type: "number", Required: false, Desc: "Page size"},
	}

	if len(params) != len(want) {
		t.Fatalf("got %d params, want %d: %+v", len(params), len(want), params)
	}
	for i, w := range want {
		if params[i] != w {
			t.Errorf("params[%d] = %+v, want %+v", i, params[i], w)
		}
	}
}
