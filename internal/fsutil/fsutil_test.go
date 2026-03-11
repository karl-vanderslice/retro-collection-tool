package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirExistingDirectory(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dir := filepath.Join(base, "a", "b")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdirall: %v", err)
	}

	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir existing directory returned error: %v", err)
	}
}

func TestEnsureDirPathBlockedByFile(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	blocked := filepath.Join(base, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := EnsureDir(blocked)
	if err == nil {
		t.Fatal("expected error when directory path is blocked by file")
	}
}
