package tools

const listIssueCommentsDesc = `List issue/PR comments. Returns all comments; the API does not support pagination for this endpoint.`

const listRepoLabelsDesc = `List repository labels. When the owner is an organization and include_org_labels is true (default), org-level labels are merged into the response. Each label carries a scope field of "repo" or "org".`

type GetIssueByIndexArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"Issue index"`
}

type ListRepoIssuesArgs struct {
	Owner      string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo       string `json:"repo"       required:"true" desc:"Repository name"`
	State      string `json:"state"      desc:"State (open|closed|all)" default:"open"`
	Type       string `json:"type"       desc:"Type (issues|pulls)"`
	Milestones string `json:"milestones" desc:"Milestone names/IDs (comma-separated)"`
	Labels     string `json:"labels"     desc:"Labels (comma-separated)"`
	Page       int    `json:"page"       desc:"Page number (1-based)" default:"1"`
	Limit      int    `json:"limit"      desc:"Page size"             default:"20"`
}

type CreateIssueArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Title string `json:"title" required:"true" desc:"Title"`
	Body  string `json:"body"  desc:"Content body"`
}

type CreateIssueCommentArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"Issue/PR index"`
	Body  string `json:"body"  required:"true" desc:"Content body"`
}

type UpdateIssueArgs struct {
	Owner     string `json:"owner"     required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"      required:"true" desc:"Repository name"`
	Index     int    `json:"index"     required:"true" desc:"Issue index"`
	Title     string `json:"title"     desc:"Title"`
	Body      string `json:"body"      desc:"Content body"`
	Assignee  string `json:"assignee"  desc:"Assignee username (convenience for a single user; equivalent to a one-element 'assignees')"`
	Assignees string `json:"assignees" desc:"Assignee usernames (comma-separated). Overrides 'assignee' if both are set. Pass an empty string to clear all assignees."`
	Milestone string `json:"milestone" desc:"Milestone ID"`
}

type AddIssueLabelsArgs struct {
	Owner  string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo   string `json:"repo"   required:"true" desc:"Repository name"`
	Index  int    `json:"index"  required:"true" desc:"Issue index"`
	Labels string `json:"labels" required:"true" desc:"Labels to add (comma-separated)"`
}

type RemoveIssueLabelsArgs struct {
	Owner  string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo   string `json:"repo"   required:"true" desc:"Repository name"`
	Index  int    `json:"index"  required:"true" desc:"Issue index"`
	Labels string `json:"labels" required:"true" desc:"Labels to remove (comma-separated label IDs)"`
}

type IssueStateChangeArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"Issue index"`
	State string `json:"state" required:"true" desc:"State (open|closed)"`
}

type ListIssueCommentsArgs struct {
	Owner  string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo   string `json:"repo"   required:"true" desc:"Repository name"`
	Index  int    `json:"index"  required:"true" desc:"Issue/PR index"`
	Since  string `json:"since"  desc:"After (RFC 3339)"`
	Before string `json:"before" desc:"Before (RFC 3339)"`
}

type GetIssueCommentArgs struct {
	Owner     string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"       required:"true" desc:"Repository name"`
	CommentID int    `json:"comment_id" required:"true" desc:"Comment ID"`
}

type EditIssueCommentArgs struct {
	Owner     string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"       required:"true" desc:"Repository name"`
	CommentID int    `json:"comment_id" required:"true" desc:"Comment ID"`
	Body      string `json:"body"       required:"true" desc:"Content body"`
}

type DeleteIssueCommentArgs struct {
	Owner     string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"       required:"true" desc:"Repository name"`
	CommentID int    `json:"comment_id" required:"true" desc:"Comment ID"`
}

type ListRepoMilestonesArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Page  int    `json:"page"  desc:"Page number (1-based)"          default:"1"`
	Limit int    `json:"limit" desc:"Page size"                      default:"100"`
	State string `json:"state" desc:"Milestone state (open|closed|all)" default:"open"`
}

type ListRepoLabelsArgs struct {
	Owner            string `json:"owner"              required:"true" desc:"Repository owner"`
	Repo             string `json:"repo"               required:"true" desc:"Repository name"`
	Page             int    `json:"page"               desc:"Page number (1-based)" default:"1"`
	Limit            int    `json:"limit"              desc:"Page size"             default:"100"`
	IncludeOrgLabels bool   `json:"include_org_labels" desc:"Merge org-level labels into the response when the owner is an organization. Default true." default:"true"`
}

type CreateLabelArgs struct {
	Owner       string `json:"owner"       required:"true" desc:"Repository owner or organization name"`
	Name        string `json:"name"        required:"true" desc:"Label name"`
	Color       string `json:"color"       required:"true" desc:"Label color hex code (e.g. #00aabb)"`
	Repo        string `json:"repo"        desc:"Repository name. Omit to create an org-level label."`
	Description string `json:"description" desc:"Label description"`
	Exclusive   bool   `json:"exclusive"   desc:"Whether this label is exclusive (only one label with the same scope can be applied)" default:"false"`
}

type EditLabelArgs struct {
	Owner       string `json:"owner"       required:"true" desc:"Repository owner or organization name"`
	ID          int    `json:"id"          required:"true" desc:"Label ID"`
	Repo        string `json:"repo"        desc:"Repository name. Omit to edit an org-level label."`
	Name        string `json:"name"        desc:"New label name"`
	Color       string `json:"color"       desc:"New label color hex code"`
	Description string `json:"description" desc:"New label description"`
	Exclusive   bool   `json:"exclusive"   desc:"Whether this label is exclusive"`
}

type DeleteLabelArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner or organization name"`
	ID    int    `json:"id"    required:"true" desc:"Label ID"`
	Repo  string `json:"repo"  desc:"Repository name. Omit to delete an org-level label."`
}

func issueTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "get_issue_by_index",
			Group:       "issue",
			CLIName:     "get",
			Description: "Get issue by index",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}",
			Args:        GetIssueByIndexArgs{},
		},
		{
			Name:        "list_repo_issues",
			Group:       "issue",
			CLIName:     "list",
			Description: "List repo issues",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues",
			Args:        ListRepoIssuesArgs{},
		},
		{
			Name:        "create_issue",
			Group:       "issue",
			CLIName:     "create",
			Description: "Create issue",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues",
			Args:        CreateIssueArgs{},
		},
		{
			Name:        "create_issue_comment",
			Group:       "issue",
			CLIName:     "comment",
			Description: "Create issue comment",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}/comments",
			Args:        CreateIssueCommentArgs{},
		},
		{
			Name:        "update_issue",
			Group:       "issue",
			CLIName:     "update",
			Description: "Update issue",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}",
			Args:        UpdateIssueArgs{},
		},
		{
			Name:        "add_issue_labels",
			Group:       "issue",
			CLIName:     "add-labels",
			Description: "Add labels to issue",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}/labels",
			Args:        AddIssueLabelsArgs{},
		},
		{
			Name:        "remove_issue_labels",
			Group:       "issue",
			CLIName:     "remove-labels",
			Description: "Remove labels from issue",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}/labels",
			Args:        RemoveIssueLabelsArgs{},
		},
		{
			Name:        "issue_state_change",
			Group:       "issue",
			CLIName:     "state",
			Description: "Change issue state",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}",
			Args:        IssueStateChangeArgs{},
		},
		{
			Name:        "list_issue_comments",
			Group:       "issue",
			CLIName:     "comments",
			Description: listIssueCommentsDesc,
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}/comments",
			Args:        ListIssueCommentsArgs{},
		},
		{
			Name:        "get_issue_comment",
			Group:       "issue",
			CLIName:     "get-comment",
			Description: "Get comment by ID",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/comments/{{.CommentID}}",
			Args:        GetIssueCommentArgs{},
		},
		{
			Name:        "edit_issue_comment",
			Group:       "issue",
			CLIName:     "edit-comment",
			Description: "Edit issue/PR comment",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/comments/{{.CommentID}}",
			Args:        EditIssueCommentArgs{},
		},
		{
			Name:        "delete_issue_comment",
			Group:       "issue",
			CLIName:     "delete-comment",
			Description: "Delete issue/PR comment",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/issues/comments/{{.CommentID}}",
			Args:        DeleteIssueCommentArgs{},
		},
		{
			Name:        "list_repo_milestones",
			Group:       "issue",
			CLIName:     "milestones",
			Description: "List repository milestones",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/milestones",
			Args:        ListRepoMilestonesArgs{},
		},
		{
			Name:        "list_repo_labels",
			Group:       "issue",
			CLIName:     "labels",
			Description: listRepoLabelsDesc,
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/labels",
			Args:        ListRepoLabelsArgs{},
		},
		{
			Name:        "create_label",
			Group:       "issue",
			CLIName:     "create-label",
			Description: "Create a label on a repository or organization",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/labels",
			Args:        CreateLabelArgs{},
		},
		{
			Name:        "edit_label",
			Group:       "issue",
			CLIName:     "edit-label",
			Description: "Edit a label on a repository or organization",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/labels/{{.ID}}",
			Args:        EditLabelArgs{},
		},
		{
			Name:        "delete_label",
			Group:       "issue",
			CLIName:     "delete-label",
			Description: "Delete a label from a repository or organization",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/labels/{{.ID}}",
			Args:        DeleteLabelArgs{},
		},
	}
}
