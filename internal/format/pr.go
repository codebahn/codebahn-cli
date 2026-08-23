package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codebahn/codebahn-cli/internal/output"
	"github.com/codebahn/codebahn-cli/tools"
)

func init() {
	Register("list_repo_pull_requests", fmtPRList)
	Register("get_pull_request_by_index", fmtPRGet)
	Register("create_pull_request", fmtPRCreate)
	Register("update_pull_request", fmtPRUpdate)
	Register("merge_pull_request", fmtPRMerge)
	Register("list_pull_request_files", fmtPRFiles)
	Register("list_pull_reviews", fmtPRReviews)
	Register("get_pull_review", fmtPRGetReview)
	Register("list_pull_review_comments", fmtPRReviewComments)
	Register("create_pull_review", fmtPRCreateReview)
	Register("submit_pull_review", fmtPRSubmitReview)
	Register("dismiss_pull_review", fmtPRDismissReview)
	Register("delete_pull_review", fmtPRDeleteReview)
	Register("create_review_requests", fmtPRRequestReview)
	Register("delete_review_requests", fmtPRCancelReview)
}

type prRow struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Head      ref    `json:"head"`
	CreatedAt string `json:"created_at"`
	Merged    bool   `json:"merged"`
}

type ref struct {
	Ref string `json:"ref"`
}

func prState(state string, merged bool) string {
	if merged {
		return "merged"
	}
	return state
}

func prStateColor(state string) string {
	switch state {
	case "open":
		return output.Green(state)
	case "closed":
		return output.Red(state)
	case "merged":
		return output.Magenta(state)
	default:
		return state
	}
}

func fmtPRList(raw json.RawMessage, _ any, p *output.Printer) error {
	var prs []prRow
	if err := json.Unmarshal(raw, &prs); err != nil {
		return err
	}
	isTTY := output.IsTTY()
	var rows [][]string
	for _, pr := range prs {
		state := prState(pr.State, pr.Merged)
		id := colorPRID(pr.Number, state)
		title := RemoveExcessiveWhitespace(pr.Title)
		branch := output.Cyan(pr.Head.Ref)
		created := output.Dim(TimeAgo(pr.CreatedAt))
		var row []string
		if isTTY {
			row = []string{id, title, branch, created}
		} else {
			row = []string{id, title, state, branch, created}
		}
		rows = append(rows, row)
	}
	p.Table(nil, rows)
	return nil
}

func colorPRID(number int, state string) string {
	id := fmt.Sprintf("#%d", number)
	switch state {
	case "open":
		return output.Green(id)
	case "closed":
		return output.Red(id)
	case "merged":
		return output.Magenta(id)
	default:
		return id
	}
}

type prDetail struct {
	Number     int          `json:"number"`
	Title      string       `json:"title"`
	State      string       `json:"state"`
	Body       string       `json:"body"`
	User       userRef      `json:"user"`
	Head       ref          `json:"head"`
	Base       ref          `json:"base"`
	Labels     []labelRef   `json:"labels"`
	Assignees  []userRef    `json:"assignees"`
	Milestone  milestoneRef `json:"milestone"`
	Additions  int          `json:"additions"`
	Deletions  int          `json:"deletions"`
	Merged     bool         `json:"merged"`
	CreatedAt  string       `json:"created_at"`
	HTMLURL    string       `json:"html_url"`
	Repository repoRef      `json:"repository"`
}

type userRef struct {
	Login string `json:"login"`
}

type labelRef struct {
	Name string `json:"name"`
}

type milestoneRef struct {
	Title string `json:"title"`
}

type repoRef struct {
	FullName string `json:"full_name"`
}

func fmtPRGet(raw json.RawMessage, _ any, p *output.Printer) error {
	var pr prDetail
	if err := json.Unmarshal(raw, &pr); err != nil {
		return err
	}
	state := prState(pr.State, pr.Merged)
	repoName := pr.Repository.FullName

	p.Text(fmt.Sprintf("%s  %s#%d", output.Bold(pr.Title), repoName, pr.Number))
	p.Text(fmt.Sprintf("%s · %s wants to merge into %s from %s · %s",
		prStateColor(state), pr.User.Login, pr.Base.Ref, pr.Head.Ref, TimeAgo(pr.CreatedAt)))
	p.Text(fmt.Sprintf("%s %s", output.Green(fmt.Sprintf("+%d", pr.Additions)), output.Red(fmt.Sprintf("-%d", pr.Deletions))))

	if len(pr.Assignees) > 0 {
		names := make([]string, len(pr.Assignees))
		for i, a := range pr.Assignees {
			names[i] = a.Login
		}
		p.Text(fmt.Sprintf("\nAssignees:  %s", strings.Join(names, ", ")))
	}
	if len(pr.Labels) > 0 {
		names := make([]string, len(pr.Labels))
		for i, l := range pr.Labels {
			names[i] = l.Name
		}
		if len(pr.Assignees) == 0 {
			p.Text("")
		}
		p.Text(fmt.Sprintf("Labels:     %s", strings.Join(names, ", ")))
	}
	if pr.Milestone.Title != "" {
		if len(pr.Assignees) == 0 && len(pr.Labels) == 0 {
			p.Text("")
		}
		p.Text(fmt.Sprintf("Milestone:  %s", pr.Milestone.Title))
	}

	if pr.Body != "" {
		p.Text("")
		for _, line := range strings.Split(pr.Body, "\n") {
			p.Text("  " + line)
		}
	}

	if pr.HTMLURL != "" {
		p.Text(fmt.Sprintf("\nView this pull request on Codebahn: %s", pr.HTMLURL))
	}
	return nil
}

func fmtPRCreate(raw json.RawMessage, _ any, p *output.Printer) error {
	var pr struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(raw, &pr); err != nil {
		return err
	}
	p.Text(pr.HTMLURL)
	return nil
}

func fmtPRUpdate(raw json.RawMessage, args any, p *output.Printer) error {
	var pr struct {
		Number     int     `json:"number"`
		Repository repoRef `json:"repository"`
	}
	if err := json.Unmarshal(raw, &pr); err != nil {
		return err
	}
	repo := pr.Repository.FullName
	if repo == "" {
		if a, ok := args.(*tools.UpdatePullRequestArgs); ok {
			repo = a.Owner + "/" + a.Repo
		}
	}
	successf("Updated pull request %s#%d", repo, pr.Number)
	return nil
}

func fmtPRMerge(_ json.RawMessage, args any, p *output.Printer) error {
	a, ok := args.(*tools.MergePullRequestArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for merge_pull_request")
	}
	verb := "Merged"
	switch a.Style {
	case "rebase", "rebase-merge":
		verb = "Rebased and merged"
	case "squash":
		verb = "Squashed and merged"
	}
	successf("%s pull request %s/%s#%d", verb, a.Owner, a.Repo, a.Index)
	return nil
}

func fmtPRFiles(raw json.RawMessage, _ any, p *output.Printer) error {
	var files []struct {
		Filename  string `json:"filename"`
		Status    string `json:"status"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
	}
	if err := json.Unmarshal(raw, &files); err != nil {
		return err
	}
	var rows [][]string
	for _, f := range files {
		status := f.Status
		switch status {
		case "added":
			status = output.Green(status)
		case "modified", "renamed":
			status = output.Yellow(status)
		case "deleted":
			status = output.Red(status)
		}
		add := output.Green(fmt.Sprintf("+%d", f.Additions))
		del := output.Red(fmt.Sprintf("-%d", f.Deletions))
		rows = append(rows, []string{status, f.Filename, add, del})
	}
	p.Table(nil, rows)
	return nil
}

func reviewStateColor(state string) string {
	s := strings.ToLower(state)
	switch s {
	case "approved":
		return output.Green(s)
	case "request_changes", "rejected":
		return output.Red(s)
	case "requested":
		return output.Yellow(s)
	default:
		return s
	}
}

func fmtPRReviews(raw json.RawMessage, _ any, p *output.Printer) error {
	var reviews []struct {
		ID          int     `json:"id"`
		User        userRef `json:"user"`
		State       string  `json:"state"`
		Body        string  `json:"body"`
		SubmittedAt string  `json:"submitted_at"`
	}
	if err := json.Unmarshal(raw, &reviews); err != nil {
		return err
	}
	var rows [][]string
	for _, r := range reviews {
		state := reviewStateColor(r.State)
		body := output.Truncate(FirstLine(r.Body), 50)
		date := output.Dim(TimeAgo(r.SubmittedAt))
		rows = append(rows, []string{r.User.Login, state, body, date})
	}
	p.Table(nil, rows)
	return nil
}

func fmtPRGetReview(raw json.RawMessage, _ any, p *output.Printer) error {
	var review struct {
		ID          int     `json:"id"`
		User        userRef `json:"user"`
		State       string  `json:"state"`
		Body        string  `json:"body"`
		SubmittedAt string  `json:"submitted_at"`
	}
	if err := json.Unmarshal(raw, &review); err != nil {
		return err
	}
	state := reviewStateColor(review.State)
	p.Text(fmt.Sprintf("Review by %s", review.User.Login))
	p.Text(fmt.Sprintf("%s · %s", state, TimeAgo(review.SubmittedAt)))
	if review.Body != "" {
		p.Text("")
		for _, line := range strings.Split(review.Body, "\n") {
			p.Text("  " + line)
		}
	}
	return nil
}

func fmtPRReviewComments(raw json.RawMessage, _ any, p *output.Printer) error {
	var comments []struct {
		Path      string  `json:"path"`
		Line      int     `json:"line"`
		Body      string  `json:"body"`
		User      userRef `json:"user"`
		CreatedAt string  `json:"created_at"`
	}
	if err := json.Unmarshal(raw, &comments); err != nil {
		return err
	}
	var rows [][]string
	for _, c := range comments {
		loc := fmt.Sprintf("%s:%d", c.Path, c.Line)
		body := output.Truncate(FirstLine(c.Body), 50)
		date := output.Dim(TimeAgo(c.CreatedAt))
		rows = append(rows, []string{loc, body, c.User.Login, date})
	}
	p.Table(nil, rows)
	return nil
}

func fmtPRCreateReview(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.CreatePullReviewArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for create_pull_review")
	}
	successf("Created review on %s/%s#%d", a.Owner, a.Repo, a.Index)
	return nil
}

func fmtPRSubmitReview(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.SubmitPullReviewArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for submit_pull_review")
	}
	successf("Submitted review on %s/%s#%d", a.Owner, a.Repo, a.Index)
	return nil
}

func fmtPRDismissReview(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DismissPullReviewArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for dismiss_pull_review")
	}
	successf("Dismissed review #%d on %s/%s#%d", a.ID, a.Owner, a.Repo, a.Index)
	return nil
}

func fmtPRDeleteReview(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DeletePullReviewArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for delete_pull_review")
	}
	successf("Deleted review #%d on %s/%s#%d", a.ID, a.Owner, a.Repo, a.Index)
	return nil
}

func fmtPRRequestReview(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.CreateReviewRequestsArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for create_review_requests")
	}
	successf("Requested review on %s/%s#%d", a.Owner, a.Repo, a.Index)
	return nil
}

func fmtPRCancelReview(_ json.RawMessage, args any, _ *output.Printer) error {
	a, ok := args.(*tools.DeleteReviewRequestsArgs)
	if !ok {
		return fmt.Errorf("unexpected args type for delete_review_requests")
	}
	successf("Cancelled review request on %s/%s#%d", a.Owner, a.Repo, a.Index)
	return nil
}
