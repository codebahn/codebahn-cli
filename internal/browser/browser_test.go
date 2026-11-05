package browser

import (
	"runtime"
	"testing"
)

func TestOpen_NoError(t *testing.T) {
	// We can't verify the browser actually opens, but we can verify the
	// function exists with the expected signature and returns no error on
	// the current platform. On CI / headless environments, the underlying
	// command may fail to start; that is acceptable.
	err := Open("https://example.com")
	if err != nil && runtime.GOOS != "linux" {
		// On Linux (CI), xdg-open may not be installed.
		t.Fatalf("Open returned unexpected error: %v", err)
	}
}

func TestOpen_RejectsNonHTTPURLs(t *testing.T) {
	// A value starting with "-" would be parsed as an option by xdg-open or open, and
	// other schemes (file:, javascript:) are never a login or verification page.
	for _, u := range []string{"", "-h", "--version", "file:///etc/passwd", "javascript:alert(1)", "ftp://example.com"} {
		if err := Open(u); err == nil {
			t.Errorf("Open(%q) = nil, want an error", u)
		}
	}
}
