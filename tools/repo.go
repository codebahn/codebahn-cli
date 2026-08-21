package tools

const getFileContentDesc = `Get file content as plain text by default. Set with_metadata=true for binary files, or when you need the SHA/encoding/links from the full ContentsResponse (e.g. before a follow-up update_file call). Optional start_line and end_line request a 1-indexed inclusive line range; out-of-range values clamp to the file extent. Range parameters are ignored when with_metadata=true.`

const listRepoContentsDesc = `List the files and directories at a given path in a repository. Use path="" to list the repository root. Returns one level of entries at the specified path; for a full recursive file tree use get_repo_tree with recursive=true.`

const getRepoTreeDesc = `Get the Git tree of a repository. With recursive=true, returns the complete file tree in a single response (subject to the server's tree-endpoint size cap); use this when you need all paths at once. With recursive=false (default), returns only the top-level entries of the tree.`

type CreateRepoArgs struct {
	Name          string `json:"name"           required:"true" desc:"Repo name"`
	Description   string `json:"description"    desc:"Description"`
	Owner         string `json:"owner"          desc:"Owner/org name"`
	Private       bool   `json:"private"        desc:"Private repo"`
	IssueLabels   string `json:"issue_labels"   desc:"Issue label set"`
	AutoInit      bool   `json:"auto_init"      desc:"Auto-initialize"`
	Template      bool   `json:"template"       desc:"Template repo"`
	Gitignores    string `json:"gitignores"     desc:"Gitignore templates"`
	License       string `json:"license"        desc:"License"`
	Readme        string `json:"readme"         desc:"README content"`
	DefaultBranch string `json:"default_branch" desc:"Default branch"`
}

type ListMyReposArgs struct {
	Page  int `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int `json:"limit" required:"true" desc:"Page size"             default:"100"`
}

type GetFileContentArgs struct {
	Owner        string `json:"owner"         required:"true" desc:"Repository owner"`
	Repo         string `json:"repo"          required:"true" desc:"Repository name"`
	Ref          string `json:"ref"           required:"true" desc:"Ref (branch/tag/commit)"`
	FilePath     string `json:"filePath"      required:"true" desc:"File path"`
	WithMetadata bool   `json:"with_metadata" desc:"Return the full ContentsResponse (sha, encoding, links, type, size, base64 content) instead of plain text."`
	StartLine    int    `json:"start_line"    desc:"Optional 1-indexed first line of the slice (inclusive). Defaults to 1 when only end_line is set."`
	EndLine      int    `json:"end_line"      desc:"Optional 1-indexed last line of the slice (inclusive). Defaults to the file's last line when only start_line is set."`
}

type CreateFileArgs struct {
	Owner         string `json:"owner"           required:"true" desc:"Repository owner"`
	Repo          string `json:"repo"            required:"true" desc:"Repository name"`
	FilePath      string `json:"filePath"        required:"true" desc:"File path"`
	Content       string `json:"content"         required:"true" desc:"Content (plain text, base64-encoded automatically)"`
	Message       string `json:"message"         required:"true" desc:"Commit message"`
	BranchName    string `json:"branch_name"     required:"true" desc:"Branch name"`
	NewBranchName string `json:"new_branch_name" desc:"New branch name"`
}

type UpdateFileArgs struct {
	Owner         string `json:"owner"           required:"true" desc:"Repository owner"`
	Repo          string `json:"repo"            required:"true" desc:"Repository name"`
	FilePath      string `json:"filePath"        required:"true" desc:"File path"`
	Content       string `json:"content"         required:"true" desc:"Content (plain text, base64-encoded automatically)"`
	Message       string `json:"message"         required:"true" desc:"Commit message"`
	BranchName    string `json:"branch_name"     required:"true" desc:"Branch name"`
	SHA           string `json:"sha"             required:"true" desc:"File SHA"`
	NewBranchName string `json:"new_branch_name" desc:"New branch name"`
}

type DeleteFileArgs struct {
	Owner         string `json:"owner"           required:"true" desc:"Repository owner"`
	Repo          string `json:"repo"            required:"true" desc:"Repository name"`
	FilePath      string `json:"filePath"        required:"true" desc:"File path"`
	Message       string `json:"message"         required:"true" desc:"Commit message"`
	BranchName    string `json:"branch_name"     required:"true" desc:"Branch name"`
	SHA           string `json:"sha"             required:"true" desc:"File SHA"`
	NewBranchName string `json:"new_branch_name" desc:"New branch name"`
}

type CreateBranchArgs struct {
	Owner     string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"       required:"true" desc:"Repository name"`
	Branch    string `json:"branch"     required:"true" desc:"Branch name"`
	OldBranch string `json:"old_branch" required:"true" desc:"Source branch"`
}

type DeleteBranchArgs struct {
	Owner  string `json:"owner"  required:"true" desc:"Repository owner"`
	Repo   string `json:"repo"   required:"true" desc:"Repository name"`
	Branch string `json:"branch" required:"true" desc:"Branch name"`
}

type ListBranchesArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"100"`
}

type ListRepoCommitsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Path  string `json:"path"  desc:"File/dir path"`
	SHA   string `json:"sha"   desc:"SHA/branch to start from"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"100"`
}

type ListRepoContentsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Ref   string `json:"ref"   required:"true" desc:"Ref (branch/tag/commit)"`
	Path  string `json:"path"  required:"true" desc:"Directory path within the repository (empty string lists the repository root)"`
}

type GetRepoTreeArgs struct {
	Owner     string `json:"owner"     required:"true" desc:"Repository owner"`
	Repo      string `json:"repo"      required:"true" desc:"Repository name"`
	Ref       string `json:"ref"       required:"true" desc:"Ref (branch/tag/commit)"`
	Recursive bool   `json:"recursive" desc:"Return the complete file tree in one response (subject to server cap); default false returns top-level entries only"`
	Page      int    `json:"page"      required:"true" desc:"Page number (1-based)" default:"1"`
	Limit     int    `json:"limit"     required:"true" desc:"Page size"             default:"1000"`
}

func repoTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "create_repo",
			Group:       "repo",
			CLIName:     "create",
			Description: "Create repo",
			Method:      "POST",
			PathTmpl:    "/user/repos",
			Args:        CreateRepoArgs{},
		},
		{
			Name:        "list_my_repos",
			Group:       "repo",
			CLIName:     "list",
			Description: "List my repos",
			Method:      "GET",
			PathTmpl:    "/user/repos",
			Args:        ListMyReposArgs{},
		},
		{
			Name:        "get_file_content",
			Group:       "repo",
			CLIName:     "cat",
			Description: getFileContentDesc,
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/raw/{{.FilePath}}",
			Args:        GetFileContentArgs{},
		},
		{
			Name:        "create_file",
			Group:       "repo",
			CLIName:     "create-file",
			Description: "Create file",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/contents/{{.FilePath}}",
			Args:        CreateFileArgs{},
		},
		{
			Name:        "update_file",
			Group:       "repo",
			CLIName:     "update-file",
			Description: "Update file",
			Method:      "PUT",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/contents/{{.FilePath}}",
			Args:        UpdateFileArgs{},
		},
		{
			Name:        "delete_file",
			Group:       "repo",
			CLIName:     "delete-file",
			Description: "Delete file",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/contents/{{.FilePath}}",
			Args:        DeleteFileArgs{},
		},
		{
			Name:        "create_branch",
			Group:       "repo",
			CLIName:     "create-branch",
			Description: "Create branch",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branches",
			Args:        CreateBranchArgs{},
		},
		{
			Name:        "delete_branch",
			Group:       "repo",
			CLIName:     "delete-branch",
			Description: "Delete branch",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branches/{{.Branch}}",
			Args:        DeleteBranchArgs{},
		},
		{
			Name:        "list_branches",
			Group:       "repo",
			CLIName:     "branches",
			Description: "List branches",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/branches",
			Args:        ListBranchesArgs{},
		},
		{
			Name:        "list_repo_commits",
			Group:       "repo",
			CLIName:     "log",
			Description: "List repo commits",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/commits",
			Args:        ListRepoCommitsArgs{},
		},
		{
			Name:        "list_repo_contents",
			Group:       "repo",
			CLIName:     "ls",
			Description: listRepoContentsDesc,
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/contents/{{.Path}}",
			Args:        ListRepoContentsArgs{},
		},
		{
			Name:        "get_repo_tree",
			Group:       "repo",
			CLIName:     "tree",
			Description: getRepoTreeDesc,
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/git/trees/{{.Ref}}",
			Args:        GetRepoTreeArgs{},
		},
	}
}
