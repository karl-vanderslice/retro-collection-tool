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
	Bios      BiosConfig               `yaml:"bios"`
	Arcade    ArcadeConfig             `yaml:"arcade"`
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
	VaultBios       string `yaml:"vault_bios"`
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
	RetailDatSource  string `yaml:"retail_dat_source"`
	DatPattern       string `yaml:"dat_pattern"`
	RetailDatPattern string `yaml:"retail_dat_pattern"`
	HackDatPattern   string `yaml:"hack_dat_pattern"`
}

const (
	RetailDatSourceNoIntro = "nointro"
	RetailDatSourceRedump  = "redump"
)

type FeatureToggles struct {
	EnableBios   bool `yaml:"enable_bios"`
	EnableRedump bool `yaml:"enable_redump"`
	EnableArcade bool `yaml:"enable_arcade"`
}

type BiosConfig struct {
	CatalogFile string   `yaml:"catalog_file"`
	SourceRoots []string `yaml:"source_roots"`
}

type ArcadeConfig struct {
	VaultMAME2003Plus string   `yaml:"vault_mame_2003_plus"`
	VaultFBNeo        string   `yaml:"vault_fbneo"`
	LibraryMAME2003   string   `yaml:"library_mame_2003_plus"`
	LibraryFBNeo      string   `yaml:"library_fbneo"`
	DatMAME2003URL    string   `yaml:"dat_mame_2003_plus_url"`
	DatFBNeoURL       string   `yaml:"dat_fbneo_url"`
	DatMAME2003File   string   `yaml:"dat_mame_2003_plus_file"`
	DatFBNeoFile      string   `yaml:"dat_fbneo_file"`
	ExcludeKeywords   []string `yaml:"exclude_keywords"`
}

type BootstrapDirectoryLayout struct {
	RommRoms []string `yaml:"romm_roms"`
	RommBios []string `yaml:"romm_bios"`
}

type EnvOverrides struct {
	Root string
}

func Load(path string) (*Config, error) {
	return LoadMerged([]string{path}, EnvOverrides{})
}

func LoadMerged(paths []string, env EnvOverrides) (*Config, error) {
	if len(paths) == 0 {
		return nil, errors.New("no config files provided")
	}

	merged := map[string]any{}

	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", path, err)
		}

		var part map[string]any
		if err := yaml.Unmarshal(b, &part); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}

		mergeMaps(merged, part)
	}

	if strings.TrimSpace(env.Root) != "" {
		merged["root"] = strings.TrimSpace(env.Root)
	}

	finalBytes, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(finalBytes, &cfg); err != nil {
		return nil, fmt.Errorf("parse merged config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func mergeMaps(dst map[string]any, src map[string]any) {
	for k, v := range src {
		srcMap, srcIsMap := v.(map[string]any)
		dstMap, dstIsMap := dst[k].(map[string]any)
		if srcIsMap && dstIsMap {
			mergeMaps(dstMap, srcMap)
			dst[k] = dstMap
			continue
		}
		dst[k] = v
	}
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

	usesRedump := false
	for k, v := range c.Systems {
		if strings.TrimSpace(v.RommSlug) == "" {
			return fmt.Errorf("config.systems.%s.romm_slug is required", k)
		}
		source := v.EffectiveRetailDatSource()
		if source != RetailDatSourceNoIntro && source != RetailDatSourceRedump {
			return fmt.Errorf("config.systems.%s.retail_dat_source must be one of: %s, %s", k, RetailDatSourceNoIntro, RetailDatSourceRedump)
		}
		if source == RetailDatSourceRedump {
			usesRedump = true
		}
		if strings.TrimSpace(v.DatPattern) == "" && strings.TrimSpace(v.RetailDatPattern) == "" && strings.TrimSpace(v.HackDatPattern) == "" {
			return fmt.Errorf("config.systems.%s requires dat_pattern or retail_dat_pattern or hack_dat_pattern", k)
		}
	}

	if usesRedump {
		if strings.TrimSpace(c.Paths.DatsRedump1G1R) == "" {
			return errors.New("config.paths.dats_redump_1g1r is required when any system uses retail_dat_source=redump")
		}
		if strings.TrimSpace(c.Paths.VaultRedump) == "" {
			return errors.New("config.paths.vault_redump is required when any system uses retail_dat_source=redump")
		}
	}
	return nil
}

func (s SystemConfig) EffectiveRetailDatPattern() string {
	if p := strings.TrimSpace(s.RetailDatPattern); p != "" {
		return p
	}
	return strings.TrimSpace(s.DatPattern)
}

func (s SystemConfig) EffectiveRetailDatSource() string {
	if source := strings.ToLower(strings.TrimSpace(s.RetailDatSource)); source != "" {
		return source
	}
	return RetailDatSourceNoIntro
}

func (s SystemConfig) EffectiveHackDatPattern() string {
	if p := strings.TrimSpace(s.HackDatPattern); p != "" {
		return p
	}
	return strings.TrimSpace(s.DatPattern)
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

func (c *Config) ArcadeVaultMAME2003Plus() string {
	if v := strings.TrimSpace(c.Arcade.VaultMAME2003Plus); v != "" {
		return c.ResolvePath(v)
	}
	if v := strings.TrimSpace(c.Paths.MediaRoot); v != "" {
		return filepath.Join(c.ResolvePath(v), "roms", "Vault", "Arcade", "mame-2003-plus-reference-set", "roms")
	}
	return c.ResolvePath("roms/Vault/Arcade/mame-2003-plus-reference-set/roms")
}

func (c *Config) ArcadeVaultFBNeo() string {
	if v := strings.TrimSpace(c.Arcade.VaultFBNeo); v != "" {
		return c.ResolvePath(v)
	}
	if v := strings.TrimSpace(c.Paths.MediaRoot); v != "" {
		return filepath.Join(c.ResolvePath(v), "roms", "Vault", "Arcade", "fbneo_1003_bestset", "fbneo_1_0_0_3_best", "games")
	}
	return c.ResolvePath("roms/Vault/Arcade/fbneo_1003_bestset/fbneo_1_0_0_3_best/games")
}

func (c *Config) ArcadeLibraryMAME2003Plus() string {
	if v := strings.TrimSpace(c.Arcade.LibraryMAME2003); v != "" {
		return c.ResolvePath(v)
	}
	return filepath.Join(c.ResolvePath(c.Paths.RommLibraryRoms), "arcade", "mame-2003-plus")
}

func (c *Config) ArcadeLibraryFBNeo() string {
	if v := strings.TrimSpace(c.Arcade.LibraryFBNeo); v != "" {
		return c.ResolvePath(v)
	}
	return filepath.Join(c.ResolvePath(c.Paths.RommLibraryRoms), "arcade", "fbneo")
}

func (c *Config) ArcadeDatMAME2003PlusURL() string {
	if v := strings.TrimSpace(c.Arcade.DatMAME2003URL); v != "" {
		return v
	}
	return "https://raw.githubusercontent.com/libretro/libretro-database/master/metadat/mame-nonmerged/MAME%202003-Plus.dat"
}

func (c *Config) ArcadeDatFBNeoURL() string {
	if v := strings.TrimSpace(c.Arcade.DatFBNeoURL); v != "" {
		return v
	}
	return "https://git.libretro.com/libretro/FBNeo/-/raw/master/dats/FinalBurn%20Neo%20%28ClrMame%20Pro%20XML%2C%20Arcade%20only%29.dat"
}

func (c *Config) ArcadeDatMAME2003PlusFile() string {
	if v := strings.TrimSpace(c.Arcade.DatMAME2003File); v != "" {
		return v
	}
	return "arcade-mame-2003-plus.dat"
}

func (c *Config) ArcadeDatFBNeoFile() string {
	if v := strings.TrimSpace(c.Arcade.DatFBNeoFile); v != "" {
		return v
	}
	return "arcade-fbneo.dat"
}

func (c *Config) ArcadeExcludeKeywords() []string {
	if len(c.Arcade.ExcludeKeywords) > 0 {
		return c.Arcade.ExcludeKeywords
	}
	return []string{"mahjong", "medal", "gambl", "adult", "hentai", "electromechanical"}
}
