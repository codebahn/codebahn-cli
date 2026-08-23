package format

import (
	"testing"
	"time"
)

func TestTimeAgo(t *testing.T) {
	tests := []struct {
		offset time.Duration
		want   string
	}{
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{2 * time.Hour, "2h ago"},
		{24 * time.Hour, "1d ago"},
		{3 * 24 * time.Hour, "3d ago"},
		{14 * 24 * time.Hour, "about 2 weeks ago"},
		{60 * 24 * time.Hour, "about 2 months ago"},
		{400 * 24 * time.Hour, "about 1 year ago"},
	}

	for _, tt := range tests {
		ts := time.Now().Add(-tt.offset).Format(time.RFC3339)
		got := TimeAgo(ts)
		if got != tt.want {
			t.Errorf("TimeAgo(%s ago) = %q, want %q", tt.offset, got, tt.want)
		}
	}
}

func TestTimeAgoEmpty(t *testing.T) {
	if got := TimeAgo(""); got != "" {
		t.Errorf("TimeAgo(\"\") = %q, want \"\"", got)
	}
}

func TestTimeAgoInvalid(t *testing.T) {
	if got := TimeAgo("not a date"); got != "" {
		t.Errorf("TimeAgo(\"not a date\") = %q, want \"\"", got)
	}
}

func TestRemoveExcessiveWhitespace(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"  hello   world\n  foo  ", "hello world foo"},
		{"normal text", "normal text"},
		{"", ""},
		{"  \n\t  ", ""},
		{"a\n\nb\n\nc", "a b c"},
	}
	for _, tt := range tests {
		got := RemoveExcessiveWhitespace(tt.input)
		if got != tt.want {
			t.Errorf("RemoveExcessiveWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello\nworld", "hello"},
		{"  hello  ", "hello"},
		{"", ""},
		{"single line", "single line"},
		{"first\nsecond\nthird", "first"},
		{"\nleading newline", ""},
	}
	for _, tt := range tests {
		got := FirstLine(tt.input)
		if got != tt.want {
			t.Errorf("FirstLine(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
