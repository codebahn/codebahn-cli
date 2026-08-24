package output

import (
	"bytes"
	"strings"
	"testing"
)

func init() {
	SetNoColor(true)
}

func TestNewPrinter(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, false)
	if p == nil {
		t.Fatal("NewPrinter returned nil")
	}
	if p.IsJSON() {
		t.Error("expected jsonMode=false")
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, false)
	p.Table(
		[]string{"NAME", "AGE"},
		[][]string{
			{"alice", "30"},
			{"bob", "25"},
		},
	)
	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header NAME, got %q", out)
	}
	if !strings.Contains(out, "alice") {
		t.Errorf("expected alice in output, got %q", out)
	}
	if !strings.Contains(out, "bob") {
		t.Errorf("expected bob in output, got %q", out)
	}
}

func TestTableNoHeaders(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, false)
	p.Table(nil, [][]string{{"x", "y"}})
	out := buf.String()
	if strings.Contains(out, "NAME") {
		t.Errorf("no headers expected, got %q", out)
	}
	if !strings.Contains(out, "x") {
		t.Errorf("expected x in output, got %q", out)
	}
}

func TestTableAsJSON(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, true)
	p.Table(
		[]string{"name", "age"},
		[][]string{{"alice", "30"}},
	)
	out := buf.String()
	if !strings.Contains(out, `"name"`) {
		t.Errorf("expected JSON key name, got %q", out)
	}
	if !strings.Contains(out, `"alice"`) {
		t.Errorf("expected JSON value alice, got %q", out)
	}
}

func TestText(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, false)
	p.Text("hello")
	if got := buf.String(); got != "hello\n" {
		t.Errorf("Text() = %q, want %q", got, "hello\n")
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, false)
	p.JSON(map[string]string{"a": "b"})
	out := buf.String()
	if !strings.Contains(out, `"a"`) {
		t.Errorf("expected JSON key a, got %q", out)
	}
	if !strings.Contains(out, `"b"`) {
		t.Errorf("expected JSON value b, got %q", out)
	}
}

func TestFprintf(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, false)
	p.Fprintf("hello %s", "world")
	if got := buf.String(); got != "hello world" {
		t.Errorf("Fprintf() = %q, want %q", got, "hello world")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"hello", 3, "hel"},
		{"日本語のテスト", 5, "日本..."},
		{"日本", 5, "日本"},
		{"abcdef", 6, "abcdef"},
	}
	for _, tt := range tests {
		got := Truncate(tt.in, tt.n)
		if got != tt.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
		}
	}
}

func TestCyan(t *testing.T) {
	got := Cyan("test")
	if got != "test" {
		t.Errorf("Cyan() with noColor = %q, want %q", got, "test")
	}
}
