package app

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	overlayProviderKrutzotrem = "krutzotrem"
	overlayProviderSkywalker  = "skywalker541"
)

type githubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubReleaseListEntry struct {
	TagName string               `json:"tag_name"`
	Assets  []githubReleaseAsset `json:"assets"`
}

func installNextUIOverlayProviders(rawProviders, destination string, logf func(string, ...any)) (int, error) {
	providers, err := parseNextUIOverlayProviders(rawProviders)
	if err != nil {
		return 0, err
	}
	if len(providers) == 0 {
		return 0, nil
	}

	cacheDir, err := nextUIOverlaysDefaultCacheDir()
	if err != nil {
		return 0, err
	}

	totalCopied := 0
	for _, provider := range providers {
		zipPath, resolved, err := downloadOverlayProviderArchive(provider, cacheDir, logf)
		if err != nil {
			return totalCopied, err
		}
		logf("nextui overlays %s: extracting %s", provider, resolved)
		copied, err := extractOverlayArchive(zipPath, destination)
		if err != nil {
			return totalCopied, err
		}
		totalCopied += copied
		logf("nextui overlays %s: copied %d files", provider, copied)
	}

	return totalCopied, nil
}

func parseNextUIOverlayProviders(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	providerSet := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(part)
		norm := normalizeOverlayProviderName(trimmed)
		if norm == "" {
			if trimmed != "" {
				return nil, fmt.Errorf("unknown --nextui-overlays provider %q (supported: %s,%s)", trimmed, overlayProviderKrutzotrem, overlayProviderSkywalker)
			}
			continue
		}
		switch norm {
		case overlayProviderKrutzotrem, overlayProviderSkywalker:
			providerSet[norm] = true
		default:
			return nil, fmt.Errorf("unknown --nextui-overlays provider %q (supported: %s,%s)", trimmed, overlayProviderKrutzotrem, overlayProviderSkywalker)
		}
	}

	providers := make([]string, 0, len(providerSet))
	for p := range providerSet {
		providers = append(providers, p)
	}
	sort.Strings(providers)
	return providers, nil
}

func normalizeOverlayProviderName(v string) string {
	norm := strings.ToLower(strings.TrimSpace(v))
	norm = strings.ReplaceAll(norm, "_", "")
	norm = strings.ReplaceAll(norm, "-", "")
	norm = strings.ReplaceAll(norm, " ", "")
	switch norm {
	case "krutzotrem", "krutz":
		return overlayProviderKrutzotrem
	case "skywalker541", "skywalker":
		return overlayProviderSkywalker
	default:
		return ""
	}
}

func downloadOverlayProviderArchive(provider, cacheDir string, logf func(string, ...any)) (zipPath string, resolved string, err error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create overlays cache dir %s: %w", cacheDir, err)
	}

	switch provider {
	case overlayProviderSkywalker:
		resolved = "main"
		zipPath = filepath.Join(cacheDir, "nextui-overlays-skywalker541-main.zip")
		if _, statErr := os.Stat(zipPath); statErr == nil {
			logf("nextui overlays %s: using cached zip %s", provider, zipPath)
			return zipPath, resolved, nil
		}
		url := "https://api.github.com/repos/SkyWalker541/TrimUI-Brick-Overlays/zipball/main"
		if err := downloadURLToPath(url, zipPath); err != nil {
			return "", "", fmt.Errorf("download overlays %s: %w", provider, err)
		}
		return zipPath, resolved, nil
	case overlayProviderKrutzotrem:
		tag, url, err := resolveKrutzotremOverlayAsset()
		if err != nil {
			return "", "", err
		}
		resolved = tag
		safeTag := strings.ReplaceAll(tag, "/", "_")
		zipPath = filepath.Join(cacheDir, "nextui-overlays-krutzotrem-"+safeTag+".zip")
		if _, statErr := os.Stat(zipPath); statErr == nil {
			logf("nextui overlays %s: using cached zip %s", provider, zipPath)
			return zipPath, resolved, nil
		}
		if err := downloadURLToPath(url, zipPath); err != nil {
			return "", "", fmt.Errorf("download overlays %s: %w", provider, err)
		}
		return zipPath, resolved, nil
	default:
		return "", "", fmt.Errorf("unsupported overlays provider: %s", provider)
	}
}

func resolveKrutzotremOverlayAsset() (tag string, downloadURL string, err error) {
	const apiURL = "https://api.github.com/repos/KrutzOtrem/Trimui-Brick-Overlays/releases"

	client := &http.Client{Timeout: httpClientTimeout}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("build overlays release request: %w", err)
	}
	req.Header.Set("User-Agent", nextUIUserAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("request overlays releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("overlays release API returned HTTP %d", resp.StatusCode)
	}

	var releases []githubReleaseListEntry
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", "", fmt.Errorf("decode overlays release response: %w", err)
	}

	for _, rel := range releases {
		for _, asset := range rel.Assets {
			name := strings.ToLower(strings.TrimSpace(asset.Name))
			if !strings.HasSuffix(name, ".zip") {
				continue
			}
			if strings.Contains(name, "psd") {
				continue
			}
			if strings.TrimSpace(asset.BrowserDownloadURL) == "" {
				continue
			}
			return rel.TagName, asset.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("no usable zip overlay asset found in %s releases", overlayProviderKrutzotrem)
}

func downloadURLToPath(url, outPath string) (err error) {
	tmpPath := outPath + ".tmp"

	client := &http.Client{Timeout: httpClientTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build download request: %w", err)
	}
	req.Header.Set("User-Agent", nextUIUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s returned HTTP %d", url, resp.StatusCode)
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file %s: %w", tmpPath, err)
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write download %s: %w", tmpPath, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync download %s: %w", tmpPath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close download %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, outPath); err != nil {
		return fmt.Errorf("move download to %s: %w", outPath, err)
	}
	return nil
}

func extractOverlayArchive(zipPath, destination string) (int, error) {
	overlayIndex, err := existingOverlayFolders(destination)
	if err != nil {
		return 0, err
	}
	if len(overlayIndex) == 0 {
		return 0, nil
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("open overlay zip %s: %w", zipPath, err)
	}
	defer func() { _ = r.Close() }()

	stripPrefix := detectZipStripPrefix(r.File)
	copied := 0

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		entry := filepath.ToSlash(strings.TrimPrefix(f.Name, stripPrefix))
		rel, ok := mapOverlayEntryPath(entry, overlayIndex)
		if !ok {
			continue
		}

		dstPath := filepath.Join(destination, rel)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return copied, err
		}
		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return copied, fmt.Errorf("open overlay zip entry %s: %w", f.Name, err)
		}

		out, err := os.Create(dstPath)
		if err != nil {
			_ = rc.Close()
			return copied, fmt.Errorf("create overlay path %s: %w", dstPath, err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			_ = rc.Close()
			_ = out.Close()
			_ = os.Remove(dstPath)
			return copied, fmt.Errorf("copy overlay file %s: %w", dstPath, err)
		}
		if err := rc.Close(); err != nil {
			_ = out.Close()
			return copied, fmt.Errorf("close overlay zip entry %s: %w", f.Name, err)
		}
		if err := out.Close(); err != nil {
			return copied, fmt.Errorf("close overlay output %s: %w", dstPath, err)
		}
		copied++
	}

	return copied, nil
}

func existingOverlayFolders(destination string) (map[string]string, error) {
	overlaysRoot := filepath.Join(destination, "Overlays")
	entries, err := os.ReadDir(overlaysRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("read overlay root %s: %w", overlaysRoot, err)
	}

	index := map[string]string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name == "" {
			continue
		}
		index[normalizeOverlayKey(name)] = name
	}
	return index, nil
}

func mapOverlayEntryPath(entryName string, overlayIndex map[string]string) (string, bool) {
	clean := path.Clean("/" + strings.TrimSpace(filepath.ToSlash(entryName)))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || clean == "." || strings.HasPrefix(clean, "..") {
		return "", false
	}

	parts := strings.Split(clean, "/")
	if len(parts) < 2 {
		return "", false
	}

	if strings.EqualFold(parts[0], "Overlays") {
		if len(parts) < 3 {
			return "", false
		}
		folder, ok := overlayIndex[normalizeOverlayKey(parts[1])]
		if !ok {
			return "", false
		}
		out := append([]string{"Overlays", folder}, parts[2:]...)
		return filepath.FromSlash(path.Clean(strings.Join(out, "/"))), true
	}

	folder, ok := overlayIndex[normalizeOverlayKey(parts[0])]
	if !ok {
		return "", false
	}
	out := append([]string{"Overlays", folder}, parts[1:]...)
	return filepath.FromSlash(path.Clean(strings.Join(out, "/"))), true
}

func normalizeOverlayKey(v string) string {
	up := strings.ToUpper(strings.TrimSpace(v))
	if up == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range up {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func nextUIOverlaysDefaultCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache dir: %w", err)
	}
	return filepath.Join(base, "retro-collection-tool", "nextui-overlays"), nil
}
