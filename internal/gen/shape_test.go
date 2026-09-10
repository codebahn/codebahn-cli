package gen

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/codebahn/codebahn-cli/tools"
)

const twoFileDiff = "diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1 @@\n-old\n+new\ndiff --git a/b.go b/b.go\n--- a/b.go\n+++ b/b.go\n@@ -1 +1 @@\n-x\n+y\n"

func TestShapeDiffFilePath(t *testing.T) {
	cases := []struct {
		td   tools.ToolDef
		args any
	}{
		{tools.ByName("get_pull_request_diff"), &tools.GetPullRequestDiffArgs{FilePath: "b.go"}},
		{tools.ByName("get_commit_diff"), &tools.GetCommitDiffArgs{FilePath: "b.go"}},
	}
	for _, tc := range cases {
		t.Run(tc.td.Name, func(t *testing.T) {
			got, err := shapeResponse(tc.td, json.RawMessage(twoFileDiff), tc.args)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(got), "diff --git a/b.go b/b.go") {
				t.Errorf("shaped diff should start at b.go, got %q", got)
			}
			if strings.Contains(string(got), "a.go") {
				t.Errorf("shaped diff should not contain a.go, got %q", got)
			}
		})
	}
}

func TestShapeDiffFilePathNotFound(t *testing.T) {
	td := tools.ByName("get_commit_diff")
	_, err := shapeResponse(td, json.RawMessage(twoFileDiff), &tools.GetCommitDiffArgs{FilePath: "nope.go"})
	if err == nil {
		t.Fatal("expected an error for a file_path that is not in the diff")
	}
	if !strings.Contains(err.Error(), "nope.go") {
		t.Errorf("error should name the missing file, got %q", err)
	}
}

func TestShapeDiffWithoutFilePathIsPassthrough(t *testing.T) {
	td := tools.ByName("get_pull_request_diff")
	got, err := shapeResponse(td, json.RawMessage(twoFileDiff), &tools.GetPullRequestDiffArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != twoFileDiff {
		t.Errorf("without file_path the diff must pass through unchanged, got %q", got)
	}
}

func TestShapeUnknownToolIsPassthrough(t *testing.T) {
	td := tools.ByName("list_repo_issues")
	raw := json.RawMessage(`[{"number":1}]`)
	got, err := shapeResponse(td, raw, &tools.ListRepoIssuesArgs{})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Errorf("unshaped tools must pass through unchanged, got %q", got)
	}
}

// Every field the REST API does not understand (api:"-") needs a client-side
// shaper, otherwise the CLI flag exists but silently does nothing.
func TestLocalFieldsHaveShaper(t *testing.T) {
	for _, td := range tools.All {
		rt := reflect.TypeOf(td.Args)
		for i := range rt.NumField() {
			if rt.Field(i).Tag.Get("api") != "-" {
				continue
			}
			if !hasShaper(td.Name) {
				t.Errorf("tool %s: field %s is api:\"-\" but shapeResponse does not handle the tool", td.Name, rt.Field(i).Name)
			}
		}
	}
}
