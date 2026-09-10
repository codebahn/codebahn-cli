// Package unidiff extracts per-file sections from unified diff output.
//
// Forgejo's .diff endpoints return the whole diff; the file_path parameter of
// get_pull_request_diff and get_commit_diff is applied on top of that by the
// tool implementation (MCP handler or CLI), which is why it lives here in the
// shared module.
package unidiff

import "strings"

const marker = "\ndiff --git "

// ExtractFile returns the section of a unified diff that describes filePath.
// It matches on either the pre-rename (a/) or post-rename (b/) side.
// Returns the section including its diff --git header, and whether it was found.
func ExtractFile(raw, filePath string) (string, bool) {
	if filePath == "" || raw == "" {
		return "", false
	}

	// Normalize: ensure raw starts with \n so marker search works uniformly.
	search := "\n" + raw
	pos := 0

	for {
		idx := strings.Index(search[pos:], marker)
		if idx < 0 {
			return "", false
		}
		start := pos + idx + 1 // skip the leading \n

		// Find end of this header line.
		headerEnd := strings.IndexByte(search[start:], '\n')
		if headerEnd < 0 {
			headerEnd = len(search) - start
		}
		header := search[start : start+headerEnd]

		if matchesPath(header, filePath) {
			// Find next diff --git or end of input.
			end := strings.Index(search[start+1:], marker)
			if end < 0 {
				return strings.TrimRight(search[start:], "\n"), true
			}
			return strings.TrimRight(search[start:start+1+end], "\n"), true
		}

		pos = start + 1
	}
}

func matchesPath(header, filePath string) bool {
	const prefix = "diff --git "
	rest := strings.TrimPrefix(header, prefix)

	sep := " b/"
	idx := strings.Index(rest, sep)
	if idx < 0 {
		return false
	}

	a := strings.TrimPrefix(rest[:idx], "a/")
	b := strings.TrimRight(rest[idx+len(sep):], " \r")

	return a == filePath || b == filePath
}
