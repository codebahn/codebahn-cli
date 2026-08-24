package format

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/codebahn/codebahn-cli/internal/output"
)

func TestFormatUserInfo(t *testing.T) {
	raw := json.RawMessage(`{"login":"admin","full_name":"Admin User","email":"admin@example.com","is_admin":true}`)
	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, ok := Get("get_my_user_info")
	if !ok {
		t.Fatal("formatter not registered for get_my_user_info")
	}
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !contains(got, "admin") {
		t.Errorf("expected login, got: %s", got)
	}
	if !contains(got, "Admin User") {
		t.Errorf("expected full_name, got: %s", got)
	}
	if !contains(got, "admin@example.com") {
		t.Errorf("expected email, got: %s", got)
	}
}

func TestFormatUserInfoNoFullName(t *testing.T) {
	raw := json.RawMessage(`{"login":"admin","full_name":"","email":"admin@example.com"}`)
	var buf bytes.Buffer
	output.SetNoColor(true)
	defer output.SetNoColor(false)

	p := output.NewPrinter(&buf, false)
	f, _ := Get("get_my_user_info")
	if err := f(raw, nil, p); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if contains(got, "()") {
		t.Errorf("should not show empty parens, got: %s", got)
	}
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
