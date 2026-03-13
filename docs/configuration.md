# Configuration

Config is merged in layers (low to high precedence):

1. XDG user config:
   - `$XDG_CONFIG_HOME/retro-collection-tool/config.yaml`
   - `$XDG_CONFIG_HOME/retro-collection-tool/config.yml`
   - `$XDG_CONFIG_HOME/retro-collection-tool/retro-collection-tool.yaml`
2. Project config (first found):
   - `./retro-collection-tool.yaml`
   - `./.retro-collection-tool.yaml`
   - `./config/retro-collection-tool.yaml`
3. `RETRO_COLLECTION_TOOL_CONFIG` (optional explicit layer)
4. `--config <path>` (optional explicit layer, highest file precedence)

Finally, `RETRO_COLLECTION_TOOL_ROOT` overrides merged `root`.

When `XDG_CONFIG_HOME` is unset, `~/.config` is used.

## Core

- `root`: absolute path to RetroLibrary root. Recommended to keep in XDG user config or `RETRO_COLLECTION_TOOL_ROOT`, not in repo config.
- `cache_dir`: cache directory relative to root, absolute path, or omitted.

If `cache_dir` is omitted, cache defaults to `$XDG_CACHE_HOME/retro-collection-tool` (or `~/.cache/retro-collection-tool`).

## Igir

- `binary`: preferred executable (default `igir`).
- `use_npx_fallback`: use `npx --yes igir@latest` if `igir` is unavailable.
- `prefer_region`, `prefer_language`: applied to sync operations.
- `input_checksum_min`: checksum floor, e.g. `CRC32`.
- `cache_retail_file`, `cache_hacks_file`: cache file names.
- `allow_compression_zip`: gate for `sync --compress`.

## Paths

All paths are relative to `root` unless absolute.

`paths.hacks_source` can be set to an absolute path (for example `/mnt/media-emulation/RetroLibrary/roms/Hacks`) to keep curated hacks outside the active merge root.

`paths.vault_bios` can point to a BIOS vault location used as one of the BIOS source roots.

## BIOS

BIOS imports are catalog-driven and enforce strict hash matching for known files.

- `bios.catalog_file`: optional path to a custom BIOS catalog YAML.
  - If omitted, the built-in default catalog is used.
  - Relative paths are resolved from current working directory first, then from `root`.
- `bios.source_roots`: directories scanned for BIOS files and zip packs.

Only files that match both configured filename and MD5 are imported.
Unknown files are skipped and reported.

## Systems

Each system includes:

- `enabled`
- `romm_slug`
- `dat_pattern`

Optional overrides:

- `retail_dat_pattern`

If `retail_dat_pattern` is omitted, `dat_pattern` is used for retail sync.

Hacks do not require DAT files. The hacks workflow uses `rompatcherjs` and applies all discovered patch files in sorted filename order.

ReDump and Arcade are feature-gated stubs and do not require per-system DAT keys yet.

Retail DAT selection always picks the latest matching `.dat` by modification time.

### No-Intro Mapping Reference

The default config maps the following No-Intro folder names to internal system keys and ROMM slugs:

| No-Intro Folder                                     | System Key           | ROMM Slug            |
| --------------------------------------------------- | -------------------- | -------------------- |
| Microsoft - MSX                                     | msx                  | msx                  |
| Microsoft - MSX2                                    | msx2                 | msx2                 |
| NEC - PC Engine - TurboGrafx-16                     | tg16                 | tg16                 |
| NEC - PC Engine SuperGrafx                          | supergrafx           | supergrafx           |
| Nintendo - Game Boy                                 | gb                   | gb                   |
| Nintendo - Game Boy Advance                         | gba                  | gba                  |
| Nintendo - Game Boy Color                           | gbc                  | gbc                  |
| Nintendo - Nintendo 64 (BigEndian)                  | n64                  | n64                  |
| Nintendo - Nintendo Entertainment System (Headered) | nes                  | nes                  |
| Nintendo - Super Nintendo Entertainment System      | snes                 | snes                 |
| SNK - NeoGeo Pocket                                 | neo-geo-pocket       | neo-geo-pocket       |
| SNK - NeoGeo Pocket Color                           | neo-geo-pocket-color | neo-geo-pocket-color |
| Sega - 32X                                          | sega32               | sega32               |
| Sega - Game Gear                                    | gamegear             | gamegear             |
| Sega - Master System - Mark III                     | sms                  | sms                  |
| Sega - Mega Drive - Genesis                         | genesis              | genesis              |
