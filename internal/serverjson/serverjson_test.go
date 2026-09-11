package serverjson

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

type serverJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repository  struct {
		URL    string `json:"url"`
		Source string `json:"source"`
	} `json:"repository"`
	Version string `json:"version"`
	Remotes []struct {
		URL           string `json:"url"`
		TransportType string `json:"transportType"`
	} `json:"remotes"`
	Packages json.RawMessage `json:"packages"`
}

func loadServerJSON(t *testing.T) serverJSON {
	t.Helper()
	data, err := os.ReadFile("../../server.json")
	if err != nil {
		t.Fatalf("failed to read server.json: %v", err)
	}
	var sj serverJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		t.Fatalf("failed to parse server.json: %v", err)
	}
	return sj
}

func TestName(t *testing.T) {
	sj := loadServerJSON(t)
	want := "io.github.codebahn/codebahn"
	if sj.Name != want {
		t.Errorf("name = %q, want %q", sj.Name, want)
	}
}

func TestDescription(t *testing.T) {
	sj := loadServerJSON(t)
	want := "Codebahn is the private GitHub alternative: fast Git and CI for small teams."
	if sj.Description != want {
		t.Errorf("description = %q, want %q", sj.Description, want)
	}
	if len(sj.Description) >= 100 {
		t.Errorf("description length = %d, must be under 100 chars", len(sj.Description))
	}
}

func TestRepository(t *testing.T) {
	sj := loadServerJSON(t)
	if sj.Repository.URL != "https://github.com/codebahn/codebahn-cli" {
		t.Errorf("repository.url = %q, want %q", sj.Repository.URL, "https://github.com/codebahn/codebahn-cli")
	}
	if sj.Repository.Source != "github" {
		t.Errorf("repository.source = %q, want %q", sj.Repository.Source, "github")
	}
}

func TestVersion(t *testing.T) {
	sj := loadServerJSON(t)
	matched, err := regexp.MatchString(`^\d+\.\d+\.\d+$`, sj.Version)
	if err != nil {
		t.Fatalf("regexp error: %v", err)
	}
	if !matched {
		t.Errorf("version = %q, does not match bare semver pattern", sj.Version)
	}
}

func TestRemotes(t *testing.T) {
	sj := loadServerJSON(t)
	if len(sj.Remotes) == 0 {
		t.Fatal("remotes is empty")
	}
	if sj.Remotes[0].URL != "https://codebahn.net/mcp" {
		t.Errorf("remotes[0].url = %q, want %q", sj.Remotes[0].URL, "https://codebahn.net/mcp")
	}
	if sj.Remotes[0].TransportType != "streamable-http" {
		t.Errorf("remotes[0].transportType = %q, want %q", sj.Remotes[0].TransportType, "streamable-http")
	}
}

func TestPackagesAbsent(t *testing.T) {
	data, err := os.ReadFile("../../server.json")
	if err != nil {
		t.Fatalf("failed to read server.json: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse server.json: %v", err)
	}
	if _, ok := raw["packages"]; ok {
		t.Error("packages key should be absent from server.json")
	}
}
