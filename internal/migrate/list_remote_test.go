package migrate

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/codebahn/codebahn-cli/internal/output"
)

func init() {
	output.SetNoColor(true)
}

func TestListRemoteTable(t *testing.T) {
	repos := []SourceRepo{
		{Name: "api", FullName: "acme/api", Description: "The API", Private: false},
		{Name: "web", FullName: "acme/web", Description: "Frontend app", Private: true},
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	printRemoteRepos(repos, p)

	out := buf.String()
	if !containsAll(out, "REPOSITORY", "VISIBILITY", "DESCRIPTION") {
		t.Errorf("missing headers in output:\n%s", out)
	}
	if !containsAll(out, "acme/api", "public", "The API") {
		t.Errorf("missing acme/api row in output:\n%s", out)
	}
	if !containsAll(out, "acme/web", "private", "Frontend app") {
		t.Errorf("missing acme/web row in output:\n%s", out)
	}
}

func TestListRemoteTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	printRemoteRepos(nil, p)

	out := buf.String()
	if out != "No repositories found.\n" {
		t.Errorf("expected empty message, got %q", out)
	}
}

func TestListRemoteTable_TruncatesDescription(t *testing.T) {
	repos := []SourceRepo{
		{
			Name:        "long-desc",
			FullName:    "acme/long-desc",
			Description: "This is a very long description that exceeds sixty characters and should be truncated",
			Private:     false,
		},
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	printRemoteRepos(repos, p)

	out := buf.String()
	if len(repos[0].Description) <= 60 {
		t.Fatal("test setup: description should be longer than 60 chars")
	}
	if !bytes.Contains([]byte(out), []byte("...")) {
		t.Errorf("expected truncated description with ..., got:\n%s", out)
	}
}

func TestListRemoteJSON(t *testing.T) {
	repos := []SourceRepo{
		{Name: "api", FullName: "acme/api", Description: "The API", Private: false},
		{Name: "web", FullName: "acme/web", Description: "", Private: true},
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, true)
	printRemoteRepos(repos, p)

	var result []SourceRepo
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(result))
	}
	if result[0].FullName != "acme/api" {
		t.Errorf("result[0].FullName = %q, want acme/api", result[0].FullName)
	}
	if result[1].Private != true {
		t.Errorf("result[1].Private = %v, want true", result[1].Private)
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !bytes.Contains([]byte(s), []byte(sub)) {
			return false
		}
	}
	return true
}
