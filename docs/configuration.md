# Configuration

Primary config file: `config/retro-collection-tool.yaml`

Config discovery order:

1. `--config <path>`
2. `RETRO_COLLECTION_TOOL_CONFIG`
3. `./retro-collection-tool.yaml`
4. `./.retro-collection-tool.yaml`
5. `./config/retro-collection-tool.yaml`
6. `$XDG_CONFIG_HOME/retro-collection-tool/config.yaml`
7. `$XDG_CONFIG_HOME/retro-collection-tool/config.yml`
8. `$XDG_CONFIG_HOME/retro-collection-tool/retro-collection-tool.yaml`

When `XDG_CONFIG_HOME` is unset, `~/.config` is used.

## Core

- `root`: absolute path to RetroLibrary root.
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

## Systems

Each system includes:

- `enabled`
- `romm_slug`
- `retail_dat_pattern`
- `hack_dat_pattern`

BIOS, ReDump, and Arcade are currently feature-gated stubs and do not require per-system DAT keys yet.

DAT selection always picks the latest matching `.dat` by modification time.
