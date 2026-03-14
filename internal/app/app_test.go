package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNormalizeGameKeyStripsRegionTokens(t *testing.T) {
	t.Parallel()

	got := normalizeGameKey("Phantasy Star II (USA, Europe) (Rev A)")
	want := "phantasy star ii (rev a)"
	if got != want {
		t.Fatalf("normalizeGameKey mismatch: got %q want %q", got, want)
	}
}

func TestNormalizeGameKeyPreservesNonRegionGroup(t *testing.T) {
	t.Parallel()

	got := normalizeGameKey("Phantasy Star III (En) (Beta)")
	want := "phantasy star iii (en) (beta)"
	if got != want {
		t.Fatalf("normalizeGameKey mismatch: got %q want %q", got, want)
	}
}

func TestEnsureNoPositionalArgs(t *testing.T) {
	t.Parallel()

	if err := ensureNoPositionalArgs("sync", nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := ensureNoPositionalArgs("sync", []string{"extra"}); err == nil {
		t.Fatal("expected error for positional args")
	}
}

func TestResolveConfigPathsLayering(t *testing.T) {
	tmp := t.TempDir()

	configHome := filepath.Join(tmp, "xdg")
	projectDir := filepath.Join(tmp, "project")
	override := filepath.Join(tmp, "override.yaml")

	if err := os.MkdirAll(filepath.Join(configHome, "retro-collection-tool"), 0o755); err != nil {
		t.Fatalf("mkdir xdg: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "config"), 0o755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}

	xdgCfg := filepath.Join(configHome, "retro-collection-tool", "config.yaml")
	projectCfg := filepath.Join(projectDir, "config", "retro-collection-tool.yaml")

	if err := os.WriteFile(xdgCfg, []byte("root: /x\n"), 0o644); err != nil {
		t.Fatalf("write xdg cfg: %v", err)
	}
	if err := os.WriteFile(projectCfg, []byte("paths: {}\n"), 0o644); err != nil {
		t.Fatalf("write project cfg: %v", err)
	}
	if err := os.WriteFile(override, []byte("cache_dir: cache\n"), 0o644); err != nil {
		t.Fatalf("write override cfg: %v", err)
	}

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origCwd)
	})
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv(configEnvVar, override)

	paths, err := resolveConfigPaths("")
	if err != nil {
		t.Fatalf("resolveConfigPaths: %v", err)
	}

	if len(paths) != 3 {
		t.Fatalf("expected 3 config layers, got %d: %#v", len(paths), paths)
	}
	if paths[0] != xdgCfg || paths[1] != projectCfg || paths[2] != override {
		t.Fatalf("unexpected layering order: %#v", paths)
	}
}

func TestCollectPatchFilesSorted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files := []string{"02-core.ups", "readme.txt", "01-base.ips", "03-polish.xdelta"}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	got, err := collectPatchFiles(dir)
	if err != nil {
		t.Fatalf("collectPatchFiles: %v", err)
	}

	want := []string{
		filepath.Join(dir, "01-base.ips"),
		filepath.Join(dir, "02-core.ups"),
		filepath.Join(dir, "03-polish.xdelta"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected patch order:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestOrganizeRetailFilesInRootMovesIntoGameFolder(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	name := "Phantasy Star II (USA, Europe) (Rev A).md"
	src := filepath.Join(root, name)
	if err := os.WriteFile(src, []byte("rom"), 0o644); err != nil {
		t.Fatalf("write source rom: %v", err)
	}

	if err := organizeRetailFilesInRoot(root, globalFlags{}); err != nil {
		t.Fatalf("organizeRetailFilesInRoot: %v", err)
	}

	dst := filepath.Join(root, "Phantasy Star II (USA, Europe) (Rev A)", name)
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("expected organized retail file at %s: %v", dst, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("expected source file to be moved, stat err=%v", err)
	}
}

func TestOrganizeRetailFilesInRootDryRunDoesNotMove(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	name := "Sonic the Hedgehog (USA, Europe).md"
	src := filepath.Join(root, name)
	if err := os.WriteFile(src, []byte("rom"), 0o644); err != nil {
		t.Fatalf("write source rom: %v", err)
	}

	if err := organizeRetailFilesInRoot(root, globalFlags{dryRun: true}); err != nil {
		t.Fatalf("organizeRetailFilesInRoot dry-run: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Fatalf("expected source file to remain in dry-run: %v", err)
	}
}

func TestOrganizeRetailFilesInRootSupportsExpandedSystemExtensions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	romFiles := []string{
		"Sonic Blast (USA, Europe).gg",
		"Golden Axe (World).sms",
		"Pokemon Red Version (USA, Europe).gb",
		"Mario Kart 64 (USA).z64",
		"SNK vs. Capcom - Match of the Millennium (USA, Europe).ngc",
		"Metal Gear 2 - Solid Snake (Japan).rom",
		"Thexder (Japan).cas",
	}

	for _, name := range romFiles {
		if err := os.WriteFile(filepath.Join(root, name), []byte("rom"), 0o644); err != nil {
			t.Fatalf("write source rom %s: %v", name, err)
		}
	}

	nonROM := "README.txt"
	nonROMSrc := filepath.Join(root, nonROM)
	if err := os.WriteFile(nonROMSrc, []byte("note"), 0o644); err != nil {
		t.Fatalf("write non-rom file: %v", err)
	}

	if err := organizeRetailFilesInRoot(root, globalFlags{}); err != nil {
		t.Fatalf("organizeRetailFilesInRoot: %v", err)
	}

	for _, name := range romFiles {
		stem := name[:len(name)-len(filepath.Ext(name))]
		dst := filepath.Join(root, stem, name)
		if _, err := os.Stat(dst); err != nil {
			t.Fatalf("expected organized rom file at %s: %v", dst, err)
		}
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("expected source rom file %s to be moved, stat err=%v", name, err)
		}
	}

	if _, err := os.Stat(nonROMSrc); err != nil {
		t.Fatalf("expected non-rom file to remain in root: %v", err)
	}
}
