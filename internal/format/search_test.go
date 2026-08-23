package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/codebahn/codebahn-cli/internal/output"
)

func TestFormatSearchCode(t *testing.T) {
	raw := json.RawMessage(`{"ok":true,"data":[{"repository":{"full_name":"acme/app"},"name":"main.go","path":"src/main.go","sha":"abc123","html_url":"https://codebahn.net/acme/app/src/main.go"}]}`)

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, ok := Get("search_code")
	if !ok {
		t.Fatal("formatter not registered for search_code")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !contains(got, "acme/app") {
		t.Errorf("expected repo name, got: %s", got)
	}
	if !contains(got, "src/main.go") {
		t.Errorf("expected file path, got: %s", got)
	}
}

func TestFormatSearchCodeEmpty(t *testing.T) {
	raw := json.RawMessage(`{"ok":true,"data":[]}`)

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, _ := Get("search_code")
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty results, got: %s", buf.String())
	}
}

func TestFormatSearchRepos(t *testing.T) {
	ts := time.Now().Add(-3 * time.Hour).Format(time.RFC3339)
	raw := json.RawMessage(fmt.Sprintf(`{"ok":true,"data":[{"full_name":"acme/app","description":"My app","private":false,"updated_at":"%s"}]}`, ts))

	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, ok := Get("search_repos")
	if !ok {
		t.Fatal("formatter not registered for search_repos")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !contains(got, "acme/app") {
		t.Errorf("expected repo name, got: %s", got)
	}
	if !contains(got, "My app") {
		t.Errorf("expected description, got: %s", got)
	}
	if !contains(got, "public") {
		t.Errorf("expected visibility, got: %s", got)
	}
}
