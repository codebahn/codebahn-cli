package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type fetchOpts struct {
	startURL  string
	token     string
	authStyle string // "bearer" or "token"
	accept    string
	label     string // for error messages, e.g. "GitHub" or "Gitea"
}

var noRedirectClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) == 0 {
			return http.ErrUseLastResponse
		}
		if !sameOrigin(via[0].URL.String(), req.URL.String()) {
			return http.ErrUseLastResponse
		}
		return nil
	},
}

func sameOrigin(a, b string) bool {
	ua, err := url.Parse(a)
	if err != nil {
		return false
	}
	ub, err := url.Parse(b)
	if err != nil {
		return false
	}
	return ua.Scheme == ub.Scheme && ua.Host == ub.Host
}

func fetchPages[T any](ctx context.Context, opts fetchOpts) ([]T, error) {
	var all []T
	nextURL := opts.startURL

	for nextURL != "" {
		if nextURL != opts.startURL && !sameOrigin(opts.startURL, nextURL) {
			return nil, fmt.Errorf("%s API: pagination URL has different origin", opts.label)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		if opts.accept != "" {
			req.Header.Set("Accept", opts.accept)
		}
		if opts.token != "" {
			style := opts.authStyle
			if style == "" {
				style = "token"
			}
			switch style {
			case "bearer":
				req.Header.Set("Authorization", "Bearer "+opts.token)
			default:
				req.Header.Set("Authorization", "token "+opts.token)
			}
		}

		resp, err := noRedirectClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%s API request: %w", opts.label, err)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading %s response: %w", opts.label, readErr)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, &httpError{
				StatusCode: resp.StatusCode,
				Body:       string(body),
				Label:      opts.label,
				Headers:    resp.Header,
			}
		}

		var page []T
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("decoding %s response: %w", opts.label, err)
		}

		all = append(all, page...)
		nextURL = parseNextLink(resp.Header.Get("Link"))
	}

	return all, nil
}

type httpError struct {
	StatusCode int
	Body       string
	Label      string
	Headers    http.Header
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%s API: HTTP %d: %s", e.Label, e.StatusCode, e.Body)
}

func sourceRepoURL(scheme, host, pathFmt string, args ...any) string {
	escapedArgs := make([]any, len(args))
	for i, a := range args {
		escapedArgs[i] = url.PathEscape(fmt.Sprint(a))
	}
	return fmt.Sprintf("%s://%s"+pathFmt, append([]any{scheme, host}, escapedArgs...)...)
}
