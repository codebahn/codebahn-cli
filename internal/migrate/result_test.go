package migrate

import (
	"testing"
)

func TestCollectResultCarriesSkipped(t *testing.T) {
	raw := []byte(`{"full_name":"me/repo","html_url":"https://codebahn.net/me/repo","import_skipped":[{"what":"releases","reason":"source refused"}]}`)

	res := collectResult(raw, "src/repo")

	if res.URL != "https://codebahn.net/me/repo" {
		t.Fatalf("URL = %q", res.URL)
	}
	if len(res.Skipped) != 1 || res.Skipped[0].What != "releases" || res.Skipped[0].Reason != "source refused" {
		t.Fatalf("Skipped = %+v", res.Skipped)
	}
}

func TestCollectResultWithoutSkipped(t *testing.T) {
	res := collectResult([]byte(`{"html_url":"https://codebahn.net/me/repo"}`), "src/repo")

	if res.Skipped != nil {
		t.Fatalf("Skipped = %+v, want none", res.Skipped)
	}
}

func TestSkippedLines(t *testing.T) {
	got := skippedLines([]skippedData{
		{What: "releases", Reason: "source refused"},
		{What: "issues", Reason: "token cannot read issues"},
	})

	want := []string{
		"skipped releases: source refused",
		"skipped issues: token cannot read issues",
	}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSummaryCountsSkipped(t *testing.T) {
	results := []migrateResult{
		{Repo: "a", Skipped: []skippedData{{What: "releases", Reason: "x"}}},
		{Repo: "b"},
		{Repo: "c", Error: "boom"},
	}

	if got := summary(results, 2, 1); got != "2 succeeded (1 with skipped data), 1 failed" {
		t.Fatalf("summary = %q", got)
	}
	if got := summary(results[1:2], 1, 0); got != "1 succeeded, 0 failed" {
		t.Fatalf("summary = %q", got)
	}
}
