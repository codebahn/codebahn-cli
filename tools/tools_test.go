package tools

import (
	"reflect"
	"testing"
)

func TestAllCount(t *testing.T) {
	if got := len(All); got != 57 {
		t.Errorf("len(All) = %d, want 57", got)
	}
}

func TestReviewTools(t *testing.T) {
	cases := []struct {
		name, group, cliName, method, pathTmpl string
	}{
		{"list_pr_commits", "pr", "commits", "GET", "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/commits"},
		{"get_commit_diff", "repo", "show", "GET", "/repos/{{.Owner}}/{{.Repo}}/git/commits/{{.SHA}}.diff"},
		{"compare_refs", "repo", "compare", "GET", "/repos/{{.Owner}}/{{.Repo}}/compare/{{.Base}}...{{.Head}}"},
	}
	for _, tc := range cases {
		td := ByName(tc.name)
		if td.Group != tc.group {
			t.Errorf("%s: Group = %q, want %q", tc.name, td.Group, tc.group)
		}
		if td.CLIName != tc.cliName {
			t.Errorf("%s: CLIName = %q, want %q", tc.name, td.CLIName, tc.cliName)
		}
		if td.Method != tc.method {
			t.Errorf("%s: Method = %q, want %q", tc.name, td.Method, tc.method)
		}
		if td.PathTmpl != tc.pathTmpl {
			t.Errorf("%s: PathTmpl = %q, want %q", tc.name, td.PathTmpl, tc.pathTmpl)
		}
	}
}

// Fields tagged api:"-" are consumed by the tool implementation (MCP handler
// or CLI) and must never be sent to the REST API. They are still part of the
// tool's schema. The tag has exactly one valid value.
func TestAPITagValues(t *testing.T) {
	for _, td := range All {
		rt := reflect.TypeOf(td.Args)
		for i := range rt.NumField() {
			f := rt.Field(i)
			if tag, ok := f.Tag.Lookup("api"); ok && tag != "-" {
				t.Errorf("tool %s: field %s has api:%q, want api:\"-\" or no tag", td.Name, f.Name, tag)
			}
		}
	}
}

func TestDiffFilePathIsLocal(t *testing.T) {
	for _, name := range []string{"get_pull_request_diff", "get_commit_diff"} {
		rt := reflect.TypeOf(ByName(name).Args)
		f, ok := rt.FieldByName("FilePath")
		if !ok {
			t.Errorf("tool %s: missing FilePath field", name)
			continue
		}
		if f.Tag.Get("api") != "-" {
			t.Errorf("tool %s: FilePath must be tagged api:\"-\" (the REST API does not filter diffs by file)", name)
		}
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
