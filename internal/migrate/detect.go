package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func ResolveServiceType(ctx context.Context, src *Source, sourceType, tokenFlag string) {
	if sourceType != "" {
		normalized := sourceType
		if normalized == "forgejo" {
			normalized = "gitea"
		}
		src.Service = normalized
		return
	}
	if src.Service == "git" {
		token := ResolveSourceToken(tokenFlag)
		if detected := DetectService(ctx, src.Scheme, src.Host, token); detected != "" {
			src.Service = detected
		}
	}
}

func DetectService(ctx context.Context, scheme, host, token string) string {
	url := fmt.Sprintf("%s://%s/api/v1/version", scheme, host)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
