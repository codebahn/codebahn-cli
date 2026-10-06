package tools

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestAllCount(t *testing.T) {
	if got := len(All); got != 103 {
		t.Errorf("len(All) = %d, want 103", got)
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
func TestCISecretsVariablesTools(t *testing.T) {
	cases := []struct {
		name, group, cliName, method, pathTmpl string
	}{
		{"list_secrets", "ci", "list-secrets", "GET", "/repos/{{.Owner}}/{{.Repo}}/actions/secrets"},
		{"set_secret", "ci", "set-secret", "PUT", "/repos/{{.Owner}}/{{.Repo}}/actions/secrets/{{.SecretName}}"},
		{"delete_secret", "ci", "delete-secret", "DELETE", "/repos/{{.Owner}}/{{.Repo}}/actions/secrets/{{.SecretName}}"},
		{"list_variables", "ci", "list-variables", "GET", "/repos/{{.Owner}}/{{.Repo}}/actions/variables"},
		{"get_variable", "ci", "get-variable", "GET", "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}"},
		{"create_variable", "ci", "create-variable", "POST", "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}"},
		{"update_variable", "ci", "update-variable", "PUT", "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}"},
		{"delete_variable", "ci", "delete-variable", "DELETE", "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}"},
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
	setSecret := ByName("set_secret")
	rt := reflect.TypeOf(setSecret.Args)
	if _, ok := rt.FieldByName("Data"); !ok {
		t.Error("set_secret: missing Data field for upsert body")
	}
}

func TestCIOrgSecretsVariablesTools(t *testing.T) {
	cases := []struct {
		name, group, cliName, method, pathTmpl string
	}{
		{"list_org_secrets", "ci", "list-org-secrets", "GET", "/orgs/{{.Owner}}/actions/secrets"},
		{"set_org_secret", "ci", "set-org-secret", "PUT", "/orgs/{{.Owner}}/actions/secrets/{{.SecretName}}"},
		{"delete_org_secret", "ci", "delete-org-secret", "DELETE", "/orgs/{{.Owner}}/actions/secrets/{{.SecretName}}"},
		{"list_org_variables", "ci", "list-org-variables", "GET", "/orgs/{{.Owner}}/actions/variables"},
		{"get_org_variable", "ci", "get-org-variable", "GET", "/orgs/{{.Owner}}/actions/variables/{{.VariableName}}"},
		{"create_org_variable", "ci", "create-org-variable", "POST", "/orgs/{{.Owner}}/actions/variables/{{.VariableName}}"},
		{"update_org_variable", "ci", "update-org-variable", "PUT", "/orgs/{{.Owner}}/actions/variables/{{.VariableName}}"},
		{"delete_org_variable", "ci", "delete-org-variable", "DELETE", "/orgs/{{.Owner}}/actions/variables/{{.VariableName}}"},
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
	setOrgSecret := ByName("set_org_secret")
	rt := reflect.TypeOf(setOrgSecret.Args)
	if _, ok := rt.FieldByName("Data"); !ok {
		t.Error("set_org_secret: missing Data field for upsert body")
	}
}

func TestReleaseTools(t *testing.T) {
	cases := []struct {
		name, group, cliName, method, pathTmpl string
	}{
		{"list_releases", "release", "list", "GET", "/repos/{{.Owner}}/{{.Repo}}/releases"},
		{"create_release", "release", "create", "POST", "/repos/{{.Owner}}/{{.Repo}}/releases"},
		{"get_release", "release", "get", "GET", "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}"},
		{"get_latest_release", "release", "get-latest", "GET", "/repos/{{.Owner}}/{{.Repo}}/releases/latest"},
		{"get_release_by_tag", "release", "get-by-tag", "GET", "/repos/{{.Owner}}/{{.Repo}}/releases/tags/{{.Tag}}"},
		{"update_release", "release", "update", "PATCH", "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}"},
		{"delete_release", "release", "delete", "DELETE", "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}"},
		{"list_release_attachments", "release", "list-attachments", "GET", "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}/assets"},
		{"delete_release_attachment", "release", "delete-attachment", "DELETE", "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}/assets/{{.AttachmentID}}"},
		{"list_tags", "release", "list-tags", "GET", "/repos/{{.Owner}}/{{.Repo}}/tags"},
		{"create_tag", "release", "create-tag", "POST", "/repos/{{.Owner}}/{{.Repo}}/tags"},
		{"delete_tag", "release", "delete-tag", "DELETE", "/repos/{{.Owner}}/{{.Repo}}/tags/{{.Tag}}"},
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

func TestWebhookTools(t *testing.T) {
	cases := []struct {
		name, group, cliName, method, pathTmpl string
	}{
		{"list_hooks", "webhook", "list", "GET", "/repos/{{.Owner}}/{{.Repo}}/hooks"},
		{"get_hook", "webhook", "get", "GET", "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}"},
		{"create_hook", "webhook", "create", "POST", "/repos/{{.Owner}}/{{.Repo}}/hooks"},
		{"update_hook", "webhook", "update", "PATCH", "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}"},
		{"delete_hook", "webhook", "delete", "DELETE", "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}"},
		{"test_hook", "webhook", "test", "POST", "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}/tests"},
		{"list_org_hooks", "webhook", "list-org", "GET", "/orgs/{{.Owner}}/hooks"},
		{"get_org_hook", "webhook", "get-org", "GET", "/orgs/{{.Owner}}/hooks/{{.ID}}"},
		{"create_org_hook", "webhook", "create-org", "POST", "/orgs/{{.Owner}}/hooks"},
		{"update_org_hook", "webhook", "update-org", "PATCH", "/orgs/{{.Owner}}/hooks/{{.ID}}"},
		{"delete_org_hook", "webhook", "delete-org", "DELETE", "/orgs/{{.Owner}}/hooks/{{.ID}}"},
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

func TestConvertRepoTool(t *testing.T) {
	td := ByName("convert_repo")
	if td.Group != "repo" {
		t.Errorf("convert_repo: Group = %q, want %q", td.Group, "repo")
	}
	if td.CLIName != "convert" {
		t.Errorf("convert_repo: CLIName = %q, want %q", td.CLIName, "convert")
	}
	if td.Method != "POST" {
		t.Errorf("convert_repo: Method = %q, want %q", td.Method, "POST")
	}
	if td.PathTmpl != "/repos/{{.Owner}}/{{.Repo}}/convert" {
		t.Errorf("convert_repo: PathTmpl = %q, want %q", td.PathTmpl, "/repos/{{.Owner}}/{{.Repo}}/convert")
	}
}

func TestUpdateRepoTool(t *testing.T) {
	td := ByName("update_repo")
	if td.Group != "repo" {
		t.Errorf("update_repo: Group = %q, want %q", td.Group, "repo")
	}
	if td.CLIName != "update" {
		t.Errorf("update_repo: CLIName = %q, want %q", td.CLIName, "update")
	}
	if td.Method != "PATCH" {
		t.Errorf("update_repo: Method = %q, want %q", td.Method, "PATCH")
	}
	if td.PathTmpl != "/repos/{{.Owner}}/{{.Repo}}" {
		t.Errorf("update_repo: PathTmpl = %q, want %q", td.PathTmpl, "/repos/{{.Owner}}/{{.Repo}}")
	}
}

func TestProtectionTools(t *testing.T) {
	cases := []struct {
		name, group, cliName, method, pathTmpl string
	}{
		{"list_branch_protections", "protection", "list", "GET", "/repos/{{.Owner}}/{{.Repo}}/branch_protections"},
		{"get_branch_protection", "protection", "get", "GET", "/repos/{{.Owner}}/{{.Repo}}/branch_protections/{{.Name}}"},
		{"create_branch_protection", "protection", "create", "POST", "/repos/{{.Owner}}/{{.Repo}}/branch_protections"},
		{"update_branch_protection", "protection", "update", "PATCH", "/repos/{{.Owner}}/{{.Repo}}/branch_protections/{{.Name}}"},
		{"delete_branch_protection", "protection", "delete", "DELETE", "/repos/{{.Owner}}/{{.Repo}}/branch_protections/{{.Name}}"},
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

func TestReadOnlyHint(t *testing.T) {
	for _, td := range All {
		if td.Method == "GET" && !td.IsReadOnly() {
			t.Errorf("tool %s: GET but IsReadOnly() = false", td.Name)
		}
		if td.Method != "GET" && !td.ReadOnly && td.IsReadOnly() {
			t.Errorf("tool %s: non-GET without ReadOnly but IsReadOnly() = true", td.Name)
		}
	}
}

func TestDestructiveTools(t *testing.T) {
	want := map[string]bool{
		"delete_file":               true,
		"delete_branch":             true,
		"remove_issue_labels":       true,
		"delete_issue_comment":      true,
		"delete_label":              true,
		"delete_pull_review":        true,
		"delete_review_requests":    true,
		"cancel_build":              true,
		"delete_secret":             true,
		"delete_variable":           true,
		"delete_org_secret":         true,
		"delete_org_variable":       true,
		"delete_release":            true,
		"delete_release_attachment": true,
		"delete_tag":                true,
		"delete_hook":               true,
		"delete_org_hook":           true,
		"delete_branch_protection":  true,
	}
	for _, td := range All {
		if want[td.Name] && !td.Destructive {
			t.Errorf("tool %s: expected Destructive = true", td.Name)
		}
		if !want[td.Name] && td.Destructive {
			t.Errorf("tool %s: unexpected Destructive = true", td.Name)
		}
	}
}

func TestEveryDELETEIsDestructive(t *testing.T) {
	for _, td := range All {
		if td.Method == "DELETE" && !td.Destructive {
			t.Errorf("tool %s: DELETE method but Destructive = false", td.Name)
		}
	}
}

func TestReleaseJSONTags(t *testing.T) {
	wantCreate := map[string]string{
		"Target": "target_commitish",
		"Title":  "name",
		"Note":   "body",
	}
	rt := reflect.TypeOf(CreateReleaseArgs{})
	for field, wantTag := range wantCreate {
		f, _ := rt.FieldByName(field)
		if got := f.Tag.Get("json"); got != wantTag {
			t.Errorf("CreateReleaseArgs.%s json tag = %q, want %q", field, got, wantTag)
		}
	}

	wantUpdate := map[string]string{
		"Target": "target_commitish",
		"Title":  "name",
		"Note":   "body",
	}
	rt = reflect.TypeOf(UpdateReleaseArgs{})
	for field, wantTag := range wantUpdate {
		f, _ := rt.FieldByName(field)
		if got := f.Tag.Get("json"); got != wantTag {
			t.Errorf("UpdateReleaseArgs.%s json tag = %q, want %q", field, got, wantTag)
		}
	}

	for _, field := range []string{"Draft", "Prerelease"} {
		f, _ := rt.FieldByName(field)
		if f.Type.Kind() != reflect.Ptr || f.Type.Elem().Kind() != reflect.Bool {
			t.Errorf("UpdateReleaseArgs.%s type = %v, want *bool", field, f.Type)
		}
	}
}

func TestREADMEToolCount(t *testing.T) {
	data, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	want := fmt.Sprintf("%d tools", len(All))
	if !bytes.Contains(data, []byte(want)) {
		t.Errorf("README.md does not contain %q; update the tool count", want)
	}
}
