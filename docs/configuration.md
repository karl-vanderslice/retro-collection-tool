# Configuration

Configuration is merged in layers (low to high precedence):

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

## Recommended Setup

- Keep machine-specific `root` out of repository-tracked config.
- Put personal overrides in XDG user config.
- Use `RETRO_COLLECTION_TOOL_ROOT` in automation or temporary shells.
- Treat `--config` as a final override for one-off runs.

## Core

- `root`: absolute path to RetroLibrary root. Recommended to keep in XDG user config or `RETRO_COLLECTION_TOOL_ROOT`, not in repo config.
- `cache_dir`: cache directory relative to root, absolute path, or omitted.

If `cache_dir` is omitted, cache defaults to `$XDG_CACHE_HOME/retro-collection-tool` (or `~/.cache/retro-collection-tool`).

## Igir

- `binary`: preferred executable (default `igir`).
- `use_npx_fallback`: when true, use `npx --yes igir@latest` if `igir` is unavailable.

When using the provided Nix dev shell, prefer `use_npx_fallback: false` so `igir` is fully pinned via `flake.lock`.

- `prefer_region`, `prefer_language`: applied during retail sync selection.
- `input_checksum_min`: checksum floor, for example `CRC32`.
- `cache_retail_file`, `cache_hacks_file`: cache file names.
- `allow_compression_zip`: gate for `sync --compress`.

## Paths

All paths are relative to `root` unless absolute.

`paths.hacks_source` can be set to an absolute path (for example `/mnt/media-emulation/RetroLibrary/roms/Hacks`) to keep curated hacks outside the active merge root.

`paths.vault_bios` can point to a BIOS vault location used as one of the BIOS source roots.

## BIOS

BIOS imports are catalog-driven and can enforce hash matching for known files.

- `bios.catalog_file`: optional path to a custom BIOS catalog YAML.
  - If omitted, the built-in default catalog is used.
  - Relative paths are resolved from current working directory first, then from `root`.
- `bios.source_roots`: directories scanned for BIOS files and zip packs.

Only files that match configured catalog names are imported.
When a source includes signatures (`md5`, `sha1`, `sha256`, `crc32`), all provided signatures must match.
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

Arcade is feature-gated behind `features.enable_arcade`.

## Arcade

Arcade workflow configuration is defined under `arcade`:

- `vault_mame_2003_plus`: source folder for MAME 2003 Plus archives.
- `vault_fbneo`: source folder for FBNeo archives.
- `library_mame_2003_plus`: output folder for curated MAME 2003 Plus links.
- `library_fbneo`: output folder for curated FBNeo links.
- `dat_mame_2003_plus_url`: URL used by `arcade dats update`.
- `dat_fbneo_url`: URL used by `arcade dats update`.
- `dat_mame_2003_plus_file`: cached DAT filename under `<cache>/arcade/dats`.
- `dat_fbneo_file`: cached DAT filename under `<cache>/arcade/dats`.

Arcade verification and sync use Igir with DAT-driven behavior and `--no-bios --no-device` flags.

Retail DAT selection always picks the latest matching `.dat` by modification time.

## Validation Expectations

- Unknown fields are rejected.
- Invalid or missing required values fail early with descriptive messages.
- Feature-gated workflows must be explicitly enabled before use.

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
