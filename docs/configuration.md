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

## Systems

Each system includes:

- `enabled`
- `romm_slug`
- `dat_pattern`

Optional overrides:

- `retail_dat_pattern`
- `hack_dat_pattern`

If override keys are omitted, `dat_pattern` is used for both workflows.

BIOS, ReDump, and Arcade are currently feature-gated stubs and do not require per-system DAT keys yet.

DAT selection always picks the latest matching `.dat` by modification time.
