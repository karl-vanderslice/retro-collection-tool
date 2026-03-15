package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
	"github.com/karl-vanderslice/retro-collection-tool/internal/fsutil"
)

type arcadeSetSpec struct {
	SetName    string
	DatPath    string
	VaultDir   string
	LibraryDir string
}

type arcadeIgirRunner interface {
	Run(ctx context.Context, args []string) error
}

func runArcade(ctx context.Context, cfg *config.Config, runner arcadeIgirRunner, g globalFlags, args []string) error {
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
		return runArcadeVerify(ctx, cfg, runner, g)
	case "sync":
		if err := ensureNoPositionalArgs("arcade sync", args[1:]); err != nil {
			return err
		}
		return runArcadeSync(ctx, cfg, runner, g)
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
		info, err := os.Stat(spec.DatPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("dat file missing: %s (run: retro-collection-tool arcade dats update)", spec.DatPath)
			}
			return fmt.Errorf("stat dat %s: %w", spec.DatPath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("dat path is a directory: %s", spec.DatPath)
		}
		if info.Size() == 0 {
			return fmt.Errorf("dat file is empty: %s", spec.DatPath)
		}
		emitInfo(g, "arcade", "dats", "verified", outputFields{"set": spec.SetName, "path": spec.DatPath, "size_bytes": info.Size()})
	}
	return nil
}

func runArcadeVerify(ctx context.Context, cfg *config.Config, runner arcadeIgirRunner, g globalFlags) error {
	specs := arcadeSpecsFromConfig(cfg)
	for i, spec := range specs {
		args, err := buildArcadeIgirArgs(cfg, spec, true, false)
		if err != nil {
			return err
		}
		emitInfo(g, "arcade", "verify", "running igir verify", outputFields{"set": spec.SetName, "vault": spec.VaultDir, "dat": spec.DatPath, "index": fmt.Sprintf("%d/%d", i+1, len(specs))})
		if g.verbose || g.dryRun {
			emitInfo(g, "arcade", "verify", "igir command", outputFields{"set": spec.SetName, "cmd": strings.Join(args, " ")})
		}
		if err := runner.Run(ctx, args); err != nil {
			return fmt.Errorf("arcade verify (%s): %w", spec.SetName, err)
		}
	}
	emitInfo(g, "arcade", "verify", "summary", outputFields{"sets": len(specs)})
	return nil
}

func runArcadeSync(ctx context.Context, cfg *config.Config, runner arcadeIgirRunner, g globalFlags) error {
	specs := arcadeSpecsFromConfig(cfg)
	for i, spec := range specs {
		args, err := buildArcadeIgirArgs(cfg, spec, false, g.dryRun)
		if err != nil {
			return err
		}
		emitInfo(g, "arcade", "sync", "running igir sync", outputFields{"set": spec.SetName, "vault": spec.VaultDir, "library": spec.LibraryDir, "dat": spec.DatPath, "index": fmt.Sprintf("%d/%d", i+1, len(specs))})
		if g.verbose || g.dryRun {
			emitInfo(g, "arcade", "sync", "igir command", outputFields{"set": spec.SetName, "cmd": strings.Join(args, " ")})
		}
		if err := runner.Run(ctx, args); err != nil {
			return fmt.Errorf("arcade sync (%s): %w", spec.SetName, err)
		}
	}
	emitInfo(g, "arcade", "sync", "summary", outputFields{"sets": len(specs), "dry_run": g.dryRun})
	return nil
}

func buildArcadeIgirArgs(cfg *config.Config, spec arcadeSetSpec, verifyOnly bool, dryRun bool) ([]string, error) {
	outputDir := spec.LibraryDir
	args := []string{"link"}
	if verifyOnly || dryRun {
		outputDir = filepath.Join(resolveCacheRoot(cfg), "arcade", "verify", spec.SetName)
		if err := fsutil.RemoveIfExists(outputDir); err != nil {
			return nil, err
		}
		if err := fsutil.EnsureDir(outputDir); err != nil {
			return nil, err
		}
		reportPath := filepath.Join(resolveCacheRoot(cfg), "arcade", "reports", spec.SetName+"_%YYYY-%MM-%DDT%HH:%mm:%ss.csv")
		if err := fsutil.EnsureDir(filepath.Dir(reportPath)); err != nil {
			return nil, err
		}
		args = append(args, "report")
		args = append(args, "--report-output", reportPath)
	} else {
		args = append(args, "clean")
	}

	args = append(args,
		"--dat", spec.DatPath,
		"--input", spec.VaultDir,
		"--input-exclude", filepath.Join(spec.VaultDir, "**", "*.chd"),
		"--output", outputDir,
		"--link-mode", "hardlink",
		"--merge-roms", "split",
		"--exclude-disks",
		"--no-bios",
		"--no-device",
	)
	if !verifyOnly && !dryRun {
		args = append(args, "--overwrite-invalid")
	}
	if min := strings.TrimSpace(cfg.Igir.InputChecksumMin); min != "" {
		args = append(args, "--input-checksum-min", min)
	}
	return args, nil
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
