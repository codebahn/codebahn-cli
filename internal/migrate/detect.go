package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var validSourceTypes = map[string]string{
	"forgejo": "gitea",
	"gitea":   "gitea",
	"github":  "github",
	"gitlab":  "gitlab",
	"git":     "git",
}

func ValidateSourceType(sourceType string) error {
	if sourceType == "" {
		return nil
	}
	lower := strings.ToLower(sourceType)
	if _, ok := validSourceTypes[lower]; !ok {
		return fmt.Errorf("unknown --source-type %q; valid values: forgejo, gitea, github, gitlab, git", sourceType)
	}
	return nil
}

func ResolveServiceType(ctx context.Context, src *Source, sourceType, tokenFlag string) {
	if sourceType != "" {
		src.Service = validSourceTypes[strings.ToLower(sourceType)]
		return
	}
	if src.Service == "git" {
		token := ResolveSourceToken(tokenFlag)
		if detected := DetectService(ctx, src.Scheme, src.Host, token); detected != "" {
			src.Service = detected
		}
	}
}

var detectClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func DetectService(ctx context.Context, scheme, host, token string) string {
	endpoint := fmt.Sprintf("%s://%s/api/v1/version", scheme, host)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ""
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := detectClient.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return ""
	}

	var payload struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(body, &payload) != nil || payload.Version == "" {
		return ""
	}

	return "gitea"
}
