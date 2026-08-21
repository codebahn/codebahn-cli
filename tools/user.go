package tools

type GetMyUserInfoArgs struct{}

func userTools() []ToolDef {
	return []ToolDef{
		{
			Name:        "get_my_user_info",
			Group:       "user",
			CLIName:     "info",
			Description: "Get current user info",
			Method:      "GET",
			PathTmpl:    "/user",
			Args:        GetMyUserInfoArgs{},
		},
	}
}
