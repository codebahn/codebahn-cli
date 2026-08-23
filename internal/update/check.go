package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/mod/semver"
)

var baseURL = "https://releases.codebahn.net/cli"

const checkInterval = 24 * time.Hour

type Release struct {
	Version string `json:"version"`
	Newer   bool   `json:"-"`
}

func releasesURL() string {
	if u := os.Getenv("CODEBAHN_RELEASES_URL"); u != "" {
		return u
	}
	return baseURL
}

func CheckLatest(currentVersion string) (*Release, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(releasesURL() + "/latest.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}

	latest := "v" + rel.Version
	if !semver.IsValid(latest) {
		return nil, fmt.Errorf("invalid version from server: %q", rel.Version)
	}
	if semver.Compare(currentVersion, latest) < 0 {
		rel.Newer = true
	}

	return &rel, nil
}

func ShouldCheck(currentVersion string, checkUpdates *bool) bool {
	if os.Getenv("CODEBAHN_NO_UPDATE_CHECK") != "" {
		return false
	}
	if currentVersion == "dev" {
		return false
	}
	if checkUpdates != nil && !*checkUpdates {
		return false
	}

	cacheFile := filepath.Join(CacheDir(), "last-update-check")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return true
	}
	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return true
	}
	return time.Since(time.Unix(ts, 0)) >= checkInterval
}

func RecordCheck() {
	dir := CacheDir()
	os.MkdirAll(dir, 0700)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	os.WriteFile(filepath.Join(dir, "last-update-check"), []byte(ts), 0600)
}

func PrintUpdateNotice(w io.Writer, current, latest string) {
	fmt.Fprintf(w, "Update available: v%s (current: %s). Run \"codebahn update\" to upgrade.\n", latest, current)
}
