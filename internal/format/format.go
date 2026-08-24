package format

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codebahn/codebahn-cli/internal/output"
)

type Formatter func(raw json.RawMessage, args any, p *output.Printer) error

var registry = map[string]Formatter{}

func Register(toolName string, f Formatter) {
	registry[toolName] = f
}

func Get(toolName string) (Formatter, bool) {
	f, ok := registry[toolName]
	return f, ok
}

func successf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if output.IsTTY() {
		fmt.Fprintf(os.Stderr, "%s %s\n", output.Green("✓"), msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
}
