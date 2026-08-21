package tools

type SearchArgs struct {
	Keyword string `json:"keyword" required:"true" desc:"Search keyword"`
}

type SearchCodeArgs struct {
	Keyword  string `json:"keyword"  required:"true" desc:"Search keyword"`
	Language string `json:"language" desc:"Filter by programming language"`
	Filename string `json:"filename" desc:"Filter by filename or path"`
	Mode     string `json:"mode"     desc:"Search mode: exact, union, fuzzy" default:"exact"`
	Page     int    `json:"page"     desc:"Page number (1-based)"           default:"1"`
	Limit    int    `json:"limit"    desc:"Page size"                       default:"30"`
}

type SearchReposArgs struct {
	Keyword string `json:"keyword" required:"true" desc:"Search keyword"`
	Page    int    `json:"page"    desc:"Page number (1-based)" default:"1"`
	Limit   int    `json:"limit"   desc:"Page size"             default:"30"`
}

func searchTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "search",
			Group:       "search",
			CLIName:     "all",
			Description: "Search across all types (code, repos, issues, users, organizations). Returns hit counts and a preview of results for each type. Use search_code or search_repos for paginated results.",
			Method:      "GET",
			PathTmpl:    "/search",
			Args:        SearchArgs{},
		},
		{
			Name:        "search_code",
			Group:       "search",
			CLIName:     "code",
			Description: "Search code across all accessible repositories",
			Method:      "GET",
			PathTmpl:    "/search",
			Args:        SearchCodeArgs{},
		},
		{
			Name:        "search_repos",
			Group:       "search",
			CLIName:     "repos",
			Description: "Search repositories by name or description",
			Method:      "GET",
			PathTmpl:    "/search",
			Args:        SearchReposArgs{},
		},
	}
}
