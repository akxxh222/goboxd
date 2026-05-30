package runner

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const tempDirPrefix = "goboxd-"

func CleanupStaleTempDirs(parentDir string, maxAge time.Duration) (int, error) {
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	removed := 0

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), tempDirPrefix) {
			continue
		}

		path := filepath.Join(parentDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return removed, err
		}

		if now.Sub(info.ModTime()) < maxAge {
			continue
		}

		if err := os.RemoveAll(path); err != nil {
			return removed, err
		}
		removed++
	}

	return removed, nil
}
