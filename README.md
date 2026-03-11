# retro-collection-tool

Production-oriented CLI wrapper for [Igir](https://github.com/emmercm/igir), designed to curate ROM libraries for [ROMM](https://github.com/rommapp/romm).

## Highlights

- Config-driven workflows with strict validation.
- Retail sync via hardlinks using latest matching DAT.
- Curated hacks flow that patches from `roms/Hacks/<system>/<hack-name>/` and writes ROMM-compatible output.
- Export selected systems to an SD card destination.
- Explicit feature-flagged stubs for BIOS, ReDump, and Arcade.

## Quick Start

1. Enter dev shell: `direnv allow` or run commands via `nix develop path:. -c ...`.
2. Build: `nix develop path:. -c make build`.
3. Inspect systems: `bin/retro-collection-tool systems`.
4. Bootstrap directories (safe create-only):
   `bin/retro-collection-tool bootstrap`
5. Dry-run sync:
   `bin/retro-collection-tool --dry-run sync --systems nes,snes,genesis,sms`

Set library root safely outside repo config using either:

- XDG user config (recommended): `$XDG_CONFIG_HOME/retro-collection-tool/config.yaml`
- Environment variable override: `RETRO_COLLECTION_TOOL_ROOT=/mnt/media-emulation/RetroLibrary`

## Commands

- `sync --systems <csv> | --all-systems [--compress]`
- `hacks --systems <csv> | --all-systems [--no-move-retail]`
- `clean --systems <csv> | --all-systems [--include-bios]`
- `export --systems <csv> | --all-systems --destination <path>`
- `cache clean|path`
- `bootstrap`
- `systems`
- `bios` (stub)
- `redump` (stub)
- `arcade` (stub)

## Configuration

Default config: `config/retro-collection-tool.yaml`

Config is merged by precedence (low to high):

1. XDG user config (`$XDG_CONFIG_HOME/retro-collection-tool/config.yaml` etc.)
2. Project config (`./retro-collection-tool.yaml`, `./.retro-collection-tool.yaml`, or `./config/retro-collection-tool.yaml`)
3. `RETRO_COLLECTION_TOOL_CONFIG` (optional explicit layer)
4. `--config <path>` (optional explicit layer)

Then `RETRO_COLLECTION_TOOL_ROOT` overrides `root`.

- `root` should come from your user XDG config or `RETRO_COLLECTION_TOOL_ROOT`.
- `systems` maps platform slugs to DAT matching patterns and ROMM slugs.
- `features` gates unfinished workflows.
- Default enabled systems: `nes`, `snes`, `genesis`, `sms`.

If `cache_dir` is omitted, cache defaults to `$XDG_CACHE_HOME/retro-collection-tool`.

See `docs/configuration.md` for details.

## Development

- Host requirement is `nix` only.
- `nix develop path:. -c make fmt`
- `nix develop path:. -c make lint`
- `nix develop path:. -c make test`
- `nix develop path:. -c make hooks-install`
- `nix develop path:. -c make pre-commit`
- `nix develop path:. -c make docs-serve`

## Notes

- Tool is safe for automation and exits non-zero on failures.
- `--dry-run` is a plan mode: validates config/input selection and prints Igir commands without executing writes.
- Commands fail fast on unexpected trailing arguments.
- Uses conventional commit style in this repository.

## CI

GitHub Actions runs on push/PR to `master` and executes:

- `make fmt` (with diff check)
- `make lint`
- `make test`
- `make build`

It uploads `bin/retro-collection-tool` as a build artifact. No release tagging/publishing is performed.
