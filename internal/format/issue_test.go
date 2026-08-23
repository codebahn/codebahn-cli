package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func TestListRepoIssues(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	raw := json.RawMessage(`[
		{"number":3,"title":"fix: update pins","state":"open","labels":[{"name":"bug","color":"d73a4a"}],"updated_at":"` + now + `"},
		{"number":2,"title":"Add onboarding","state":"closed","labels":[],"updated_at":"` + now + `"}
	]`)

	f, ok := Get("list_repo_issues")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "#3") {
		t.Error("expected #3 in output")
	}
	if !strings.Contains(out, "fix: update pins") {
		t.Error("expected title in output")
	}
	if !strings.Contains(out, "bug") {
		t.Error("expected label in output")
	}
	if !strings.Contains(out, "#2") {
		t.Error("expected #2 in output")
	}
}

func TestGetIssueByIndex(t *testing.T) {
	created := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	raw := json.RawMessage(`{
		"number":3,
		"title":"fix: update pins",
		"state":"open",
		"body":"Updated the CI action pins.",
		"user":{"login":"admin"},
		"labels":[{"name":"bug"}],
		"assignees":[{"login":"admin"}],
		"milestone":{"title":"v1.0"},
		"comments":2,
		"created_at":"` + created + `",
		"html_url":"https://codebahn.net/owner/repo/issues/3"
	}`)

	f, ok := Get("get_issue_by_index")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "fix: update pins") {
		t.Error("expected title")
	}
	if !strings.Contains(out, "#3") {
		t.Error("expected issue number")
	}
	if !strings.Contains(out, "admin") {
		t.Error("expected author")
	}
	if !strings.Contains(out, "2 comments") {
		t.Error("expected comment count")
	}
	if !strings.Contains(out, "Labels:") {
		t.Error("expected labels section")
	}
	if !strings.Contains(out, "bug") {
		t.Error("expected label name")
	}
	if !strings.Contains(out, "Assignees:") {
		t.Error("expected assignees section")
	}
	if !strings.Contains(out, "Milestone:") {
		t.Error("expected milestone section")
	}
	if !strings.Contains(out, "v1.0") {
		t.Error("expected milestone name")
	}
	if !strings.Contains(out, "Updated the CI action pins.") {
		t.Error("expected body")
	}
	if !strings.Contains(out, "https://codebahn.net/owner/repo/issues/3") {
		t.Error("expected URL")
	}
}

func TestGetIssueByIndexNoOptionalFields(t *testing.T) {
	created := time.Now().UTC().Format(time.RFC3339)
	raw := json.RawMessage(`{
		"number":1,
		"title":"Simple issue",
		"state":"open",
		"body":"",
		"user":{"login":"admin"},
		"labels":[],
		"assignees":[],
		"comments":0,
		"created_at":"` + created + `",
		"html_url":"https://codebahn.net/owner/repo/issues/1"
	}`)

	f, ok := Get("get_issue_by_index")
	if !ok {
		t.Fatal("formatter not registered")
	}
	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "Labels:") {
		t.Error("should not show Labels: when empty")
	}
	if strings.Contains(out, "Assignees:") {
		t.Error("should not show Assignees: when empty")
	}
	if strings.Contains(out, "Milestone:") {
		t.Error("should not show Milestone: when empty")
	}
}

func TestCreateIssue(t *testing.T) {
	raw := json.RawMessage(`{"number":4,"html_url":"https://codebahn.net/owner/repo/issues/4"}`)
	f, ok := Get("create_issue")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "https://codebahn.net/owner/repo/issues/4") {
		t.Error("expected URL on stdout")
	}
}

func TestCreateIssueComment(t *testing.T) {
	raw := json.RawMessage(`{"id":101,"html_url":"https://codebahn.net/owner/repo/issues/4#issuecomment-101"}`)
	f, ok := Get("create_issue_comment")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "issuecomment-101") {
		t.Error("expected comment URL on stdout")
	}
}

func TestIssueStateChange(t *testing.T) {
	raw := json.RawMessage(`{"number":3,"title":"fix: update pins","state":"closed","repository":{"full_name":"owner/repo"}}`)
	f, ok := Get("issue_state_change")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteIssueComment(t *testing.T) {
	f, ok := Get("delete_issue_comment")
	if !ok {
		t.Fatal("formatter not registered")
	}

	args := &tools.DeleteIssueCommentArgs{Owner: "owner", Repo: "repo", CommentID: 101}
	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(nil, args, p); err != nil {
		t.Fatal(err)
	}
}

func TestListIssueComments(t *testing.T) {
	created := time.Now().UTC().Format(time.RFC3339)
	raw := json.RawMessage(`[
		{"id":101,"user":{"login":"admin"},"body":"This looks good\nmore text","created_at":"` + created + `"},
		{"id":102,"user":{"login":"user2"},"body":"Can you add a test?","created_at":"` + created + `"}
	]`)

	f, ok := Get("list_issue_comments")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "#101") {
		t.Error("expected comment ID")
	}
	if !strings.Contains(out, "admin") {
		t.Error("expected author")
	}
	if !strings.Contains(out, "This looks good") {
		t.Error("expected first line of body")
	}
}

func TestListRepoLabels(t *testing.T) {
	raw := json.RawMessage(`[
		{"id":1,"name":"bug","color":"d73a4a","description":"Bug reports"},
		{"id":2,"name":"enhancement","color":"0075ca","description":"Feature requests"}
	]`)

	f, ok := Get("list_repo_labels")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "bug") {
		t.Error("expected label name")
	}
	if !strings.Contains(out, "#d73a4a") {
		t.Error("expected color")
	}
	if !strings.Contains(out, "Bug reports") {
		t.Error("expected description")
	}
}

func TestListRepoMilestones(t *testing.T) {
	raw := json.RawMessage(`[
		{"id":1,"title":"v1.0","state":"open","open_issues":3,"closed_issues":1},
		{"id":2,"title":"v0.9","state":"closed","open_issues":0,"closed_issues":5}
	]`)

	f, ok := Get("list_repo_milestones")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "v1.0") {
		t.Error("expected milestone title")
	}
	if !strings.Contains(out, "3 open") {
		t.Error("expected open count")
	}
	if !strings.Contains(out, "1 closed") {
		t.Error("expected closed count")
	}
}

func TestDeleteLabel(t *testing.T) {
	f, ok := Get("delete_label")
	if !ok {
		t.Fatal("formatter not registered")
	}

	args := &tools.DeleteLabelArgs{Owner: "owner", ID: 42}
	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(nil, args, p); err != nil {
		t.Fatal(err)
	}
}

func TestGetIssueComment(t *testing.T) {
	created := time.Now().Add(-30 * time.Minute).UTC().Format(time.RFC3339)
	raw := json.RawMessage(`{
		"id":101,
		"user":{"login":"admin"},
		"body":"Full comment body here.",
		"created_at":"` + created + `"
	}`)

	f, ok := Get("get_issue_comment")
	if !ok {
		t.Fatal("formatter not registered")
	}

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Comment #101") {
		t.Error("expected comment header")
	}
	if !strings.Contains(out, "admin") {
		t.Error("expected author")
	}
	if !strings.Contains(out, "Full comment body here.") {
		t.Error("expected body")
	}
}
