package migrate

import (
	"fmt"
	"net/url"
	"strings"
)

type Source struct {
	Service  string
	Host     string
	Owner    string
	Repo     string
	CloneURL string
}

func (s Source) IsOrg() bool { return s.Repo == "" }

func (s Source) SupportsMetadata() bool {
	switch s.Service {
	case "github", "gitlab", "gitea":
		return true
	default:
		return false
	}
}

var hostToService = map[string]string{
	"github.com":   "github",
	"gitlab.com":   "gitlab",
	"codeberg.org": "gitea",
}

func ParseSource(input string) (Source, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Source{}, fmt.Errorf("source is required")
	}

	raw := input
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return Source{}, fmt.Errorf("could not parse source: %q", input)
	}

	if u.Host == "" {
		return Source{}, fmt.Errorf("could not parse source: %q", input)
	}

	host := strings.ToLower(u.Host)

	path := strings.Trim(u.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	if path == "" {
		return Source{}, fmt.Errorf("could not parse source: %q (missing owner)", input)
	}

	parts := strings.Split(path, "/")

	if len(parts) > 2 {
		return Source{}, fmt.Errorf("could not parse source: %q (expected owner/repo or owner)", input)
	}

	service := hostToService[host]
	if service == "" {
		service = "git"
	}

	s := Source{
		Service: service,
		Host:    host,
		Owner:   parts[0],
	}

	if len(parts) == 2 && parts[1] != "" {
		s.Repo = parts[1]
		s.CloneURL = fmt.Sprintf("https://%s/%s/%s.git", host, s.Owner, s.Repo)
	}

	return s, nil
}
