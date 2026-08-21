package schema

import (
	"encoding/json"
	"testing"

	"github.com/codebahn/codebahn-cli/tools"
)

func TestForProducesValidJSON(t *testing.T) {
	for _, td := range tools.All {
		raw := For(td)
		if !json.Valid(raw) {
			t.Errorf("tool %s: For() produced invalid JSON", td.Name)
		}
	}
}

func TestForObjectType(t *testing.T) {
	for _, td := range tools.All {
		raw := For(td)
		var schema map[string]any
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Errorf("tool %s: unmarshal error: %v", td.Name, err)
			continue
		}
		if schema["type"] != "object" {
			t.Errorf("tool %s: type = %v, want object", td.Name, schema["type"])
		}
	}
}

func TestForEmptyStruct(t *testing.T) {
	td := tools.ByName("get_my_user_info")
	raw := For(td)
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if schema["type"] != "object" {
		t.Errorf("type = %v, want object", schema["type"])
	}
}

func TestForRequiredFields(t *testing.T) {
	td := tools.ByName("create_issue")
	raw := For(td)
	var schema struct {
		Type       string         `json:"type"`
		Properties map[string]any `json:"properties"`
		Required   []string       `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	requiredSet := map[string]bool{}
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	if !requiredSet["owner"] {
		t.Error("owner should be required")
	}
	if !requiredSet["repo"] {
		t.Error("repo should be required")
	}
	if !requiredSet["title"] {
		t.Error("title should be required")
	}
	if requiredSet["body"] {
		t.Error("body should not be required")
	}
}

func TestForPropertyTypes(t *testing.T) {
	td := tools.ByName("get_file_content")
	raw := For(td)
	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	cases := []struct {
		field    string
		wantType string
	}{
		{"owner", "string"},
		{"repo", "string"},
		{"with_metadata", "boolean"},
		{"start_line", "number"},
	}
	for _, tc := range cases {
		prop, ok := schema.Properties[tc.field]
		if !ok {
			t.Errorf("missing property %s", tc.field)
			continue
		}
		if prop["type"] != tc.wantType {
			t.Errorf("property %s: type = %v, want %s", tc.field, prop["type"], tc.wantType)
		}
	}
}

func TestForDescription(t *testing.T) {
	td := tools.ByName("create_issue")
	raw := For(td)
	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	desc, ok := schema.Properties["owner"]["description"]
	if !ok || desc == "" {
		t.Error("owner property should have a description")
	}
}

func TestForDefaultValue(t *testing.T) {
	td := tools.ByName("list_repo_issues")
	raw := For(td)
	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	state := schema.Properties["state"]
	if state["default"] != "open" {
		t.Errorf("state default = %v, want open", state["default"])
	}
}

func TestForAllTools(t *testing.T) {
	for _, td := range tools.All {
		raw := For(td)
		var schema struct {
			Type       string                    `json:"type"`
			Properties map[string]map[string]any `json:"properties"`
			Required   []string                  `json:"required"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Errorf("tool %s: unmarshal: %v", td.Name, err)
			continue
		}
		if schema.Type != "object" {
			t.Errorf("tool %s: type = %s, want object", td.Name, schema.Type)
		}
		for _, prop := range schema.Properties {
			if _, ok := prop["type"]; !ok {
				t.Errorf("tool %s: property missing type", td.Name)
			}
			if _, ok := prop["description"]; !ok {
				t.Errorf("tool %s: property missing description", td.Name)
			}
		}
	}
}
