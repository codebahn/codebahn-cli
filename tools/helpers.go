package tools

import (
	"net/url"
	"strings"
)

// SplitCSV splits a comma-separated string, trims whitespace, and drops empty entries.
func SplitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// EscapePath splits a file path on "/", drops empty/"."/".." segments,
// URL-encodes each segment, and rejoins.
func EscapePath(p string) string {
	parts := strings.Split(p, "/")
	clean := make([]string, 0, len(parts))
	for _, s := range parts {
		if s == "" || s == ".." || s == "." {
			continue
		}
		clean = append(clean, url.PathEscape(s))
	}
	return strings.Join(clean, "/")
}
