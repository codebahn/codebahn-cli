package tools

import (
	"reflect"
	"testing"
)

func TestAllCount(t *testing.T) {
	if got := len(All); got != 54 {
		t.Errorf("len(All) = %d, want 54", got)
	}
}

func TestNoDuplicateNames(t *testing.T) {
	seen := map[string]bool{}
	for _, td := range All {
		if seen[td.Name] {
			t.Errorf("duplicate tool name: %s", td.Name)
		}
		seen[td.Name] = true
	}
}

func TestAllFieldsPopulated(t *testing.T) {
	for _, td := range All {
		if td.Name == "" {
			t.Error("tool with empty Name")
		}
		if td.Group == "" {
			t.Errorf("tool %s: empty Group", td.Name)
		}
		if td.CLIName == "" {
			t.Errorf("tool %s: empty CLIName", td.Name)
		}
		if td.Description == "" {
			t.Errorf("tool %s: empty Description", td.Name)
		}
		if td.Method == "" {
			t.Errorf("tool %s: empty Method", td.Name)
		}
		if td.PathTmpl == "" {
			t.Errorf("tool %s: empty PathTmpl", td.Name)
		}
		if td.Args == nil {
			t.Errorf("tool %s: nil Args", td.Name)
		}
	}
}

func TestStructTags(t *testing.T) {
	for _, td := range All {
		rt := reflect.TypeOf(td.Args)
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct {
			t.Errorf("tool %s: Args is %s, want struct", td.Name, rt.Kind())
			continue
		}
		for i := range rt.NumField() {
			f := rt.Field(i)
			if f.Tag.Get("json") == "" {
				t.Errorf("tool %s: field %s missing json tag", td.Name, f.Name)
			}
			if f.Tag.Get("desc") == "" {
				t.Errorf("tool %s: field %s missing desc tag", td.Name, f.Name)
			}
		}
	}
}

func TestByName(t *testing.T) {
	td := ByName("create_issue")
	if td.Group != "issue" {
		t.Errorf("ByName(create_issue).Group = %q, want %q", td.Group, "issue")
	}
}

func TestByNamePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ByName with unknown tool did not panic")
		}
	}()
	ByName("nonexistent_tool")
}

func TestByGroup(t *testing.T) {
	issues := ByGroup("issue")
	if len(issues) == 0 {
		t.Error("ByGroup(issue) returned empty")
	}
	for _, td := range issues {
		if td.Group != "issue" {
			t.Errorf("ByGroup(issue) returned tool %s in group %s", td.Name, td.Group)
		}
	}
}

func TestGroups(t *testing.T) {
	groups := Groups()
	if len(groups) == 0 {
		t.Error("Groups() returned empty")
	}
	seen := map[string]bool{}
	for _, g := range groups {
		if seen[g] {
			t.Errorf("duplicate group: %s", g)
		}
		seen[g] = true
	}
}

func TestMethodValues(t *testing.T) {
	valid := map[string]bool{
		"GET": true, "POST": true, "PATCH": true, "PUT": true, "DELETE": true,
	}
	for _, td := range All {
		if !valid[td.Method] {
			t.Errorf("tool %s: invalid Method %q", td.Name, td.Method)
		}
	}
}
