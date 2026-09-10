package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func init() {
	Register("list_my_repos", fmtListRepos)
	Register("list_branches", fmtListBranches)
	Register("list_repo_commits", fmtListCommits)
	Register("compare_refs", fmtCompareRefs)
	Register("list_repo_contents", fmtListContents)
	Register("get_repo_tree", fmtGetTree)
	Register("create_repo", fmtCreateRepo)
	Register("create_file", fmtCreateFile)
	Register("update_file", fmtUpdateFile)
	Register("delete_file", fmtDeleteFile)
	Register("create_branch", fmtCreateBranch)
	Register("delete_branch", fmtDeleteBranch)
}

func fmtListRepos(raw json.RawMessage, _ any, p *output.Printer) error {
	var repos []struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		Fork        bool   `json:"fork"`
		Archived    bool   `json:"archived"`
		UpdatedAt   string `json:"updated_at"`
	}
	if err := json.Unmarshal(raw, &repos); err != nil {
		return err
	}

	var rows [][]string
	for _, r := range repos {
		var info []string
		if r.Private {
			info = append(info, "private")
		} else {
			info = append(info, "public")
		}
		if r.Fork {
			info = append(info, "fork")
		}
		if r.Archived {
			info = append(info, "archived")
		}
		infoStr := strings.Join(info, ", ")
		if r.Private {
			infoStr = output.Yellow(infoStr)
		} else {
			infoStr = output.Dim(infoStr)
		}

		rows = append(rows, []string{
			output.Bold(r.FullName),
			output.Truncate(RemoveExcessiveWhitespace(r.Description), 50),
			infoStr,
			output.Dim(TimeAgo(r.UpdatedAt)),
		})
	}
	p.Table(nil, rows)
	return nil
}

func fmtListBranches(raw json.RawMessage, _ any, p *output.Printer) error {
	var branches []struct {
		Name   string `json:"name"`
		Commit struct {
			ID string `json:"id"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(raw, &branches); err != nil {
		return err
	}

	var rows [][]string
	for _, b := range branches {
		sha := b.Commit.ID
		if len(sha) > 7 {
			sha = sha[:7]
		}
		rows = append(rows, []string{
			output.Bold(b.Name),
			output.Dim(sha),
		})
	}
	p.Table(nil, rows)
	return nil
}

// commit is the subset of the API commit object the human views use.
type commit struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	Stats struct {
		Additions int `json:"additions"`
		Deletions int `json:"deletions"`
	} `json:"stats"`
}

func commitRows(commits []commit) [][]string {
	rows := make([][]string, 0, len(commits))
	for _, c := range commits {
		sha := c.SHA
		if len(sha) > 7 {
			sha = sha[:7]
		}
		rows = append(rows, []string{
			output.Yellow(sha),
			FirstLine(c.Commit.Message),
			c.Commit.Author.Name,
			output.Dim(TimeAgo(c.Commit.Author.Date)),
		})
	}
	return rows
}

func fmtListCommits(raw json.RawMessage, _ any, p *output.Printer) error {
	var commits []commit
	if err := json.Unmarshal(raw, &commits); err != nil {
		return err
	}
	p.Table(nil, commitRows(commits))
	return nil
}

func fmtCompareRefs(raw json.RawMessage, args any, p *output.Printer) error {
	a, ok := args.(*tools.CompareRefsArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for compare_refs")
	}
	var cmp struct {
		TotalCommits int      `json:"total_commits"`
		Commits      []commit `json:"commits"`
		Files        []struct {
			Filename string `json:"filename"`
			Status   string `json:"status"`
		} `json:"files"`
	}
	if err := json.Unmarshal(raw, &cmp); err != nil {
		return err
	}

	// The API concatenates every commit's affected files; a file touched by
	// several commits appears several times. Keep the first status seen.
	seen := make(map[string]bool, len(cmp.Files))
	var fileRows [][]string
	for _, f := range cmp.Files {
		if seen[f.Filename] {
			continue
		}
		seen[f.Filename] = true
		fileRows = append(fileRows, []string{fileStatus(f.Status), f.Filename})
	}

	additions, deletions := 0, 0
	for _, c := range cmp.Commits {
		additions += c.Stats.Additions
		deletions += c.Stats.Deletions
	}

	p.Text(fmt.Sprintf("%s  %d %s, %d %s, %s %s",
		output.Bold(a.Base+"..."+a.Head),
		cmp.TotalCommits, plural(cmp.TotalCommits, "commit", "commits"),
		len(fileRows), plural(len(fileRows), "file changed", "files changed"),
		output.Green(fmt.Sprintf("+%d", additions)),
		output.Red(fmt.Sprintf("-%d", deletions))))

	if len(cmp.Commits) > 0 {
		p.Text("")
		p.Table(nil, commitRows(cmp.Commits))
	}
	if len(fileRows) > 0 {
		p.Text("")
		p.Table(nil, fileRows)
	}
	return nil
}

// fileStatus colors a changed-file status the way the PR files view does.
func fileStatus(status string) string {
	switch status {
	case "added":
		return output.Green(status)
	case "modified", "renamed":
		return output.Yellow(status)
	case "deleted":
		return output.Red(status)
	default:
		return status
	}
}

func fmtListContents(raw json.RawMessage, _ any, p *output.Printer) error {
	var entries []struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	if err := json.Unmarshal(raw, &entries); err != nil {
		return err
	}

	var rows [][]string
	for _, e := range entries {
		name := e.Name
		sizeStr := ""
		if e.Type == "dir" {
			name = output.Bold(name + "/")
		} else {
			sizeStr = output.Dim(humanSize(e.Size))
		}
		rows = append(rows, []string{
			output.Dim(e.Type),
			name,
			sizeStr,
		})
	}
	p.Table(nil, rows)
	return nil
}

func fmtGetTree(raw json.RawMessage, _ any, p *output.Printer) error {
	var tree struct {
		Tree []struct {
			Path string `json:"path"`
			Type string `json:"type"`
			Size int64  `json:"size"`
		} `json:"tree"`
	}
	if err := json.Unmarshal(raw, &tree); err != nil {
		return err
	}

	var rows [][]string
	for _, e := range tree.Tree {
		typeStr := "file"
		sizeStr := output.Dim(humanSize(e.Size))
		if e.Type == "tree" {
			typeStr = "dir"
			sizeStr = ""
		}
		rows = append(rows, []string{
			output.Dim(typeStr),
			e.Path,
			sizeStr,
		})
	}
	p.Table(nil, rows)
	return nil
}

func fmtCreateRepo(raw json.RawMessage, _ any, p *output.Printer) error {
	var repo struct {
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"`
	}
	if err := json.Unmarshal(raw, &repo); err != nil {
		return err
	}
	successf("Created repository %s", repo.FullName)
	p.Text(repo.HTMLURL)
	return nil
}

func fmtCreateFile(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.CreateFileArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for create_file")
	}
	successf("Created %s/%s:%s on %s", a.Owner, a.Repo, a.FilePath, a.BranchName)
	return nil
}

func fmtUpdateFile(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.UpdateFileArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for update_file")
	}
	successf("Updated %s/%s:%s on %s", a.Owner, a.Repo, a.FilePath, a.BranchName)
	return nil
}

func fmtDeleteFile(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DeleteFileArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for delete_file")
	}
	successf("Deleted %s/%s:%s on %s", a.Owner, a.Repo, a.FilePath, a.BranchName)
	return nil
}

func fmtCreateBranch(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.CreateBranchArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for create_branch")
	}
	successf("Created branch %s from %s in %s/%s", a.Branch, a.OldBranch, a.Owner, a.Repo)
	return nil
}

func fmtDeleteBranch(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DeleteBranchArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for delete_branch")
	}
	successf("Deleted branch %s in %s/%s", a.Branch, a.Owner, a.Repo)
	return nil
}

func humanSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GiB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MiB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KiB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
