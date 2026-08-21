package tools

func buildRegistry() []ToolDef {
	var all []ToolDef
	all = append(all, userTools()...)
	all = append(all, repoTools()...)
	all = append(all, issueTools()...)
	all = append(all, prTools()...)
	all = append(all, searchTools()...)
	all = append(all, actionsTools()...)
	return all
}
