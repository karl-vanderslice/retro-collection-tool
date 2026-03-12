package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMergedPrecedenceAndEnvOverride(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	base := filepath.Join(tmp, "base.yaml")
	override := filepath.Join(tmp, "override.yaml")

	baseYAML := strings.Join([]string{
		"root: /base",
		"cache_dir: cache",
		"paths:",
		"  romm_library_roms: roms/Library/roms",
		"  hacks_source: roms/Hacks",
		"systems:",
		"  genesis:",
		"    enabled: true",
		"    romm_slug: genesis",
		"    dat_pattern: Sega - Mega Drive - Genesis",
		"",
	}, "\n")
	overrideYAML := strings.Join([]string{
		"root: /override",
		"paths:",
		"  hacks_source: roms/Hacks-Override",
		"",
	}, "\n")

	if err := os.WriteFile(base, []byte(baseYAML), 0o644); err != nil {
		t.Fatalf("write base: %v", err)
	}
	if err := os.WriteFile(override, []byte(overrideYAML), 0o644); err != nil {
		t.Fatalf("write override: %v", err)
	}

	cfg, err := LoadMerged([]string{base, override}, EnvOverrides{Root: "/env-root"})
	if err != nil {
		t.Fatalf("LoadMerged: %v", err)
	}

	if cfg.Root != "/env-root" {
		t.Fatalf("root precedence mismatch: got %q", cfg.Root)
	}
	if cfg.Paths.HacksSource != "roms/Hacks-Override" {
		t.Fatalf("override path mismatch: got %q", cfg.Paths.HacksSource)
	}
}

func TestLoadMergedRequiresRootAfterMerge(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.yaml")
	cfgYAML := strings.Join([]string{
		"cache_dir: cache",
		"paths:",
		"  romm_library_roms: roms/Library/roms",
		"  hacks_source: roms/Hacks",
		"systems:",
		"  genesis:",
		"    enabled: true",
		"    romm_slug: genesis",
		"    dat_pattern: Sega - Mega Drive - Genesis",
		"",
	}, "\n")
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadMerged([]string{cfgPath}, EnvOverrides{})
	if err == nil {
		t.Fatal("expected validation error without root")
	}
}

func TestSystemConfigEffectiveDatPattern_Default(t *testing.T) {
	t.Parallel()

	s := SystemConfig{DatPattern: "Sega - Mega Drive - Genesis"}
	if got := s.EffectiveRetailDatPattern(); got != "Sega - Mega Drive - Genesis" {
		t.Fatalf("retail pattern mismatch: %q", got)
	}
	if got := s.EffectiveHackDatPattern(); got != "Sega - Mega Drive - Genesis" {
		t.Fatalf("hack pattern mismatch: %q", got)
	}
}

func TestSystemConfigEffectiveDatPattern_Overrides(t *testing.T) {
	t.Parallel()

	s := SystemConfig{
		DatPattern:       "default",
		RetailDatPattern: "retail-specific",
		HackDatPattern:   "hack-specific",
	}
	if got := s.EffectiveRetailDatPattern(); got != "retail-specific" {
		t.Fatalf("retail override mismatch: %q", got)
	}
	if got := s.EffectiveHackDatPattern(); got != "hack-specific" {
		t.Fatalf("hack override mismatch: %q", got)
	}
}

func TestLoadDefaultConfigIncludesExpandedNoIntroSystems(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	defaultConfigPath := filepath.Join(wd, "..", "..", "config", "retro-collection-tool.yaml")
	cfg, err := LoadMerged([]string{defaultConfigPath}, EnvOverrides{Root: "/tmp/retro-library"})
	if err != nil {
		t.Fatalf("LoadMerged default config: %v", err)
	}

	testCases := map[string]struct {
		rommSlug   string
		datPattern string
	}{
		"msx": {
			rommSlug:   "msx",
			datPattern: "Microsoft - MSX",
		},
		"msx2": {
			rommSlug:   "msx2",
			datPattern: "Microsoft - MSX2",
		},
		"tg16": {
			rommSlug:   "tg16",
			datPattern: "NEC - PC Engine - TurboGrafx-16",
		},
		"supergrafx": {
			rommSlug:   "supergrafx",
			datPattern: "NEC - PC Engine SuperGrafx",
		},
		"gb": {
			rommSlug:   "gb",
			datPattern: "Nintendo - Game Boy",
		},
		"gba": {
			rommSlug:   "gba",
			datPattern: "Nintendo - Game Boy Advance",
		},
		"gbc": {
			rommSlug:   "gbc",
			datPattern: "Nintendo - Game Boy Color",
		},
		"n64": {
			rommSlug:   "n64",
			datPattern: "Nintendo - Nintendo 64 (BigEndian)",
		},
		"nes": {
			rommSlug:   "nes",
			datPattern: "Nintendo - Nintendo Entertainment System",
		},
		"snes": {
			rommSlug:   "snes",
			datPattern: "Nintendo - Super Nintendo Entertainment System",
		},
		"neo-geo-pocket": {
			rommSlug:   "neo-geo-pocket",
			datPattern: "SNK - NeoGeo Pocket",
		},
		"neo-geo-pocket-color": {
			rommSlug:   "neo-geo-pocket-color",
			datPattern: "SNK - NeoGeo Pocket Color",
		},
		"sega32": {
			rommSlug:   "sega32",
			datPattern: "Sega - 32X",
		},
		"gamegear": {
			rommSlug:   "gamegear",
			datPattern: "Sega - Game Gear",
		},
		"sms": {
			rommSlug:   "sms",
			datPattern: "Sega - Master System - Mark III",
		},
		"genesis": {
			rommSlug:   "genesis",
			datPattern: "Sega - Mega Drive - Genesis",
		},
	}

	for key, tc := range testCases {
		sysCfg, ok := cfg.Systems[key]
		if !ok {
			t.Fatalf("missing system %q in default config", key)
		}
		if !sysCfg.Enabled {
			t.Fatalf("expected system %q to be enabled", key)
		}
		if sysCfg.RommSlug != tc.rommSlug {
			t.Fatalf("romm slug mismatch for %q: got %q want %q", key, sysCfg.RommSlug, tc.rommSlug)
		}
		if sysCfg.DatPattern != tc.datPattern {
			t.Fatalf("dat pattern mismatch for %q: got %q want %q", key, sysCfg.DatPattern, tc.datPattern)
		}
	}
}
