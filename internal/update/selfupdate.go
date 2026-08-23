package update

import (
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"golang.org/x/mod/semver"
)

//go:embed pgp/release-key.asc
var publicKey string

var ErrHomebrew = fmt.Errorf("installed via Homebrew; run: brew upgrade codebahn")

func Update(currentVersion, execPath string) error {
	rel, err := CheckLatest(currentVersion)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}
	if !rel.Newer {
		return nil
	}

	if execPath == "" {
		p, err := os.Executable()
		if err != nil {
			return fmt.Errorf("finding executable: %w", err)
		}
		execPath, err = filepath.EvalSymlinks(p)
		if err != nil {
			return fmt.Errorf("resolving executable path: %w", err)
		}
	}

	if isHomebrew(execPath) {
		return ErrHomebrew
	}

	tag := "v" + rel.Version
	base := releasesURL()

	checksumData, err := fetch(base + "/" + tag + "/checksums.txt")
	if err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	sigData, err := fetch(base + "/" + tag + "/checksums.txt.asc")
	if err != nil {
		return fmt.Errorf("downloading signature: %w", err)
	}

	if err := verifySignature(checksumData, sigData); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	binaryName := fmt.Sprintf("codebahn-%s-%s", runtime.GOOS, runtime.GOARCH)
	expectedHash, err := findChecksum(checksumData, binaryName)
	if err != nil {
		return err
	}

	binaryData, err := fetch(base + "/" + tag + "/" + binaryName)
	if err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}

	actualHash := fmt.Sprintf("%x", sha256.Sum256(binaryData))
	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return replaceBinary(execPath, binaryData)
}

func isHomebrew(execPath string) bool {
	lower := strings.ToLower(execPath)
	return strings.Contains(lower, "/cellar/") || strings.Contains(lower, "/homebrew/")
}

func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

func verifySignature(data, sig []byte) error {
	keyring, err := openpgp.ReadArmoredKeyRing(strings.NewReader(publicKey))
	if err != nil {
		return fmt.Errorf("reading public key: %w", err)
	}

	_, err = openpgp.CheckArmoredDetachedSignature(keyring, strings.NewReader(string(data)), strings.NewReader(string(sig)), nil)
	return err
}

func findChecksum(checksumData []byte, filename string) (string, error) {
	for _, line := range strings.Split(string(checksumData), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == filename {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("no checksum found for %s", filename)
}

func replaceBinary(execPath string, data []byte) error {
	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("stat current binary: %w", err)
	}

	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, ".codebahn-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("setting permissions: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}

func LatestVersion() string {
	rel, err := CheckLatest("v0.0.0")
	if err != nil {
		return ""
	}
	return rel.Version
}

func IsNewer(current, latest string) bool {
	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}
	return semver.Compare(current, latest) < 0
}
