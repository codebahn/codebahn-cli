package format

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func TestListMyRepos(t *testing.T) {
	now := time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	raw := mustJSON(t, []map[string]any{
		{"full_name": "acme/web", "description": "Web app", "private": false, "fork": false, "archived": false, "updated_at": now},
		{"full_name": "acme/api", "description": "API server with a very long description that should be truncated at fifty chars", "private": true, "fork": true, "archived": false, "updated_at": now},
	})

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("list_my_repos")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	assertContains(t, out, "acme/web")
	assertContains(t, out, "Web app")
	assertContains(t, out, "public")
	assertContains(t, out, "acme/api")
	assertContains(t, out, "private, fork")
	assertContains(t, out, "2h ago")
}

func TestListBranches(t *testing.T) {
	raw := mustJSON(t, []map[string]any{
		{"name": "main", "commit": map[string]string{"id": "abc123def456789"}},
		{"name": "feature/x", "commit": map[string]string{"id": "deadbeefcafe123"}},
	})

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("list_branches")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	assertContains(t, out, "main")
	assertContains(t, out, "abc123d")
	assertContains(t, out, "feature/x")
	assertContains(t, out, "deadbee")
}

func TestListRepoCommits(t *testing.T) {
	now := time.Now().UTC().Add(-3 * time.Hour).Format(time.RFC3339)
	raw := mustJSON(t, []map[string]any{
		{
			"sha": "abc123def456789",
			"commit": map[string]any{
				"message": "fix: update pins\n\nSome details here",
				"author":  map[string]string{"name": "Admin", "date": now},
			},
		},
	})

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("list_repo_commits")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	assertContains(t, out, "abc123d")
	assertContains(t, out, "fix: update pins")
	assertContains(t, out, "Admin")
	assertContains(t, out, "3h ago")
	assertNotContains(t, out, "Some details here")
}

func TestListRepoContents(t *testing.T) {
	raw := mustJSON(t, []map[string]any{
		{"type": "dir", "name": "src", "size": 0},
		{"type": "file", "name": "README.md", "size": 1234},
		{"type": "file", "name": "go.mod", "size": 450},
	})

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("list_repo_contents")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	assertContains(t, out, "dir")
	assertContains(t, out, "src/")
	assertContains(t, out, "file")
	assertContains(t, out, "README.md")
	assertContains(t, out, "1.2 KiB")
}

func TestGetRepoTree(t *testing.T) {
	raw := mustJSON(t, map[string]any{
		"tree": []map[string]any{
			{"path": "src/main.go", "type": "blob", "size": 2048},
			{"path": "docs", "type": "tree", "size": 0},
		},
	})

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("get_repo_tree")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	assertContains(t, out, "file")
	assertContains(t, out, "src/main.go")
	assertContains(t, out, "2.0 KiB")
	assertContains(t, out, "dir")
	assertContains(t, out, "docs")
}

func TestCreateRepo(t *testing.T) {
	raw := mustJSON(t, map[string]any{
		"full_name": "acme/new-repo",
		"html_url":  "https://codebahn.net/acme/new-repo",
	})

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("create_repo")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	assertContains(t, buf.String(), "https://codebahn.net/acme/new-repo")
}

func TestCreateFile(t *testing.T) {
	raw := mustJSON(t, map[string]any{
		"content": map[string]string{"path": "src/main.go"},
	})
	args := &tools.CreateFileArgs{Owner: "acme", Repo: "web", FilePath: "src/main.go", BranchName: "main"}

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("create_file")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, args, p); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBranch(t *testing.T) {
	args := &tools.DeleteBranchArgs{Owner: "acme", Repo: "web", Branch: "old-branch"}

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)
	f, ok := Get("delete_branch")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(nil, args, p); err != nil {
		t.Fatal(err)
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{1234, "1.2 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
	}
	for _, tt := range tests {
		got := humanSize(tt.bytes)
		if got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestGetFileContentNotRegistered(t *testing.T) {
	_, ok := Get("get_file_content")
	if ok {
		t.Error("get_file_content should not have a formatter (passthrough)")
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("output %q does not contain %q", s, substr)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("output %q should not contain %q", s, substr)
	}
}
