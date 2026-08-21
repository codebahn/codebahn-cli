package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/codebahn/codebahn-cli/tools"
)

type Client struct {
	baseURL      string
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
	onRefresh    func(accessToken, refreshToken string, expiry time.Time)
	clientID     string
	httpClient   *http.Client
	streamClient *http.Client
	mu           sync.Mutex
}

type StatusCodeError struct {
	Code int
	Body string
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Body)
}

func New(baseURL, accessToken string) *Client {
	noRedirect := func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &Client{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout:       5 * time.Minute,
			CheckRedirect: noRedirect,
		},
		streamClient: &http.Client{
			Timeout:       0,
			CheckRedirect: noRedirect,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 30 * time.Second,
			},
		},
	}
}

// SetOAuth configures OAuth token refresh.
func (c *Client) SetOAuth(refreshToken, clientID string, expiry time.Time, onRefresh func(string, string, time.Time)) {
	c.refreshToken = refreshToken
	c.clientID = clientID
	c.tokenExpiry = expiry
	c.onRefresh = onRefresh
}

const refreshBuffer = 30 * time.Second

func (c *Client) ensureValidToken(ctx context.Context) error {
	if c.refreshToken == "" || c.tokenExpiry.IsZero() {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.tokenExpiry.Add(-refreshBuffer)) {
		return nil
	}

	params := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {c.clientID},
		"refresh_token": {c.refreshToken},
	}

	endpoint := c.baseURL + "/login/oauth/access_token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("building refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("reading refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			ErrorDescription string `json:"error_description"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.ErrorDescription != "" {
			return fmt.Errorf("token refresh failed: %s; run 'codebahn auth login' to re-authenticate", errResp.ErrorDescription)
		}
		return fmt.Errorf("token refresh failed (HTTP %d); run 'codebahn auth login' to re-authenticate", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("decoding refresh response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		c.refreshToken = tokenResp.RefreshToken
	}
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	if c.onRefresh != nil {
		c.onRefresh(c.accessToken, c.refreshToken, c.tokenExpiry)
	}

	return nil
}

func (c *Client) apiURL(path string) string {
	return c.baseURL + "/api/v1" + path
}

func (c *Client) doRaw(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	if err := c.ensureValidToken(ctx); err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody := make([]byte, 512)
		n, _ := io.ReadFull(resp.Body, respBody)
		return nil, &StatusCodeError{Code: resp.StatusCode, Body: string(respBody[:n])}
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	const maxResponseSize = 50 * 1024 * 1024
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return json.RawMessage(raw), nil
}

func (c *Client) GetRaw(ctx context.Context, path string) (json.RawMessage, error) {
	return c.doRaw(ctx, http.MethodGet, path, nil)
}

func (c *Client) PostRaw(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.doRaw(ctx, http.MethodPost, path, body)
}

func (c *Client) PatchRaw(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.doRaw(ctx, http.MethodPatch, path, body)
}

func (c *Client) Delete(ctx context.Context, path string) error {
	_, err := c.doRaw(ctx, http.MethodDelete, path, nil)
	return err
}

func (c *Client) Stream(ctx context.Context, path string) (io.ReadCloser, error) {
	if err := c.ensureValidToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL(path), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody := make([]byte, 512)
		n, _ := io.ReadFull(resp.Body, respBody)
		_ = resp.Body.Close()
		return nil, &StatusCodeError{Code: resp.StatusCode, Body: string(respBody[:n])}
	}
	return resp.Body, nil
}

// Execute runs a tool's API call using its ToolDef metadata.
// For GET/DELETE: non-path fields become query parameters.
// For POST/PATCH/PUT: non-path fields become the JSON body.
func (c *Client) Execute(ctx context.Context, td tools.ToolDef, args any) (json.RawMessage, error) {
	path, err := resolvePath(td.PathTmpl, args)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	inPath := pathFields(td.PathTmpl)

	switch td.Method {
	case "GET", "DELETE":
		q := buildQueryParams(args, inPath)
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
		return c.doRaw(ctx, td.Method, path, nil)
	default:
		body := buildBody(args, inPath)
		return c.doRaw(ctx, td.Method, path, body)
	}
}

func resolvePath(tmpl string, args any) (string, error) {
	if !strings.Contains(tmpl, "{{") {
		return tmpl, nil
	}
	t, err := template.New("path").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := t.Execute(&buf, args); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var tmplFieldRE = regexp.MustCompile(`\{\{\.(\w+)\}\}`)

func pathFields(tmpl string) map[string]bool {
	matches := tmplFieldRE.FindAllStringSubmatch(tmpl, -1)
	out := make(map[string]bool, len(matches))
	for _, m := range matches {
		out[m[1]] = true
	}
	return out
}

func buildQueryParams(args any, inPath map[string]bool) url.Values {
	q := url.Values{}
	rv := reflect.ValueOf(args)
	rt := rv.Type()
	if rt.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rv.Type()
	}

	for i := range rt.NumField() {
		f := rt.Field(i)
		if inPath[f.Name] {
			continue
		}
		name := f.Tag.Get("json")
		if name == "" || name == "-" {
			continue
		}
		v := rv.Field(i)
		if v.IsZero() {
			continue
		}
		q.Set(name, fmt.Sprintf("%v", v.Interface()))
	}
	return q
}

func buildBody(args any, inPath map[string]bool) any {
	rv := reflect.ValueOf(args)
	rt := rv.Type()
	if rt.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rv.Type()
	}

	body := map[string]any{}
	for i := range rt.NumField() {
		f := rt.Field(i)
		if inPath[f.Name] {
			continue
		}
		name := f.Tag.Get("json")
		if name == "" || name == "-" {
			continue
		}
		v := rv.Field(i)
		if v.IsZero() {
			continue
		}
		body[name] = v.Interface()
	}
	if len(body) == 0 {
		return nil
	}
	return body
}
