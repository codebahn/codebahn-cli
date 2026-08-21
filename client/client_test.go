package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codebahn/codebahn-cli/tools"
)

func TestResolvePath(t *testing.T) {
	type args struct {
		Owner string `json:"owner"`
		Repo  string `json:"repo"`
		Index int    `json:"index"`
	}
	tmpl := "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}"
	got, err := resolvePath(tmpl, args{Owner: "acme", Repo: "test", Index: 42})
	if err != nil {
		t.Fatal(err)
	}
	want := "/repos/acme/test/issues/42"
	if got != want {
		t.Errorf("resolvePath = %q, want %q", got, want)
	}
}

func TestResolvePathNoTemplate(t *testing.T) {
	type args struct{}
	got, err := resolvePath("/user", args{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/user" {
		t.Errorf("resolvePath = %q, want /user", got)
	}
}

func TestPathFields(t *testing.T) {
	tmpl := "/repos/{{.Owner}}/{{.Repo}}/issues/{{.Index}}"
	fields := pathFields(tmpl)
	want := map[string]bool{"Owner": true, "Repo": true, "Index": true}
	if len(fields) != len(want) {
		t.Fatalf("pathFields = %v, want %v", fields, want)
	}
	for k := range want {
		if !fields[k] {
			t.Errorf("missing field %s", k)
		}
	}
}

func TestBuildQueryParams(t *testing.T) {
	type args struct {
		Owner string `json:"owner" required:"true"`
		Repo  string `json:"repo"  required:"true"`
		Page  int    `json:"page"  default:"1"`
		Limit int    `json:"limit" default:"20"`
		State string `json:"state" default:"open"`
	}
	inPath := map[string]bool{"Owner": true, "Repo": true}
	a := args{Owner: "acme", Repo: "test", Page: 2, Limit: 50, State: "closed"}
	q := buildQueryParams(a, inPath)
	if q.Get("page") != "2" {
		t.Errorf("page = %q, want 2", q.Get("page"))
	}
	if q.Get("limit") != "50" {
		t.Errorf("limit = %q, want 50", q.Get("limit"))
	}
	if q.Get("state") != "closed" {
		t.Errorf("state = %q, want closed", q.Get("state"))
	}
	if q.Get("owner") != "" {
		t.Error("owner should not be in query params (it's a path field)")
	}
}

func TestBuildQueryParamsSkipsZeroValues(t *testing.T) {
	type args struct {
		Owner string `json:"owner"`
		Page  int    `json:"page"`
		State string `json:"state"`
		Flag  bool   `json:"flag"`
	}
	q := buildQueryParams(args{Owner: "acme"}, map[string]bool{"Owner": true})
	if q.Get("page") != "" {
		t.Error("zero int should be skipped")
	}
	if q.Get("state") != "" {
		t.Error("zero string should be skipped")
	}
	if q.Get("flag") != "" {
		t.Error("zero bool should be skipped")
	}
}

func TestBuildBody(t *testing.T) {
	type args struct {
		Owner string `json:"owner"`
		Repo  string `json:"repo"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	inPath := map[string]bool{"Owner": true, "Repo": true}
	a := args{Owner: "acme", Repo: "test", Title: "hello", Body: "world"}
	body := buildBody(a, inPath)
	m, ok := body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want map[string]any", body)
	}
	if m["title"] != "hello" {
		t.Errorf("title = %v, want hello", m["title"])
	}
	if _, ok := m["owner"]; ok {
		t.Error("owner should not be in body (it's a path field)")
	}
}

func TestNewClient(t *testing.T) {
	c := New("https://codebahn.dev", "test-token")
	if c.baseURL != "https://codebahn.dev" {
		t.Errorf("baseURL = %q", c.baseURL)
	}
}

func TestGetRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok123" {
			t.Errorf("auth = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":1}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "tok123")
	raw, err := c.GetRaw(context.Background(), "/test")
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"id":1}` {
		t.Errorf("body = %s", raw)
	}
}

func TestPostRaw(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["title"] != "hello" {
			t.Errorf("title = %v", m["title"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"number":1}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	raw, err := c.PostRaw(context.Background(), "/test", map[string]string{"title": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"number":1}` {
		t.Errorf("body = %s", raw)
	}
}

func TestStatusCodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = io.WriteString(w, "not found")
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	_, err := c.GetRaw(context.Background(), "/nope")
	sce, ok := err.(*StatusCodeError)
	if !ok {
		t.Fatalf("error type = %T, want *StatusCodeError", err)
	}
	if sce.Code != 404 {
		t.Errorf("code = %d, want 404", sce.Code)
	}
}

func TestExecuteGET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/repos/acme/test/issues" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("state") != "open" {
			t.Errorf("state = %s", r.URL.Query().Get("state"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("page = %s", r.URL.Query().Get("page"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"number":1}]`)
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	td := tools.ToolDef{
		Name:     "list_repo_issues",
		Method:   "GET",
		PathTmpl: "/repos/{{.Owner}}/{{.Repo}}/issues",
		Args:     tools.ListRepoIssuesArgs{},
	}
	result, err := c.Execute(context.Background(), td, tools.ListRepoIssuesArgs{
		Owner: "acme",
		Repo:  "test",
		State: "open",
		Page:  2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != `[{"number":1}]` {
		t.Errorf("result = %s", result)
	}
}

func TestExecutePOST(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/v1/repos/acme/test/issues" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var m map[string]any
		_ = json.Unmarshal(body, &m)
		if m["title"] != "hello" {
			t.Errorf("title = %v", m["title"])
		}
		if _, ok := m["owner"]; ok {
			t.Error("owner should not be in body")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"number":1}`)
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	td := tools.ToolDef{
		Name:     "create_issue",
		Method:   "POST",
		PathTmpl: "/repos/{{.Owner}}/{{.Repo}}/issues",
		Args:     tools.CreateIssueArgs{},
	}
	result, err := c.Execute(context.Background(), td, tools.CreateIssueArgs{
		Owner: "acme",
		Repo:  "test",
		Title: "hello",
		Body:  "world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != `{"number":1}` {
		t.Errorf("result = %s", result)
	}
}

func TestDeleteMethod(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := New(srv.URL, "tok")
	err := c.Delete(context.Background(), "/test")
	if err != nil {
		t.Fatal(err)
	}
}
