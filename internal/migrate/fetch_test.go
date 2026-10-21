package migrate

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchPages_RejectsCrossOriginLinkHeader(t *testing.T) {
	leaked := false
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			leaked = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"name":"stolen"}]`)
	}))
	defer attacker.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", fmt.Sprintf(`<%s/steal>; rel="next"`, attacker.URL))
		_, _ = fmt.Fprint(w, `[{"name":"page1"}]`)
	}))
	defer ts.Close()

	_, err := fetchPages[SourceRepo](context.Background(), fetchOpts{
		startURL: ts.URL + "/repos",
		token:    "secret-token",
		label:    "test",
	})
	if err == nil {
		t.Fatal("expected error for cross-origin Link URL, got nil")
	}
	if leaked {
		t.Fatal("auth token was leaked to a cross-origin server via Link header")
	}
}

func TestFetchPages_NoFollowRedirect(t *testing.T) {
	leaked := false
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			leaked = true
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"name":"stolen"}]`)
	}))
	defer attacker.Close()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, attacker.URL+"/steal", http.StatusFound)
	}))
	defer ts.Close()

	_, err := fetchPages[SourceRepo](context.Background(), fetchOpts{
		startURL: ts.URL + "/repos",
		token:    "secret-token",
		label:    "test",
	})
	if err == nil {
		t.Fatal("expected error when server redirects, got nil")
	}
	if leaked {
		t.Fatal("auth token was leaked via redirect")
	}
}

func TestFetchPages_SameOriginLinkAllowed(t *testing.T) {
	var tsURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") != "2" {
			w.Header().Set("Link", fmt.Sprintf(`<%s/repos?page=2>; rel="next"`, tsURL))
			_, _ = fmt.Fprint(w, `[{"name":"repo1","full_name":"o/repo1","clone_url":"https://x.com/o/repo1.git"}]`)
			return
		}
		_, _ = fmt.Fprint(w, `[{"name":"repo2","full_name":"o/repo2","clone_url":"https://x.com/o/repo2.git"}]`)
	}))
	defer ts.Close()
	tsURL = ts.URL

	repos, err := fetchPages[SourceRepo](context.Background(), fetchOpts{
		startURL: ts.URL + "/repos",
		token:    "tok",
		label:    "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
}
