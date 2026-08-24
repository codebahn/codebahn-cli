package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func TestFormatListWorkflowRuns(t *testing.T) {
	ts := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	raw := json.RawMessage(fmt.Sprintf(`{"workflow_runs":[{"id":12345,"name":"CI","head_branch":"main","event":"push","status":"completed","conclusion":"success","run_started_at":"%s","updated_at":"%s"}]}`, ts, ts))

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, ok := Get("list_workflow_runs")
	if !ok {
		t.Fatal("formatter not registered for list_workflow_runs")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !contains(got, "CI") {
		t.Errorf("expected workflow name, got: %s", got)
	}
	if !contains(got, "main") {
		t.Errorf("expected branch, got: %s", got)
	}
	if !contains(got, "12345") {
		t.Errorf("expected run ID, got: %s", got)
	}
	if !contains(got, "push") {
		t.Errorf("expected event, got: %s", got)
	}
}

func TestFormatListWorkflowRunsTopLevel(t *testing.T) {
	ts := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	raw := json.RawMessage(fmt.Sprintf(`[{"id":99,"name":"Deploy","head_branch":"release","event":"manual","status":"completed","conclusion":"failure","run_started_at":"%s","updated_at":"%s"}]`, ts, ts))

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, ok := Get("list_workflow_runs")
	if !ok {
		t.Fatal("formatter not registered for list_workflow_runs")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !contains(got, "Deploy") {
		t.Errorf("expected workflow name, got: %s", got)
	}
}

func TestFormatGetWorkflowRun(t *testing.T) {
	ts := time.Now().Add(-30 * time.Minute).Format(time.RFC3339)
	raw := json.RawMessage(fmt.Sprintf(`{"id":12345,"name":"CI","status":"completed","conclusion":"success","head_branch":"main","event":"push","run_started_at":"%s","updated_at":"%s","html_url":"https://codebahn.net/owner/repo/actions/runs/12345"}`, ts, ts))

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, ok := Get("get_workflow_run")
	if !ok {
		t.Fatal("formatter not registered for get_workflow_run")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !contains(got, "CI") {
		t.Errorf("expected name, got: %s", got)
	}
	if !contains(got, "12345") {
		t.Errorf("expected run ID, got: %s", got)
	}
	if !contains(got, "push") {
		t.Errorf("expected event, got: %s", got)
	}
}

func TestFormatListWorkflowRunsEmptyObject(t *testing.T) {
	raw := json.RawMessage(`{"total_count":0}`)

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, ok := Get("list_workflow_runs")
	if !ok {
		t.Fatal("formatter not registered for list_workflow_runs")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatalf("expected no error for empty object response, got: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for zero results, got: %s", buf.String())
	}
}

func TestFormatDispatchWorkflowBadArgs(t *testing.T) {
	f, ok := Get("dispatch_workflow")
	if !ok {
		t.Fatal("formatter not registered")
	}
	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	err := f(nil, "wrong-type", p)
	if err == nil {
		t.Fatal("expected error for wrong args type")
	}
}

func TestFormatCancelBuildBadArgs(t *testing.T) {
	f, ok := Get("cancel_build")
	if !ok {
		t.Fatal("formatter not registered")
	}
	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	err := f(nil, "wrong-type", p)
	if err == nil {
		t.Fatal("expected error for wrong args type")
	}
}

func TestFormatDispatchWorkflow(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	args := &tools.DispatchWorkflowArgs{Owner: "acme", Repo: "app", Workflow: "ci.yml"}
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	f, ok := Get("dispatch_workflow")
	if !ok {
		t.Fatal("formatter not registered for dispatch_workflow")
	}
	_ = f(nil, args, p)
	_ = w.Close()

	var stderr bytes.Buffer
	_, _ = stderr.ReadFrom(r)
	os.Stderr = oldStderr

	got := stderr.String()
	if !bytes.Contains([]byte(got), []byte("Dispatched")) {
		t.Errorf("expected dispatch confirmation on stderr, got: %s", got)
	}
	if !bytes.Contains([]byte(got), []byte("acme/app")) {
		t.Errorf("expected owner/repo on stderr, got: %s", got)
	}
}

func TestFormatCancelBuild(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	args := &tools.CancelBuildArgs{Owner: "acme", Repo: "app", RunID: 999}
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	var buf bytes.Buffer
	p := output.NewPrinter(&buf, false)
	f, ok := Get("cancel_build")
	if !ok {
		t.Fatal("formatter not registered for cancel_build")
	}
	_ = f(nil, args, p)
	_ = w.Close()

	var stderr bytes.Buffer
	_, _ = stderr.ReadFrom(r)
	os.Stderr = oldStderr

	got := stderr.String()
	if !bytes.Contains([]byte(got), []byte("Cancelled")) {
		t.Errorf("expected cancel confirmation on stderr, got: %s", got)
	}
	if !bytes.Contains([]byte(got), []byte("999")) {
		t.Errorf("expected run ID on stderr, got: %s", got)
	}
}
