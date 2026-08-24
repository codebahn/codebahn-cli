package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"golang.org/x/term"
)

var noColor bool

func init() {
	_, noColor = os.LookupEnv("NO_COLOR")
}

func SetNoColor(v bool) { noColor = v }

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func colorize(code, s string) string {
	if noColor || !IsTTY() {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func Green(s string) string   { return colorize("32", s) }
func Red(s string) string     { return colorize("31", s) }
func Yellow(s string) string  { return colorize("33", s) }
func Magenta(s string) string { return colorize("35", s) }
func Cyan(s string) string    { return colorize("36", s) }
func Bold(s string) string    { return colorize("1", s) }
func Dim(s string) string     { return colorize("2", s) }

func StatusColor(status string) string {
	switch status {
	case "success", "open":
		return Green(status)
	case "failure", "closed":
		return Red(status)
	case "running", "waiting", "blocked":
		return Yellow(status)
	case "merged":
		return Magenta(status)
	case "cancelled", "skipped":
		return Dim(status)
	default:
		return status
	}
}

type Printer struct {
	w        io.Writer
	jsonMode bool
}

func NewPrinter(w io.Writer, jsonMode bool) *Printer {
	return &Printer{w: w, jsonMode: jsonMode}
}

func (p *Printer) JSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}
	fmt.Fprintln(p.w, string(data))
}

func (p *Printer) Table(headers []string, rows [][]string) {
	if p.jsonMode {
		p.tableAsJSON(headers, rows)
		return
	}
	w := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	if len(headers) > 0 {
		_, _ = fmt.Fprintln(w, Bold(strings.Join(headers, "\t")))
	}
	for _, row := range rows {
		_, _ = fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	_ = w.Flush()
}

func (p *Printer) tableAsJSON(headers []string, rows [][]string) {
	result := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		m := make(map[string]string, len(headers))
		for i, h := range headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		result = append(result, m)
	}
	p.JSON(result)
}

func (p *Printer) Text(s string) {
	fmt.Fprintln(p.w, s)
}

func (p *Printer) Fprintf(format string, args ...any) {
	fmt.Fprintf(p.w, format, args...)
}

func (p *Printer) IsJSON() bool {
	return p.jsonMode
}

func Truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 3 {
		return string(runes[:n])
	}
	return string(runes[:n-3]) + "..."
}

func Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}
