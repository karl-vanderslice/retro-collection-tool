package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		if info, statErr := os.Stat(path); statErr == nil {
			if info.IsDir() {
				return nil
			}
			return fmt.Errorf("create dir %s: path exists and is not a directory", path)
		}
		return fmt.Errorf("create dir %s: %w", path, err)
	}
	return nil
}

func RemoveIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		return os.RemoveAll(path)
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src %s: %w", src, err)
	}
	defer func() {
		_ = in.Close()
	}()

	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst %s: %w", dst, err)
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		return fmt.Errorf("sync %s: %w", dst, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", dst, err)
	}
	return nil
}

func LinkOrCopy(src, dst string) error {
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("remove existing %s: %w", dst, err)
	}
	if err := os.Link(src, dst); err == nil {
		return nil
	}
	return CopyFile(src, dst)
}

func WalkFiles(root string, cb func(path string, d os.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return cb(path, d)
	})
}
