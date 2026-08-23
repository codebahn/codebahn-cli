package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func init() {
	Register("list_repo_issues", fmtListIssues)
	Register("get_issue_by_index", fmtGetIssue)
	Register("create_issue", fmtCreateIssue)
	Register("create_issue_comment", fmtCreateIssueComment)
	Register("update_issue", fmtUpdateIssue)
	Register("issue_state_change", fmtIssueStateChange)
	Register("add_issue_labels", fmtAddIssueLabels)
	Register("remove_issue_labels", fmtRemoveIssueLabels)
	Register("list_issue_comments", fmtListIssueComments)
	Register("get_issue_comment", fmtGetIssueComment)
	Register("edit_issue_comment", fmtEditIssueComment)
	Register("delete_issue_comment", fmtDeleteIssueComment)
	Register("list_repo_labels", fmtListLabels)
	Register("list_repo_milestones", fmtListMilestones)
	Register("create_label", fmtCreateLabel)
	Register("edit_label", fmtEditLabel)
	Register("delete_label", fmtDeleteLabel)
}

type issueRow struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	UpdatedAt string `json:"updated_at"`
	Labels    []struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"labels"`
}

func fmtListIssues(raw json.RawMessage, _ any, p *output.Printer) error {
	var issues []issueRow
	if err := json.Unmarshal(raw, &issues); err != nil {
		return err
	}
	isTTY := output.IsTTY()
	var rows [][]string
	for _, issue := range issues {
		id := issueIDColored(fmt.Sprintf("#%d", issue.Number), issue.State)

		var labelNames []string
		for _, l := range issue.Labels {
			labelNames = append(labelNames, l.Name)
		}

		row := []string{id}
		if !isTTY {
			row = append(row, issue.State)
		}
		row = append(row,
			RemoveExcessiveWhitespace(issue.Title),
			strings.Join(labelNames, ", "),
			output.Dim(TimeAgo(issue.UpdatedAt)),
		)
		rows = append(rows, row)
	}
	p.Table(nil, rows)
	return nil
}

func issueIDColored(id, state string) string {
	switch state {
	case "open":
		return output.Green(id)
	case "closed":
		return output.Red(id)
	default:
		return id
	}
}

type issueDetail struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Body      string `json:"body"`
	Comments  int    `json:"comments"`
	CreatedAt string `json:"created_at"`
	HTMLURL   string `json:"html_url"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
	Milestone *struct {
		Title string `json:"title"`
	} `json:"milestone"`
	Repository *struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func fmtGetIssue(raw json.RawMessage, _ any, p *output.Printer) error {
	var issue issueDetail
	if err := json.Unmarshal(raw, &issue); err != nil {
		return err
	}

	repoRef := ""
	if issue.Repository != nil {
		repoRef = issue.Repository.FullName
	}

	p.Text(fmt.Sprintf("%s  %s#%d", output.Bold(issue.Title), repoRef, issue.Number))

	commentWord := "comments"
	if issue.Comments == 1 {
		commentWord = "comment"
	}
	p.Text(fmt.Sprintf("%s · %s opened %s · %d %s",
		output.StatusColor(issue.State),
		issue.User.Login,
		TimeAgo(issue.CreatedAt),
		issue.Comments,
		commentWord,
	))

	p.Text("")

	if len(issue.Labels) > 0 {
		var names []string
		for _, l := range issue.Labels {
			names = append(names, l.Name)
		}
		p.Text(fmt.Sprintf("Labels:     %s", strings.Join(names, ", ")))
	}
	if len(issue.Assignees) > 0 {
		var logins []string
		for _, a := range issue.Assignees {
			logins = append(logins, a.Login)
		}
		p.Text(fmt.Sprintf("Assignees:  %s", strings.Join(logins, ", ")))
	}
	if issue.Milestone != nil && issue.Milestone.Title != "" {
		p.Text(fmt.Sprintf("Milestone:  %s", issue.Milestone.Title))
	}

	if issue.Body != "" {
		p.Text("")
		for _, line := range strings.Split(issue.Body, "\n") {
			p.Text("  " + line)
		}
	}

	if issue.HTMLURL != "" {
		p.Text("")
		p.Text("View this issue on Codebahn: " + issue.HTMLURL)
	}
	return nil
}

type mutationResult struct {
	Number     int    `json:"number"`
	ID         int    `json:"id"`
	Title      string `json:"title"`
	State      string `json:"state"`
	Name       string `json:"name"`
	HTMLURL    string `json:"html_url"`
	Repository *struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func fmtCreateIssue(raw json.RawMessage, _ any, p *output.Printer) error {
	var r mutationResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	p.Text(r.HTMLURL)
	return nil
}

func fmtCreateIssueComment(raw json.RawMessage, _ any, p *output.Printer) error {
	var r mutationResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	p.Text(r.HTMLURL)
	return nil
}

func fmtUpdateIssue(raw json.RawMessage, args any, p *output.Printer) error {
	var r mutationResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	repo := ""
	if r.Repository != nil {
		repo = r.Repository.FullName
	} else if a, ok := args.(*tools.UpdateIssueArgs); ok {
		repo = a.Owner + "/" + a.Repo
	}
	successf("Updated issue %s#%d", repo, r.Number)
	return nil
}

func fmtIssueStateChange(raw json.RawMessage, _ any, p *output.Printer) error {
	var r mutationResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	repo := ""
	if r.Repository != nil {
		repo = r.Repository.FullName
	}
	verb := "Reopened"
	if r.State == "closed" {
		verb = "Closed"
	}
	successf("%s issue %s#%d (%s)", verb, repo, r.Number, r.Title)
	return nil
}

func fmtAddIssueLabels(_ json.RawMessage, args any, p *output.Printer) error {
	a, ok := args.(*tools.AddIssueLabelsArgs)
	if !ok {
		return nil
	}
	successf("Added labels to %s/%s#%d", a.Owner, a.Repo, a.Index)
	return nil
}

func fmtRemoveIssueLabels(_ json.RawMessage, args any, p *output.Printer) error {
	a, ok := args.(*tools.RemoveIssueLabelsArgs)
	if !ok {
		return nil
	}
	successf("Removed labels from %s/%s#%d", a.Owner, a.Repo, a.Index)
	return nil
}

type commentRow struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

func fmtListIssueComments(raw json.RawMessage, _ any, p *output.Printer) error {
	var comments []commentRow
	if err := json.Unmarshal(raw, &comments); err != nil {
		return err
	}
	var rows [][]string
	for _, c := range comments {
		rows = append(rows, []string{
			output.Dim(fmt.Sprintf("#%d", c.ID)),
			output.Bold(c.User.Login),
			output.Truncate(FirstLine(c.Body), 50),
			output.Dim(TimeAgo(c.CreatedAt)),
		})
	}
	p.Table(nil, rows)
	return nil
}

func fmtGetIssueComment(raw json.RawMessage, _ any, p *output.Printer) error {
	var c commentRow
	if err := json.Unmarshal(raw, &c); err != nil {
		return err
	}
	p.Text(fmt.Sprintf("Comment #%d", c.ID))
	p.Text(fmt.Sprintf("%s commented %s", c.User.Login, TimeAgo(c.CreatedAt)))
	if c.Body != "" {
		p.Text("")
		for _, line := range strings.Split(c.Body, "\n") {
			p.Text("  " + line)
		}
	}
	return nil
}

func fmtEditIssueComment(raw json.RawMessage, args any, _ *output.Printer) error {
	var r mutationResult
	if raw != nil {
		_ = json.Unmarshal(raw, &r)
	}
	id := r.ID
	if id == 0 {
		if a, ok := args.(*tools.EditIssueCommentArgs); ok {
			id = a.CommentID
		}
	}
	successf("Edited comment #%d", id)
	return nil
}

func fmtDeleteIssueComment(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DeleteIssueCommentArgs)
	if !ok {
		return nil
	}
	successf("Deleted comment #%d", a.CommentID)
	return nil
}

type labelRow struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

func fmtListLabels(raw json.RawMessage, _ any, p *output.Printer) error {
	var labels []labelRow
	if err := json.Unmarshal(raw, &labels); err != nil {
		return err
	}
	var rows [][]string
	for _, l := range labels {
		rows = append(rows, []string{
			l.Name,
			output.Dim("#" + l.Color),
			l.Description,
		})
	}
	p.Table(nil, rows)
	return nil
}

type milestoneRow struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	State        string `json:"state"`
	OpenIssues   int    `json:"open_issues"`
	ClosedIssues int    `json:"closed_issues"`
}

func fmtListMilestones(raw json.RawMessage, _ any, p *output.Printer) error {
	var milestones []milestoneRow
	if err := json.Unmarshal(raw, &milestones); err != nil {
		return err
	}
	var rows [][]string
	for _, m := range milestones {
		rows = append(rows, []string{
			output.Bold(m.Title),
			output.StatusColor(m.State),
			fmt.Sprintf("%d open", m.OpenIssues),
			fmt.Sprintf("%d closed", m.ClosedIssues),
		})
	}
	p.Table(nil, rows)
	return nil
}

func fmtCreateLabel(raw json.RawMessage, _ any, _ *output.Printer) error {
	var r mutationResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	successf("Created label %q", r.Name)
	return nil
}

func fmtEditLabel(raw json.RawMessage, _ any, _ *output.Printer) error {
	var r mutationResult
	if err := json.Unmarshal(raw, &r); err != nil {
		return err
	}
	successf("Updated label %q", r.Name)
	return nil
}

func fmtDeleteLabel(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DeleteLabelArgs)
	if !ok {
		return nil
	}
	successf("Deleted label #%d", a.ID)
	return nil
}
