package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
)

type fakeArcadeRunner struct {
	calls [][]string
	err   error
}

func (f *fakeArcadeRunner) Run(_ context.Context, args []string) error {
	clone := append([]string(nil), args...)
	f.calls = append(f.calls, clone)
	return f.err
}

func TestRunArcadeVerifyUsesIgirDryRun(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfg := testArcadeConfig(tmp)
	runner := &fakeArcadeRunner{}

	specs := arcadeSpecsFromConfig(cfg)
	for _, spec := range specs {
		if err := os.MkdirAll(filepath.Dir(spec.DatPath), 0o755); err != nil {
			t.Fatalf("mkdir dat dir: %v", err)
		}
		if err := os.WriteFile(spec.DatPath, []byte("dummy"), 0o644); err != nil {
			t.Fatalf("write dat: %v", err)
		}
	}

	if err := runArcade(context.Background(), cfg, runner, globalFlags{}, []string{"verify"}); err != nil {
		t.Fatalf("runArcade verify: %v", err)
	}

	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 igir calls, got %d", len(runner.calls))
	}
	for _, args := range runner.calls {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "link report") {
			t.Fatalf("expected link report command, got %q", joined)
		}
		if !strings.Contains(joined, "--report-output") {
			t.Fatalf("expected --report-output in verify command, got %q", joined)
		}
		if !strings.Contains(joined, "--no-bios") || !strings.Contains(joined, "--no-device") {
			t.Fatalf("expected bios/device exclusion flags, got %q", joined)
		}
	}
}

func TestRunArcadeSyncUsesIgir(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfg := testArcadeConfig(tmp)
	runner := &fakeArcadeRunner{}

	specs := arcadeSpecsFromConfig(cfg)
	for _, spec := range specs {
		if err := os.MkdirAll(filepath.Dir(spec.DatPath), 0o755); err != nil {
			t.Fatalf("mkdir dat dir: %v", err)
		}
		if err := os.WriteFile(spec.DatPath, []byte("dummy"), 0o644); err != nil {
			t.Fatalf("write dat: %v", err)
		}
	}

	if err := runArcade(context.Background(), cfg, runner, globalFlags{dryRun: true}, []string{"sync"}); err != nil {
		t.Fatalf("runArcade sync: %v", err)
	}

	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 igir calls, got %d", len(runner.calls))
	}
	for _, args := range runner.calls {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "link report") {
			t.Fatalf("expected link report command for dry-run sync, got %q", joined)
		}
		if !strings.Contains(joined, "--report-output") {
			t.Fatalf("expected --report-output in dry-run sync command, got %q", joined)
		}
	}
}

func TestRunArcadeSyncUsesIgirCleanWhenNotDryRun(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfg := testArcadeConfig(tmp)
	runner := &fakeArcadeRunner{}

	specs := arcadeSpecsFromConfig(cfg)
	for _, spec := range specs {
		if err := os.MkdirAll(filepath.Dir(spec.DatPath), 0o755); err != nil {
			t.Fatalf("mkdir dat dir: %v", err)
		}
		if err := os.WriteFile(spec.DatPath, []byte("dummy"), 0o644); err != nil {
			t.Fatalf("write dat: %v", err)
		}
	}

	if err := runArcade(context.Background(), cfg, runner, globalFlags{}, []string{"sync"}); err != nil {
		t.Fatalf("runArcade sync: %v", err)
	}

	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 igir calls, got %d", len(runner.calls))
	}
	for _, args := range runner.calls {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "link clean") {
			t.Fatalf("expected link clean command, got %q", joined)
		}
		if strings.Contains(joined, "--report-output") {
			t.Fatalf("did not expect --report-output in non-dry sync command, got %q", joined)
		}
	}
}

func TestRunArcadeDatsUpdateDownloads(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/mame.dat":
			_, _ = w.Write([]byte("clrmamepro (\n  name \"mame\"\n)"))
		case "/fbneo.dat":
			_, _ = w.Write([]byte("<datafile></datafile>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	cfg := testArcadeConfig(tmp)
	cfg.Arcade.DatMAME2003URL = server.URL + "/mame.dat"
	cfg.Arcade.DatFBNeoURL = server.URL + "/fbneo.dat"

	if err := runArcade(context.Background(), cfg, &fakeArcadeRunner{}, globalFlags{}, []string{"dats", "update"}); err != nil {
		t.Fatalf("runArcade dats update: %v", err)
	}

	specs := arcadeSpecsFromConfig(cfg)
	for _, spec := range specs {
		if _, err := os.Stat(spec.DatPath); err != nil {
			t.Fatalf("expected dat at %s: %v", spec.DatPath, err)
		}
	}
}

func TestRunArcadeDatsVerifyRequiresExistingDats(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfg := testArcadeConfig(tmp)
	err := runArcade(context.Background(), cfg, &fakeArcadeRunner{}, globalFlags{}, []string{"dats", "verify"})
	if err == nil {
		t.Fatalf("expected missing dat error")
	}
	if !strings.Contains(err.Error(), "run: retro-collection-tool arcade dats update") {
		t.Fatalf("expected update guidance in error: %v", err)
	}
}

func testArcadeConfig(root string) *config.Config {
	return &config.Config{
		Root:     root,
		CacheDir: "cache",
		Igir: config.IgirConfig{
			InputChecksumMin: "CRC32",
		},
		Paths: config.PathsConfig{
			RommLibraryRoms: "roms/Library/roms",
		},
		Features: config.FeatureToggles{
			EnableArcade: true,
		},
		Arcade: config.ArcadeConfig{
			VaultMAME2003Plus: "roms/Vault/Arcade/mame-2003-plus-reference-set/roms",
			VaultFBNeo:        "roms/Vault/Arcade/fbneo_1003_bestset/fbneo_1_0_0_3_best/games",
			LibraryMAME2003:   "roms/Library/roms/arcade/mame-2003-plus",
			LibraryFBNeo:      "roms/Library/roms/arcade/fbneo",
			DatMAME2003File:   "arcade-mame-2003-plus.dat",
			DatFBNeoFile:      "arcade-fbneo.dat",
		},
	}
}
