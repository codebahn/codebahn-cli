package format

import (
	"encoding/json"
	"fmt"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func init() {
	Register("list_workflow_runs", formatListWorkflowRuns)
	Register("get_workflow_run", formatGetWorkflowRun)
	Register("dispatch_workflow", formatDispatchWorkflow)
	Register("cancel_build", formatCancelBuild)
}

type workflowRun struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	HeadBranch string `json:"head_branch"`
	Event      string `json:"event"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	StartedAt  string `json:"run_started_at"`
	UpdatedAt  string `json:"updated_at"`
	HTMLURL    string `json:"html_url"`
}

func ciIcon(status, conclusion string) string {
	if !output.IsTTY() {
		return status + "/" + conclusion
	}
	if status != "completed" {
		return output.Yellow("*")
	}
	switch conclusion {
	case "success":
		return output.Green("✓")
	case "failure", "timed_out":
		return output.Red("✗")
	default:
		return output.Dim("-")
	}
}

func formatListWorkflowRuns(raw json.RawMessage, _ any, p *output.Printer) error {
	var runs []workflowRun

	var wrapped struct {
		Runs []workflowRun `json:"workflow_runs"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil {
		runs = wrapped.Runs
	} else if err := json.Unmarshal(raw, &runs); err != nil {
		return err
	}

	if len(runs) == 0 {
		return nil
	}

	var rows [][]string
	for _, r := range runs {
		ts := r.StartedAt
		if ts == "" {
			ts = r.UpdatedAt
		}
		rows = append(rows, []string{
			ciIcon(r.Status, r.Conclusion),
			output.Bold(r.Name),
			r.Event,
			output.Bold(r.HeadBranch),
			output.Cyan(fmt.Sprintf("%d", r.ID)),
			output.Dim(TimeAgo(ts)),
		})
	}
	p.Table(nil, rows)
	return nil
}

func formatGetWorkflowRun(raw json.RawMessage, _ any, p *output.Printer) error {
	var r workflowRun
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}

	ts := r.StartedAt
	if ts == "" {
		ts = r.UpdatedAt
	}

	p.Text(fmt.Sprintf("%s %s · %d", ciIcon(r.Status, r.Conclusion), r.Name, r.ID))
	p.Text(fmt.Sprintf("Triggered via %s · %s · %s", r.Event, r.HeadBranch, TimeAgo(ts)))
	return nil
}

func formatDispatchWorkflow(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DispatchWorkflowArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for dispatch_workflow")
	}
	successf("Dispatched workflow in %s/%s", a.Owner, a.Repo)
	return nil
}

func formatCancelBuild(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.CancelBuildArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for cancel_build")
	}
	successf("Cancelled run %d in %s/%s", a.RunID, a.Owner, a.Repo)
	return nil
}
