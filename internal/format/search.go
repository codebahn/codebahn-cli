package format

import (
	"encoding/json"
	"fmt"

	"github.com/codebahn/codebahn-cli/internal/output"
)

func init() {
	Register("search_code", formatSearchCode)
	Register("search_repos", formatSearchRepos)
}

func formatSearchCode(raw json.RawMessage, _ any, p *output.Printer) error {
	var resp struct {
		Data []struct {
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			Path string `json:"path"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return err
	}
	if len(resp.Data) == 0 {
		return nil
	}

	for _, item := range resp.Data {
		p.Text(fmt.Sprintf("%s %s", output.Bold(item.Repository.FullName), output.Green(item.Path)))
	}
	return nil
}

func formatSearchRepos(raw json.RawMessage, _ any, p *output.Printer) error {
	var resp struct {
		Data []struct {
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			Private     bool   `json:"private"`
			UpdatedAt   string `json:"updated_at"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return err
	}
	if len(resp.Data) == 0 {
		return nil
	}

	var rows [][]string
	for _, r := range resp.Data {
		vis := output.Dim("public")
		if r.Private {
			vis = output.Yellow("private")
		}
		rows = append(rows, []string{
			output.Bold(r.FullName),
			output.Truncate(RemoveExcessiveWhitespace(r.Description), 50),
			vis,
			output.Dim(TimeAgo(r.UpdatedAt)),
		})
	}
	p.Table(nil, rows)
	return nil
}
