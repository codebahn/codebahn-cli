package tools

type ListReleasesArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"30"`
}

type CreateReleaseArgs struct {
	Owner      string `json:"owner"       required:"true" desc:"Repository owner"`
	Repo       string `json:"repo"        required:"true" desc:"Repository name"`
	TagName    string `json:"tag_name"    required:"true" desc:"Tag name for the release"`
	Target     string `json:"target"      desc:"Target branch or commit SHA (defaults to the repo default branch)"`
	Title      string `json:"title"       desc:"Release title"`
	Note       string `json:"note"        desc:"Release notes body"`
	Draft      bool   `json:"draft"       desc:"Create as a draft release"`
	Prerelease bool   `json:"prerelease"  desc:"Mark as a prerelease"`
}

type GetReleaseArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	ID    int    `json:"id"    required:"true" desc:"Release ID"`
}

type GetLatestReleaseArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
}

type GetReleaseByTagArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Tag   string `json:"tag"   required:"true" desc:"Tag name"`
}

type UpdateReleaseArgs struct {
	Owner      string `json:"owner"      required:"true" desc:"Repository owner"`
	Repo       string `json:"repo"       required:"true" desc:"Repository name"`
	ID         int    `json:"id"         required:"true" desc:"Release ID"`
	TagName    string `json:"tag_name"   desc:"Tag name for the release"`
	Target     string `json:"target"     desc:"Target branch or commit SHA"`
	Title      string `json:"title"      desc:"Release title"`
	Note       string `json:"note"       desc:"Release notes body"`
	Draft      bool   `json:"draft"      desc:"Mark as a draft release"`
	Prerelease bool   `json:"prerelease" desc:"Mark as a prerelease"`
}

type DeleteReleaseArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	ID    int    `json:"id"    required:"true" desc:"Release ID"`
}

type ListReleaseAttachmentsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	ID    int    `json:"id"    required:"true" desc:"Release ID"`
}

type DeleteReleaseAttachmentArgs struct {
	Owner        string `json:"owner"         required:"true" desc:"Repository owner"`
	Repo         string `json:"repo"          required:"true" desc:"Repository name"`
	ID           int    `json:"id"            required:"true" desc:"Release ID"`
	AttachmentID int    `json:"attachment_id" required:"true" desc:"Attachment (asset) ID"`
}

type ListTagsArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Page  int    `json:"page"  required:"true" desc:"Page number (1-based)" default:"1"`
	Limit int    `json:"limit" required:"true" desc:"Page size"             default:"30"`
}

type CreateTagArgs struct {
	Owner   string `json:"owner"   required:"true" desc:"Repository owner"`
	Repo    string `json:"repo"    required:"true" desc:"Repository name"`
	TagName string `json:"tag_name" required:"true" desc:"Tag name"`
	Target  string `json:"target"  desc:"Target branch or commit SHA (defaults to the repo default branch)"`
	Message string `json:"message" desc:"Tag message"`
}

type DeleteTagArgs struct {
	Owner string `json:"owner" required:"true" desc:"Repository owner"`
	Repo  string `json:"repo"  required:"true" desc:"Repository name"`
	Tag   string `json:"tag"   required:"true" desc:"Tag name"`
}

func releaseTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "list_releases",
			Group:       "release",
			CLIName:     "list",
			Description: "List releases for a repository",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases",
			Args:        ListReleasesArgs{},
		},
		{
			Name:        "create_release",
			Group:       "release",
			CLIName:     "create",
			Description: "Create a release",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases",
			Args:        CreateReleaseArgs{},
		},
		{
			Name:        "get_release",
			Group:       "release",
			CLIName:     "get",
			Description: "Get a release by ID",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}",
			Args:        GetReleaseArgs{},
		},
		{
			Name:        "get_latest_release",
			Group:       "release",
			CLIName:     "get-latest",
			Description: "Get the latest release",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases/latest",
			Args:        GetLatestReleaseArgs{},
		},
		{
			Name:        "get_release_by_tag",
			Group:       "release",
			CLIName:     "get-by-tag",
			Description: "Get a release by tag name",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases/tags/{{.Tag}}",
			Args:        GetReleaseByTagArgs{},
		},
		{
			Name:        "update_release",
			Group:       "release",
			CLIName:     "update",
			Description: "Update a release",
			Method:      "PATCH",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}",
			Args:        UpdateReleaseArgs{},
		},
		{
			Name:        "delete_release",
			Group:       "release",
			CLIName:     "delete",
			Description: "Delete a release",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}",
			Args:        DeleteReleaseArgs{},
			Destructive: true,
		},
		{
			Name:        "list_release_attachments",
			Group:       "release",
			CLIName:     "list-attachments",
			Description: "List attachments (assets) of a release",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}/assets",
			Args:        ListReleaseAttachmentsArgs{},
		},
		{
			Name:        "delete_release_attachment",
			Group:       "release",
			CLIName:     "delete-attachment",
			Description: "Delete a release attachment (asset)",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/releases/{{.ID}}/assets/{{.AttachmentID}}",
			Args:        DeleteReleaseAttachmentArgs{},
			Destructive: true,
		},
		{
			Name:        "list_tags",
			Group:       "release",
			CLIName:     "list-tags",
			Description: "List tags for a repository",
			Method:      "GET",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/tags",
			Args:        ListTagsArgs{},
		},
		{
			Name:        "create_tag",
			Group:       "release",
			CLIName:     "create-tag",
			Description: "Create a tag",
			Method:      "POST",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/tags",
			Args:        CreateTagArgs{},
		},
		{
			Name:        "delete_tag",
			Group:       "release",
			CLIName:     "delete-tag",
			Description: "Delete a tag",
			Method:      "DELETE",
			PathTmpl:    "/repos/{{.Owner}}/{{.Repo}}/tags/{{.Tag}}",
			Args:        DeleteTagArgs{},
			Destructive: true,
		},
	}
}
