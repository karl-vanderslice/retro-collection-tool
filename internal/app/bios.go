package app

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
	"github.com/karl-vanderslice/retro-collection-tool/internal/fsutil"
	"github.com/karl-vanderslice/retro-collection-tool/internal/platform"
	"gopkg.in/yaml.v3"
)

//go:embed bios_catalog_default.yaml
var defaultBiosCatalogYAML []byte

type biosFlags struct {
	systemsCSV string
	allSystems bool
	strict     bool
}

type biosCatalog struct {
	Entries []biosCatalogEntry `yaml:"entries"`
}

type biosCatalogEntry struct {
	System      string              `yaml:"system"`
	Destination string              `yaml:"destination"`
	Required    bool                `yaml:"required"`
	Sources     []biosCatalogSource `yaml:"sources"`
}

type biosCatalogSource struct {
	Name   string `yaml:"name"`
	MD5    string `yaml:"md5"`
	SHA1   string `yaml:"sha1"`
	SHA256 string `yaml:"sha256"`
	CRC32  string `yaml:"crc32"`
}

type biosCandidate struct {
	Display  string
	Name     string
	MD5      string
	SHA1     string
	SHA256   string
	CRC32    string
	FilePath string
	ZipPath  string
	ZipEntry string
}

type biosSummary struct {
	Imported []string
	Missing  []string
	Unknown  []string
}

type biosScanStats struct {
	Scanned    int
	CacheHits  int
	CacheMiss  int
	ZipEntries int
	BadZips    int
}

var errBiosScanComplete = errors.New("bios scan complete")

type biosScanPlan struct {
	targetNames          map[string]bool
	entries              []biosCatalogEntry
	nameToEntryIndices   map[string][]int
	unresolvedEntryCount int
	resolvedEntries      map[int]bool
}

type biosHashCache struct {
	Version int                           `yaml:"version"`
	Entries map[string]biosHashCacheEntry `yaml:"entries"`
}

type biosHashCacheEntry struct {
	Size    int64  `yaml:"size"`
	ModUnix int64  `yaml:"mod_unix"`
	MD5     string `yaml:"md5"`
	SHA1    string `yaml:"sha1"`
	SHA256  string `yaml:"sha256"`
	CRC32   string `yaml:"crc32"`
}

func runBios(cfg *config.Config, g globalFlags, args []string) error {
	if !cfg.Features.EnableBios {
		return errors.New("bios workflow disabled in config.features.enable_bios")
	}

	var bf biosFlags
	fs := flag.NewFlagSet("bios", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&bf.systemsCSV, "systems", "", "comma-separated system slugs")
	fs.BoolVar(&bf.allSystems, "all-systems", false, "run all enabled systems")
	fs.BoolVar(&bf.strict, "strict", false, "fail when required BIOS entries are missing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ensureNoPositionalArgs("bios", fs.Args()); err != nil {
		return err
	}

	systems, err := platform.ExpandSystems([]string{bf.systemsCSV}, bf.allSystems, cfg)
	if err != nil {
		return err
	}

	emitInfo(g, "bios", "", "accepted", outputFields{"systems": strings.Join(systems, ","), "strict": bf.strict, "dry_run": g.dryRun})

	loadSpinner := newCommandSpinner(g, "bios", "catalog", "loading BIOS catalog")
	catalog, err := loadBiosCatalog(cfg)
	if err != nil {
		loadSpinner.Stop(false, err.Error())
		emitError(g, "bios", "catalog", "load failed", outputFields{"error": err.Error()})
		return err
	}
	if len(catalog.Entries) == 0 {
		loadSpinner.Stop(false, "catalog has no entries")
		emitError(g, "bios", "catalog", "catalog has no entries", nil)
		return errors.New("bios catalog has no entries")
	}
	loadSpinner.Stop(true, fmt.Sprintf("loaded entries=%d", len(catalog.Entries)))

	sourceRoots := resolveBiosSourceRoots(cfg)
	if len(sourceRoots) == 0 {
		return errors.New("bios.source_roots must include at least one directory")
	}
	plan := buildBiosScanPlan(catalog, systems)
	if plan.unresolvedEntryCount == 0 {
		emitInfo(g, "bios", "", "summary", outputFields{"imported": 0, "missing": 0, "unknown": 0})
		return nil
	}

	if g.verbose {
		emitInfo(g, "bios", "scan", "source roots", outputFields{"roots": strings.Join(sourceRoots, ", "), "recursive": true})
	} else {
		emitInfo(g, "bios", "scan", "source roots", outputFields{"count": len(sourceRoots), "recursive": true})
	}

	cachePath := filepath.Join(resolveCacheRoot(cfg), "bios_md5_cache.yaml")
	cacheSpinner := newCommandSpinner(g, "bios", "cache", "loading hash cache")
	hashCache, err := loadBiosHashCache(cachePath)
	if err != nil {
		cacheSpinner.Stop(false, err.Error())
		emitError(g, "bios", "cache", "load failed", outputFields{"error": err.Error(), "path": cachePath})
		return err
	}
	cacheSpinner.Stop(true, fmt.Sprintf("entries=%d path=%s", len(hashCache.Entries), cachePath))

	entriesNeedingScan, err := countEntriesNeedingScan(cfg, systems, catalog)
	if err != nil {
		emitError(g, "bios", "scan", "pre-check failed", outputFields{"error": err.Error()})
		return err
	}
	if entriesNeedingScan == 0 {
		emitInfo(g, "bios", "scan", "early check complete; all selected entries already satisfied in vault", nil)
		matchSpinner := newCommandSpinner(g, "bios", "match", "linking existing vault files")
		summary, err := syncBiosEntries(cfg, systems, catalog, nil, g, nil)
		if err != nil {
			matchSpinner.Stop(false, err.Error())
			emitError(g, "bios", "match", "import failed", outputFields{"error": err.Error()})
			return err
		}
		matchSpinner.Stop(true, fmt.Sprintf("imported=%d missing=%d", len(summary.Imported), len(summary.Missing)))
		emitInfo(g, "bios", "", "summary", outputFields{"imported": len(summary.Imported), "missing": len(summary.Missing), "unknown": len(summary.Unknown)})
		if bf.strict && len(summary.Missing) > 0 {
			return fmt.Errorf("bios strict mode failed: %d required entries missing", len(summary.Missing))
		}
		return nil
	}

	scanSpinner := newCommandSpinner(g, "bios", "scan", "walking files and hashing candidates")
	progress := func(stats biosScanStats, path string) {
		if !g.verbose {
			return
		}
		if stats.Scanned == 1 || stats.Scanned%500 == 0 {
			scanSpinner.Update(fmt.Sprintf("candidates=%d cache-hit=%d cache-miss=%d latest=%s", stats.Scanned, stats.CacheHits, stats.CacheMiss, filepath.Base(path)))
		}
	}
	onArchiveError := func(path string, err error) {
		if !g.verbose {
			return
		}
		emitInfo(g, "bios", "scan", "skipping unreadable archive", outputFields{"path": path, "error": err.Error()})
	}

	candidates, cacheDirty, scanStats, err := collectBiosCandidates(sourceRoots, g.verbose, progress, hashCache, plan, onArchiveError)
	if err != nil {
		scanSpinner.Stop(false, err.Error())
		emitError(g, "bios", "scan", "scan failed", outputFields{"error": err.Error()})
		return err
	}
	scanSpinner.Stop(true, fmt.Sprintf("candidates=%d cache-hit=%d cache-miss=%d zip-entries=%d bad-zips=%d", scanStats.Scanned, scanStats.CacheHits, scanStats.CacheMiss, scanStats.ZipEntries, scanStats.BadZips))
	if cacheDirty {
		saveCacheSpinner := newCommandSpinner(g, "bios", "cache", "writing md5 cache")
		if err := saveBiosHashCache(cachePath, hashCache); err != nil {
			saveCacheSpinner.Stop(false, err.Error())
			emitError(g, "bios", "cache", "write failed", outputFields{"error": err.Error(), "path": cachePath})
			return err
		}
		saveCacheSpinner.Stop(true, fmt.Sprintf("updated %s", cachePath))
	}

	matchSpinner := newCommandSpinner(g, "bios", "match", "matching catalog entries and importing")
	matchProgress := func(processed, total int, system, destination string) {
		if !g.verbose {
			return
		}
		if processed == 1 || processed%10 == 0 || processed == total {
			matchSpinner.Update(fmt.Sprintf("processed=%d/%d system=%s target=%s", processed, total, system, destination))
		}
	}

	summary, err := syncBiosEntries(cfg, systems, catalog, candidates, g, matchProgress)
	if err != nil {
		matchSpinner.Stop(false, err.Error())
		emitError(g, "bios", "match", "import failed", outputFields{"error": err.Error()})
		return err
	}
	matchSpinner.Stop(true, fmt.Sprintf("imported=%d missing=%d", len(summary.Imported), len(summary.Missing)))

	if g.verbose {
		for _, line := range summary.Imported {
			fmt.Println(line)
		}
		for _, line := range summary.Missing {
			fmt.Println(line)
		}
		for _, line := range summary.Unknown {
			fmt.Println(line)
		}
	} else if len(summary.Unknown) > 0 {
		emitInfo(g, "bios", "match", "skipped unknown candidates", outputFields{"count": len(summary.Unknown), "hint": "use --verbose for details"})
	}
	if len(summary.Missing) > 0 {
		emitInfo(g, "bios", "match", "missing catalog entries", outputFields{"count": len(summary.Missing), "hint": "use --verbose for details"})
	}

	emitInfo(g, "bios", "", "summary", outputFields{"imported": len(summary.Imported), "missing": len(summary.Missing), "unknown": len(summary.Unknown)})

	if bf.strict && len(summary.Missing) > 0 {
		return fmt.Errorf("bios strict mode failed: %d required entries missing", len(summary.Missing))
	}

	return nil
}

func loadBiosCatalog(cfg *config.Config) (*biosCatalog, error) {
	if p := strings.TrimSpace(cfg.Bios.CatalogFile); p != "" {
		data, err := readCatalogOverride(cfg, p)
		if err != nil {
			return nil, err
		}
		catalog, err := parseBiosCatalog(data)
		if err != nil {
			return nil, err
		}
		return catalog, nil
	}
	return parseBiosCatalog(defaultBiosCatalogYAML)
}

func readCatalogOverride(cfg *config.Config, p string) ([]byte, error) {
	if filepath.IsAbs(p) {
		return os.ReadFile(p)
	}
	if b, err := os.ReadFile(p); err == nil {
		return b, nil
	}
	resolved := cfg.ResolvePath(p)
	return os.ReadFile(resolved)
}

func parseBiosCatalog(data []byte) (*biosCatalog, error) {
	var catalog biosCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("parse bios catalog: %w", err)
	}

	for i, e := range catalog.Entries {
		if strings.TrimSpace(e.System) == "" {
			return nil, fmt.Errorf("bios catalog entry %d missing system", i)
		}
		if strings.TrimSpace(e.Destination) == "" {
			return nil, fmt.Errorf("bios catalog entry %d missing destination", i)
		}
		if !isSafeRelativePath(e.Destination) {
			return nil, fmt.Errorf("bios catalog entry %d has unsafe destination path: %s", i, e.Destination)
		}
		if len(e.Sources) == 0 {
			return nil, fmt.Errorf("bios catalog entry %d has no sources", i)
		}
		for j, s := range e.Sources {
			if strings.TrimSpace(s.Name) == "" {
				return nil, fmt.Errorf("bios catalog entry %d source %d missing name", i, j)
			}
			md5Value := strings.TrimSpace(s.MD5)
			if md5Value != "" && !isMD5Hex(md5Value) {
				return nil, fmt.Errorf("bios catalog entry %d source %d has invalid md5", i, j)
			}
			sha1Value := strings.TrimSpace(s.SHA1)
			if sha1Value != "" && !isSHA1Hex(sha1Value) {
				return nil, fmt.Errorf("bios catalog entry %d source %d has invalid sha1", i, j)
			}
			sha256Value := strings.TrimSpace(s.SHA256)
			if sha256Value != "" && !isSHA256Hex(sha256Value) {
				return nil, fmt.Errorf("bios catalog entry %d source %d has invalid sha256", i, j)
			}
			crc32Value := strings.TrimSpace(s.CRC32)
			if crc32Value != "" && !isCRC32Hex(crc32Value) {
				return nil, fmt.Errorf("bios catalog entry %d source %d has invalid crc32", i, j)
			}
		}
	}

	return &catalog, nil
}

func resolveBiosSourceRoots(cfg *config.Config) []string {
	out := make([]string, 0, len(cfg.Bios.SourceRoots)+2)
	out = append(out, cfg.ResolvePath("bios"))
	for _, root := range cfg.Bios.SourceRoots {
		trimmed := strings.TrimSpace(root)
		if trimmed == "" {
			continue
		}
		out = append(out, cfg.ResolvePath(trimmed))
	}
	return dedupePreserveOrder(out)
}

func collectBiosCandidates(
	sourceRoots []string,
	verbose bool,
	progress func(stats biosScanStats, path string),
	hashCache *biosHashCache,
	plan biosScanPlan,
	onArchiveError func(path string, err error),
) ([]biosCandidate, bool, biosScanStats, error) {
	out := make([]biosCandidate, 0)
	stats := biosScanStats{}
	scanned := 0
	cacheDirty := false

	for _, root := range sourceRoots {
		info, err := os.Stat(root)
		if os.IsNotExist(err) {
			if verbose {
				fmt.Printf("[bios] source root missing, skipping: %s\n", root)
			}
			continue
		}
		if err != nil {
			return nil, cacheDirty, stats, err
		}
		if !info.IsDir() {
			continue
		}

		err = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".zip" {
				zipItems, err := collectBiosCandidatesFromZip(path, plan.targetNames)
				if err != nil {
					stats.BadZips++
					if onArchiveError != nil {
						onArchiveError(path, err)
					}
					return nil
				}
				out = append(out, zipItems...)
				for _, candidate := range zipItems {
					scanned++
					stats.ZipEntries++
					markBiosPlanCandidate(&plan, candidate)
					if biosPlanComplete(plan) {
						stats.Scanned = scanned
						if progress != nil {
							progress(stats, path)
						}
						return errBiosScanComplete
					}
				}
				stats.Scanned = scanned
				if progress != nil {
					progress(stats, path)
				}
				return nil
			}

			name := strings.ToLower(filepath.Base(path))
			if !plan.targetNames[name] {
				return nil
			}

			hashes, changed, err := hashesPathCached(path, hashCache)
			if err != nil {
				return err
			}
			if changed {
				cacheDirty = true
				stats.CacheMiss++
			} else {
				stats.CacheHits++
			}
			out = append(out, biosCandidate{
				Display:  path,
				Name:     name,
				MD5:      hashes.MD5,
				SHA1:     hashes.SHA1,
				SHA256:   hashes.SHA256,
				CRC32:    hashes.CRC32,
				FilePath: path,
			})
			markBiosPlanCandidate(&plan, out[len(out)-1])
			scanned++
			stats.Scanned = scanned
			if progress != nil {
				progress(stats, path)
			}
			if biosPlanComplete(plan) {
				return errBiosScanComplete
			}
			return nil
		})
		if err != nil {
			if errors.Is(err, errBiosScanComplete) {
				break
			}
			return nil, cacheDirty, stats, err
		}
		if biosPlanComplete(plan) {
			break
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Display < out[j].Display
	})
	return out, cacheDirty, stats, nil
}

func collectBiosCandidatesFromZip(zipPath string, targetNames map[string]bool) ([]biosCandidate, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open zip %s: %w", zipPath, err)
	}
	defer func() {
		_ = r.Close()
	}()

	items := make([]biosCandidate, 0, len(r.File))
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		cleanName := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return nil, fmt.Errorf("zip %s contains unsafe path: %s", zipPath, f.Name)
		}
		base := strings.ToLower(filepath.Base(cleanName))
		if !targetNames[base] {
			continue
		}

		in, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip entry %s in %s: %w", f.Name, zipPath, err)
		}
		hashes, err := computeHashesFromReader(in)
		if err != nil {
			_ = in.Close()
			return nil, fmt.Errorf("hash zip entry %s in %s: %w", f.Name, zipPath, err)
		}
		if err := in.Close(); err != nil {
			return nil, fmt.Errorf("close zip entry %s in %s: %w", f.Name, zipPath, err)
		}

		items = append(items, biosCandidate{
			Display:  fmt.Sprintf("%s:%s", zipPath, cleanName),
			Name:     base,
			MD5:      hashes.MD5,
			SHA1:     hashes.SHA1,
			SHA256:   hashes.SHA256,
			CRC32:    hashes.CRC32,
			ZipPath:  zipPath,
			ZipEntry: cleanName,
		})
	}
	return items, nil
}

func buildBiosScanPlan(catalog *biosCatalog, systems []string) biosScanPlan {
	systemSet := map[string]bool{}
	for _, s := range systems {
		systemSet[strings.ToLower(strings.TrimSpace(s))] = true
	}
	plan := biosScanPlan{
		targetNames:          map[string]bool{},
		entries:              make([]biosCatalogEntry, 0),
		nameToEntryIndices:   map[string][]int{},
		unresolvedEntryCount: 0,
		resolvedEntries:      map[int]bool{},
	}
	for _, entry := range catalog.Entries {
		systemKey := strings.ToLower(strings.TrimSpace(entry.System))
		if !systemSet[systemKey] {
			continue
		}
		entryIdx := len(plan.entries)
		plan.entries = append(plan.entries, entry)
		plan.unresolvedEntryCount++
		for _, src := range entry.Sources {
			name := strings.ToLower(strings.TrimSpace(src.Name))
			plan.targetNames[name] = true
			plan.nameToEntryIndices[name] = append(plan.nameToEntryIndices[name], entryIdx)
		}
	}
	return plan
}

func markBiosPlanCandidate(plan *biosScanPlan, candidate biosCandidate) {
	if plan == nil {
		return
	}
	indices := plan.nameToEntryIndices[strings.ToLower(candidate.Name)]
	for _, idx := range indices {
		if plan.resolvedEntries[idx] {
			continue
		}
		if !entryHasMatchingSource(plan.entries[idx], candidate) {
			continue
		}
		plan.resolvedEntries[idx] = true
		if plan.unresolvedEntryCount > 0 {
			plan.unresolvedEntryCount--
		}
	}
}

func biosPlanComplete(plan biosScanPlan) bool {
	return plan.unresolvedEntryCount == 0
}

func findExistingVaultMatch(vaultPath string, entry biosCatalogEntry) (*biosCandidate, error) {
	info, err := os.Stat(vaultPath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, nil
	}

	hashes, err := hashesPath(vaultPath)
	if err != nil {
		return nil, err
	}
	candidate := biosCandidate{
		Display:  vaultPath,
		Name:     filepath.Base(vaultPath),
		MD5:      hashes.MD5,
		SHA1:     hashes.SHA1,
		SHA256:   hashes.SHA256,
		CRC32:    hashes.CRC32,
		FilePath: vaultPath,
	}
	for _, src := range entry.Sources {
		if sourceMatchesCandidate(src, candidate) {
			return &candidate, nil
		}
	}
	return nil, nil
}

func countEntriesNeedingScan(cfg *config.Config, systems []string, catalog *biosCatalog) (int, error) {
	systemSet := map[string]bool{}
	for _, s := range systems {
		systemSet[strings.ToLower(strings.TrimSpace(s))] = true
	}

	count := 0
	for _, entry := range catalog.Entries {
		systemKey := strings.ToLower(strings.TrimSpace(entry.System))
		if !systemSet[systemKey] {
			continue
		}
		if _, ok := cfg.Systems[systemKey]; !ok {
			continue
		}
		vaultDst := filepath.Join(cfg.ResolvePath(cfg.Paths.VaultBios), entry.Destination)
		match, err := findExistingVaultMatch(vaultDst, entry)
		if err != nil {
			return 0, err
		}
		if match == nil {
			count++
		}
	}

	return count, nil
}

func syncBiosEntries(cfg *config.Config, systems []string, catalog *biosCatalog, candidates []biosCandidate, g globalFlags, progress func(processed, total int, system, destination string)) (*biosSummary, error) {
	systemSet := make(map[string]bool, len(systems))
	for _, s := range systems {
		systemSet[s] = true
	}

	nameToCandidates := map[string][]biosCandidate{}
	for _, c := range candidates {
		nameToCandidates[strings.ToLower(c.Name)] = append(nameToCandidates[strings.ToLower(c.Name)], c)
	}

	usedCandidates := map[string]bool{}
	summary := &biosSummary{}

	total := 0
	for _, entry := range catalog.Entries {
		systemKey := strings.ToLower(strings.TrimSpace(entry.System))
		if systemSet[systemKey] {
			total++
		}
	}
	processed := 0

	for _, entry := range catalog.Entries {
		systemKey := strings.ToLower(strings.TrimSpace(entry.System))
		if !systemSet[systemKey] {
			continue
		}
		processed++
		if progress != nil {
			progress(processed, total, systemKey, entry.Destination)
		}

		sysCfg, ok := cfg.Systems[systemKey]
		if !ok {
			return nil, fmt.Errorf("bios catalog references unknown system: %s", systemKey)
		}

		vaultDst := filepath.Join(cfg.ResolvePath(cfg.Paths.VaultBios), entry.Destination)
		libraryName := filepath.Base(filepath.Clean(entry.Destination))
		libraryDst := filepath.Join(cfg.ResolvePath(cfg.Paths.RommLibraryBios), sysCfg.RommSlug, libraryName)
		vaultMatch, err := findExistingVaultMatch(vaultDst, entry)
		if err != nil {
			return nil, err
		}
		if vaultMatch != nil {
			if g.dryRun {
				summary.Imported = append(summary.Imported, fmt.Sprintf("[dry-run] bios link %s -> %s", vaultDst, libraryDst))
			} else {
				if err := fsutil.LinkOrCopy(vaultDst, libraryDst); err != nil {
					return nil, err
				}
				summary.Imported = append(summary.Imported, fmt.Sprintf("[bios] reused %s (linked %s)", vaultDst, libraryDst))
			}
			continue
		}

		match, mismatchedNames := findCatalogEntryMatch(entry, nameToCandidates)
		if match == nil {
			if entry.Required {
				summary.Missing = append(summary.Missing, fmt.Sprintf("[bios] missing required %s/%s", sysCfg.RommSlug, libraryName))
			}
			summary.Missing = append(summary.Missing, mismatchedNames...)
			continue
		}

		usedCandidates[match.Display] = true
		if g.dryRun {
			summary.Imported = append(summary.Imported, fmt.Sprintf("[dry-run] bios import %s -> %s (link %s)", match.Display, vaultDst, libraryDst))
			continue
		}

		if err := copyBiosCandidate(*match, vaultDst); err != nil {
			return nil, err
		}
		if err := fsutil.LinkOrCopy(vaultDst, libraryDst); err != nil {
			return nil, err
		}
		summary.Imported = append(summary.Imported, fmt.Sprintf("[bios] imported %s -> %s (linked %s)", match.Display, vaultDst, libraryDst))
	}

	for _, c := range candidates {
		if usedCandidates[c.Display] {
			continue
		}
		summary.Unknown = append(summary.Unknown, fmt.Sprintf("[bios] skipped unknown %s", c.Display))
	}

	return summary, nil
}

func findCatalogEntryMatch(entry biosCatalogEntry, byName map[string][]biosCandidate) (*biosCandidate, []string) {
	mismatches := make([]string, 0)
	for _, src := range entry.Sources {
		name := strings.ToLower(strings.TrimSpace(src.Name))
		for _, candidate := range byName[name] {
			if sourceMatchesCandidate(src, candidate) {
				m := candidate
				return &m, mismatches
			}
			if sourceHasHashes(src) {
				mismatches = append(mismatches, fmt.Sprintf("[bios] hash mismatch for %s: got %s expected %s (%s)", src.Name, candidateHashSummary(candidate), sourceHashSummary(src), candidate.Display))
			}
		}
	}
	return nil, mismatches
}

func copyBiosCandidate(c biosCandidate, dst string) error {
	if c.FilePath != "" {
		return fsutil.CopyFile(c.FilePath, dst)
	}
	if c.ZipPath != "" {
		return copyZipEntry(c.ZipPath, c.ZipEntry, dst)
	}
	return errors.New("invalid bios candidate source")
}

func copyZipEntry(zipPath, entryName, dst string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip %s: %w", zipPath, err)
	}
	defer func() {
		_ = r.Close()
	}()

	for _, f := range r.File {
		if filepath.Clean(f.Name) != entryName {
			continue
		}
		in, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s in %s: %w", entryName, zipPath, err)
		}
		defer func() {
			_ = in.Close()
		}()

		if err := fsutil.EnsureDir(filepath.Dir(dst)); err != nil {
			return err
		}
		out, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("create dst %s: %w", dst, err)
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return fmt.Errorf("copy zip entry %s to %s: %w", entryName, dst, err)
		}
		if err := out.Close(); err != nil {
			return fmt.Errorf("close dst %s: %w", dst, err)
		}
		return nil
	}

	return fmt.Errorf("zip entry not found: %s in %s", entryName, zipPath)
}

func hashesPath(path string) (biosHashCacheEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return biosHashCacheEntry{}, err
	}
	defer func() {
		_ = f.Close()
	}()
	return computeHashesFromReader(f)
}

func computeHashesFromReader(r io.Reader) (biosHashCacheEntry, error) {
	hMD5 := md5.New()
	hSHA1 := sha1.New()
	hSHA256 := sha256.New()
	hCRC32 := crc32.NewIEEE()
	if _, err := io.Copy(io.MultiWriter(hMD5, hSHA1, hSHA256, hCRC32), r); err != nil {
		return biosHashCacheEntry{}, err
	}
	return biosHashCacheEntry{
		MD5:    hex.EncodeToString(hMD5.Sum(nil)),
		SHA1:   hex.EncodeToString(hSHA1.Sum(nil)),
		SHA256: hex.EncodeToString(hSHA256.Sum(nil)),
		CRC32:  fmt.Sprintf("%08x", hCRC32.Sum32()),
	}, nil
}

func isMD5Hex(v string) bool {
	return isHexLen(v, 32)
}

func isSHA1Hex(v string) bool {
	return isHexLen(v, 40)
}

func isSHA256Hex(v string) bool {
	return isHexLen(v, 64)
}

func isCRC32Hex(v string) bool {
	return isHexLen(v, 8)
}

func isHexLen(v string, length int) bool {
	v = strings.TrimSpace(v)
	if len(v) != length {
		return false
	}
	for _, r := range v {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func sourceHasHashes(src biosCatalogSource) bool {
	return strings.TrimSpace(src.MD5) != "" || strings.TrimSpace(src.SHA1) != "" || strings.TrimSpace(src.SHA256) != "" || strings.TrimSpace(src.CRC32) != ""
}

func sourceMatchesCandidate(src biosCatalogSource, candidate biosCandidate) bool {
	if !strings.EqualFold(strings.TrimSpace(src.Name), strings.TrimSpace(candidate.Name)) {
		return false
	}
	if v := strings.TrimSpace(src.MD5); v != "" && !strings.EqualFold(v, strings.TrimSpace(candidate.MD5)) {
		return false
	}
	if v := strings.TrimSpace(src.SHA1); v != "" && !strings.EqualFold(v, strings.TrimSpace(candidate.SHA1)) {
		return false
	}
	if v := strings.TrimSpace(src.SHA256); v != "" && !strings.EqualFold(v, strings.TrimSpace(candidate.SHA256)) {
		return false
	}
	if v := strings.TrimSpace(src.CRC32); v != "" && !strings.EqualFold(v, strings.TrimSpace(candidate.CRC32)) {
		return false
	}
	return true
}

func sourceHashSummary(src biosCatalogSource) string {
	parts := make([]string, 0, 4)
	if v := strings.TrimSpace(src.MD5); v != "" {
		parts = append(parts, "md5="+strings.ToLower(v))
	}
	if v := strings.TrimSpace(src.SHA1); v != "" {
		parts = append(parts, "sha1="+strings.ToLower(v))
	}
	if v := strings.TrimSpace(src.SHA256); v != "" {
		parts = append(parts, "sha256="+strings.ToLower(v))
	}
	if v := strings.TrimSpace(src.CRC32); v != "" {
		parts = append(parts, "crc32="+strings.ToLower(v))
	}
	if len(parts) == 0 {
		return "name-only"
	}
	return strings.Join(parts, " ")
}

func candidateHashSummary(c biosCandidate) string {
	parts := make([]string, 0, 4)
	if v := strings.TrimSpace(c.MD5); v != "" {
		parts = append(parts, "md5="+strings.ToLower(v))
	}
	if v := strings.TrimSpace(c.SHA1); v != "" {
		parts = append(parts, "sha1="+strings.ToLower(v))
	}
	if v := strings.TrimSpace(c.SHA256); v != "" {
		parts = append(parts, "sha256="+strings.ToLower(v))
	}
	if v := strings.TrimSpace(c.CRC32); v != "" {
		parts = append(parts, "crc32="+strings.ToLower(v))
	}
	if len(parts) == 0 {
		return "no-hash"
	}
	return strings.Join(parts, " ")
}

func loadBiosHashCache(path string) (*biosHashCache, error) {
	cache := &biosHashCache{Version: 2, Entries: map[string]biosHashCacheEntry{}}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cache, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read bios hash cache %s: %w", path, err)
	}
	if err := yaml.Unmarshal(b, cache); err != nil {
		return nil, fmt.Errorf("parse bios hash cache %s: %w", path, err)
	}
	if cache.Entries == nil {
		cache.Entries = map[string]biosHashCacheEntry{}
	}
	if cache.Version != 2 {
		cache.Version = 2
		cache.Entries = map[string]biosHashCacheEntry{}
	}
	return cache, nil
}

func saveBiosHashCache(path string, cache *biosHashCache) error {
	if err := fsutil.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	b, err := yaml.Marshal(cache)
	if err != nil {
		return fmt.Errorf("marshal bios hash cache: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write bios hash cache %s: %w", path, err)
	}
	return nil
}

func hashesPathCached(path string, cache *biosHashCache) (biosHashCacheEntry, bool, error) {
	if cache == nil {
		h, err := hashesPath(path)
		return h, false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return biosHashCacheEntry{}, false, err
	}
	key := filepath.Clean(path)
	if cached, ok := cache.Entries[key]; ok {
		if cached.Size == info.Size() && cached.ModUnix == info.ModTime().Unix() && isMD5Hex(cached.MD5) && isSHA1Hex(cached.SHA1) && isSHA256Hex(cached.SHA256) && isCRC32Hex(cached.CRC32) {
			return biosHashCacheEntry{
				MD5:    strings.ToLower(cached.MD5),
				SHA1:   strings.ToLower(cached.SHA1),
				SHA256: strings.ToLower(cached.SHA256),
				CRC32:  strings.ToLower(cached.CRC32),
			}, false, nil
		}
	}
	hashes, err := hashesPath(path)
	if err != nil {
		return biosHashCacheEntry{}, false, err
	}
	cache.Entries[key] = biosHashCacheEntry{
		Size:    info.Size(),
		ModUnix: info.ModTime().Unix(),
		MD5:     hashes.MD5,
		SHA1:    hashes.SHA1,
		SHA256:  hashes.SHA256,
		CRC32:   hashes.CRC32,
	}
	return hashes, true, nil
}

func entryHasMatchingSource(entry biosCatalogEntry, candidate biosCandidate) bool {
	for _, src := range entry.Sources {
		if sourceMatchesCandidate(src, candidate) {
			return true
		}
	}
	return false
}

func isSafeRelativePath(p string) bool {
	clean := filepath.Clean(strings.TrimSpace(p))
	if clean == "." || clean == "" {
		return false
	}
	if filepath.IsAbs(clean) {
		return false
	}
	return !strings.HasPrefix(clean, "..")
}
