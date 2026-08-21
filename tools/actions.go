package tools

type DispatchWorkflowArgs struct {
	Owner    string `json:"owner"    required:"true" desc:"Repository owner"`
	Repo     string `json:"repo"     required:"true" desc:"Repository name"`
	Workflow string `json:"workflow" required:"true" desc:"Workflow file (e.g. ci.yml)"`
	Ref      string `json:"ref"      desc:"Git ref to build (branch/tag). Defaults to repo default branch."`
	Inputs   string `json:"inputs"   desc:"Workflow inputs as JSON object, e.g. {\"key\": \"value\"}"`
}

type ListWorkflowRunsArgs struct {
	Owner     string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"       required:"true" desc:"Repository name"`
	Status    string `json:"status"     desc:"Filter by status (waiting, running, success, failure, cancelled)"`
	Event     string `json:"event"      desc:"Filter by event (push, pull_request, workflow_dispatch)"`
	RunNumber int    `json:"run_number" desc:"Filter by run number"`
	HeadSHA   string `json:"head_sha"   desc:"Filter by HEAD SHA"`
	Page      int    `json:"page"       desc:"Page number (1-based)" default:"1"`
	Limit     int    `json:"limit"      desc:"Page size"             default:"30"`
}

type GetWorkflowRunArgs struct {
	Owner string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"   required:"true" desc:"Repository name"`
	RunID int    `json:"run_id" required:"true" desc:"Run ID"`
}

type GetJobLogsArgs struct {
	Owner     string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"       required:"true" desc:"Repository name"`
	RunID     int    `json:"run_id"     required:"true" desc:"Workflow run ID"`
	Job       string `json:"job"        desc:"Filter logs to a specific job name"`
	TailLines int    `json:"tail_lines" desc:"Max log lines per job (default 200, max 500)"`
}

type CancelBuildArgs struct {
	Owner string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"   required:"true" desc:"Repository name"`
	RunID int    `json:"run_id" required:"true" desc:"Workflow run ID"`
}

func actionsTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "dispatch_workflow",
			Group:       "ci",
			CLIName:     "dispatch",
			Description: "Trigger a workflow run. Returns the run ID.",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/workflows/{{.Workflow}}/dispatches",
			Args:        DispatchWorkflowArgs{},
		},
		{
			Name:        "list_workflow_runs",
			Group:       "ci",
			CLIName:     "list",
			Description: "List workflow runs for a repository",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/runs",
			Args:        ListWorkflowRunsArgs{},
		},
		{
			Name:        "get_workflow_run",
			Group:       "ci",
			CLIName:     "get",
			Description: "Get details of a specific workflow run",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/runs/{{.RunID}}",
			Args:        GetWorkflowRunArgs{},
		},
		{
			Name:        "get_job_logs",
			Group:       "ci",
			CLIName:     "logs",
			Description: "Get logs for a CI build run, grouped by job",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/runs/{{.RunID}}/logs",
			Args:        GetJobLogsArgs{},
		},
		{
			Name:        "cancel_build",
			Group:       "ci",
			CLIName:     "cancel",
			Description: "Cancel a running CI build",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/runs/{{.RunID}}/cancel",
			Args:        CancelBuildArgs{},
		},
	}
}
