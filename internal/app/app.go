package app

import (
	"archive/zip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
	"github.com/karl-vanderslice/retro-collection-tool/internal/fsutil"
	"github.com/karl-vanderslice/retro-collection-tool/internal/igir"
	"github.com/karl-vanderslice/retro-collection-tool/internal/platform"
	"github.com/karl-vanderslice/retro-collection-tool/internal/xdg"
)

const (
	configEnvVar = "RETRO_COLLECTION_TOOL_CONFIG"
	appName      = "retro-collection-tool"
)

var regionGroupRe = regexp.MustCompile(`\([^)]*\)`)

type globalFlags struct {
	configPath string
	dryRun     bool
	verbose    bool
}

func Run(args []string) error {
	if len(args) == 0 {
		printRootUsage()
		return errors.New("no command provided")
	}

	globals, rest, err := parseGlobalFlags(args)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		printRootUsage()
		return errors.New("no command provided")
	}

	configPath, err := resolveConfigPath(globals.configPath)
	if err != nil {
		return err
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config %s: %w", configPath, err)
	}
	if globals.verbose {
		fmt.Printf("[config] using %s\n", configPath)
	}

	ctx := context.Background()
	command := rest[0]
	runner := igir.NewRunner(cfg)

	switch command {
	case "sync":
		return runSync(ctx, cfg, runner, globals, rest[1:])
	case "hacks":
		return runHacks(ctx, cfg, runner, globals, rest[1:])
	case "bios":
		return runBiosStub(cfg)
	case "redump":
		return runRedumpStub(cfg)
	case "arcade":
		return runArcadeStub(cfg)
	case "cache":
		return runCache(cfg, rest[1:])
	case "export":
		return runExport(cfg, globals, rest[1:])
	case "bootstrap":
		return runBootstrap(cfg)
	case "systems":
		return runSystems(cfg)
	case "version":
		fmt.Println("retro-collection-tool dev")
		return nil
	case "help", "-h", "--help":
		printRootUsage()
		return nil
	default:
		printRootUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func parseGlobalFlags(args []string) (globalFlags, []string, error) {
	var g globalFlags
	fs := flag.NewFlagSet("retro-collection-tool", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&g.configPath, "config", "", "path to YAML config (default discovery: cwd then XDG)")
	fs.BoolVar(&g.dryRun, "dry-run", false, "print planned actions without making changes")
	fs.BoolVar(&g.verbose, "verbose", false, "verbose logs")

	if err := fs.Parse(args); err != nil {
		return g, nil, err
	}
	return g, fs.Args(), nil
}

func resolveConfigPath(flagPath string) (string, error) {
	if p := strings.TrimSpace(flagPath); p != "" {
		return p, nil
	}
	if p := strings.TrimSpace(os.Getenv(configEnvVar)); p != "" {
		return p, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get cwd: %w", err)
	}

	candidates := []string{
		filepath.Join(cwd, "retro-collection-tool.yaml"),
		filepath.Join(cwd, ".retro-collection-tool.yaml"),
		filepath.Join(cwd, "config", "retro-collection-tool.yaml"),
	}

	if configHome := xdg.ConfigHome(); configHome != "" {
		candidates = append(candidates,
			filepath.Join(configHome, appName, "config.yaml"),
			filepath.Join(configHome, appName, "config.yml"),
			filepath.Join(configHome, appName, "retro-collection-tool.yaml"),
		)
	}

	for _, c := range candidates {
		if fileExists(c) {
			return c, nil
		}
	}

	return "", fmt.Errorf("no config found; set --config or %s (searched: %s)", configEnvVar, strings.Join(candidates, ", "))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

type syncFlags struct {
	systemsCSV string
	allSystems bool
	compress   bool
}

func runSync(ctx context.Context, cfg *config.Config, runner *igir.Runner, g globalFlags, args []string) error {
	var sf syncFlags
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&sf.systemsCSV, "systems", "", "comma-separated system slugs")
	fs.BoolVar(&sf.allSystems, "all-systems", false, "run all enabled systems")
	fs.BoolVar(&sf.compress, "compress", false, "enable zip output if configured")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ensureNoPositionalArgs("sync", fs.Args()); err != nil {
		return err
	}

	systems, err := platform.ExpandSystems([]string{sf.systemsCSV}, sf.allSystems, cfg)
	if err != nil {
		return err
	}

	for _, system := range systems {
		if err := syncRetailSystem(ctx, cfg, runner, g, system, sf.compress); err != nil {
			return err
		}
	}
	return nil
}

func syncRetailSystem(ctx context.Context, cfg *config.Config, runner *igir.Runner, g globalFlags, system string, compress bool) error {
	sysCfg := cfg.Systems[system]
	rommDir := filepath.Join(cfg.ResolvePath(cfg.Paths.RommLibraryRoms), sysCfg.RommSlug)
	if err := fsutil.EnsureDir(rommDir); err != nil {
		return err
	}

	datDir := cfg.ResolvePath(cfg.Paths.DatsNoIntro1G1R)
	datPath, err := fsutil.FindLatestDAT(datDir, sysCfg.RetailDatPattern)
	if err != nil {
		return err
	}

	cachePath := filepath.Join(resolveCacheRoot(cfg), cfg.Igir.CacheRetailFile)
	if err := fsutil.EnsureDir(filepath.Dir(cachePath)); err != nil {
		return err
	}

	args := []string{"link", "playlist", "clean", "--dat", datPath}
	args = append(args,
		"--input", cfg.ResolvePath(cfg.Paths.VaultNoIntro),
		"--input", rommDir,
		"--output", rommDir,
		"--clean-exclude", "hack/**",
		"--link-mode", "hardlink",
		"--single",
		"--only-retail",
		"--no-bios",
		"--no-unlicensed",
		"--no-homebrew",
		"--no-aftermarket",
		"--no-program",
		"--prefer-parent",
		"--prefer-region", strings.Join(cfg.Igir.PreferRegion, ","),
		"--prefer-language", strings.Join(cfg.Igir.PreferLanguage, ","),
		"--merge-discs",
		"--overwrite-invalid",
		"--input-checksum-min", cfg.Igir.InputChecksumMin,
		"--cache-path", cachePath,
	)

	for _, a := range cfg.Igir.RetailBaseArgs {
		if strings.TrimSpace(a) != "" {
			args = append(args, a)
		}
	}

	if compress {
		if !cfg.Igir.AllowCompressionZip {
			return errors.New("compression requested but config.igir.allow_compression_zip is false")
		}
		args = append(args, "--zip")
	}
	if g.verbose {
		fmt.Printf("[sync:%s] dat=%s output=%s\n", system, datPath, rommDir)
	}
	if g.dryRun {
		fmt.Printf("[dry-run] igir %s\n", strings.Join(args, " "))
		return nil
	}
	return runner.Run(ctx, args)
}

type hacksFlags struct {
	systemsCSV string
	allSystems bool
}

func runHacks(ctx context.Context, cfg *config.Config, runner *igir.Runner, g globalFlags, args []string) error {
	var hf hacksFlags
	fs := flag.NewFlagSet("hacks", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&hf.systemsCSV, "systems", "", "comma-separated system slugs")
	fs.BoolVar(&hf.allSystems, "all-systems", false, "run all enabled systems")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ensureNoPositionalArgs("hacks", fs.Args()); err != nil {
		return err
	}

	systems, err := platform.ExpandSystems([]string{hf.systemsCSV}, hf.allSystems, cfg)
	if err != nil {
		return err
	}

	for _, system := range systems {
		if err := runHacksSystem(ctx, cfg, runner, g, system); err != nil {
			return err
		}
	}
	return nil
}

func runHacksSystem(ctx context.Context, cfg *config.Config, runner *igir.Runner, g globalFlags, system string) error {
	sysCfg := cfg.Systems[system]
	systemHacksDir := filepath.Join(cfg.ResolvePath(cfg.Paths.HacksSource), system)
	if _, statErr := os.Stat(systemHacksDir); os.IsNotExist(statErr) {
		if g.verbose {
			fmt.Printf("[hacks:%s] no hacks directory, skipping\n", system)
		}
		return nil
	} else if statErr != nil {
		return statErr
	}

	baseOutput := filepath.Join(cfg.ResolvePath(cfg.Paths.RommLibraryRoms), sysCfg.RommSlug)
	if err := fsutil.EnsureDir(baseOutput); err != nil {
		return err
	}

	datPath, err := fsutil.FindLatestDAT(cfg.ResolvePath(cfg.Paths.DatsNoIntroRaw), sysCfg.HackDatPattern)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(systemHacksDir)
	if err != nil {
		return fmt.Errorf("read hacks dir %s: %w", systemHacksDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		hackName := entry.Name()
		hackPath := filepath.Join(systemHacksDir, hackName)

		workRoot := filepath.Join(resolveCacheRoot(cfg), "work", system, sanitizeName(hackName))
		inDir := filepath.Join(workRoot, "in")
		patchDir := filepath.Join(workRoot, "patch")
		outDir := filepath.Join(workRoot, "out")

		if err := fsutil.RemoveIfExists(workRoot); err != nil {
			return err
		}
		if err := fsutil.EnsureDir(inDir); err != nil {
			return err
		}
		if err := fsutil.EnsureDir(patchDir); err != nil {
			return err
		}
		if err := fsutil.EnsureDir(outDir); err != nil {
			return err
		}

		if err := stageHackFiles(hackPath, inDir, patchDir); err != nil {
			return err
		}

		cachePath := filepath.Join(resolveCacheRoot(cfg), cfg.Igir.CacheHacksFile)
		args := []string{
			"copy",
			"--dat", datPath,
			"--input", inDir,
			"--patch", patchDir,
			"--output", outDir,
			"--patch-only",
			"--overwrite-invalid",
			"--cache-path", cachePath,
		}
		if g.verbose {
			fmt.Printf("[hacks:%s] processing %s with dat=%s\n", system, hackName, datPath)
		}
		if g.dryRun {
			gameDir, err := resolveHackGameDir(baseOutput, inDir, hackName)
			if err != nil {
				return err
			}
			fmt.Printf("[dry-run] target hack path %s\n", filepath.Join(gameDir, "hack", sanitizeName(hackName)))
			fmt.Printf("[dry-run] igir %s\n", strings.Join(args, " "))
			continue
		}
		if err := runner.Run(ctx, args); err != nil {
			return err
		}

		patched, err := firstPatchedROMInDir(outDir)
		if err != nil {
			return fmt.Errorf("hack %s produced no output: %w", hackName, err)
		}

		gameDir, err := resolveHackGameDir(baseOutput, inDir, hackName)
		if err != nil {
			return err
		}

		ext := filepath.Ext(patched)
		targetDir := filepath.Join(gameDir, "hack")
		targetFile := filepath.Join(targetDir, sanitizeName(hackName)+ext)
		if err := fsutil.CopyFile(patched, targetFile); err != nil {
			return err
		}

		// Place an unaltered ROM alongside the hack to satisfy ROMM hack folder expectations.
		baseROM, err := firstROMInDir(inDir)
		if err == nil {
			baseDst := filepath.Join(gameDir, filepath.Base(baseROM))
			if err := fsutil.LinkOrCopy(baseROM, baseDst); err != nil {
				return err
			}
		}
	}
	return nil
}

func stageHackFiles(srcDir, inDir, patchDir string) error {
	if err := fsutil.WalkFiles(srcDir, func(path string, _ os.DirEntry) error {
		if strings.ToLower(filepath.Ext(path)) != ".zip" {
			return nil
		}
		return unzipInto(path, srcDir)
	}); err != nil {
		return err
	}

	return fsutil.WalkFiles(srcDir, func(path string, _ os.DirEntry) error {
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".zip":
			return nil
		case ".ips", ".bps", ".ups", ".xdelta":
			return fsutil.CopyFile(path, filepath.Join(patchDir, filepath.Base(path)))
		default:
			return fsutil.CopyFile(path, filepath.Join(inDir, filepath.Base(path)))
		}
	})
}

func firstPatchedROMInDir(dir string) (string, error) {
	if rom, err := firstROMInDir(dir); err == nil {
		return rom, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		return filepath.Join(dir, e.Name()), nil
	}
	return "", errors.New("no file found")
}

func unzipInto(zipPath, dstRoot string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip %s: %w", zipPath, err)
	}
	defer func() {
		_ = r.Close()
	}()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		cleanName := filepath.Clean(f.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return fmt.Errorf("zip %s contains unsafe path: %s", zipPath, f.Name)
		}

		targetPath := filepath.Join(dstRoot, cleanName)
		if err := fsutil.EnsureDir(filepath.Dir(targetPath)); err != nil {
			return err
		}

		src, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry %s in %s: %w", f.Name, zipPath, err)
		}

		dst, err := os.Create(targetPath)
		if err != nil {
			_ = src.Close()
			return fmt.Errorf("create extracted file %s: %w", targetPath, err)
		}

		if _, err := io.Copy(dst, src); err != nil {
			_ = dst.Close()
			_ = src.Close()
			return fmt.Errorf("extract %s from %s: %w", f.Name, zipPath, err)
		}

		if err := dst.Close(); err != nil {
			_ = src.Close()
			return fmt.Errorf("close extracted file %s: %w", targetPath, err)
		}
		if err := src.Close(); err != nil {
			return fmt.Errorf("close zip entry %s in %s: %w", f.Name, zipPath, err)
		}
	}

	return nil
}

func firstROMInDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		switch ext {
		case ".nes", ".sfc", ".smc", ".gb", ".gbc", ".gba", ".sms", ".md", ".gen", ".pce", ".bin", ".chd", ".cue", ".iso":
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", errors.New("no base ROM found")
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	if name == "" {
		return "unnamed"
	}
	return name
}

func resolveHackGameDir(systemOutputRoot, hackInputDir, hackName string) (string, error) {
	baseROM, err := firstROMInDir(hackInputDir)
	if err != nil {
		// Fall back to a predictable location when no base ROM is staged.
		return filepath.Join(systemOutputRoot, sanitizeName(hackName)), nil
	}

	baseName := strings.TrimSuffix(filepath.Base(baseROM), filepath.Ext(baseROM))
	baseKey := normalizeGameKey(baseName)

	if existingDir, err := findExistingGameDir(systemOutputRoot, baseKey); err == nil {
		return existingDir, nil
	}

	if retailStem, err := findRetailStemMatch(systemOutputRoot, baseKey); err == nil {
		return filepath.Join(systemOutputRoot, sanitizeName(retailStem)), nil
	}

	return filepath.Join(systemOutputRoot, sanitizeName(baseName)), nil
}

func findExistingGameDir(root, gameKey string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if normalizeGameKey(entry.Name()) == gameKey {
			return filepath.Join(root, entry.Name()), nil
		}
	}
	return "", errors.New("no existing game dir match")
}

func findRetailStemMatch(root, gameKey string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if normalizeGameKey(stem) == gameKey {
			return stem, nil
		}
	}
	return "", errors.New("no retail file match")
}

func normalizeGameKey(s string) string {
	s = strings.TrimSpace(s)
	s = removeRegionGroups(s)
	s = strings.ToLower(s)
	replacer := strings.NewReplacer("_", " ", "-", " ", ".", " ")
	s = replacer.Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func removeRegionGroups(name string) string {
	return regionGroupRe.ReplaceAllStringFunc(name, func(group string) string {
		content := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(group, "("), ")"))
		if isRegionGroup(content) {
			return ""
		}
		return group
	})
}

func isRegionGroup(content string) bool {
	if content == "" {
		return false
	}
	regionTokens := []string{
		"usa", "europe", "eur", "japan", "jpn", "world", "korea", "asia", "australia", "brazil",
		"spain", "france", "germany", "italy", "netherlands", "sweden", "russia", "taiwan", "china", "uk",
	}
	norm := strings.ToLower(content)
	for _, token := range strings.Split(norm, ",") {
		t := strings.TrimSpace(token)
		for _, region := range regionTokens {
			if t == region {
				return true
			}
		}
	}
	return false
}

func runBiosStub(cfg *config.Config) error {
	if !cfg.Features.EnableBios {
		return errors.New("bios workflow disabled in config.features.enable_bios")
	}
	return errors.New("bios workflow is stubbed; implementation planned in next phase")
}

func runRedumpStub(cfg *config.Config) error {
	if !cfg.Features.EnableRedump {
		return errors.New("redump workflow disabled in config.features.enable_redump")
	}
	return errors.New("redump workflow is stubbed; implementation planned in next phase")
}

func runArcadeStub(cfg *config.Config) error {
	if !cfg.Features.EnableArcade {
		return errors.New("arcade workflow disabled in config.features.enable_arcade")
	}
	return errors.New("arcade workflow is stubbed; implementation planned in next phase")
}

func runCache(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return errors.New("cache requires subcommand: clean|path")
	}
	if len(args) > 1 {
		return fmt.Errorf("cache: unexpected arguments: %s", strings.Join(args[1:], ", "))
	}
	cacheRoot := cfg.ResolvePath(cfg.CacheDir)
	if strings.TrimSpace(cfg.CacheDir) == "" {
		cacheRoot = resolveCacheRoot(cfg)
	}
	switch args[0] {
	case "clean":
		return fsutil.RemoveIfExists(cacheRoot)
	case "path":
		fmt.Println(cacheRoot)
		return nil
	default:
		return fmt.Errorf("unknown cache subcommand: %s", args[0])
	}
}

func runExport(cfg *config.Config, g globalFlags, args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	systemsCSV := fs.String("systems", "", "comma-separated system slugs")
	allSystems := fs.Bool("all-systems", false, "export all enabled systems")
	destination := fs.String("destination", "", "destination root path (e.g., mounted SD card)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ensureNoPositionalArgs("export", fs.Args()); err != nil {
		return err
	}
	if strings.TrimSpace(*destination) == "" {
		return errors.New("export requires --destination")
	}

	systems, err := platform.ExpandSystems([]string{*systemsCSV}, *allSystems, cfg)
	if err != nil {
		return err
	}

	dstRoot := filepath.Clean(*destination)
	for _, s := range systems {
		sysCfg := cfg.Systems[s]
		src := filepath.Join(cfg.ResolvePath(cfg.Paths.RommLibraryRoms), sysCfg.RommSlug)
		dst := filepath.Join(dstRoot, sysCfg.RommSlug)

		if g.dryRun {
			fmt.Printf("[dry-run] export %s -> %s\n", src, dst)
			continue
		}
		if err := copyDirRecursive(src, dst); err != nil {
			return err
		}
	}
	return nil
}

func copyDirRecursive(src, dst string) error {
	if err := fsutil.EnsureDir(dst); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return fsutil.EnsureDir(target)
		}
		return fsutil.CopyFile(path, target)
	})
}

func runBootstrap(cfg *config.Config) error {
	dirs := []string{
		cfg.ResolvePath(cfg.Paths.RommLibraryRoms),
		cfg.ResolvePath(cfg.Paths.RommLibraryBios),
		cfg.ResolvePath(cfg.Paths.HacksSource),
		cfg.ResolvePath(cfg.Paths.ToSort),
		cfg.ResolvePath(cfg.Paths.VaultNoIntro),
		cfg.ResolvePath(cfg.CacheDir),
	}
	for _, d := range dirs {
		if d == "" {
			continue
		}
		if err := fsutil.EnsureDir(d); err != nil {
			return err
		}
	}

	for _, slug := range cfg.Bootstrap.RommRoms {
		if err := fsutil.EnsureDir(filepath.Join(cfg.ResolvePath(cfg.Paths.RommLibraryRoms), slug)); err != nil {
			return err
		}
	}
	for _, slug := range cfg.Bootstrap.RommBios {
		if err := fsutil.EnsureDir(filepath.Join(cfg.ResolvePath(cfg.Paths.RommLibraryBios), slug)); err != nil {
			return err
		}
	}
	return nil
}

func runSystems(cfg *config.Config) error {
	enabled := cfg.EnabledSystems()
	sort.Strings(enabled)
	for _, s := range enabled {
		fmt.Println(s)
	}
	return nil
}

func printRootUsage() {
	fmt.Println(`retro-collection-tool: Igir workflow wrapper for ROMM

Usage:
  retro-collection-tool [global flags] <command> [command flags]

Global flags:
	--config <path>       YAML config path (default discovery via cwd, then XDG)
  --dry-run             Plan-only mode
  --verbose             Verbose output

Commands:
  sync        Run retail sync with Igir
  hacks       Run curated ROM hacks patch workflow
  bios        BIOS workflow (stub)
  redump      ReDump workflow (stub)
  arcade      Arcade workflow (stub)
  export      Copy selected ROMM systems to destination (e.g., SD card)
  cache       Cache controls: clean|path
  bootstrap   Create expected directory structure
  systems     List enabled systems
  version     Print version`) //nolint:lll
}

func ensureNoPositionalArgs(command string, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return fmt.Errorf("%s: unexpected arguments: %s", command, strings.Join(args, ", "))
}

func resolveCacheRoot(cfg *config.Config) string {
	if strings.TrimSpace(cfg.CacheDir) != "" {
		return cfg.ResolvePath(cfg.CacheDir)
	}
	if cacheHome := xdg.CacheHome(); cacheHome != "" {
		return filepath.Join(cacheHome, appName)
	}
	return filepath.Join(os.TempDir(), appName)
}
