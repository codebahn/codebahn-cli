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

type ListSecretsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"30"`
}

type SetSecretArgs struct {
	Owner      string `json:"owner"       required:"true" desc:"Repository owner"`
	Repo       string `json:"repo"        required:"true" desc:"Repository name"`
	SecretName string `json:"secret_name" required:"true" desc:"Secret name"`
	Data       string `json:"data"        required:"true" desc:"Secret value. Write-only: there is no way to read it back once set."`
}

type DeleteSecretArgs struct {
	Owner      string `json:"owner"       required:"true" desc:"Repository owner"`
	Repo       string `json:"repo"        required:"true" desc:"Repository name"`
	SecretName string `json:"secret_name" required:"true" desc:"Secret name"`
}

type ListVariablesArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"30"`
}

type GetVariableArgs struct {
	Owner        string `json:"owner"         required:"true" desc:"Repository owner"`
	Repo         string `json:"repo"          required:"true" desc:"Repository name"`
	VariableName string `json:"variable_name" required:"true" desc:"Variable name"`
}

type CreateVariableArgs struct {
	Owner        string `json:"owner"         required:"true" desc:"Repository owner"`
	Repo         string `json:"repo"          required:"true" desc:"Repository name"`
	VariableName string `json:"variable_name" required:"true" desc:"Variable name"`
	Value        string `json:"value"         required:"true" desc:"Variable value"`
}

type UpdateVariableArgs struct {
	Owner        string `json:"owner"         required:"true" desc:"Repository owner"`
	Repo         string `json:"repo"          required:"true" desc:"Repository name"`
	VariableName string `json:"variable_name" required:"true" desc:"Variable name"`
	Value        string `json:"value"         required:"true" desc:"Variable value"`
}

type DeleteVariableArgs struct {
	Owner        string `json:"owner"         required:"true" desc:"Repository owner"`
	Repo         string `json:"repo"          required:"true" desc:"Repository name"`
	VariableName string `json:"variable_name" required:"true" desc:"Variable name"`
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
			Destructive: true,
		},
		{
			Name:        "list_secrets",
			Group:       "ci",
			CLIName:     "list-secrets",
			Description: "List repository Actions secret names (values are write-only and never returned)",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/secrets",
			Args:        ListSecretsArgs{},
		},
		{
			Name:        "set_secret",
			Group:       "ci",
			CLIName:     "set-secret",
			Description: "Create or update a repository Actions secret. Secrets are write-only: there is no get_secret tool because the value cannot be read back.",
			Method:      "PUT",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/secrets/{{.SecretName}}",
			Args:        SetSecretArgs{},
		},
		{
			Name:        "delete_secret",
			Group:       "ci",
			CLIName:     "delete-secret",
			Description: "Delete a repository Actions secret",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/secrets/{{.SecretName}}",
			Args:        DeleteSecretArgs{},
			Destructive: true,
		},
		{
			Name:        "list_variables",
			Group:       "ci",
			CLIName:     "list-variables",
			Description: "List repository Actions variables",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/variables",
			Args:        ListVariablesArgs{},
		},
		{
			Name:        "get_variable",
			Group:       "ci",
			CLIName:     "get-variable",
			Description: "Get a repository Actions variable",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}",
			Args:        GetVariableArgs{},
		},
		{
			Name:        "create_variable",
			Group:       "ci",
			CLIName:     "create-variable",
			Description: "Create a repository Actions variable",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}",
			Args:        CreateVariableArgs{},
		},
		{
			Name:        "update_variable",
			Group:       "ci",
			CLIName:     "update-variable",
			Description: "Update a repository Actions variable",
			Method:      "PUT",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}",
			Args:        UpdateVariableArgs{},
		},
		{
			Name:        "delete_variable",
			Group:       "ci",
			CLIName:     "delete-variable",
			Description: "Delete a repository Actions variable",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/actions/variables/{{.VariableName}}",
			Args:        DeleteVariableArgs{},
			Destructive: true,
		},
	}
}
