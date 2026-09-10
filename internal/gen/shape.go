package gen

import (
	"encoding/json"
	"fmt"

	"github.com/codebahn/codebahn-cli/tools"
	"github.com/codebahn/codebahn-cli/tools/unidiff"
)

// shaper applies the parameters the REST API does not understand (struct
// fields tagged api:"-") to a raw response. The MCP endpoint does the same
// work in its handlers; the CLI does it here so both surfaces behave alike.
type shaper func(raw json.RawMessage, args any) (json.RawMessage, error)

var shapers = map[string]shaper{
	"get_pull_request_diff": shapeDiff,
	"get_commit_diff":       shapeDiff,
}

func hasShaper(toolName string) bool {
	_, ok := shapers[toolName]
	return ok
}

// shapeResponse runs the tool's shaper, if any. Tools without one pass through.
func shapeResponse(td tools.ToolDef, raw json.RawMessage, args any) (json.RawMessage, error) {
	s, ok := shapers[td.Name]
	if !ok {
		return raw, nil
	}
	return s(raw, args)
}

// shapeDiff narrows a unified diff to the file named by file_path.
func shapeDiff(raw json.RawMessage, args any) (json.RawMessage, error) {
	var filePath string
	switch a := args.(type) {
	case *tools.GetPullRequestDiffArgs:
		filePath = a.FilePath
	case *tools.GetCommitDiffArgs:
		filePath = a.FilePath
	default:
		return nil, fmt.Errorf("unexpected args type %T for diff tool", args)
	}
	if filePath == "" {
		return raw, nil
	}
	section, found := unidiff.ExtractFile(string(raw), filePath)
	if !found {
		return nil, fmt.Errorf("file_path %q not found in diff", filePath)
	}
	return json.RawMessage(section + "\n"), nil
}
