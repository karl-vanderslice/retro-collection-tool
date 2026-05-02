package app

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	nextUIOwner     = "LoveRetro"
	nextUIRepo      = "NextUI"
	nextUIAPIBase   = "https://api.github.com/repos/" + nextUIOwner + "/" + nextUIRepo + "/releases"
	nextUIUserAgent = "retro-collection-tool/1.0"

	// httpClientTimeout caps GitHub API and release download calls.
	httpClientTimeout = 5 * time.Minute
)

// nextUIRelease is the subset of GitHub release fields we care about.
type nextUIRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []nextUIAsset `json:"assets"`
}

type nextUIAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// resolveNextUIRelease returns the tag name and all.zip download URL for the
// requested release tag ("latest" or a specific tag like "v6.10.0").
func resolveNextUIRelease(tag string) (resolvedTag, downloadURL string, err error) {
	var apiURL string
	if strings.ToLower(strings.TrimSpace(tag)) == "latest" {
		apiURL = nextUIAPIBase + "/latest"
	} else {
		apiURL = nextUIAPIBase + "/tags/" + strings.TrimSpace(tag)
	}

	client := &http.Client{Timeout: httpClientTimeout}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("build GitHub API request: %w", err)
	}
	req.Header.Set("User-Agent", nextUIUserAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("GitHub API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned %d for %s", resp.StatusCode, apiURL)
	}

	var rel nextUIRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", "", fmt.Errorf("decode GitHub release: %w", err)
	}
	if rel.TagName == "" {
		return "", "", errors.New("GitHub release response missing tag_name")
	}

	for _, asset := range rel.Assets {
		if strings.HasSuffix(asset.Name, "-all.zip") {
			return rel.TagName, asset.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("NextUI release %s has no *-all.zip asset", rel.TagName)
}

// cachedNextUIZipPath returns the path where the zip for the given tag would be cached.
func cachedNextUIZipPath(cacheDir, tag string) string {
	safe := strings.ReplaceAll(tag, "/", "_")
	return filepath.Join(cacheDir, "nextui-"+safe+".zip")
}

// downloadNextUIRelease fetches (or returns a cached copy of) the NextUI all.zip
// for the requested tag and returns the local zip path.
func downloadNextUIRelease(tag, cacheDir string, logf func(string, ...any)) (zipPath string, resolvedTag string, err error) {
	resolvedTag, downloadURL, err := resolveNextUIRelease(tag)
	if err != nil {
		return "", "", err
	}

	zipPath = cachedNextUIZipPath(cacheDir, resolvedTag)
	if _, statErr := os.Stat(zipPath); statErr == nil {
		logf("nextui release %s: using cached zip %s", resolvedTag, zipPath)
		return zipPath, resolvedTag, nil
	}

	logf("nextui release %s: downloading %s", resolvedTag, downloadURL)

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create cache dir %s: %w", cacheDir, err)
	}

	// Download to a temp file first so a failed download doesn't poison the cache.
	tmpFile, err := os.CreateTemp(cacheDir, "nextui-download-*.zip.tmp")
	if err != nil {
		return "", "", fmt.Errorf("create temp file for download: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	client := &http.Client{Timeout: httpClientTimeout}
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("build download request: %w", err)
	}
	req.Header.Set("User-Agent", nextUIUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("download NextUI release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	if _, err = io.Copy(tmpFile, resp.Body); err != nil {
		return "", "", fmt.Errorf("write download: %w", err)
	}
	if err = tmpFile.Sync(); err != nil {
		return "", "", fmt.Errorf("sync download: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return "", "", fmt.Errorf("close download: %w", err)
	}

	if err = os.Rename(tmpPath, zipPath); err != nil {
		return "", "", fmt.Errorf("move download to cache: %w", err)
	}

	logf("nextui release %s: cached to %s", resolvedTag, zipPath)
	return zipPath, resolvedTag, nil
}

// extractNextUIRelease extracts the all.zip into destination, stripping a
// single top-level directory if the zip contains one (so the SD-card files land
// directly in destination).
func extractNextUIRelease(zipPath, destination string, logf func(string, ...any)) error {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("create destination %s: %w", destination, err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip %s: %w", zipPath, err)
	}
	defer func() { _ = r.Close() }()

	// Detect whether all entries share a single top-level directory prefix.
	stripPrefix := detectZipStripPrefix(r.File)
	logf("nextui: extracting %s → %s (strip prefix: %q)", zipPath, destination, stripPrefix)

	for _, f := range r.File {
		if err := extractZipEntry(f, destination, stripPrefix); err != nil {
			return err
		}
	}
	return nil
}

// detectZipStripPrefix returns the common top-level directory prefix to strip
// from all zip entries, or "" if no common prefix is found.
func detectZipStripPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}

	// Collect the first path component of every entry.
	var prefix string
	for _, f := range files {
		name := filepath.ToSlash(f.Name)
		parts := strings.SplitN(name, "/", 2)
		if len(parts) < 2 || !f.FileInfo().IsDir() && parts[1] == "" {
			// There's at least one entry at the root level — no stripping.
			return ""
		}
		if prefix == "" {
			prefix = parts[0]
		} else if parts[0] != prefix {
			return ""
		}
	}

	if prefix != "" {
		return prefix + "/"
	}
	return ""
}

// extractZipEntry extracts a single zip file entry to destination, stripping
// the given prefix from its name.
func extractZipEntry(f *zip.File, destination, stripPrefix string) error {
	// Sanitize the entry name to prevent zip-slip path traversal.
	name := filepath.ToSlash(f.Name)
	name = strings.TrimPrefix(name, stripPrefix)
	if name == "" || strings.HasPrefix(name, "..") {
		return nil
	}
	// Reject absolute paths and path components with traversal sequences.
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return fmt.Errorf("zip entry %q contains path traversal", f.Name)
		}
	}

	destPath := filepath.Join(destination, filepath.FromSlash(name))

	if f.FileInfo().IsDir() {
		return os.MkdirAll(destPath, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create dir for %s: %w", destPath, err)
	}

	// Do not overwrite existing files (incremental-safe).
	if _, err := os.Stat(destPath); err == nil {
		return nil
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %s: %w", f.Name, err)
	}
	defer func() { _ = rc.Close() }()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", destPath, err)
	}
	if _, err := io.Copy(out, rc); err != nil {
		_ = out.Close()
		_ = os.Remove(destPath)
		return fmt.Errorf("extract %s: %w", f.Name, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", destPath, err)
	}
	return nil
}

// nextUIDefaultCacheDir returns the OS cache directory for NextUI zips.
func nextUIDefaultCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}
	return filepath.Join(base, "retro-collection-tool", "nextui"), nil
}
