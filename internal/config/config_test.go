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

func TestSystemConfigEffectiveRetailDatSource_Default(t *testing.T) {
	t.Parallel()

	s := SystemConfig{}
	if got := s.EffectiveRetailDatSource(); got != RetailDatSourceNoIntro {
		t.Fatalf("retail source default mismatch: %q", got)
	}
}

func TestSystemConfigEffectiveRetailDatSource_Override(t *testing.T) {
	t.Parallel()

	s := SystemConfig{RetailDatSource: "ReDump"}
	if got := s.EffectiveRetailDatSource(); got != RetailDatSourceRedump {
		t.Fatalf("retail source override mismatch: %q", got)
	}
}

func TestLoadMergedRejectsUnknownRetailDatSource(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "config.yaml")
	cfgYAML := strings.Join([]string{
		"root: /tmp/root",
		"cache_dir: cache",
		"paths:",
		"  romm_library_roms: roms/Library/roms",
		"  hacks_source: roms/Hacks",
		"systems:",
		"  psx:",
		"    enabled: true",
		"    romm_slug: psx",
		"    retail_dat_source: invalid-source",
		"    retail_dat_pattern: Sony - PlayStation",
		"",
	}, "\n")
	if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadMerged([]string{cfgPath}, EnvOverrides{})
	if err == nil {
		t.Fatal("expected validation error for invalid retail_dat_source")
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
		"3ds": {
			rommSlug:   "3ds",
			datPattern: "Nintendo - Nintendo 3DS (Decrypted)",
		},
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
		"jaguar": {
			rommSlug:   "jaguar",
			datPattern: "Atari - Atari Jaguar (ROM)",
		},
		"lynx": {
			rommSlug:   "lynx",
			datPattern: "Atari - Atari Lynx (LYX)",
		},
		"nds": {
			rommSlug:   "nds",
			datPattern: "Nintendo - Nintendo DS (Decrypted)",
		},
		"new-nintendo-3ds": {
			rommSlug:   "new-nintendo-3ds",
			datPattern: "Nintendo - New Nintendo 3DS (Decrypted)",
		},
		"nintendo-dsi": {
			rommSlug:   "nintendo-dsi",
			datPattern: "Nintendo - Nintendo DSi (Decrypted)",
		},
		"wiiu": {
			rommSlug:   "wiiu",
			datPattern: "Nintendo - Wii U (Digital) (CDN)",
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

	redumpSystems := map[string]struct {
		rommSlug string
		pattern  string
	}{
		"3do": {
			rommSlug: "3do",
			pattern:  "Panasonic - 3DO Interactive Multiplayer",
		},
		"dreamcast": {
			rommSlug: "dreamcast",
			pattern:  "Sega - Dreamcast",
		},
		"jaguar-cd": {
			rommSlug: "atari-jaguar-cd",
			pattern:  "Atari - Jaguar CD Interactive Multimedia System",
		},
		"neo-geo-cd": {
			rommSlug: "neo-geo-cd",
			pattern:  "SNK - Neo Geo CD",
		},
		"gamecube": {
			rommSlug: "gamecube",
			pattern:  "Nintendo - GameCube",
		},
		"ps3": {
			rommSlug: "ps3",
			pattern:  "Sony - PlayStation 3",
		},
		"psp": {
			rommSlug: "psp",
			pattern:  "Sony - PlayStation Portable",
		},
		"psx": {
			rommSlug: "psx",
			pattern:  "Sony - PlayStation",
		},
		"ps2": {
			rommSlug: "ps2",
			pattern:  "Sony - PlayStation 2",
		},
		"saturn": {
			rommSlug: "saturn",
			pattern:  "Sega - Saturn",
		},
		"wii": {
			rommSlug: "wii",
			pattern:  "Nintendo - Wii",
		},
		"xbox360": {
			rommSlug: "xbox360",
			pattern:  "Microsoft - Xbox 360",
		},
		"xbox": {
			rommSlug: "xbox",
			pattern:  "Microsoft - Xbox",
		},
		"turbografx-cd": {
			rommSlug: "turbografx-cd",
			pattern:  "NEC - PC Engine CD & TurboGrafx CD",
		},
	}

	for key, tc := range redumpSystems {
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
		if sysCfg.EffectiveRetailDatPattern() != tc.pattern {
			t.Fatalf("retail dat pattern mismatch for %q: got %q want %q", key, sysCfg.EffectiveRetailDatPattern(), tc.pattern)
		}
		if sysCfg.EffectiveRetailDatSource() != RetailDatSourceRedump {
			t.Fatalf("retail dat source mismatch for %q: got %q want %q", key, sysCfg.EffectiveRetailDatSource(), RetailDatSourceRedump)
		}
	}
}
