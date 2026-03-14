package app

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/karl-vanderslice/retro-collection-tool/internal/config"
)

func TestRunBiosImportsKnownHashMatch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "bios-source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}

	content := []byte("abc")
	src := filepath.Join(sourceRoot, "custom_gba.bin")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("write source bios: %v", err)
	}

	catalog := "entries:\n  - system: gba\n    required: true\n    destination: gba_bios.bin\n    sources:\n      - name: custom_gba.bin\n        md5: 900150983cd24fb0d6963f7d28e17f72\n"
	catalogPath := filepath.Join(root, "bios-catalog.yaml")
	if err := os.WriteFile(catalogPath, []byte(catalog), 0o644); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	cfg := biosTestConfig(root, sourceRoot, catalogPath)
	if err := runBios(cfg, globalFlags{}, []string{"--systems", "gba"}); err != nil {
		t.Fatalf("runBios: %v", err)
	}

	vaultDst := filepath.Join(root, "roms", "Vault", "BIOS", "gba_bios.bin")
	libraryDst := filepath.Join(root, "roms", "Library", "bios", "gba", "gba_bios.bin")
	vaultInfo, err := os.Stat(vaultDst)
	if err != nil {
		t.Fatalf("expected BIOS vault output at %s: %v", vaultDst, err)
	}
	libraryInfo, err := os.Stat(libraryDst)
	if err != nil {
		t.Fatalf("expected BIOS library output at %s: %v", libraryDst, err)
	}
	if !os.SameFile(vaultInfo, libraryInfo) {
		t.Fatalf("expected hardlinked files for vault and library outputs")
	}
}

func TestRunBiosStrictFailsOnHashMismatch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "bios-source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}

	src := filepath.Join(sourceRoot, "gba_bios.bin")
	if err := os.WriteFile(src, []byte("wrong-content"), 0o644); err != nil {
		t.Fatalf("write source bios: %v", err)
	}

	catalog := "entries:\n  - system: gba\n    required: true\n    destination: gba_bios.bin\n    sources:\n      - name: gba_bios.bin\n        md5: a860e8c0b6d573d191e4ec7db1b1e4f6\n"
	catalogPath := filepath.Join(root, "bios-catalog.yaml")
	if err := os.WriteFile(catalogPath, []byte(catalog), 0o644); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	cfg := biosTestConfig(root, sourceRoot, catalogPath)
	err := runBios(cfg, globalFlags{}, []string{"--systems", "gba", "--strict"})
	if err == nil {
		t.Fatal("expected strict mode failure for hash mismatch")
	}
}

func TestRunBiosImportsKnownHashFromZipPack(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "bios-source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}

	zipPath := filepath.Join(sourceRoot, "bios-pack.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(zf)
	w, err := zw.Create("packs/gba_bios.bin")
	if err != nil {
		t.Fatalf("zip create entry: %v", err)
	}
	if _, err := w.Write([]byte("abc")); err != nil {
		t.Fatalf("zip write entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close writer: %v", err)
	}
	if err := zf.Close(); err != nil {
		t.Fatalf("zip close file: %v", err)
	}

	catalog := "entries:\n  - system: gba\n    required: true\n    destination: gba_bios.bin\n    sources:\n      - name: gba_bios.bin\n        md5: 900150983cd24fb0d6963f7d28e17f72\n"
	catalogPath := filepath.Join(root, "bios-catalog.yaml")
	if err := os.WriteFile(catalogPath, []byte(catalog), 0o644); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	cfg := biosTestConfig(root, sourceRoot, catalogPath)
	if err := runBios(cfg, globalFlags{}, []string{"--systems", "gba"}); err != nil {
		t.Fatalf("runBios from zip: %v", err)
	}

	vaultDst := filepath.Join(root, "roms", "Vault", "BIOS", "gba_bios.bin")
	libraryDst := filepath.Join(root, "roms", "Library", "bios", "gba", "gba_bios.bin")
	vaultInfo, err := os.Stat(vaultDst)
	if err != nil {
		t.Fatalf("expected BIOS vault output at %s: %v", vaultDst, err)
	}
	libraryInfo, err := os.Stat(libraryDst)
	if err != nil {
		t.Fatalf("expected BIOS library output at %s: %v", libraryDst, err)
	}
	if !os.SameFile(vaultInfo, libraryInfo) {
		t.Fatalf("expected hardlinked files for vault and library outputs")
	}
}

func TestRunBiosSkipsInvalidZipAndContinues(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "bios-source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}

	badZip := filepath.Join(sourceRoot, "bad.zip")
	if err := os.WriteFile(badZip, []byte("not-a-valid-zip"), 0o644); err != nil {
		t.Fatalf("write bad zip: %v", err)
	}

	good := filepath.Join(sourceRoot, "gba_bios.bin")
	if err := os.WriteFile(good, []byte("abc"), 0o644); err != nil {
		t.Fatalf("write good bios: %v", err)
	}

	catalog := "entries:\n  - system: gba\n    required: true\n    destination: gba_bios.bin\n    sources:\n      - name: gba_bios.bin\n        md5: 900150983cd24fb0d6963f7d28e17f72\n"
	catalogPath := filepath.Join(root, "bios-catalog.yaml")
	if err := os.WriteFile(catalogPath, []byte(catalog), 0o644); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	cfg := biosTestConfig(root, sourceRoot, catalogPath)
	if err := runBios(cfg, globalFlags{}, []string{"--systems", "gba"}); err != nil {
		t.Fatalf("runBios should continue past invalid zip: %v", err)
	}

	vaultDst := filepath.Join(root, "roms", "Vault", "BIOS", "gba_bios.bin")
	if _, err := os.Stat(vaultDst); err != nil {
		t.Fatalf("expected BIOS output at %s: %v", vaultDst, err)
	}
}

func TestRunBiosEarlyCheckUsesExistingVaultFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "bios-source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}

	vaultDir := filepath.Join(root, "roms", "Vault", "BIOS")
	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		t.Fatalf("mkdir vault: %v", err)
	}
	vaultFile := filepath.Join(vaultDir, "gba_bios.bin")
	if err := os.WriteFile(vaultFile, []byte("abc"), 0o644); err != nil {
		t.Fatalf("write vault file: %v", err)
	}

	catalog := "entries:\n  - system: gba\n    required: true\n    destination: gba_bios.bin\n    sources:\n      - name: gba_bios.bin\n        md5: 900150983cd24fb0d6963f7d28e17f72\n"
	catalogPath := filepath.Join(root, "bios-catalog.yaml")
	if err := os.WriteFile(catalogPath, []byte(catalog), 0o644); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	cfg := biosTestConfig(root, sourceRoot, catalogPath)
	if err := runBios(cfg, globalFlags{}, []string{"--systems", "gba"}); err != nil {
		t.Fatalf("runBios should reuse vault file: %v", err)
	}

	libraryDst := filepath.Join(root, "roms", "Library", "bios", "gba", "gba_bios.bin")
	vaultInfo, err := os.Stat(vaultFile)
	if err != nil {
		t.Fatalf("stat vault: %v", err)
	}
	libraryInfo, err := os.Stat(libraryDst)
	if err != nil {
		t.Fatalf("stat library: %v", err)
	}
	if !os.SameFile(vaultInfo, libraryInfo) {
		t.Fatalf("expected library to hardlink/copy from existing vault file")
	}
}

func TestRunBiosImportsWithNameOnlyCatalogSource(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourceRoot := filepath.Join(root, "bios-source")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}

	src := filepath.Join(sourceRoot, "prod.keys")
	if err := os.WriteFile(src, []byte("unique-device-key-content"), 0o644); err != nil {
		t.Fatalf("write source bios: %v", err)
	}

	catalog := "entries:\n  - system: switch\n    required: true\n    destination: ryujinx/keys/prod.keys\n    sources:\n      - name: prod.keys\n"
	catalogPath := filepath.Join(root, "bios-catalog.yaml")
	if err := os.WriteFile(catalogPath, []byte(catalog), 0o644); err != nil {
		t.Fatalf("write catalog: %v", err)
	}

	cfg := &config.Config{
		Root: root,
		Paths: config.PathsConfig{
			RommLibraryBios: "roms/Library/bios",
			VaultBios:       "roms/Vault/BIOS",
		},
		Bios: config.BiosConfig{
			CatalogFile: catalogPath,
			SourceRoots: []string{sourceRoot},
		},
		Features: config.FeatureToggles{
			EnableBios: true,
		},
		Systems: map[string]config.SystemConfig{
			"switch": {
				Enabled:  true,
				RommSlug: "switch",
			},
		},
	}

	if err := runBios(cfg, globalFlags{}, []string{"--systems", "switch"}); err != nil {
		t.Fatalf("runBios: %v", err)
	}

	vaultDst := filepath.Join(root, "roms", "Vault", "BIOS", "ryujinx", "keys", "prod.keys")
	libraryDst := filepath.Join(root, "roms", "Library", "bios", "switch", "prod.keys")
	if _, err := os.Stat(vaultDst); err != nil {
		t.Fatalf("expected BIOS vault output at %s: %v", vaultDst, err)
	}
	if _, err := os.Stat(libraryDst); err != nil {
		t.Fatalf("expected BIOS library output at %s: %v", libraryDst, err)
	}
}

func biosTestConfig(root, sourceRoot, catalogPath string) *config.Config {
	return &config.Config{
		Root: root,
		Paths: config.PathsConfig{
			RommLibraryBios: "roms/Library/bios",
			VaultBios:       "roms/Vault/BIOS",
		},
		Bios: config.BiosConfig{
			CatalogFile: catalogPath,
			SourceRoots: []string{sourceRoot},
		},
		Features: config.FeatureToggles{
			EnableBios: true,
		},
		Systems: map[string]config.SystemConfig{
			"gba": {
				Enabled:  true,
				RommSlug: "gba",
			},
		},
	}
}

func TestDefaultBiosCatalogCoversEnabledSystems(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	repoRoot := filepath.Join(wd, "..", "..")
	configPath := filepath.Join(repoRoot, "config", "retro-collection-tool.yaml")
	catalogPath := filepath.Join(repoRoot, "config", "bios-catalog.yaml")

	cfg, err := config.LoadMerged([]string{configPath}, config.EnvOverrides{Root: "/tmp/retro-library"})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	b, err := os.ReadFile(catalogPath)
	if err != nil {
		t.Fatalf("read bios catalog: %v", err)
	}
	catalog, err := parseBiosCatalog(b)
	if err != nil {
		t.Fatalf("parse bios catalog: %v", err)
	}

	hasCatalog := map[string]bool{}
	for _, entry := range catalog.Entries {
		hasCatalog[entry.System] = true
		if _, ok := cfg.Systems[entry.System]; !ok {
			t.Fatalf("bios catalog references unknown system %q", entry.System)
		}
	}

	for _, system := range cfg.EnabledSystems() {
		if !hasCatalog[system] {
			t.Fatalf("enabled system %q missing BIOS catalog entries", system)
		}
	}
}
