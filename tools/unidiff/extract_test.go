package unidiff

import (
	"strings"
	"testing"
)

// Three files: a Go source, a markdown, and a binary.
var multiFileDiff = strings.Join([]string{
	"diff --git a/cmd/main.go b/cmd/main.go",
	"index 1234..5678 100644",
	"--- a/cmd/main.go",
	"+++ b/cmd/main.go",
	"@@ -1,3 +1,4 @@",
	" package main",
	"",
	`+import "fmt"`,
	" func main() {}",
	"diff --git a/README.md b/README.md",
	"index abcd..ef01 100644",
	"--- a/README.md",
	"+++ b/README.md",
	"@@ -10,2 +10,3 @@",
	" line ten",
	"+line eleven",
	" line twelve",
	"diff --git a/img.png b/img.png",
	"Binary files a/img.png and b/img.png differ",
}, "\n")

var singleFileDiff = strings.Join([]string{
	"diff --git a/only.go b/only.go",
	"index 1234..5678 100644",
	"--- a/only.go",
	"+++ b/only.go",
	"@@ -1 +1 @@",
	"-old",
	"+new",
}, "\n")

var renameDiff = strings.Join([]string{
	"diff --git a/old.go b/new.go",
	"similarity index 95%",
	"rename from old.go",
	"rename to new.go",
	"--- a/old.go",
	"+++ b/new.go",
	"@@ -1 +1 @@",
	"-package old",
	"+package new",
}, "\n")

func TestExtractFile(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		path     string
		wantOK   bool
		contains []string
		excludes []string
	}{
		{
			name:     "first file in multi-file diff",
			diff:     multiFileDiff,
			path:     "cmd/main.go",
			wantOK:   true,
			contains: []string{"diff --git a/cmd/main.go", `+import "fmt"`},
			excludes: []string{"README.md", "img.png"},
		},
		{
			name:     "middle file in multi-file diff",
			diff:     multiFileDiff,
			path:     "README.md",
			wantOK:   true,
			contains: []string{"diff --git a/README.md", "+line eleven"},
			excludes: []string{"cmd/main.go", "img.png"},
		},
		{
			name:     "binary file",
			diff:     multiFileDiff,
			path:     "img.png",
			wantOK:   true,
			contains: []string{"Binary files a/img.png and b/img.png differ"},
		},
		{
			name:   "single file equals input",
			diff:   singleFileDiff,
			path:   "only.go",
			wantOK: true,
		},
		{
			name:   "rename matches pre-rename path",
			diff:   renameDiff,
			path:   "old.go",
			wantOK: true,
		},
		{
			name:   "rename matches post-rename path",
			diff:   renameDiff,
			path:   "new.go",
			wantOK: true,
		},
		{
			name: "rename does not match unrelated path",
			diff: renameDiff,
			path: "neither.go",
		},
		{
			name: "not found",
			diff: multiFileDiff,
			path: "does/not/exist.go",
		},
		{
			name: "empty diff",
			diff: "",
			path: "x",
		},
		{
			name: "empty path",
			diff: multiFileDiff,
			path: "",
		},
		{
			name: "no diff markers",
			diff: "just some text\n",
			path: "anything",
		},
		{
			name: "prefix of a longer path does not match",
			diff: multiFileDiff,
			path: "main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok := ExtractFile(tt.diff, tt.path)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v\nout: %q", ok, tt.wantOK, out)
			}
			if !ok {
				if out != "" {
					t.Fatalf("expected empty output for not-found, got: %q", out)
				}
				return
			}
			for _, s := range tt.contains {
				if !strings.Contains(out, s) {
					t.Errorf("missing %q in output:\n%s", s, out)
				}
			}
			for _, s := range tt.excludes {
				if strings.Contains(out, s) {
					t.Errorf("unexpected %q in output:\n%s", s, out)
				}
			}
		})
	}
}

func TestExtractFileSingleFileIsWholeInput(t *testing.T) {
	out, ok := ExtractFile(singleFileDiff, "only.go")
	if !ok {
		t.Fatal("expected only.go to be found")
	}
	if out != singleFileDiff {
		t.Errorf("single-file extraction should return the whole diff\n got: %q\nwant: %q", out, singleFileDiff)
	}
}

func TestExtractFileWithTrailingNewline(t *testing.T) {
	out, ok := ExtractFile(multiFileDiff+"\n", "README.md")
	if !ok {
		t.Fatal("expected README.md to be found")
	}
	if strings.HasSuffix(out, "\n") {
		t.Errorf("section should be trimmed of trailing newlines, got %q", out)
	}
	if !strings.HasPrefix(out, "diff --git a/README.md b/README.md") {
		t.Errorf("section should start at its own header, got %q", out)
	}
}
