package update

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

var testEntity *openpgp.Entity

func TestMain(m *testing.M) {
	entity, err := openpgp.NewEntity("Test", "", "test@test.com", nil)
	if err != nil {
		panic(err)
	}
	testEntity = entity

	var pubBuf bytes.Buffer
	w, _ := armor.Encode(&pubBuf, "PGP PUBLIC KEY BLOCK", nil)
	entity.Serialize(w)
	w.Close()
	publicKey = pubBuf.String()

	os.Exit(m.Run())
}

func signBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var sigBuf bytes.Buffer
	err := openpgp.ArmoredDetachSign(&sigBuf, testEntity, bytes.NewReader(data), nil)
	if err != nil {
		t.Fatalf("signing: %v", err)
	}
	return sigBuf.Bytes()
}

func setupReleaseServer(t *testing.T, version, binaryContent string) *httptest.Server {
	t.Helper()

	binaryName := fmt.Sprintf("codebahn-%s-%s", runtime.GOOS, runtime.GOARCH)

	h := sha256.Sum256([]byte(binaryContent))
	checksumLine := fmt.Sprintf("%x  %s\n", h, binaryName)
	checksumData := []byte(checksumLine)
	sigData := signBytes(t, checksumData)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/cli/latest.json":
			fmt.Fprintf(w, `{"version":%q}`, version)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			w.Write(checksumData)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt.asc"):
			w.Write(sigData)
		case strings.HasSuffix(r.URL.Path, "/"+binaryName):
			w.Write([]byte(binaryContent))
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestUpdate_Success(t *testing.T) {
	srv := setupReleaseServer(t, "2.0.0", "#!/bin/sh\necho updated\n")
	defer srv.Close()
	t.Setenv("CODEBAHN_RELEASES_URL", srv.URL+"/cli")

	fakeExe := filepath.Join(t.TempDir(), "codebahn")
	os.WriteFile(fakeExe, []byte("old binary"), 0755)

	rel := &Release{Version: "2.0.0", Newer: true}
	err := Update(rel, fakeExe)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(fakeExe)
	if string(data) != "#!/bin/sh\necho updated\n" {
		t.Errorf("binary not updated, got %q", data)
	}
}

func TestUpdate_BadChecksum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		binaryName := fmt.Sprintf("codebahn-%s-%s", runtime.GOOS, runtime.GOARCH)
		switch {
		case r.URL.Path == "/cli/latest.json":
			fmt.Fprint(w, `{"version":"2.0.0"}`)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			fmt.Fprintf(w, "0000000000000000000000000000000000000000000000000000000000000000  %s\n", binaryName)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt.asc"):
			w.Write([]byte("not a real signature"))
		case strings.HasSuffix(r.URL.Path, "/"+binaryName):
			w.Write([]byte("binary content"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	t.Setenv("CODEBAHN_RELEASES_URL", srv.URL+"/cli")

	fakeExe := filepath.Join(t.TempDir(), "codebahn")
	os.WriteFile(fakeExe, []byte("old"), 0755)

	rel := &Release{Version: "2.0.0", Newer: true}
	err := Update(rel, fakeExe)
	if err == nil {
		t.Fatal("expected error for bad signature")
	}

	data, _ := os.ReadFile(fakeExe)
	if string(data) != "old" {
		t.Error("binary should not have been replaced")
	}
}

func TestUpdate_HomebrewDetection(t *testing.T) {
	srv := setupReleaseServer(t, "2.0.0", "new binary")
	defer srv.Close()
	t.Setenv("CODEBAHN_RELEASES_URL", srv.URL+"/cli")

	brewPath := filepath.Join(t.TempDir(), "Cellar", "codebahn", "1.0.0", "bin", "codebahn")
	os.MkdirAll(filepath.Dir(brewPath), 0755)
	os.WriteFile(brewPath, []byte("brew binary"), 0755)

	rel := &Release{Version: "2.0.0", Newer: true}
	err := Update(rel, brewPath)
	if err == nil || !strings.Contains(err.Error(), "brew upgrade") {
		t.Errorf("expected Homebrew detection error, got: %v", err)
	}
}
