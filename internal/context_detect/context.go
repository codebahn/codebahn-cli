package context_detect

import (
	"net/url"
	"os/exec"
	"strings"
)

type RepoContext struct {
	Owner string
	Repo  string
}

func DetectRepo(instanceURL string) (RepoContext, bool) {
	for _, remote := range gitRemoteURLs() {
		if ctx, ok := parseRemote(remote, instanceURL); ok {
			return ctx, true
		}
	}
	return RepoContext{}, false
}

func gitRemoteURLs() []string {
	out, err := exec.Command("git", "remote").Output()
	if err != nil {
		return nil
	}
	var urls []string
	for _, name := range strings.Fields(string(out)) {
		u, err := exec.Command("git", "remote", "get-url", name).Output()
		if err == nil {
			urls = append(urls, strings.TrimSpace(string(u)))
		}
	}
	return urls
}

func parseRemote(remote, instanceURL string) (RepoContext, bool) {
	instanceHost := hostFromURL(instanceURL)
	if instanceHost == "" {
		return RepoContext{}, false
	}

	if strings.HasPrefix(remote, "git@") {
		after, ok := strings.CutPrefix(remote, "git@")
		if !ok {
			return RepoContext{}, false
		}
		colonIdx := strings.Index(after, ":")
		if colonIdx < 0 {
			return RepoContext{}, false
		}
		host := after[:colonIdx]
		if !strings.EqualFold(host, instanceHost) {
			return RepoContext{}, false
		}
		path := after[colonIdx+1:]
		return ownerRepoFromPath(path)
	}

	u, err := url.Parse(remote)
	if err != nil {
		return RepoContext{}, false
	}
	if !strings.EqualFold(u.Hostname(), instanceHost) {
		return RepoContext{}, false
	}
	return ownerRepoFromPath(u.Path)
}

func ownerRepoFromPath(path string) (RepoContext, bool) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, ".git")
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return RepoContext{}, false
	}
	return RepoContext{Owner: parts[0], Repo: parts[1]}, true
}

func hostFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
