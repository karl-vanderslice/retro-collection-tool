# Configuration

Primary config file: `config/retro-collection-tool.yaml`

## Core

- `root`: absolute path to RetroLibrary root.
- `cache_dir`: cache directory relative to root.

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
- optional placeholders for BIOS/ReDump/Arcade patterns

DAT selection always picks the latest matching `.dat` by modification time.
