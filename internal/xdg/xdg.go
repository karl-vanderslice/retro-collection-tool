package xdg

import (
	"os"
	"path/filepath"
	"strings"
)

func ConfigHome() string {
	if p := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); p != "" {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config")
	}
	return ""
}

func CacheHome() string {
	if p := strings.TrimSpace(os.Getenv("XDG_CACHE_HOME")); p != "" {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache")
	}
	return ""
}
