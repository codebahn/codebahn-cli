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
		lower := strings.ToLower(sourceType)
		if mapped, ok := validSourceTypes[lower]; ok {
			src.Service = mapped
		} else {
			src.Service = lower
		}
		return
	}
	if src.Service == "git" {
		token := ResolveSourceToken(tokenFlag)
		if detected := DetectService(ctx, src.Scheme, src.Host, token); detected != "" {
			src.Service = detected
		}
	}
}

const detectTimeout = 5 * time.Second

func DetectService(ctx context.Context, scheme, host, token string) string {
	probeCtx, cancel := context.WithTimeout(ctx, detectTimeout)
	defer cancel()

	url := fmt.Sprintf("%s://%s/api/v1/version", scheme, host)
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
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
