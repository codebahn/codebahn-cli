package format

import (
	"fmt"
	"strings"
	"time"
)

func TimeAgo(iso8601 string) string {
	if iso8601 == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, iso8601)
	if err != nil {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / 24 / 7)
		return fmt.Sprintf("about %d %s ago", weeks, plural(weeks, "week", "weeks"))
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		return fmt.Sprintf("about %d %s ago", months, plural(months, "month", "months"))
	default:
		years := int(d.Hours() / 24 / 365)
		return fmt.Sprintf("about %d %s ago", years, plural(years, "year", "years"))
	}
}

func RemoveExcessiveWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func FirstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return strings.TrimSpace(line)
}
