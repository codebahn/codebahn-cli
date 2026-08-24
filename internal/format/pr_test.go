package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func TestListPullRequests(t *testing.T) {
	now := time.Now().UTC()
	created := now.Add(-2 * time.Hour).Format(time.RFC3339)
	raw := json.RawMessage(fmt.Sprintf(`[
		{"number":7,"title":"Add OAuth support","state":"open","head":{"ref":"feature/oauth"},"created_at":%q,"merged":false},
		{"number":5,"title":"Fix CI pipeline","state":"closed","head":{"ref":"fix/ci"},"created_at":%q,"merged":true}
	]`, created, created))

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)

	f, ok := Get("list_repo_pull_requests")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "#7") {
		t.Errorf("expected #7 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Add OAuth support") {
		t.Errorf("expected title in output, got:\n%s", out)
	}
	if !strings.Contains(out, "feature/oauth") {
		t.Errorf("expected branch in output, got:\n%s", out)
	}
	if !strings.Contains(out, "#5") {
		t.Errorf("expected #5 in output, got:\n%s", out)
	}
}

func TestGetPullRequest(t *testing.T) {
	now := time.Now().UTC()
	created := now.Add(-2 * time.Hour).Format(time.RFC3339)
	raw := json.RawMessage(fmt.Sprintf(`{
		"number":7,
		"title":"Add OAuth support",
		"state":"open",
		"body":"Implements OAuth2 with PKCE.",
		"user":{"login":"admin"},
		"head":{"ref":"feature/oauth"},
		"base":{"ref":"main"},
		"labels":[{"name":"feature"}],
		"assignees":[{"login":"admin"}],
		"milestone":{"title":"v1.0"},
		"additions":150,
		"deletions":23,
		"merged":false,
		"created_at":%q,
		"html_url":"https://codebahn.net/owner/repo/pulls/7",
		"repository":{"full_name":"owner/repo"}
	}`, created))

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)

	f, ok := Get("get_pull_request_by_index")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Add OAuth support") {
		t.Errorf("expected title, got:\n%s", out)
	}
	if !strings.Contains(out, "open") {
		t.Errorf("expected state, got:\n%s", out)
	}
	if !strings.Contains(out, "admin") {
		t.Errorf("expected user, got:\n%s", out)
	}
	if !strings.Contains(out, "main") {
		t.Errorf("expected base branch, got:\n%s", out)
	}
	if !strings.Contains(out, "feature/oauth") {
		t.Errorf("expected head branch, got:\n%s", out)
	}
	if !strings.Contains(out, "+150") {
		t.Errorf("expected additions, got:\n%s", out)
	}
	if !strings.Contains(out, "-23") {
		t.Errorf("expected deletions, got:\n%s", out)
	}
	if !strings.Contains(out, "feature") {
		t.Errorf("expected label, got:\n%s", out)
	}
	if !strings.Contains(out, "v1.0") {
		t.Errorf("expected milestone, got:\n%s", out)
	}
	if !strings.Contains(out, "Implements OAuth2 with PKCE.") {
		t.Errorf("expected body, got:\n%s", out)
	}
	if !strings.Contains(out, "https://codebahn.net/owner/repo/pulls/7") {
		t.Errorf("expected URL, got:\n%s", out)
	}
}

func TestCreatePullRequest(t *testing.T) {
	raw := json.RawMessage(`{"number":7,"html_url":"https://codebahn.net/owner/repo/pulls/7"}`)

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)

	f, ok := Get("create_pull_request")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "https://codebahn.net/owner/repo/pulls/7") {
		t.Errorf("expected URL on stdout, got:\n%s", out)
	}
}

func TestMergePullRequest(t *testing.T) {
	raw := json.RawMessage(`{"merged":true}`)
	args := &tools.MergePullRequestArgs{
		Owner: "owner",
		Repo:  "repo",
		Index: 7,
		Style: "squash",
	}

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	output.SetNoColor(true)
	defer output.SetNoColor(false)

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)

	f, ok := Get("merge_pull_request")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, args, p); err != nil {
		t.Fatal(err)
	}

	_ = w.Close()
	var stderrBuf bytes.Buffer
	_, _ = stderrBuf.ReadFrom(r)
	os.Stderr = oldStderr

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "Squashed and merged") {
		t.Errorf("expected squash verb in stderr, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "owner/repo#7") {
		t.Errorf("expected PR ref in stderr, got:\n%s", stderr)
	}
}

func TestListPullRequestFiles(t *testing.T) {
	raw := json.RawMessage(`[
		{"filename":"src/main.go","status":"modified","additions":10,"deletions":3},
		{"filename":"src/new.go","status":"added","additions":45,"deletions":0}
	]`)

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)

	f, ok := Get("list_pull_request_files")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "src/main.go") {
		t.Errorf("expected filename, got:\n%s", out)
	}
	if !strings.Contains(out, "modified") {
		t.Errorf("expected status, got:\n%s", out)
	}
	if !strings.Contains(out, "+10") {
		t.Errorf("expected additions, got:\n%s", out)
	}
	if !strings.Contains(out, "-3") {
		t.Errorf("expected deletions, got:\n%s", out)
	}
}

func TestListPullReviews(t *testing.T) {
	now := time.Now().UTC()
	submitted := now.Add(-2 * time.Hour).Format(time.RFC3339)
	raw := json.RawMessage(fmt.Sprintf(`[
		{"id":1,"user":{"login":"user2"},"state":"APPROVED","body":"Great work!","submitted_at":%q}
	]`, submitted))

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)
	p := output.NewPrinter(&buf, false)

	f, ok := Get("list_pull_reviews")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "user2") {
		t.Errorf("expected reviewer, got:\n%s", out)
	}
	if !strings.Contains(out, "approved") {
		t.Errorf("expected state (lowercase), got:\n%s", out)
	}
	if !strings.Contains(out, "Great work!") {
		t.Errorf("expected body, got:\n%s", out)
	}
}

func TestDeletePullReviewNilResponse(t *testing.T) {
	args := &tools.DeletePullReviewArgs{
		Owner: "owner",
		Repo:  "repo",
		Index: 7,
		ID:    42,
	}

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	output.SetNoColor(true)
	defer output.SetNoColor(false)

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)

	f, ok := Get("delete_pull_review")
	if !ok {
		t.Fatal("formatter not registered")
	}
	if err := f(nil, args, p); err != nil {
		t.Fatal(err)
	}

	_ = w.Close()
	var stderrBuf bytes.Buffer
	_, _ = stderrBuf.ReadFrom(r)
	os.Stderr = oldStderr

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "Deleted review #42") {
		t.Errorf("expected delete confirmation, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "owner/repo#7") {
		t.Errorf("expected PR ref, got:\n%s", stderr)
	}
}
