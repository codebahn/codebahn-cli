package migrate

import "testing"

func TestParseSource(t *testing.T) {
	tests := []struct {
		input       string
		wantService string
		wantHost    string
		wantOwner   string
		wantRepo    string
		wantClone   string
		wantOrg     bool
		wantErr     bool
	}{
		{
			input:       "github.com/acme/api",
			wantService: "github", wantHost: "github.com",
			wantOwner: "acme", wantRepo: "api",
			wantClone: "https://github.com/acme/api.git",
		},
		{
			input:       "https://github.com/acme/api",
			wantService: "github", wantHost: "github.com",
			wantOwner: "acme", wantRepo: "api",
			wantClone: "https://github.com/acme/api.git",
		},
		{
			input:       "https://github.com/acme/api.git",
			wantService: "github", wantHost: "github.com",
			wantOwner: "acme", wantRepo: "api",
			wantClone: "https://github.com/acme/api.git",
		},
		{
			input:       "github.com/acme",
			wantService: "github", wantHost: "github.com",
			wantOwner: "acme",
			wantOrg:   true,
		},
		{
			input:       "gitlab.com/acme/api",
			wantService: "gitlab", wantHost: "gitlab.com",
			wantOwner: "acme", wantRepo: "api",
			wantClone: "https://gitlab.com/acme/api.git",
		},
		{
			input:       "codeberg.org/acme/api",
			wantService: "gitea", wantHost: "codeberg.org",
			wantOwner: "acme", wantRepo: "api",
			wantClone: "https://codeberg.org/acme/api.git",
		},
		{
			input:       "git.example.com/foo/bar",
			wantService: "git", wantHost: "git.example.com",
			wantOwner: "foo", wantRepo: "bar",
			wantClone: "https://git.example.com/foo/bar.git",
		},
		{
			input:       "GITHUB.COM/Acme/API",
			wantService: "github", wantHost: "github.com",
			wantOwner: "Acme", wantRepo: "API",
			wantClone: "https://github.com/Acme/API.git",
		},
		{
			input:       "github.com/acme/api/",
			wantService: "github", wantHost: "github.com",
			wantOwner: "acme", wantRepo: "api",
			wantClone: "https://github.com/acme/api.git",
		},
		{input: "", wantErr: true},
		{input: "not-a-url", wantErr: true},
		{input: "github.com/a/b/c/d", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			src, err := ParseSource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSource(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if src.Service != tt.wantService {
				t.Errorf("Service = %q, want %q", src.Service, tt.wantService)
			}
			if src.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", src.Host, tt.wantHost)
			}
			if src.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", src.Owner, tt.wantOwner)
			}
			if src.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", src.Repo, tt.wantRepo)
			}
			if src.CloneURL != tt.wantClone {
				t.Errorf("CloneURL = %q, want %q", src.CloneURL, tt.wantClone)
			}
			if src.IsOrg() != tt.wantOrg {
				t.Errorf("IsOrg() = %v, want %v", src.IsOrg(), tt.wantOrg)
			}
		})
	}
}

func TestSourceSupportsMetadata(t *testing.T) {
	tests := []struct {
		service string
		want    bool
	}{
		{"github", true},
		{"gitlab", true},
		{"gitea", true},
		{"git", false},
	}
	for _, tt := range tests {
		s := Source{Service: tt.service}
		if got := s.SupportsMetadata(); got != tt.want {
			t.Errorf("Source{Service: %q}.SupportsMetadata() = %v, want %v", tt.service, got, tt.want)
		}
	}
}
