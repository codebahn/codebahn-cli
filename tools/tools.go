package tools

// ToolDef ties a tool's arg struct to its metadata. The Args field holds a
// zero-value struct whose fields (via reflect) define the tool's parameters.
// Struct tags: json (wire name), required ("true"), desc (description),
// default (default value).
type ToolDef struct {
	Name        string // "create_issue"
	Group       string // "issue" (cobra parent command)
	CLIName     string // "create" (subcommand name under group)
	Description string // tool description for MCP and CLI help
	Method      string // "POST"
	PathTmpl    string // "/repos/{{.Owner}}/{{.Repo}}/issues"
	Args        any    // zero-value of the args struct (for reflection)
}

// All is the single registry of every tool. Both the CLI binary and the
// embedded MCP endpoint derive their tool definitions from this slice.
var All []ToolDef

// ByName returns the ToolDef with the given name, or panics if not found.
func ByName(name string) ToolDef {
	for _, td := range All {
		if td.Name == name {
			return td
		}
	}
	panic("tools: unknown tool " + name)
}

// ByGroup returns all ToolDefs in the given group.
func ByGroup(group string) []ToolDef {
	var out []ToolDef
	for _, td := range All {
		if td.Group == group {
			out = append(out, td)
		}
	}
	return out
}

// Groups returns the unique group names in registration order.
func Groups() []string {
	seen := map[string]bool{}
	var out []string
	for _, td := range All {
		if !seen[td.Group] {
			seen[td.Group] = true
			out = append(out, td.Group)
		}
	}
	return out
}

func init() {
	All = buildRegistry()
}
