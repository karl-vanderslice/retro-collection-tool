# retro-collection-tool

Production-oriented CLI wrapper for [Igir](https://github.com/emmercm/igir), designed to curate ROM libraries for [ROMM](https://github.com/rommapp/romm).

## Highlights

- Config-driven workflows with strict validation.
- Retail sync via hardlinks using latest matching DAT from No-Intro or ReDump.
- Curated hacks flow that patches from `roms/Hacks/<system>/<hack-name>/` and writes ROMM-compatible output.
- BIOS import workflow with strict catalog-based filename and MD5 matching.
- Export selected systems to an SD card destination.
- Explicit feature-flagged stubs for ReDump and Arcade.

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
- `bios`
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
- Default enabled systems: `gb`, `gba`, `gbc`, `gamegear`, `genesis`, `msx`, `msx2`, `n64`, `nes`, `neo-geo-pocket`, `neo-geo-pocket-color`, `sega32`, `sms`, `snes`, `supergrafx`, `tg16`, `dreamcast`, `gamecube`, `psx`, `ps2`, `saturn`, `wii`, `xbox`.

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

On pushes to `master` after a successful build, CI also runs:

- `mkdocs gh-deploy --force`

It uploads `bin/retro-collection-tool` as a build artifact. No release tagging/publishing is performed.
