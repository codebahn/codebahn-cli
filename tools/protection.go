package tools

type ListBranchProtectionsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
}

type GetBranchProtectionArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Name  string `json:"name"  required:"true" desc:"Branch protection rule name (glob pattern)"`
}

type CreateBranchProtectionArgs struct {
	Owner                         string `json:"owner"                               required:"true" desc:"Repository owner"`
	Repo                          string `json:"repo"                                required:"true" desc:"Repository name"`
	RuleName                      string `json:"rule_name"                           required:"true" desc:"Branch name glob pattern this rule protects"`
	EnablePush                    *bool  `json:"enable_push"                         desc:"Allow whitelisted users/teams to push directly"`
	EnablePushWhitelist           *bool  `json:"enable_push_whitelist"               desc:"Restrict direct pushes to the push whitelist"`
	PushWhitelistUsernames        string `json:"push_whitelist_usernames"            desc:"Usernames allowed to push directly (comma-separated)" body:"csv"`
	PushWhitelistTeams            string `json:"push_whitelist_teams"                desc:"Teams allowed to push directly (comma-separated)" body:"csv"`
	PushWhitelistDeployKeys       *bool  `json:"push_whitelist_deploy_keys"          desc:"Allow deploy keys to push directly"`
	EnableMergeWhitelist          *bool  `json:"enable_merge_whitelist"              desc:"Restrict merging to the merge whitelist"`
	MergeWhitelistUsernames       string `json:"merge_whitelist_usernames"           desc:"Usernames allowed to merge (comma-separated)" body:"csv"`
	MergeWhitelistTeams           string `json:"merge_whitelist_teams"               desc:"Teams allowed to merge (comma-separated)" body:"csv"`
	EnableStatusCheck             *bool  `json:"enable_status_check"                 desc:"Require status checks to pass before merging"`
	StatusCheckContexts           string `json:"status_check_contexts"               desc:"Required status check contexts (comma-separated)" body:"csv"`
	RequiredApprovals             *int   `json:"required_approvals"                  desc:"Minimum number of approving reviews required"`
	EnableApprovalsWhitelist      *bool  `json:"enable_approvals_whitelist"          desc:"Restrict qualifying approvals to the approvals whitelist"`
	ApprovalsWhitelistUsernames   string `json:"approvals_whitelist_username"        desc:"Usernames whose approvals count (comma-separated)" body:"csv"`
	ApprovalsWhitelistTeams       string `json:"approvals_whitelist_teams"           desc:"Teams whose approvals count (comma-separated)" body:"csv"`
	BlockOnRejectedReviews        *bool  `json:"block_on_rejected_reviews"            desc:"Block merging when there are pending rejected reviews"`
	BlockOnOfficialReviewRequests *bool  `json:"block_on_official_review_requests"   desc:"Block merging when there are pending official review requests"`
	BlockOnOutdatedBranch         *bool  `json:"block_on_outdated_branch"            desc:"Block merging when the head branch is outdated"`
	DismissStaleApprovals         *bool  `json:"dismiss_stale_approvals"             desc:"Dismiss approvals when new commits are pushed"`
	RequireSignedCommits          *bool  `json:"require_signed_commits"              desc:"Require commits to be signed"`
	ProtectedFilePatterns         string `json:"protected_file_patterns"             desc:"Glob patterns of files that cannot be changed directly (semicolon-separated)"`
	UnprotectedFilePatterns       string `json:"unprotected_file_patterns"           desc:"Glob patterns of files exempt from protection (semicolon-separated)"`
}

type UpdateBranchProtectionArgs struct {
	Owner                         string `json:"owner"                               required:"true" desc:"Repository owner"`
	Repo                          string `json:"repo"                                required:"true" desc:"Repository name"`
	Name                          string `json:"name"                                required:"true" desc:"Branch protection rule name (glob pattern)"`
	EnablePush                    *bool  `json:"enable_push"                         desc:"Allow whitelisted users/teams to push directly"`
	EnablePushWhitelist           *bool  `json:"enable_push_whitelist"               desc:"Restrict direct pushes to the push whitelist"`
	PushWhitelistUsernames        string `json:"push_whitelist_usernames"            desc:"Usernames allowed to push directly (comma-separated)" body:"csv"`
	PushWhitelistTeams            string `json:"push_whitelist_teams"                desc:"Teams allowed to push directly (comma-separated)" body:"csv"`
	PushWhitelistDeployKeys       *bool  `json:"push_whitelist_deploy_keys"          desc:"Allow deploy keys to push directly"`
	EnableMergeWhitelist          *bool  `json:"enable_merge_whitelist"              desc:"Restrict merging to the merge whitelist"`
	MergeWhitelistUsernames       string `json:"merge_whitelist_usernames"           desc:"Usernames allowed to merge (comma-separated)" body:"csv"`
	MergeWhitelistTeams           string `json:"merge_whitelist_teams"               desc:"Teams allowed to merge (comma-separated)" body:"csv"`
	EnableStatusCheck             *bool  `json:"enable_status_check"                 desc:"Require status checks to pass before merging"`
	StatusCheckContexts           string `json:"status_check_contexts"               desc:"Required status check contexts (comma-separated)" body:"csv"`
	RequiredApprovals             *int   `json:"required_approvals"                  desc:"Minimum number of approving reviews required"`
	EnableApprovalsWhitelist      *bool  `json:"enable_approvals_whitelist"          desc:"Restrict qualifying approvals to the approvals whitelist"`
	ApprovalsWhitelistUsernames   string `json:"approvals_whitelist_username"        desc:"Usernames whose approvals count (comma-separated)" body:"csv"`
	ApprovalsWhitelistTeams       string `json:"approvals_whitelist_teams"           desc:"Teams whose approvals count (comma-separated)" body:"csv"`
	BlockOnRejectedReviews        *bool  `json:"block_on_rejected_reviews"            desc:"Block merging when there are pending rejected reviews"`
	BlockOnOfficialReviewRequests *bool  `json:"block_on_official_review_requests"   desc:"Block merging when there are pending official review requests"`
	BlockOnOutdatedBranch         *bool  `json:"block_on_outdated_branch"            desc:"Block merging when the head branch is outdated"`
	DismissStaleApprovals         *bool  `json:"dismiss_stale_approvals"             desc:"Dismiss approvals when new commits are pushed"`
	RequireSignedCommits          *bool  `json:"require_signed_commits"              desc:"Require commits to be signed"`
	ProtectedFilePatterns         string `json:"protected_file_patterns"             desc:"Glob patterns of files that cannot be changed directly (semicolon-separated)"`
	UnprotectedFilePatterns       string `json:"unprotected_file_patterns"           desc:"Glob patterns of files exempt from protection (semicolon-separated)"`
}

type DeleteBranchProtectionArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Name  string `json:"name"  required:"true" desc:"Branch protection rule name (glob pattern)"`
}

func protectionTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "list_branch_protections",
			Group:       "protection",
			CLIName:     "list",
			Description: "List branch protection rules for a repository",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branch_protections",
			Args:        ListBranchProtectionsArgs{},
		},
		{
			Name:        "get_branch_protection",
			Group:       "protection",
			CLIName:     "get",
			Description: "Get a branch protection rule by name",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branch_protections/{{.Name}}",
			Args:        GetBranchProtectionArgs{},
		},
		{
			Name:        "create_branch_protection",
			Group:       "protection",
			CLIName:     "create",
			Description: "Create a branch protection rule",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branch_protections",
			Args:        CreateBranchProtectionArgs{},
		},
		{
			Name:        "update_branch_protection",
			Group:       "protection",
			CLIName:     "update",
			Description: "Update a branch protection rule",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branch_protections/{{.Name}}",
			Args:        UpdateBranchProtectionArgs{},
		},
		{
			Name:        "delete_branch_protection",
			Group:       "protection",
			CLIName:     "delete",
			Description: "Delete a branch protection rule",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branch_protections/{{.Name}}",
			Args:        DeleteBranchProtectionArgs{},
			Destructive: true,
		},
	}
}
