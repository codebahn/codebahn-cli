package tools

type CreateHookArgs struct {
	Owner               string `json:"owner"                 required:"true" desc:"Repository owner"`
	Repo                string `json:"repo"                  required:"true" desc:"Repository name"`
	Type                string `json:"type"                  required:"true" desc:"Webhook type (e.g. gitea, slack, discord)" default:"gitea"`
	URL                 string `json:"url"                   required:"true" desc:"Target URL the webhook payload is delivered to" body:"nest" nest:"config"`
	ContentType         string `json:"content_type"          desc:"Payload content type (json or form)" default:"json" body:"nest" nest:"config"`
	Secret              string `json:"secret"                desc:"Secret used to sign the payload" body:"nest" nest:"config"`
	Events              string `json:"events"                required:"true" desc:"Events that trigger the webhook (comma-separated, e.g. push,pull_request)" body:"csv"`
	Active              *bool  `json:"active"                desc:"Whether the webhook is active"`
	BranchFilter        string `json:"branch_filter"         desc:"Glob pattern restricting which branches trigger the webhook"`
	AuthorizationHeader string `json:"authorization_header"  desc:"Authorization header value sent with each delivery"`
}

type UpdateHookArgs struct {
	Owner               string `json:"owner"                required:"true" desc:"Repository owner"`
	Repo                string `json:"repo"                 required:"true" desc:"Repository name"`
	ID                  int    `json:"id"                   required:"true" desc:"Webhook ID"`
	Type                string `json:"type"                 desc:"Webhook type (e.g. gitea, slack, discord)"`
	URL                 string `json:"url"                  desc:"Target URL the webhook payload is delivered to" body:"nest" nest:"config"`
	ContentType         string `json:"content_type"         desc:"Payload content type (json or form)" body:"nest" nest:"config"`
	Secret              string `json:"secret"               desc:"Secret used to sign the payload" body:"nest" nest:"config"`
	Events              string `json:"events"               desc:"Events that trigger the webhook (comma-separated, e.g. push,pull_request)" body:"csv"`
	Active              *bool  `json:"active"               desc:"Whether the webhook is active"`
	BranchFilter        string `json:"branch_filter"        desc:"Glob pattern restricting which branches trigger the webhook"`
	AuthorizationHeader string `json:"authorization_header" desc:"Authorization header value sent with each delivery"`
}

type ListHooksArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"30"`
}

type GetHookArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	ID    int    `json:"id"    required:"true" desc:"Webhook ID"`
}

type DeleteHookArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	ID    int    `json:"id"    required:"true" desc:"Webhook ID"`
}

type TestHookArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	ID    int    `json:"id"    required:"true" desc:"Webhook ID"`
}

type ListOrgHooksArgs struct {
	Owner string `json:"owner" required:"true" desc:"Organization name"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"30"`
}

type GetOrgHookArgs struct {
	Owner string `json:"owner" required:"true" desc:"Organization name"`
	ID    int    `json:"id"    required:"true" desc:"Webhook ID"`
}

type CreateOrgHookArgs struct {
	Owner               string `json:"owner"                required:"true" desc:"Organization name"`
	Type                string `json:"type"                 required:"true" desc:"Webhook type (e.g. gitea, slack, discord)" default:"gitea"`
	URL                 string `json:"url"                  required:"true" desc:"Target URL the webhook payload is delivered to" body:"nest" nest:"config"`
	ContentType         string `json:"content_type"         desc:"Payload content type (json or form)" default:"json" body:"nest" nest:"config"`
	Secret              string `json:"secret"               desc:"Secret used to sign the payload" body:"nest" nest:"config"`
	Events              string `json:"events"               required:"true" desc:"Events that trigger the webhook (comma-separated, e.g. push,pull_request)" body:"csv"`
	Active              *bool  `json:"active"               desc:"Whether the webhook is active"`
	BranchFilter        string `json:"branch_filter"        desc:"Glob pattern restricting which branches trigger the webhook"`
	AuthorizationHeader string `json:"authorization_header" desc:"Authorization header value sent with each delivery"`
}

type UpdateOrgHookArgs struct {
	Owner               string `json:"owner"                required:"true" desc:"Organization name"`
	ID                  int    `json:"id"                   required:"true" desc:"Webhook ID"`
	Type                string `json:"type"                 desc:"Webhook type (e.g. gitea, slack, discord)"`
	URL                 string `json:"url"                  desc:"Target URL the webhook payload is delivered to" body:"nest" nest:"config"`
	ContentType         string `json:"content_type"         desc:"Payload content type (json or form)" body:"nest" nest:"config"`
	Secret              string `json:"secret"               desc:"Secret used to sign the payload" body:"nest" nest:"config"`
	Events              string `json:"events"               desc:"Events that trigger the webhook (comma-separated, e.g. push,pull_request)" body:"csv"`
	Active              *bool  `json:"active"               desc:"Whether the webhook is active"`
	BranchFilter        string `json:"branch_filter"        desc:"Glob pattern restricting which branches trigger the webhook"`
	AuthorizationHeader string `json:"authorization_header" desc:"Authorization header value sent with each delivery"`
}

type DeleteOrgHookArgs struct {
	Owner string `json:"owner" required:"true" desc:"Organization name"`
	ID    int    `json:"id"    required:"true" desc:"Webhook ID"`
}

func webhookTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "list_hooks",
			Group:       "webhook",
			CLIName:     "list",
			Description: "List webhooks for a repository",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/hooks",
			Args:        ListHooksArgs{},
		},
		{
			Name:        "get_hook",
			Group:       "webhook",
			CLIName:     "get",
			Description: "Get a repository webhook by ID",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}",
			Args:        GetHookArgs{},
		},
		{
			Name:        "create_hook",
			Group:       "webhook",
			CLIName:     "create",
			Description: "Create a repository webhook",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/hooks",
			Args:        CreateHookArgs{},
		},
		{
			Name:        "update_hook",
			Group:       "webhook",
			CLIName:     "update",
			Description: "Update a repository webhook",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}",
			Args:        UpdateHookArgs{},
		},
		{
			Name:        "delete_hook",
			Group:       "webhook",
			CLIName:     "delete",
			Description: "Delete a repository webhook",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}",
			Args:        DeleteHookArgs{},
			Destructive: true,
		},
		{
			Name:        "test_hook",
			Group:       "webhook",
			CLIName:     "test",
			Description: "Send a test delivery to a repository webhook",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/hooks/{{.ID}}/tests",
			Args:        TestHookArgs{},
		},
		{
			Name:        "list_org_hooks",
			Group:       "webhook",
			CLIName:     "list-org",
			Description: "List webhooks for an organization",
			Method:      "GET",
			PathTmpl:    "/orgs/{{.Owner}}/hooks",
			Args:        ListOrgHooksArgs{},
		},
		{
			Name:        "get_org_hook",
			Group:       "webhook",
			CLIName:     "get-org",
			Description: "Get an organization webhook by ID",
			Method:      "GET",
			PathTmpl:    "/orgs/{{.Owner}}/hooks/{{.ID}}",
			Args:        GetOrgHookArgs{},
		},
		{
			Name:        "create_org_hook",
			Group:       "webhook",
			CLIName:     "create-org",
			Description: "Create an organization webhook",
			Method:      "POST",
			PathTmpl:    "/orgs/{{.Owner}}/hooks",
			Args:        CreateOrgHookArgs{},
		},
		{
			Name:        "update_org_hook",
			Group:       "webhook",
			CLIName:     "update-org",
			Description: "Update an organization webhook",
			Method:      "PATCH",
			PathTmpl:    "/orgs/{{.Owner}}/hooks/{{.ID}}",
			Args:        UpdateOrgHookArgs{},
		},
		{
			Name:        "delete_org_hook",
			Group:       "webhook",
			CLIName:     "delete-org",
			Description: "Delete an organization webhook",
			Method:      "DELETE",
			PathTmpl:    "/orgs/{{.Owner}}/hooks/{{.ID}}",
			Args:        DeleteOrgHookArgs{},
			Destructive: true,
		},
	}
}
