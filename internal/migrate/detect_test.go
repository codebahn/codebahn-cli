package migrate

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDetectService(t *testing.T) {
	t.Run("forgejo", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/v1/version" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"version":"1.20.0+gitea-1.20.0"}`)
		}))
		defer ts.Close()

		scheme, host := splitTestURL(t, ts.URL)
		got := DetectService(context.Background(), scheme, host, "")
		if got != "gitea" {
			t.Fatalf("expected gitea, got %q", got)
		}
	})

	t.Run("gitea", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"version":"1.21.0"}`)
		}))
		defer ts.Close()

		scheme, host := splitTestURL(t, ts.URL)
		got := DetectService(context.Background(), scheme, host, "")
		if got != "gitea" {
			t.Fatalf("expected gitea, got %q", got)
		}
	})

	t.Run("sends token", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "token test-token" {
				t.Errorf("missing or wrong auth header: %q", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"version":"1.21.0"}`)
		}))
		defer ts.Close()

		scheme, host := splitTestURL(t, ts.URL)
		got := DetectService(context.Background(), scheme, host, "test-token")
		if got != "gitea" {
			t.Fatalf("expected gitea, got %q", got)
		}
	})

	t.Run("non-forgejo 200 missing version key", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		}))
		defer ts.Close()

		scheme, host := splitTestURL(t, ts.URL)
		got := DetectService(context.Background(), scheme, host, "")
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("404", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		scheme, host := splitTestURL(t, ts.URL)
		got := DetectService(context.Background(), scheme, host, "")
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("connection refused", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to allocate port: %v", err)
		}
		addr := ln.Addr().String()
		if err := ln.Close(); err != nil {
			t.Fatalf("failed to close listener: %v", err)
		}

		got := DetectService(context.Background(), "http", addr, "")
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"version":"1.21.0"}`)
		}))
		defer ts.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		scheme, host := splitTestURL(t, ts.URL)
		got := DetectService(ctx, scheme, host, "")
		if got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}

func splitTestURL(t *testing.T, rawURL string) (scheme, host string) {
	t.Helper()
	const httpPrefix = "http://"
	if len(rawURL) <= len(httpPrefix) || rawURL[:len(httpPrefix)] != httpPrefix {
		t.Fatalf("unexpected test server URL: %q", rawURL)
	}
	return "http", rawURL[len(httpPrefix):]
}
