package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codebahn/codebahn-cli/client"
)

func TestFetchGitHubAppConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/github-app/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"enabled":true,"client_id":"Iv1.abc123","app_slug":"codebahn"}`)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-token")
	cfg, err := FetchGitHubAppConfig(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Enabled {
		t.Error("expected enabled=true")
	}
	if cfg.ClientID != "Iv1.abc123" {
		t.Errorf("ClientID = %q, want Iv1.abc123", cfg.ClientID)
	}
}

func TestFetchGitHubAppConfig_Disabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"enabled":false}`)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-token")
	cfg, err := FetchGitHubAppConfig(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Enabled {
		t.Error("expected enabled=false")
	}
}

func TestConnectGitHubApp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/github-app/connect" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("X-GitHub-Token"); got != "ghu_test" {
			t.Errorf("X-GitHub-Token = %q, want ghu_test", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"installation_id":42,"account_login":"acme","account_type":"Organization"}`)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-token")
	conn, err := ConnectGitHubApp(context.Background(), c, "ghu_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn.InstallationID != 42 {
		t.Errorf("InstallationID = %d, want 42", conn.InstallationID)
	}
	if conn.AccountLogin != "acme" {
		t.Errorf("AccountLogin = %q, want acme", conn.AccountLogin)
	}
}

func TestConnectGitHubApp_NotInstalled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error":"not_installed","install_url":"https://github.com/apps/codebahn/installations/new"}`)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-token")
	conn, err := ConnectGitHubApp(context.Background(), c, "ghu_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn.Error != "not_installed" {
		t.Errorf("Error = %q, want not_installed", conn.Error)
	}
	if conn.InstallURL == "" {
		t.Error("expected install_url")
	}
}

func TestListGitHubAppRepos(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/github-app/repos" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("installation_id") != "42" {
			t.Errorf("installation_id = %q, want 42", r.URL.Query().Get("installation_id"))
		}
		if got := r.Header.Get("X-GitHub-Token"); got != "ghu_test" {
			t.Errorf("X-GitHub-Token = %q, want ghu_test", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"repos":[{"full_name":"acme/api","description":"API","private":true}],"permissions":{"issues":true}}`)
	}))
	defer ts.Close()

	c := client.New(ts.URL, "test-token")
	repos, err := ListGitHubAppRepos(context.Background(), c, 42, "ghu_test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(repos))
	}
	if repos[0].Name != "api" {
		t.Errorf("Name = %q, want api", repos[0].Name)
	}
	if repos[0].FullName != "acme/api" {
		t.Errorf("FullName = %q, want acme/api", repos[0].FullName)
	}
}
