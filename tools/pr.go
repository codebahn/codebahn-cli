package tools

const getPullRequestDiffDesc = `Get the unified diff of a pull request. Pass an optional file_path to receive only the hunks for that file (match is exact on either the pre- or post-rename path). Use list_pull_request_files first to discover the file paths in the PR.`

type GetPullRequestByIndexArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"PR index"`
}

type ListRepoPullRequestsArgs struct {
	Owner     string `json:"owner"     required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"      required:"true" desc:"Repository name"`
	State     string `json:"state"     desc:"State (open|closed|all)"                                       default:"open"`
	Sort      string `json:"sort"      desc:"Sort (oldest|recentupdate|leastupdate|mostcomment)"`
	Milestone string `json:"milestone" desc:"Milestone ID"`
	Labels    string `json:"labels"    desc:"Label IDs"`
	Page      int    `json:"page"      desc:"Page number (1-based)" default:"1"`
	Limit     int    `json:"limit"     desc:"Page size"             default:"20"`
}

type CreatePullRequestArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Head  string `json:"head"  required:"true" desc:"Head branch"`
	Base  string `json:"base"  required:"true" desc:"Base branch"`
	Title string `json:"title" required:"true" desc:"Title"`
	Body  string `json:"body"  desc:"Content body"`
}

type UpdatePullRequestArgs struct {
	Owner     string `json:"owner"     required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"      required:"true" desc:"Repository name"`
	Index     int    `json:"index"     required:"true" desc:"PR index"`
	Title     string `json:"title"     desc:"Title"`
	Body      string `json:"body"      desc:"Content body"`
	Base      string `json:"base"      desc:"Base branch"`
	Assignee  string `json:"assignee"  desc:"Assignee username"`
	Milestone string `json:"milestone" desc:"Milestone ID"`
}

type ListPullReviewsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"PR index"`
	Page  int    `json:"page"  desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" desc:"Page size"             default:"20"`
}

type GetPullReviewArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"PR index"`
	ID    int    `json:"id"    required:"true" desc:"Review ID"`
}

type ListPullReviewCommentsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"PR index"`
	ID    int    `json:"id"    required:"true" desc:"Review ID"`
}

type MergePullRequestArgs struct {
	Owner                  string `json:"owner"                     required:"true" desc:"Repository owner"`
	Repo                   string `json:"repo"                      required:"true" desc:"Repository name"`
	Index                  int    `json:"index"                     required:"true" desc:"PR index"`
	Style                  string `json:"style"                     required:"true" desc:"Merge style (merge, rebase, rebase-merge, squash)"`
	Title                  string `json:"title"                     desc:"Merge commit title"`
	Message                string `json:"message"                   desc:"Merge commit message"`
	DeleteBranchAfterMerge bool   `json:"delete_branch_after_merge" desc:"Delete head branch after merge"`
	ForceMerge             bool   `json:"force_merge"               desc:"Force merge even if checks have not passed"`
	MergeWhenChecksSucceed bool   `json:"merge_when_checks_succeed" desc:"Schedule merge for when all checks succeed"`
}

type ListPullRequestFilesArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"PR index"`
	Page  int    `json:"page"  desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" desc:"Page size"             default:"50"`
}

type GetPullRequestDiffArgs struct {
	Owner    string `json:"owner"     required:"true" desc:"Repository owner"`
	Repo     string `json:"repo"      required:"true" desc:"Repository name"`
	Index    int    `json:"index"     required:"true" desc:"PR index"`
	FilePath string `json:"file_path" desc:"Optional. Return only the diff section for this file (matched on the diff --git boundary). Omit for the full diff."`
}

type CreatePullReviewArgs struct {
	Owner    string `json:"owner"    required:"true" desc:"Repository owner"`
	Repo     string `json:"repo"     required:"true" desc:"Repository name"`
	Index    int    `json:"index"    required:"true" desc:"PR index"`
	State    string `json:"state"    required:"true" desc:"Review state (APPROVED, REQUEST_CHANGES, COMMENT)"`
	Body     string `json:"body"     desc:"Review body"`
	Comments string `json:"comments" desc:"Inline comments as JSON array, e.g. [{\"path\":\"file.go\",\"body\":\"Fix this\",\"new_position\":10}]"`
}

type SubmitPullReviewArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"PR index"`
	ID    int    `json:"id"    required:"true" desc:"Review ID"`
	State string `json:"state" required:"true" desc:"Review state (APPROVED, REQUEST_CHANGES, COMMENT)"`
	Body  string `json:"body"  desc:"Review body"`
}

type DismissPullReviewArgs struct {
	Owner   string `json:"owner"   required:"true" desc:"Repository owner"`
	Repo    string `json:"repo"    required:"true" desc:"Repository name"`
	Index   int    `json:"index"   required:"true" desc:"PR index"`
	ID      int    `json:"id"      required:"true" desc:"Review ID"`
	Message string `json:"message" required:"true" desc:"Dismissal message"`
}

type DeletePullReviewArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Index int    `json:"index" required:"true" desc:"PR index"`
	ID    int    `json:"id"    required:"true" desc:"Review ID"`
}

type CreateReviewRequestsArgs struct {
	Owner         string `json:"owner"          required:"true" desc:"Repository owner"`
	Repo          string `json:"repo"           required:"true" desc:"Repository name"`
	Index         int    `json:"index"          required:"true" desc:"PR index"`
	Reviewers     string `json:"reviewers"      desc:"Reviewer usernames (comma-separated)"`
	TeamReviewers string `json:"team_reviewers" desc:"Team reviewer names (comma-separated)"`
}

type DeleteReviewRequestsArgs struct {
	Owner         string `json:"owner"          required:"true" desc:"Repository owner"`
	Repo          string `json:"repo"           required:"true" desc:"Repository name"`
	Index         int    `json:"index"          required:"true" desc:"PR index"`
	Reviewers     string `json:"reviewers"      desc:"Reviewer usernames (comma-separated)"`
	TeamReviewers string `json:"team_reviewers" desc:"Team reviewer names (comma-separated)"`
}

func prTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "get_pull_request_by_index",
			Group:       "pr",
			CLIName:     "get",
			Description: "Get pull request by index",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}",
			Args:        GetPullRequestByIndexArgs{},
		},
		{
			Name:        "list_repo_pull_requests",
			Group:       "pr",
			CLIName:     "list",
			Description: "List repo pull requests",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls",
			Args:        ListRepoPullRequestsArgs{},
		},
		{
			Name:        "create_pull_request",
			Group:       "pr",
			CLIName:     "create",
			Description: "Create pull request",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls",
			Args:        CreatePullRequestArgs{},
		},
		{
			Name:        "update_pull_request",
			Group:       "pr",
			CLIName:     "update",
			Description: "Update pull request",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}",
			Args:        UpdatePullRequestArgs{},
		},
		{
			Name:        "list_pull_reviews",
			Group:       "pr",
			CLIName:     "reviews",
			Description: "List reviews for a pull request",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/reviews",
			Args:        ListPullReviewsArgs{},
		},
		{
			Name:        "get_pull_review",
			Group:       "pr",
			CLIName:     "get-review",
			Description: "Get a specific pull request review",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/reviews/{{.ID}}",
			Args:        GetPullReviewArgs{},
		},
		{
			Name:        "list_pull_review_comments",
			Group:       "pr",
			CLIName:     "review-comments",
			Description: "List comments on a pull request review",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/reviews/{{.ID}}/comments",
			Args:        ListPullReviewCommentsArgs{},
		},
		{
			Name:        "merge_pull_request",
			Group:       "pr",
			CLIName:     "merge",
			Description: "Merge a pull request",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/merge",
			Args:        MergePullRequestArgs{},
		},
		{
			Name:        "list_pull_request_files",
			Group:       "pr",
			CLIName:     "files",
			Description: "List changed files in a pull request",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/files",
			Args:        ListPullRequestFilesArgs{},
		},
		{
			Name:        "get_pull_request_diff",
			Group:       "pr",
			CLIName:     "diff",
			Description: getPullRequestDiffDesc,
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}.diff",
			Args:        GetPullRequestDiffArgs{},
		},
		{
			Name:        "create_pull_review",
			Group:       "pr",
			CLIName:     "create-review",
			Description: "Create a pull request review with optional inline comments",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/reviews",
			Args:        CreatePullReviewArgs{},
		},
		{
			Name:        "submit_pull_review",
			Group:       "pr",
			CLIName:     "submit-review",
			Description: "Submit a pending pull request review",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/reviews/{{.ID}}",
			Args:        SubmitPullReviewArgs{},
		},
		{
			Name:        "dismiss_pull_review",
			Group:       "pr",
			CLIName:     "dismiss-review",
			Description: "Dismiss a pull request review",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/reviews/{{.ID}}/dismissals",
			Args:        DismissPullReviewArgs{},
		},
		{
			Name:        "delete_pull_review",
			Group:       "pr",
			CLIName:     "delete-review",
			Description: "Delete a pending pull request review",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/reviews/{{.ID}}",
			Args:        DeletePullReviewArgs{},
		},
		{
			Name:        "create_review_requests",
			Group:       "pr",
			CLIName:     "request-review",
			Description: "Request reviews from specific users or teams",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/requested_reviewers",
			Args:        CreateReviewRequestsArgs{},
		},
		{
			Name:        "delete_review_requests",
			Group:       "pr",
			CLIName:     "cancel-review",
			Description: "Cancel pending review requests",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/pulls/{{.Index}}/requested_reviewers",
			Args:        DeleteReviewRequestsArgs{},
		},
	}
}
