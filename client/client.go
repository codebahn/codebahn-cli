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

// DoRefreshFunc performs a locked, config-aware token refresh and returns
// the new access token and expiry. Implementations should acquire any
// cross-process lock, re-read the latest token state (another process may
// have already refreshed), perform the HTTP refresh only if still needed,
// and persist the result.
type DoRefreshFunc func(ctx context.Context) (accessToken string, expiry time.Time, err error)

type Client struct {
	baseURL      string
	accessToken  string
	tokenExpiry  time.Time
	doRefresh    DoRefreshFunc
	httpClient   *http.Client
	streamClient *http.Client
	mu           sync.Mutex
}

type StatusCodeError struct {
	Code int
	Body string
}

// scopeRefusal is how the server words a 403 caused by the token's scope
// ("token does not have at least one of required scope(s): [write:user]").
// A login made on the consent page gets only the areas the user chose there,
// so this is the refusal a narrowed connection meets.
const (
	scopeRefusal     = "required scope(s)"
	scopeRefusalHint = "The login or token in use does not have that access. " +
		"Change the connection under Settings > Applications on Codebahn, or use a token with that scope."
)

// Error reports the server's JSON "message" when the body has one, the raw
// body otherwise, and a fixed phrase when the server gave no detail at all.
// A scope refusal also says where to widen the access.
func (e *StatusCodeError) Error() string {
	detail := strings.TrimSpace(e.Body)
	var payload struct {
		Message *string `json:"message"`
	}
	if json.Unmarshal([]byte(detail), &payload) == nil && payload.Message != nil {
		detail = strings.TrimSpace(*payload.Message)
	}
	if detail == "" {
		detail = "no detail from server"
	}
	if e.Code == http.StatusForbidden && strings.Contains(detail, scopeRefusal) {
		detail = strings.TrimSuffix(detail, ".") + ". " + scopeRefusalHint
	}
	return fmt.Sprintf("HTTP %d: %s", e.Code, detail)
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

// BaseURL returns the Codebahn server the client talks to.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// SetOAuth configures OAuth token refresh. The doRefresh callback handles the
// entire refresh cycle: locking, re-reading config, performing the HTTP
// refresh only if still needed, and persisting. This keeps the client
// decoupled from the config layer while letting the caller coordinate with
// cross-process locks.
func (c *Client) SetOAuth(expiry time.Time, doRefresh DoRefreshFunc) {
	c.tokenExpiry = expiry
	c.doRefresh = doRefresh
}

const refreshBuffer = 30 * time.Second

func (c *Client) ensureValidToken(ctx context.Context) error {
	if c.doRefresh == nil || c.tokenExpiry.IsZero() {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.tokenExpiry.Add(-refreshBuffer)) {
		return nil
	}

	access, expiry, err := c.doRefresh(ctx)
	if err != nil {
		return err
	}

	c.accessToken = access
	c.tokenExpiry = expiry
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

// DoRawWithHeaders is like doRaw but sets additional request headers.
func (c *Client) DoRawWithHeaders(ctx context.Context, method, path string, body any, headers map[string]string) (json.RawMessage, error) {
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
	for k, v := range headers {
		req.Header.Set(k, v)
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
	case "GET":
		q := buildQueryParams(args, inPath)
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
		return c.doRaw(ctx, td.Method, path, nil)
	case "DELETE":
		body := buildBody(args, inPath)
		if body != nil {
			return c.doRaw(ctx, td.Method, path, body)
		}
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
		if inPath[f.Name] || !apiField(f) {
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

// apiField reports whether a field is sent to the REST API. Fields tagged
// api:"-" are tool parameters that the CLI applies client-side instead.
func apiField(f reflect.StructField) bool {
	return f.Tag.Get("api") != "-"
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
		if inPath[f.Name] || !apiField(f) {
			continue
		}
		jsonName := f.Tag.Get("json")
		if jsonName == "" || jsonName == "-" {
			continue
		}
		v := rv.Field(i)
		if v.IsZero() {
			continue
		}

		name := jsonName
		if override := f.Tag.Get("apiname"); override != "" {
			name = override
		}

		val := v.Interface()

		switch f.Tag.Get("body") {
		case "csv":
			if s, ok := val.(string); ok {
				val = tools.SplitCSV(s)
			}
		case "json":
			if s, ok := val.(string); ok {
				var raw json.RawMessage
				if json.Unmarshal([]byte(s), &raw) == nil {
					val = raw
				}
			}
		case "nest":
			nest := f.Tag.Get("nest")
			if nest == "" {
				panic(fmt.Sprintf("field %s has body:\"nest\" but no nest tag", f.Name))
			}
			sub, _ := body[nest].(map[string]any)
			if sub == nil {
				sub = map[string]any{}
			}
			sub[name] = val
			body[nest] = sub
			continue
		}

		body[name] = val
	}
	if len(body) == 0 {
		return nil
	}
	return body
}
