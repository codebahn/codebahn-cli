package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const ClientID = "2499dc67-192d-412e-827a-d319288ca085"

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type callbackResult struct {
	Code  string
	State string
}

type tokenErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

func generateCodeVerifier() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating state: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func buildAuthorizeURL(instanceURL, redirectURI, codeChallenge, state string) (string, error) {
	u, err := url.Parse(strings.TrimSuffix(instanceURL, "/"))
	if err != nil {
		return "", fmt.Errorf("parsing instance URL: %w", err)
	}
	u.Path = "/login/oauth/authorize"
	q := u.Query()
	q.Set("client_id", ClientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func startCallbackServer() (*http.Server, chan callbackResult, chan error, error) {
	codeCh := make(chan callbackResult, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !r.URL.Query().Has("state") {
			http.NotFound(w, r)
			return
		}

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			desc := r.URL.Query().Get("error_description")
			if desc == "" {
				desc = errMsg
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprint(w, cancelledHTML)
			errCh <- fmt.Errorf("authorization denied: %s", desc)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(w, "Missing authorization code")
			errCh <- fmt.Errorf("callback missing authorization code")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, callbackHTML)

		codeCh <- callbackResult{
			Code:  code,
			State: r.URL.Query().Get("state"),
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("starting callback server: %w", err)
	}
	srv := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: mux,
	}
	go func() { _ = srv.Serve(listener) }()

	return srv, codeCh, errCh, nil
}

func exchangeToken(ctx context.Context, instanceURL, code, codeVerifier, redirectURI string) (*TokenResponse, error) {
	return doTokenRequest(ctx, instanceURL, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {ClientID},
		"code":          {code},
		"code_verifier": {codeVerifier},
		"redirect_uri":  {redirectURI},
	})
}

func Refresh(ctx context.Context, instanceURL, refreshToken string) (*TokenResponse, error) {
	return doTokenRequest(ctx, instanceURL, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {ClientID},
		"refresh_token": {refreshToken},
	})
}

func doTokenRequest(ctx context.Context, instanceURL string, params url.Values) (*TokenResponse, error) {
	endpoint := strings.TrimSuffix(instanceURL, "/") + "/login/oauth/access_token"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp tokenErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("token error: %s: %s", errResp.Error, errResp.Description)
		}
		return nil, fmt.Errorf("token request failed: HTTP %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}
	return &token, nil
}

func Login(ctx context.Context, instanceURL string, openBrowser func(string) error) (*TokenResponse, error) {
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, err
	}
	challenge := generateCodeChallenge(verifier)

	state, err := generateState()
	if err != nil {
		return nil, err
	}

	srv, codeCh, errCh, err2 := startCallbackServer()
	if err2 != nil {
		return nil, err2
	}
	defer func() { _ = srv.Close() }()

	redirectURI := fmt.Sprintf("http://%s", srv.Addr)
	authorizeURL, err2 := buildAuthorizeURL(instanceURL, redirectURI, challenge, state)
	if err2 != nil {
		return nil, err2
	}

	fmt.Fprintf(os.Stderr, "\nOpen this URL in your browser:\n\n  %s\n\n", authorizeURL)
	if openBrowser != nil {
		_ = openBrowser(authorizeURL)
	}

	select {
	case result := <-codeCh:
		if subtle.ConstantTimeCompare([]byte(result.State), []byte(state)) != 1 {
			return nil, fmt.Errorf("state mismatch (possible CSRF attack)")
		}
		return exchangeToken(ctx, instanceURL, result.Code, verifier, redirectURI)
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

const callbackPageStyle = `
  body {
    background: #15161A;
    color: #E8E6E1;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 100vh;
    margin: 0;
  }
  .card { text-align: center; padding: 3rem; }
  .icon { font-size: 2.5rem; margin-bottom: 1rem; }
  h1 { font-size: 1.1rem; font-weight: 600; margin-bottom: 0.5rem; }
  p { color: #8A867D; font-size: 0.875rem; }
`

const callbackHTML = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>codebahn</title>
<style>` + callbackPageStyle + `
  .icon { color: #C2521D; }
  h1 { color: #E8E6E1; }
</style>
</head>
<body>
<div class="card">
  <div class="icon">&#10003;</div>
  <h1>Authorization successful</h1>
  <p>You can close this tab and return to the terminal.</p>
</div>
</body>
</html>`

const cancelledHTML = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>codebahn</title>
<style>` + callbackPageStyle + `
  .icon { color: #5A574F; }
  h1 { color: #8A867D; }
</style>
</head>
<body>
<div class="card">
  <div class="icon">&#10005;</div>
  <h1>Authorization cancelled</h1>
  <p>You can close this tab and return to the terminal.</p>
</div>
</body>
</html>`
