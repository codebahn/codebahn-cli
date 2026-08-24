package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestRequestDeviceCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login/device/code" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("missing Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"device_code": "dc_123",
			"user_code": "ABCD-1234",
			"verification_uri": "https://github.com/login/device",
			"interval": 5,
			"expires_in": 900
		}`)
	}))
	defer ts.Close()

	orig := githubDeviceBase
	SetGitHubDeviceBase(ts.URL)
	defer SetGitHubDeviceBase(orig)

	code, err := RequestDeviceCode(context.Background(), "test-client-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code.DeviceCode != "dc_123" {
		t.Errorf("DeviceCode = %q, want dc_123", code.DeviceCode)
	}
	if code.UserCode != "ABCD-1234" {
		t.Errorf("UserCode = %q, want ABCD-1234", code.UserCode)
	}
	if code.Interval != 5 {
		t.Errorf("Interval = %d, want 5", code.Interval)
	}
}

func TestPollForToken_Success(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n < 3 {
			fmt.Fprint(w, `{"error": "authorization_pending"}`)
			return
		}
		fmt.Fprint(w, `{"access_token": "ghu_test123", "token_type": "bearer"}`)
	}))
	defer ts.Close()

	orig := githubDeviceBase
	SetGitHubDeviceBase(ts.URL)
	defer SetGitHubDeviceBase(orig)

	token, err := PollForToken(context.Background(), "client", "device", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ghu_test123" {
		t.Errorf("token = %q, want ghu_test123", token)
	}
	if calls.Load() != 3 {
		t.Errorf("expected 3 calls, got %d", calls.Load())
	}
}

func TestPollForToken_Expired(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error": "expired_token"}`)
	}))
	defer ts.Close()

	orig := githubDeviceBase
	SetGitHubDeviceBase(ts.URL)
	defer SetGitHubDeviceBase(orig)

	_, err := PollForToken(context.Background(), "client", "device", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "authorization timed out; re-run the command to try again" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestPollForToken_Denied(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error": "access_denied"}`)
	}))
	defer ts.Close()

	orig := githubDeviceBase
	SetGitHubDeviceBase(ts.URL)
	defer SetGitHubDeviceBase(orig)

	_, err := PollForToken(context.Background(), "client", "device", 0)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "authorization denied by user" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestPollForToken_SlowDown(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			fmt.Fprint(w, `{"error": "slow_down"}`)
			return
		}
		fmt.Fprint(w, `{"access_token": "ghu_ok", "token_type": "bearer"}`)
	}))
	defer ts.Close()

	orig := githubDeviceBase
	SetGitHubDeviceBase(ts.URL)
	defer SetGitHubDeviceBase(orig)

	token, err := PollForToken(context.Background(), "client", "device", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ghu_ok" {
		t.Errorf("token = %q, want ghu_ok", token)
	}
}
