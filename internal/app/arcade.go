package app

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
	"github.com/karl-vanderslice/retro-collection-tool/internal/fsutil"
)

type arcadeDATEntry struct {
	Name        string
	Description string
	CloneOf     string
	IsBios      bool
}

type arcadeSetSelection struct {
	Games []string
}

type arcadeSetReport struct {
	SetName        string
	VaultDir       string
	LibraryDir     string
	TotalGames     int
	PresentGames   int
	MissingGames   int
	LinkedGames    int
	MissingSamples []string
}

type arcadeSetSpec struct {
	SetName    string
	DatPath    string
	VaultDir   string
	LibraryDir string
}

func runArcade(cfg *config.Config, g globalFlags, args []string) error {
	if !cfg.Features.EnableArcade {
		return errors.New("arcade workflow disabled in config.features.enable_arcade")
	}
	if len(args) == 0 {
		return errors.New("arcade requires subcommand: dats|verify|sync")
	}

	sub := strings.ToLower(strings.TrimSpace(args[0]))
	switch sub {
	case "dats", "dat":
		if len(args) < 2 {
			return errors.New("arcade dats requires subcommand: update|verify")
		}
		if err := ensureNoPositionalArgs("arcade dats", args[2:]); err != nil {
			return err
		}
		action := strings.ToLower(strings.TrimSpace(args[1]))
		switch action {
		case "update", "refresh":
			return runArcadeDATUpdate(cfg, g)
		case "verify":
			return runArcadeDATVerify(cfg, g)
		default:
			return fmt.Errorf("unknown arcade dats subcommand: %s", args[1])
		}
	case "verify":
		if err := ensureNoPositionalArgs("arcade verify", args[1:]); err != nil {
			return err
		}
		return runArcadeVerify(cfg, g)
	case "sync":
		if err := ensureNoPositionalArgs("arcade sync", args[1:]); err != nil {
			return err
		}
		return runArcadeSync(cfg, g)
	default:
		return fmt.Errorf("unknown arcade subcommand: %s", args[0])
	}
}

func runArcadeDATUpdate(cfg *config.Config, g globalFlags) error {
	cacheRoot := resolveCacheRoot(cfg)
	datDir := filepath.Join(cacheRoot, "arcade", "dats")
	if err := fsutil.EnsureDir(datDir); err != nil {
		return err
	}

	jobs := []struct {
		name string
		url  string
		out  string
	}{
		{
			name: "mame-2003-plus",
			url:  cfg.ArcadeDatMAME2003PlusURL(),
			out:  filepath.Join(datDir, cfg.ArcadeDatMAME2003PlusFile()),
		},
		{
			name: "fbneo",
			url:  cfg.ArcadeDatFBNeoURL(),
			out:  filepath.Join(datDir, cfg.ArcadeDatFBNeoFile()),
		},
	}

	for _, job := range jobs {
		emitInfo(g, "arcade", "dats", "updating", outputFields{"set": job.name, "url": job.url, "path": job.out, "dry_run": g.dryRun})
		if g.dryRun {
			continue
		}
		if err := downloadFile(job.url, job.out); err != nil {
			return err
		}
	}

	emitInfo(g, "arcade", "dats", "updated", outputFields{"path": datDir, "count": len(jobs)})
	return nil
}

func runArcadeDATVerify(cfg *config.Config, g globalFlags) error {
	specs := arcadeSpecsFromConfig(cfg)
	for _, spec := range specs {
		entries, err := parseArcadeDAT(spec.DatPath)
		if err != nil {
			return fmt.Errorf("verify %s dat: %w", spec.SetName, err)
		}
		sel := selectArcadeEntries(entries, cfg.ArcadeExcludeKeywords())
		emitInfo(g, "arcade", "dats", "verified", outputFields{
			"set":          spec.SetName,
			"path":         spec.DatPath,
			"entries":      len(entries),
			"games":        len(sel.Games),
			"exclude_keys": strings.Join(cfg.ArcadeExcludeKeywords(), ","),
		})
	}
	return nil
}

func runArcadeVerify(cfg *config.Config, g globalFlags) error {
	specs := arcadeSpecsFromConfig(cfg)
	totalGames := 0
	presentGames := 0
	for _, spec := range specs {
		report, err := verifyArcadeVaultSet(spec, cfg)
		if err != nil {
			return err
		}
		totalGames += report.TotalGames
		presentGames += report.PresentGames
		emitInfo(g, "arcade", "verify", "set report", outputFields{
			"set":            report.SetName,
			"vault":          report.VaultDir,
			"total_games":    report.TotalGames,
			"present_games":  report.PresentGames,
			"missing_games":  report.MissingGames,
			"missing_sample": strings.Join(report.MissingSamples, ","),
		})
	}
	emitInfo(g, "arcade", "verify", "summary", outputFields{
		"sets":          len(specs),
		"total_games":   totalGames,
		"present_games": presentGames,
		"missing_games": totalGames - presentGames,
	})
	return nil
}

func runArcadeSync(cfg *config.Config, g globalFlags) error {
	specs := arcadeSpecsFromConfig(cfg)
	for _, spec := range specs {
		report, err := linkArcadeSet(spec, cfg, g)
		if err != nil {
			return err
		}
		emitInfo(g, "arcade", "sync", "set synced", outputFields{
			"set":            report.SetName,
			"vault":          report.VaultDir,
			"library":        report.LibraryDir,
			"linked_games":   report.LinkedGames,
			"missing_games":  report.MissingGames,
			"missing_sample": strings.Join(report.MissingSamples, ","),
			"dry_run":        g.dryRun,
		})
	}
	return nil
}

func arcadeSpecsFromConfig(cfg *config.Config) []arcadeSetSpec {
	datDir := filepath.Join(resolveCacheRoot(cfg), "arcade", "dats")
	return []arcadeSetSpec{
		{
			SetName:    "mame-2003-plus",
			DatPath:    filepath.Join(datDir, cfg.ArcadeDatMAME2003PlusFile()),
			VaultDir:   cfg.ArcadeVaultMAME2003Plus(),
			LibraryDir: cfg.ArcadeLibraryMAME2003Plus(),
		},
		{
			SetName:    "fbneo",
			DatPath:    filepath.Join(datDir, cfg.ArcadeDatFBNeoFile()),
			VaultDir:   cfg.ArcadeVaultFBNeo(),
			LibraryDir: cfg.ArcadeLibraryFBNeo(),
		},
	}
}

func verifyArcadeVaultSet(spec arcadeSetSpec, cfg *config.Config) (arcadeSetReport, error) {
	entries, err := parseArcadeDAT(spec.DatPath)
	if err != nil {
		return arcadeSetReport{}, fmt.Errorf("load %s dat %s: %w", spec.SetName, spec.DatPath, err)
	}
	sel := selectArcadeEntries(entries, cfg.ArcadeExcludeKeywords())
	report := arcadeSetReport{
		SetName:    spec.SetName,
		VaultDir:   spec.VaultDir,
		LibraryDir: spec.LibraryDir,
		TotalGames: len(sel.Games),
	}
	for _, name := range sel.Games {
		if _, ok := findArcadeArchive(spec.VaultDir, name); ok {
			report.PresentGames++
		} else {
			report.MissingGames++
			if len(report.MissingSamples) < 12 {
				report.MissingSamples = append(report.MissingSamples, name)
			}
		}
	}
	return report, nil
}

func linkArcadeSet(spec arcadeSetSpec, cfg *config.Config, g globalFlags) (arcadeSetReport, error) {
	report, err := verifyArcadeVaultSet(spec, cfg)
	if err != nil {
		return arcadeSetReport{}, err
	}
	if err := fsutil.EnsureDir(spec.LibraryDir); err != nil {
		return arcadeSetReport{}, err
	}

	entries, err := parseArcadeDAT(spec.DatPath)
	if err != nil {
		return arcadeSetReport{}, fmt.Errorf("load %s dat %s: %w", spec.SetName, spec.DatPath, err)
	}
	sel := selectArcadeEntries(entries, cfg.ArcadeExcludeKeywords())
	linkOne := func(name string) error {
		src, ok := findArcadeArchive(spec.VaultDir, name)
		if !ok {
			return nil
		}
		dst := filepath.Join(spec.LibraryDir, filepath.Base(src))
		if g.dryRun {
			emitInfo(g, "arcade", "sync", "dry-run link", outputFields{"set": spec.SetName, "from": src, "to": dst})
			report.LinkedGames++
			return nil
		}
		if err := fsutil.LinkOrCopy(src, dst); err != nil {
			return err
		}
		report.LinkedGames++
		return nil
	}

	for _, name := range sel.Games {
		if err := linkOne(name); err != nil {
			return arcadeSetReport{}, err
		}
	}
	return report, nil
}

func selectArcadeEntries(entries []arcadeDATEntry, excludeKeywords []string) arcadeSetSelection {
	games := map[string]bool{}
	for _, e := range entries {
		if strings.TrimSpace(e.Name) == "" {
			continue
		}
		if e.IsBios {
			continue
		}
		if strings.TrimSpace(e.CloneOf) != "" {
			continue
		}
		haystack := strings.ToLower(e.Name + " " + e.Description)
		excluded := false
		for _, raw := range excludeKeywords {
			k := strings.ToLower(strings.TrimSpace(raw))
			if k == "" {
				continue
			}
			if strings.Contains(haystack, k) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}
		games[e.Name] = true
	}

	gameNames := make([]string, 0, len(games))
	for name := range games {
		gameNames = append(gameNames, name)
	}
	sort.Strings(gameNames)

	return arcadeSetSelection{Games: gameNames}
}

func parseArcadeDAT(path string) ([]arcadeDATEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("dat file missing: %s (run: retro-collection-tool arcade dats update)", path)
		}
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	dec := xml.NewDecoder(f)
	entries := make([]arcadeDATEntry, 0, 1024)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse dat xml %s: %w", path, err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if start.Name.Local != "game" && start.Name.Local != "machine" {
			continue
		}

		var node struct {
			Name        string `xml:"name,attr"`
			CloneOf     string `xml:"cloneof,attr"`
			IsBios      string `xml:"isbios,attr"`
			Description string `xml:"description"`
		}
		if err := dec.DecodeElement(&node, &start); err != nil {
			return nil, fmt.Errorf("decode dat entry in %s: %w", path, err)
		}
		entries = append(entries, arcadeDATEntry{
			Name:        strings.TrimSpace(node.Name),
			Description: strings.TrimSpace(node.Description),
			CloneOf:     strings.TrimSpace(node.CloneOf),
			IsBios:      strings.EqualFold(strings.TrimSpace(node.IsBios), "yes"),
		})
	}
	return entries, nil
}

func findArcadeArchive(root, name string) (string, bool) {
	candidates := []string{
		filepath.Join(root, name+".zip"),
		filepath.Join(root, name+".7z"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c, true
		}
	}
	return "", false
}

func downloadFile(url, out string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download %s: unexpected status %s", url, resp.Status)
	}
	if err := fsutil.EnsureDir(filepath.Dir(out)); err != nil {
		return err
	}
	tmp := out + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create %s: %w", tmp, err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, out); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmp, out, err)
	}
	return nil
}
