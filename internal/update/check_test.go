package update

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestCheckLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cli/latest.json" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"version":"1.2.3"}`)
	}))
	defer srv.Close()
	t.Setenv("CODEBAHN_RELEASES_URL", srv.URL+"/cli")

	rel, err := CheckLatest("v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if rel == nil {
		t.Fatal("expected non-nil release")
	}
	if rel.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", rel.Version, "1.2.3")
	}
	if !rel.Newer {
		t.Error("expected Newer = true")
	}
}

func TestCheckLatest_AlreadyCurrent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"version":"1.0.0"}`)
	}))
	defer srv.Close()
	t.Setenv("CODEBAHN_RELEASES_URL", srv.URL+"/cli")

	rel, err := CheckLatest("v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if rel.Newer {
		t.Error("expected Newer = false when versions match")
	}
}

func TestShouldCheck_EnvVar(t *testing.T) {
	t.Setenv("CODEBAHN_NO_UPDATE_CHECK", "1")
	if ShouldCheck("v1.0.0", nil) {
		t.Error("ShouldCheck should return false when env var is set")
	}
}

func TestShouldCheck_DevVersion(t *testing.T) {
	t.Setenv("CODEBAHN_NO_UPDATE_CHECK", "")
	if ShouldCheck("dev", nil) {
		t.Error("ShouldCheck should return false for dev version")
	}
}

func TestShouldCheck_ConfigDisabled(t *testing.T) {
	t.Setenv("CODEBAHN_NO_UPDATE_CHECK", "")
	disabled := false
	if ShouldCheck("v1.0.0", &disabled) {
		t.Error("ShouldCheck should return false when config disables it")
	}
}

func TestShouldCheck_RecentCheck(t *testing.T) {
	t.Setenv("CODEBAHN_NO_UPDATE_CHECK", "")
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	cacheFile := filepath.Join(dir, "codebahn", "last-update-check")
	os.MkdirAll(filepath.Dir(cacheFile), 0700)
	os.WriteFile(cacheFile, []byte(strconv.FormatInt(time.Now().Unix(), 10)), 0600)

	if ShouldCheck("v1.0.0", nil) {
		t.Error("ShouldCheck should return false when checked recently")
	}
}

func TestShouldCheck_StaleCheck(t *testing.T) {
	t.Setenv("CODEBAHN_NO_UPDATE_CHECK", "")
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	cacheFile := filepath.Join(dir, "codebahn", "last-update-check")
	os.MkdirAll(filepath.Dir(cacheFile), 0700)
	stale := time.Now().Add(-25 * time.Hour).Unix()
	os.WriteFile(cacheFile, []byte(strconv.FormatInt(stale, 10)), 0600)

	if !ShouldCheck("v1.0.0", nil) {
		t.Error("ShouldCheck should return true when last check is >24h ago")
	}
}

func TestRecordCheck(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	RecordCheck()

	data, err := os.ReadFile(filepath.Join(dir, "codebahn", "last-update-check"))
	if err != nil {
		t.Fatal(err)
	}
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		t.Fatalf("expected unix timestamp, got %q", data)
	}
	if time.Since(time.Unix(ts, 0)) > time.Minute {
		t.Error("recorded timestamp is too old")
	}
}

func TestPrintUpdateNotice(t *testing.T) {
	var buf bytes.Buffer
	PrintUpdateNotice(&buf, "v0.1.0", "0.2.0")
	got := buf.String()
	if got == "" {
		t.Fatal("expected non-empty notice")
	}
	want := "codebahn update"
	if !bytes.Contains(buf.Bytes(), []byte(want)) {
		t.Errorf("notice should mention %q, got %q", want, got)
	}
}
