package update

import (
	"os"
	"path/filepath"
	"runtime"
)

func CacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "codebahn")
	}
	if runtime.GOOS == "windows" {
		dir, _ := os.UserCacheDir()
		return filepath.Join(dir, "codebahn")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "codebahn")
}
