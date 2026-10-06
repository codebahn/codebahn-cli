package tools

// MCPTool defines a consolidated MCP tool that dispatches to one or more
// underlying ToolDefs based on the action parameter.
type MCPTool struct {
	Name        string
	Description string
	Args        any
	Actions     []MCPAction
}

// MCPAction maps an action value (plus optional discriminators) to a ToolDef.
type MCPAction struct {
	Action   string // "list", "get", "create", etc.
	Scope    string // "repo" or "org" (empty if not scoped)
	Type     string // "variable" or "secret" (empty if not typed)
	ToolName string // -> tools.ByName()
}

// Key returns the dispatch lookup key for this action.
func (a MCPAction) Key() string {
	k := a.Action
	if a.Type != "" {
		k += "_" + a.Type
	}
	if a.Scope != "" {
		k += "_" + a.Scope
	}
	return k
}

// MCPAll is the registry of consolidated MCP tools.
var MCPAll []MCPTool

func init() {
	MCPAll = buildMCPRegistry()
}

// MCPByName returns the MCPTool with the given name, or panics.
func MCPByName(name string) MCPTool {
	for _, mt := range MCPAll {
		if mt.Name == name {
			return mt
		}
	}
	panic("mcp: unknown tool " + name)
}

func buildMCPRegistry() []MCPTool {
	return []MCPTool{
		mcpUserInfo(),
		mcpSearch(),
		mcpRepos(),
		mcpFiles(),
		mcpBranches(),
		mcpCommits(),
		mcpIssues(),
		mcpLabels(),
		mcpComments(),
		mcpPullRequests(),
		mcpReviews(),
		mcpCI(),
		mcpCIConfig(),
		mcpWebhooks(),
		mcpReleases(),
	}
}

// --- Arg structs for consolidated tools ---

type MCPUserInfoArgs struct{}

type MCPSearchArgs struct {
	Action   string `json:"action"   required:"true" desc:"Search scope to query" enum:"all,code,repos" default:"all"`
	Keyword  string `json:"keyword"  required:"true" desc:"Search query"`
	Language string `json:"language" desc:"Filter by programming language. Only for action=code"`
	Filename string `json:"filename" desc:"Filter by filename or path. Only for action=code"`
	Mode     string `json:"mode"     desc:"Search mode: exact, union, fuzzy. Only for action=code" default:"exact"`
	Page     int    `json:"page"     desc:"Page number (1-based). Ignored for action=all" default:"1"`
	Limit    int    `json:"limit"    desc:"Page size. Ignored for action=all" default:"30"`
}

type MCPReposArgs struct {
	Action        string `json:"action"        required:"true" desc:"Operation to perform" enum:"list,create,update"`
	Owner         string `json:"owner"         desc:"Repository owner or org name. Required for create (under org) and update"`
	Repo          string `json:"repo"          desc:"Repository name. Required for update"`
	Name          string `json:"name"          desc:"New repository name. Required for create"`
	Description   string `json:"description"   desc:"Repository description"`
	Private       *bool  `json:"private"       desc:"Whether repository is private"`
	DefaultBranch string `json:"default_branch" desc:"Default branch name"`
	AutoInit      bool   `json:"auto_init"     desc:"Initialize with README"`
	Template      *bool  `json:"template"      desc:"Whether this is a template repository"`
	Gitignores    string `json:"gitignores"    desc:"Gitignore template names"`
	License       string `json:"license"       desc:"License template"`
	Readme        string `json:"readme"        desc:"README content"`
	IssueLabels   string `json:"issue_labels"  desc:"Issue label set"`

	DefaultMergeStyle             string `json:"default_merge_style"               desc:"Default merge style (merge, rebase, rebase-merge, squash, fast-forward-only). For update"`
	Archived                      *bool  `json:"archived"                          desc:"Archive the repository. For update"`
	HasIssues                     *bool  `json:"has_issues"                        desc:"Enable issue tracker. For update"`
	HasWiki                       *bool  `json:"has_wiki"                          desc:"Enable wiki. For update"`
	HasPullRequests               *bool  `json:"has_pull_requests"                 desc:"Enable pull requests. For update"`
	HasProjects                   *bool  `json:"has_projects"                      desc:"Enable projects. For update"`
	HasReleases                   *bool  `json:"has_releases"                      desc:"Enable releases. For update"`
	HasPackages                   *bool  `json:"has_packages"                      desc:"Enable package registry. For update"`
	HasActions                    *bool  `json:"has_actions"                       desc:"Enable Actions CI. For update"`
	AllowMergeCommits             *bool  `json:"allow_merge_commits"               desc:"Allow merge commits. For update"`
	AllowRebase                   *bool  `json:"allow_rebase"                      desc:"Allow rebase merging. For update"`
	AllowSquashMerge              *bool  `json:"allow_squash_merge"                desc:"Allow squash merging. For update"`
	DefaultDeleteBranchAfterMerge *bool  `json:"default_delete_branch_after_merge" desc:"Delete head branch by default after merge. For update"`

	Page  int `json:"page"  desc:"Page number (1-based). For list" default:"1"`
	Limit int `json:"limit" desc:"Page size. For list" default:"100"`
}

type MCPFilesArgs struct {
	Action        string `json:"action"       required:"true" desc:"Operation to perform" enum:"read,list,tree,create,update,delete"`
	Owner         string `json:"owner"        required:"true" desc:"Repository owner"`
	Repo          string `json:"repo"         required:"true" desc:"Repository name"`
	Ref           string `json:"ref"          desc:"Branch, tag, or commit SHA. Required for read/list/tree"`
	Path          string `json:"path"         desc:"File or directory path. Required for read, list (empty string for root), create, update, delete"`
	Content       string `json:"content"      desc:"File content (plain text, base64-encoded automatically). Required for create/update"`
	Message       string `json:"message"      desc:"Commit message. Required for create/update/delete"`
	BranchName    string `json:"branch_name"  desc:"Target branch. Required for create/update/delete"`
	SHA           string `json:"sha"          desc:"File SHA from a prior read with with_metadata=true. Required for update/delete"`
	NewBranchName string `json:"new_branch_name" desc:"Create a new branch from branch_name and commit there"`
	WithMetadata  bool   `json:"with_metadata" desc:"Return full metadata (SHA, encoding, links) instead of plain text. For read"`
	StartLine     int    `json:"start_line"   desc:"First line to return (1-indexed). For read"`
	EndLine       int    `json:"end_line"     desc:"Last line to return (1-indexed). For read"`
	Recursive     bool   `json:"recursive"    desc:"Return complete file tree. For tree"`
	Page          int    `json:"page"         desc:"Page number (1-based). For tree" default:"1"`
	Limit         int    `json:"limit"        desc:"Page size. For tree" default:"1000"`
}

type MCPBranchesArgs struct {
	Action    string `json:"action"    required:"true" desc:"Operation to perform" enum:"list,create,delete,compare,list_protections,get_protection,create_protection,update_protection,delete_protection"`
	Owner     string `json:"owner"     required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"      required:"true" desc:"Repository name"`
	Branch    string `json:"branch"    desc:"Branch name. For create/delete"`
	OldBranch string `json:"old_branch" desc:"Source branch to create from. For create"`
	Base      string `json:"base"      desc:"Base ref. For compare"`
	Head      string `json:"head"      desc:"Head ref. For compare"`
	Name      string `json:"name"      desc:"Protection rule name (glob pattern). For get/update/delete_protection"`
	RuleName  string `json:"rule_name" desc:"Glob pattern for the new rule. For create_protection"`
	Page      int    `json:"page"      desc:"Page number. For list" default:"1"`
	Limit     int    `json:"limit"     desc:"Page size. For list" default:"100"`

	EnablePush                    *bool  `json:"enable_push"                       desc:"Allow whitelisted users/teams to push. For create/update_protection"`
	EnablePushWhitelist           *bool  `json:"enable_push_whitelist"             desc:"Restrict direct pushes to whitelist. For create/update_protection"`
	PushWhitelistUsernames        string `json:"push_whitelist_usernames"          desc:"Comma-separated usernames allowed to push. For create/update_protection"`
	PushWhitelistTeams            string `json:"push_whitelist_teams"              desc:"Comma-separated teams allowed to push. For create/update_protection"`
	PushWhitelistDeployKeys       *bool  `json:"push_whitelist_deploy_keys"        desc:"Allow deploy keys to push. For create/update_protection"`
	EnableMergeWhitelist          *bool  `json:"enable_merge_whitelist"            desc:"Restrict merging to whitelist. For create/update_protection"`
	MergeWhitelistUsernames       string `json:"merge_whitelist_usernames"         desc:"Comma-separated usernames allowed to merge. For create/update_protection"`
	MergeWhitelistTeams           string `json:"merge_whitelist_teams"             desc:"Comma-separated teams allowed to merge. For create/update_protection"`
	EnableStatusCheck             *bool  `json:"enable_status_check"               desc:"Require status checks before merging. For create/update_protection"`
	StatusCheckContexts           string `json:"status_check_contexts"             desc:"Comma-separated required status check contexts. For create/update_protection"`
	RequiredApprovals             *int   `json:"required_approvals"                desc:"Minimum approving reviews required. For create/update_protection"`
	EnableApprovalsWhitelist      *bool  `json:"enable_approvals_whitelist"        desc:"Restrict approvals to whitelist. For create/update_protection"`
	ApprovalsWhitelistUsernames   string `json:"approvals_whitelist_username"      desc:"Comma-separated usernames whose approvals count. For create/update_protection"`
	ApprovalsWhitelistTeams       string `json:"approvals_whitelist_teams"         desc:"Comma-separated teams whose approvals count. For create/update_protection"`
	BlockOnRejectedReviews        *bool  `json:"block_on_rejected_reviews"         desc:"Block merge on rejected reviews. For create/update_protection"`
	BlockOnOfficialReviewRequests *bool  `json:"block_on_official_review_requests" desc:"Block merge on pending review requests. For create/update_protection"`
	BlockOnOutdatedBranch         *bool  `json:"block_on_outdated_branch"          desc:"Block merge when branch is outdated. For create/update_protection"`
	DismissStaleApprovals         *bool  `json:"dismiss_stale_approvals"           desc:"Dismiss approvals on new pushes. For create/update_protection"`
	RequireSignedCommits          *bool  `json:"require_signed_commits"            desc:"Require signed commits. For create/update_protection"`
	ProtectedFilePatterns         string `json:"protected_file_patterns"           desc:"Semicolon-separated glob patterns of protected files. For create/update_protection"`
	UnprotectedFilePatterns       string `json:"unprotected_file_patterns"         desc:"Semicolon-separated glob patterns exempt from protection. For create/update_protection"`
}

type MCPCommitsArgs struct {
	Action string `json:"action" required:"true" desc:"Operation to perform" enum:"list,diff"`
	Owner  string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo   string `json:"repo"   required:"true" desc:"Repository name"`
	SHA    string `json:"sha"    desc:"Branch, tag, or commit SHA. Required for diff; optional start point for list"`
	Path   string `json:"path"   desc:"File or directory path. For list: only commits touching this path. For diff: only this file's diff"`
	Page   int    `json:"page"   desc:"Page number (1-based). For list" default:"1"`
	Limit  int    `json:"limit"  desc:"Page size. For list" default:"100"`
}

type MCPIssuesArgs struct {
	Action    string `json:"action"    required:"true" desc:"Operation to perform" enum:"list,get,create,update,list_milestones"`
	Owner     string `json:"owner"     required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"      required:"true" desc:"Repository name"`
	Index     int    `json:"index"     desc:"Issue index. Required for get/update"`
	Title     string `json:"title"     desc:"Issue title. Required for create"`
	Body      string `json:"body"      desc:"Issue content body"`
	State     string `json:"state"     desc:"Issue state. For list: filter (open/closed/all). For update: change state (open/closed)" default:"open"`
	Assignee  string `json:"assignee"  desc:"Assignee username. For update"`
	Assignees string `json:"assignees" desc:"Comma-separated assignee usernames. For update"`
	Milestone string `json:"milestone" desc:"Milestone ID. For list/update"`
	Type      string `json:"type"      desc:"Filter: issues or pulls. For list"`
	Labels    string `json:"labels"    desc:"Comma-separated label names. For list"`
	Page      int    `json:"page"      desc:"Page number. For list/list_milestones" default:"1"`
	Limit     int    `json:"limit"     desc:"Page size. For list/list_milestones" default:"20"`
}

type MCPLabelsArgs struct {
	Action           string `json:"action"             required:"true" desc:"Operation to perform" enum:"list,create,edit,delete,add_to_issue,remove_from_issue"`
	Owner            string `json:"owner"              required:"true" desc:"Repository owner or organization name"`
	Repo             string `json:"repo"               desc:"Repository name. Omit for org-level labels"`
	ID               int    `json:"id"                 desc:"Label ID. Required for edit/delete"`
	Index            int    `json:"index"              desc:"Issue index. Required for add_to_issue/remove_from_issue"`
	Name             string `json:"name"               desc:"Label name. Required for create"`
	Color            string `json:"color"              desc:"Hex color code (e.g. #00aabb). Required for create"`
	Labels           string `json:"labels"             desc:"Comma-separated labels to add (names) or remove (IDs). For add_to_issue/remove_from_issue"`
	Description      string `json:"description"        desc:"Label description"`
	Exclusive        bool   `json:"exclusive"          desc:"Whether label is exclusive within its scope"`
	IncludeOrgLabels bool   `json:"include_org_labels" desc:"Include org labels in list response" default:"true"`
	Page             int    `json:"page"               desc:"Page number. For list" default:"1"`
	Limit            int    `json:"limit"              desc:"Page size. For list" default:"100"`
}

type MCPCommentsArgs struct {
	Action string `json:"action" required:"true" desc:"Operation to perform" enum:"list,get,create,edit,delete"`
	Owner  string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo   string `json:"repo"   required:"true" desc:"Repository name"`
	Index  int    `json:"index"  desc:"Issue or PR index. Required for list/create"`
	ID     int    `json:"id"     desc:"Comment ID. Required for get/edit/delete"`
	Body   string `json:"body"   desc:"Comment body (markdown). Required for create/edit"`
	Since  string `json:"since"  desc:"ISO 8601 timestamp. Only list comments updated after this"`
	Before string `json:"before" desc:"ISO 8601 timestamp. Only list comments updated before this"`
}

type MCPPullRequestsArgs struct {
	Action                 string `json:"action"                     required:"true" desc:"Operation to perform" enum:"list,get,create,update,merge,diff,list_files,list_commits"`
	Owner                  string `json:"owner"                      required:"true" desc:"Repository owner"`
	Repo                   string `json:"repo"                       required:"true" desc:"Repository name"`
	Index                  int    `json:"index"                      desc:"PR index. Required for all actions except list/create"`
	Title                  string `json:"title"                      desc:"PR title. Required for create"`
	Head                   string `json:"head"                       desc:"Head branch. Required for create"`
	Base                   string `json:"base"                       desc:"Base branch. Required for create"`
	Body                   string `json:"body"                       desc:"PR description body"`
	State                  string `json:"state"                      desc:"Filter: open, closed, all. For list" default:"open"`
	Style                  string `json:"style"                      desc:"Merge style: merge, rebase, rebase-merge, squash. Required for merge"`
	Message                string `json:"message"                    desc:"Merge commit message. For merge"`
	DeleteBranchAfterMerge bool   `json:"delete_branch_after_merge"  desc:"Delete head branch after merge"`
	ForceMerge             bool   `json:"force_merge"                desc:"Force merge even if checks have not passed"`
	MergeWhenChecksSucceed bool   `json:"merge_when_checks_succeed"  desc:"Schedule merge for when all checks pass"`
	Assignee               string `json:"assignee"                   desc:"Assignee username. For update"`
	Milestone              string `json:"milestone"                  desc:"Milestone ID. For list/update"`
	FilePath               string `json:"file_path"                  desc:"Return only this file's diff. For diff"`
	Sort                   string `json:"sort"                       desc:"Sort order. For list"`
	Labels                 string `json:"labels"                     desc:"Label IDs filter. For list"`
	Page                   int    `json:"page"                       desc:"Page number" default:"1"`
	Limit                  int    `json:"limit"                      desc:"Page size" default:"20"`
}

type MCPReviewsArgs struct {
	Action        string `json:"action"         required:"true" desc:"Operation to perform" enum:"list,get,list_comments,create,submit,dismiss,delete,request,cancel_request"`
	Owner         string `json:"owner"          required:"true" desc:"Repository owner"`
	Repo          string `json:"repo"           required:"true" desc:"Repository name"`
	Index         int    `json:"index"          required:"true" desc:"PR index"`
	ID            int    `json:"id"             desc:"Review ID. Required for get/list_comments/submit/dismiss/delete"`
	State         string `json:"state"          desc:"Review verdict: APPROVED, REQUEST_CHANGES, or COMMENT. For create/submit"`
	Body          string `json:"body"           desc:"Review body text"`
	Message       string `json:"message"        desc:"Dismissal message. Required for dismiss"`
	Comments      string `json:"comments"       desc:"Inline comments as JSON array. For create"`
	Reviewers     string `json:"reviewers"      desc:"Comma-separated reviewer usernames. For request/cancel_request"`
	TeamReviewers string `json:"team_reviewers" desc:"Comma-separated team reviewer names. For request/cancel_request"`
	Page          int    `json:"page"           desc:"Page number. For list" default:"1"`
	Limit         int    `json:"limit"          desc:"Page size. For list" default:"20"`
}

type MCPCIArgs struct {
	Action     string `json:"action"      required:"true" desc:"Operation to perform" enum:"dispatch,cancel,list_runs,get_run,get_logs"`
	Owner      string `json:"owner"       required:"true" desc:"Repository owner"`
	Repo       string `json:"repo"        required:"true" desc:"Repository name"`
	WorkflowID string `json:"workflow_id" desc:"Workflow filename (e.g. ci.yml). Required for dispatch"`
	RunID      int    `json:"run_id"      desc:"Workflow run ID. Required for cancel/get_run/get_logs"`
	Ref        string `json:"ref"         desc:"Branch or tag to run on. For dispatch"`
	Inputs     string `json:"inputs"      desc:"Workflow inputs as JSON object. For dispatch"`
	Page       int    `json:"page"        desc:"Page number. For list_runs" default:"1"`
	Limit      int    `json:"limit"       desc:"Page size. For list_runs" default:"20"`
}

type MCPCIConfigArgs struct {
	Action string `json:"action" required:"true" desc:"Operation to perform. Note: get is only available for variables (secrets are write-only)" enum:"list,get,set,delete"`
	Type   string `json:"type"   required:"true" desc:"Config type" enum:"variable,secret"`
	Scope  string `json:"scope"  required:"true" desc:"Scope level" enum:"repo,org"`
	Owner  string `json:"owner"  required:"true" desc:"Repository owner or organization name"`
	Repo   string `json:"repo"   desc:"Repository name. Required when scope=repo"`
	Name   string `json:"name"   desc:"Variable or secret name. Required for get/set/delete"`
	Value  string `json:"value"  desc:"Value to store. For secrets, this is write-only. Required for set"`
	Page   int    `json:"page"   desc:"Page number. For list" default:"1"`
	Limit  int    `json:"limit"  desc:"Page size. For list" default:"30"`
}

type MCPWebhooksArgs struct {
	Action              string `json:"action"               required:"true" desc:"Operation to perform. Note: test is only available for repo-scope webhooks" enum:"list,get,create,update,delete,test"`
	Scope               string `json:"scope"                required:"true" desc:"Scope level" enum:"repo,org"`
	Owner               string `json:"owner"                required:"true" desc:"Repository owner or organization name"`
	Repo                string `json:"repo"                 desc:"Repository name. Required when scope=repo"`
	ID                  int    `json:"id"                   desc:"Webhook ID. Required for get/update/delete/test"`
	WebhookType         string `json:"type"                 desc:"Webhook type: gitea, slack, discord. Default: gitea. For create" default:"gitea"`
	URL                 string `json:"url"                  desc:"Target URL for payload delivery. Required for create"`
	ContentType         string `json:"content_type"         desc:"Payload format: json (default) or form" default:"json"`
	Secret              string `json:"secret"               desc:"Secret used to sign payloads"`
	Events              string `json:"events"               desc:"Comma-separated events (e.g. push,pull_request). Required for create"`
	Active              *bool  `json:"active"               desc:"Whether the webhook is active"`
	BranchFilter        string `json:"branch_filter"        desc:"Glob pattern restricting which branches trigger delivery"`
	AuthorizationHeader string `json:"authorization_header" desc:"Authorization header sent with each delivery"`
	Page                int    `json:"page"                 desc:"Page number. For list" default:"1"`
	Limit               int    `json:"limit"                desc:"Page size. For list" default:"30"`
}

type MCPReleasesArgs struct {
	Action       string `json:"action"        required:"true" desc:"Operation to perform" enum:"list,get,get_latest,get_by_tag,create,update,delete,list_tags,create_tag,delete_tag,list_attachments,delete_attachment"`
	Owner        string `json:"owner"         required:"true" desc:"Repository owner"`
	Repo         string `json:"repo"          required:"true" desc:"Repository name"`
	ID           int    `json:"id"            desc:"Release ID. For get/update/delete/list_attachments"`
	Tag          string `json:"tag"           desc:"Tag name. For get_by_tag/delete_tag"`
	TagName      string `json:"tag_name"      desc:"Tag name to create. For create/create_tag"`
	Target       string `json:"target"        desc:"Branch or commit SHA for the tag"`
	ReleaseName  string `json:"name"          desc:"Release title"`
	Body         string `json:"body"          desc:"Release description (markdown)"`
	TagMessage   string `json:"message"       desc:"Tag message. For create_tag"`
	Draft        *bool  `json:"draft"         desc:"Whether release is a draft"`
	Prerelease   *bool  `json:"prerelease"    desc:"Whether release is a pre-release"`
	AttachmentID int    `json:"attachment_id" desc:"Attachment ID. For delete_attachment"`
	Page         int    `json:"page"          desc:"Page number. For list/list_tags" default:"1"`
	Limit        int    `json:"limit"         desc:"Page size. For list/list_tags" default:"30"`
}

// --- MCPTool builders ---

func mcpUserInfo() MCPTool {
	return MCPTool{
		Name:        "user_info",
		Description: "Get the authenticated user's profile: username, email, permissions, and account settings.",
		Args:        MCPUserInfoArgs{},
		Actions: []MCPAction{
			{Action: "", ToolName: "get_my_user_info"},
		},
	}
}

func mcpSearch() MCPTool {
	return MCPTool{
		Name:        "search",
		Description: "Search across repositories, code, issues, users, and organizations. Action 'all' returns hit counts across types; 'code' and 'repos' return paginated results with type-specific filters.",
		Args:        MCPSearchArgs{},
		Actions: []MCPAction{
			{Action: "all", ToolName: "search"},
			{Action: "code", ToolName: "search_code"},
			{Action: "repos", ToolName: "search_repos"},
		},
	}
}

func mcpRepos() MCPTool {
	return MCPTool{
		Name:        "repos",
		Description: "List, create, or update repositories. Use 'list' for your repositories, 'create' to make a new one (under your account or an org), 'update' to change settings like description, visibility, merge options, or enabled features.",
		Args:        MCPReposArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_my_repos"},
			{Action: "create", ToolName: "create_repo"},
			{Action: "update", ToolName: "update_repo"},
		},
	}
}

func mcpFiles() MCPTool {
	return MCPTool{
		Name:        "files",
		Description: "Read, list, or write files in a repository. Write actions (create, update, delete) each produce a commit. Use 'read' for file content (plain text by default; set with_metadata=true for the SHA needed by update/delete). Use 'list' for a single directory level, 'tree' for a recursive file listing.",
		Args:        MCPFilesArgs{},
		Actions: []MCPAction{
			{Action: "read", ToolName: "get_file_content"},
			{Action: "list", ToolName: "list_repo_contents"},
			{Action: "tree", ToolName: "get_repo_tree"},
			{Action: "create", ToolName: "create_file"},
			{Action: "update", ToolName: "update_file"},
			{Action: "delete", ToolName: "delete_file"},
		},
	}
}

func mcpBranches() MCPTool {
	return MCPTool{
		Name:        "branches",
		Description: "Manage branches, compare refs, and configure branch protection rules. Use 'compare' to see commits reachable from head but not base. Protection actions manage merge requirements, status checks, and push restrictions.",
		Args:        MCPBranchesArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_branches"},
			{Action: "create", ToolName: "create_branch"},
			{Action: "delete", ToolName: "delete_branch"},
			{Action: "compare", ToolName: "compare_refs"},
			{Action: "list_protections", ToolName: "list_branch_protections"},
			{Action: "get_protection", ToolName: "get_branch_protection"},
			{Action: "create_protection", ToolName: "create_branch_protection"},
			{Action: "update_protection", ToolName: "update_branch_protection"},
			{Action: "delete_protection", ToolName: "delete_branch_protection"},
		},
	}
}

func mcpCommits() MCPTool {
	return MCPTool{
		Name:        "commits",
		Description: "Browse commit history and view diffs. Use 'list' to list commits (optionally filtered to a file path). Use 'diff' to get the unified diff of a single commit.",
		Args:        MCPCommitsArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_repo_commits"},
			{Action: "diff", ToolName: "get_commit_diff"},
		},
	}
}

func mcpIssues() MCPTool {
	return MCPTool{
		Name:        "issues",
		Description: "Manage issues and milestones. Use 'update' with the state field to open or close an issue. Use 'list_milestones' to see available milestones for filtering or assignment.",
		Args:        MCPIssuesArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_repo_issues"},
			{Action: "get", ToolName: "get_issue_by_index"},
			{Action: "create", ToolName: "create_issue"},
			{Action: "update", ToolName: "update_issue"},
			{Action: "list_milestones", ToolName: "list_repo_milestones"},
		},
	}
}

func mcpLabels() MCPTool {
	return MCPTool{
		Name:        "labels",
		Description: "Manage labels at repository or organization scope, and add or remove labels from issues. Omit repo to operate on org-level labels.",
		Args:        MCPLabelsArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_repo_labels"},
			{Action: "create", ToolName: "create_label"},
			{Action: "edit", ToolName: "edit_label"},
			{Action: "delete", ToolName: "delete_label"},
			{Action: "add_to_issue", ToolName: "add_issue_labels"},
			{Action: "remove_from_issue", ToolName: "remove_issue_labels"},
		},
	}
}

func mcpComments() MCPTool {
	return MCPTool{
		Name:        "comments",
		Description: "Manage comments on issues and pull requests. Comments are shared between issues and PRs; use the same index for either.",
		Args:        MCPCommentsArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_issue_comments"},
			{Action: "get", ToolName: "get_issue_comment"},
			{Action: "create", ToolName: "create_issue_comment"},
			{Action: "edit", ToolName: "edit_issue_comment"},
			{Action: "delete", ToolName: "delete_issue_comment"},
		},
	}
}

func mcpPullRequests() MCPTool {
	return MCPTool{
		Name:        "pull_requests",
		Description: "Manage pull requests: create, update, merge, and inspect changes. Use 'diff' for the unified diff, 'list_files' for changed file paths, 'list_commits' for individual commits in the PR.",
		Args:        MCPPullRequestsArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_repo_pull_requests"},
			{Action: "get", ToolName: "get_pull_request_by_index"},
			{Action: "create", ToolName: "create_pull_request"},
			{Action: "update", ToolName: "update_pull_request"},
			{Action: "merge", ToolName: "merge_pull_request"},
			{Action: "diff", ToolName: "get_pull_request_diff"},
			{Action: "list_files", ToolName: "list_pull_request_files"},
			{Action: "list_commits", ToolName: "list_pr_commits"},
		},
	}
}

func mcpReviews() MCPTool {
	return MCPTool{
		Name:        "reviews",
		Description: "Manage PR reviews, review comments, and review requests. Create a review with inline comments and a verdict (APPROVED, REQUEST_CHANGES, COMMENT). Use 'request' to ask users or teams to review, 'cancel_request' to withdraw.",
		Args:        MCPReviewsArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_pull_reviews"},
			{Action: "get", ToolName: "get_pull_review"},
			{Action: "list_comments", ToolName: "list_pull_review_comments"},
			{Action: "create", ToolName: "create_pull_review"},
			{Action: "submit", ToolName: "submit_pull_review"},
			{Action: "dismiss", ToolName: "dismiss_pull_review"},
			{Action: "delete", ToolName: "delete_pull_review"},
			{Action: "request", ToolName: "create_review_requests"},
			{Action: "cancel_request", ToolName: "delete_review_requests"},
		},
	}
}

func mcpCI() MCPTool {
	return MCPTool{
		Name:        "ci",
		Description: "Trigger and monitor CI workflows. Use 'dispatch' to start a workflow run, 'cancel' to abort a running build, 'list_runs' for run history, 'get_run' for details, 'get_logs' for job output grouped by step.",
		Args:        MCPCIArgs{},
		Actions: []MCPAction{
			{Action: "dispatch", ToolName: "dispatch_workflow"},
			{Action: "cancel", ToolName: "cancel_build"},
			{Action: "list_runs", ToolName: "list_workflow_runs"},
			{Action: "get_run", ToolName: "get_workflow_run"},
			{Action: "get_logs", ToolName: "get_job_logs"},
		},
	}
}

func mcpCIConfig() MCPTool {
	return MCPTool{
		Name:        "ci_config",
		Description: "Manage Actions variables and secrets at repository or organization scope. Variables store plaintext config; secrets store encrypted values (write-only, cannot be read back). The 'set' action creates or updates (upsert). The 'get' action is available for variables only.",
		Args:        MCPCIConfigArgs{},
		Actions: []MCPAction{
			{Action: "list", Type: "variable", Scope: "repo", ToolName: "list_variables"},
			{Action: "list", Type: "variable", Scope: "org", ToolName: "list_org_variables"},
			{Action: "list", Type: "secret", Scope: "repo", ToolName: "list_secrets"},
			{Action: "list", Type: "secret", Scope: "org", ToolName: "list_org_secrets"},
			{Action: "get", Type: "variable", Scope: "repo", ToolName: "get_variable"},
			{Action: "get", Type: "variable", Scope: "org", ToolName: "get_org_variable"},
			{Action: "set", Type: "variable", Scope: "repo", ToolName: "create_variable"},
			{Action: "set", Type: "variable", Scope: "org", ToolName: "create_org_variable"},
			{Action: "set", Type: "secret", Scope: "repo", ToolName: "set_secret"},
			{Action: "set", Type: "secret", Scope: "org", ToolName: "set_org_secret"},
			{Action: "delete", Type: "variable", Scope: "repo", ToolName: "delete_variable"},
			{Action: "delete", Type: "variable", Scope: "org", ToolName: "delete_org_variable"},
			{Action: "delete", Type: "secret", Scope: "repo", ToolName: "delete_secret"},
			{Action: "delete", Type: "secret", Scope: "org", ToolName: "delete_org_secret"},
		},
	}
}

func mcpWebhooks() MCPTool {
	return MCPTool{
		Name:        "webhooks",
		Description: "Manage webhooks at repository or organization scope. Use scope 'repo' for repository webhooks, 'org' for organization-wide hooks. The 'test' action is only available for repository webhooks.",
		Args:        MCPWebhooksArgs{},
		Actions: []MCPAction{
			{Action: "list", Scope: "repo", ToolName: "list_hooks"},
			{Action: "list", Scope: "org", ToolName: "list_org_hooks"},
			{Action: "get", Scope: "repo", ToolName: "get_hook"},
			{Action: "get", Scope: "org", ToolName: "get_org_hook"},
			{Action: "create", Scope: "repo", ToolName: "create_hook"},
			{Action: "create", Scope: "org", ToolName: "create_org_hook"},
			{Action: "update", Scope: "repo", ToolName: "update_hook"},
			{Action: "update", Scope: "org", ToolName: "update_org_hook"},
			{Action: "delete", Scope: "repo", ToolName: "delete_hook"},
			{Action: "delete", Scope: "org", ToolName: "delete_org_hook"},
			{Action: "test", Scope: "repo", ToolName: "test_hook"},
		},
	}
}

func mcpReleases() MCPTool {
	return MCPTool{
		Name:        "releases",
		Description: "Manage releases, git tags, and release attachments. Use 'get_latest' for the most recent release or 'get_by_tag' to look up by tag name. Tag actions (list_tags, create_tag, delete_tag) manage git tags independent of releases.",
		Args:        MCPReleasesArgs{},
		Actions: []MCPAction{
			{Action: "list", ToolName: "list_releases"},
			{Action: "get", ToolName: "get_release"},
			{Action: "get_latest", ToolName: "get_latest_release"},
			{Action: "get_by_tag", ToolName: "get_release_by_tag"},
			{Action: "create", ToolName: "create_release"},
			{Action: "update", ToolName: "update_release"},
			{Action: "delete", ToolName: "delete_release"},
			{Action: "list_tags", ToolName: "list_tags"},
			{Action: "create_tag", ToolName: "create_tag"},
			{Action: "delete_tag", ToolName: "delete_tag"},
			{Action: "list_attachments", ToolName: "list_release_attachments"},
			{Action: "delete_attachment", ToolName: "delete_release_attachment"},
		},
	}
}
