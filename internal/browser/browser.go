package browser

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Open opens the given http or https URL in the user's default browser. Anything else is
// rejected: a value starting with "-" would be read as an option by the opener, and other
// schemes are never a login or verification page.
func Open(url string) error {
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return fmt.Errorf("refusing to open %q: not an http or https URL", url)
	}
	// The URL is a single argument (no shell) and is validated above.
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start() // #nosec G204 -- argument array, http(s) only
	case "darwin":
		return exec.Command("open", url).Start() // #nosec G204 -- argument array, http(s) only
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start() // #nosec G204 -- argument array, http(s) only
	default:
		return nil
	}
}
