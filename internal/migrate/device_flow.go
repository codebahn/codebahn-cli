package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var githubDeviceBase = "https://github.com"

func SetGitHubDeviceBase(u string) { githubDeviceBase = u }

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
	ExpiresIn       int    `json:"expires_in"`
}

func RequestDeviceCode(ctx context.Context, clientID string) (*DeviceCodeResponse, error) {
	form := url.Values{
		"client_id": {clientID},
		"scope":     {""},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		githubDeviceBase+"/login/device/code",
		nil)
	if err != nil {
		return nil, fmt.Errorf("building device code request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(strings.NewReader(form.Encode()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device code request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return nil, fmt.Errorf("reading device code response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed (HTTP %d): %s", resp.StatusCode, body)
	}

	var code DeviceCodeResponse
	if err := json.Unmarshal(body, &code); err != nil {
		return nil, fmt.Errorf("decoding device code response: %w", err)
	}
	if code.DeviceCode == "" {
		return nil, fmt.Errorf("GitHub returned empty device code")
	}
	if code.Interval == 0 {
		code.Interval = 5
	}
	return &code, nil
}

type TokenResponse struct {
	AccessToken string
	ExpiresIn   int
}

func PollForToken(ctx context.Context, clientID, deviceCode string, interval int) (*TokenResponse, error) {
	form := url.Values{
		"client_id":   {clientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(interval) * time.Second):
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			githubDeviceBase+"/login/oauth/access_token",
			nil)
		if err != nil {
			return nil, fmt.Errorf("building token request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Body = io.NopCloser(strings.NewReader(form.Encode()))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("token request: %w", err)
		}

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading token response: %w", readErr)
		}

		var tokenResp struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			Error       string `json:"error"`
		}
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			return nil, fmt.Errorf("decoding token response: %w", err)
		}

		switch tokenResp.Error {
		case "":
			if tokenResp.AccessToken != "" {
				return &TokenResponse{
					AccessToken: tokenResp.AccessToken,
					ExpiresIn:   tokenResp.ExpiresIn,
				}, nil
			}
			return nil, fmt.Errorf("GitHub returned empty access token")
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5
			continue
		case "expired_token":
			return nil, fmt.Errorf("authorization timed out; re-run the command to try again")
		case "access_denied":
			return nil, fmt.Errorf("authorization denied by user")
		default:
			return nil, fmt.Errorf("GitHub OAuth error: %s", tokenResp.Error)
		}
	}
}

func openBrowser(u string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", u).Start()
	case "darwin":
		return exec.Command("open", u).Start()
	default:
		return nil
	}
}
