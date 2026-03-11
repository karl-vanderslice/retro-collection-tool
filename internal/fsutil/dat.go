package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type datCandidate struct {
	path    string
	modTime time.Time
}

func FindLatestDAT(dir, contains string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read dat dir %s: %w", dir, err)
	}

	needle := strings.ToLower(strings.TrimSpace(contains))
	var matches []datCandidate

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".dat" {
			continue
		}
		if needle != "" && !strings.Contains(strings.ToLower(name), needle) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return "", fmt.Errorf("stat dat %s: %w", name, err)
		}
		matches = append(matches, datCandidate{
			path:    filepath.Join(dir, name),
			modTime: info.ModTime(),
		})
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no DAT file found in %s matching %q", dir, contains)
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].modTime.Equal(matches[j].modTime) {
			return matches[i].path > matches[j].path
		}
		return matches[i].modTime.After(matches[j].modTime)
	})

	return matches[0].path, nil
}
