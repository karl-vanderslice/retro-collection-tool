package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Root      string                   `yaml:"root"`
	CacheDir  string                   `yaml:"cache_dir"`
	Igir      IgirConfig               `yaml:"igir"`
	Paths     PathsConfig              `yaml:"paths"`
	Systems   map[string]SystemConfig  `yaml:"systems"`
	Features  FeatureToggles           `yaml:"features"`
	Bootstrap BootstrapDirectoryLayout `yaml:"bootstrap"`
}

type IgirConfig struct {
	Binary              string   `yaml:"binary"`
	UseNpxFallback      bool     `yaml:"use_npx_fallback"`
	PreferRegion        []string `yaml:"prefer_region"`
	PreferLanguage      []string `yaml:"prefer_language"`
	InputChecksumMin    string   `yaml:"input_checksum_min"`
	CacheRetailFile     string   `yaml:"cache_retail_file"`
	CacheHacksFile      string   `yaml:"cache_hacks_file"`
	RetailBaseArgs      []string `yaml:"retail_base_args"`
	AllowCompressionZip bool     `yaml:"allow_compression_zip"`
}

type PathsConfig struct {
	VaultNoIntro    string `yaml:"vault_no_intro"`
	VaultRedump     string `yaml:"vault_redump"`
	ToSort          string `yaml:"to_sort"`
	RommLibraryRoms string `yaml:"romm_library_roms"`
	RommLibraryBios string `yaml:"romm_library_bios"`
	HacksSource     string `yaml:"hacks_source"`
	DatsNoIntro1G1R string `yaml:"dats_no_intro_1g1r"`
	DatsNoIntroRaw  string `yaml:"dats_no_intro_raw"`
	DatsRedump1G1R  string `yaml:"dats_redump_1g1r"`
	MediaRoot       string `yaml:"media_root"`
}

type SystemConfig struct {
	Enabled          bool   `yaml:"enabled"`
	RommSlug         string `yaml:"romm_slug"`
	RetailDatPattern string `yaml:"retail_dat_pattern"`
	HackDatPattern   string `yaml:"hack_dat_pattern"`
}

type FeatureToggles struct {
	EnableBios   bool `yaml:"enable_bios"`
	EnableRedump bool `yaml:"enable_redump"`
	EnableArcade bool `yaml:"enable_arcade"`
}

type BootstrapDirectoryLayout struct {
	RommRoms []string `yaml:"romm_roms"`
	RommBios []string `yaml:"romm_bios"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Root) == "" {
		return errors.New("config.root is required")
	}
	if strings.TrimSpace(c.Paths.RommLibraryRoms) == "" {
		return errors.New("config.paths.romm_library_roms is required")
	}
	if strings.TrimSpace(c.Paths.HacksSource) == "" {
		return errors.New("config.paths.hacks_source is required")
	}
	if len(c.Systems) == 0 {
		return errors.New("config.systems must define at least one system")
	}

	for k, v := range c.Systems {
		if strings.TrimSpace(v.RommSlug) == "" {
			return fmt.Errorf("config.systems.%s.romm_slug is required", k)
		}
		if strings.TrimSpace(v.RetailDatPattern) == "" {
			return fmt.Errorf("config.systems.%s.retail_dat_pattern is required", k)
		}
	}
	return nil
}

func (c *Config) ResolvePath(p string) string {
	if p == "" {
		return ""
	}
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(c.Root, p)
}

func (c *Config) EnabledSystems() []string {
	keys := make([]string, 0, len(c.Systems))
	for k, v := range c.Systems {
		if v.Enabled {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}
