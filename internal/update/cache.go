package update

import (
	"os"
	"path/filepath"
)

func CacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "codebahn")
	}
	dir, _ := os.UserCacheDir()
	return filepath.Join(dir, "codebahn")
}
