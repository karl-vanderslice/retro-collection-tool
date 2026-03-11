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

Config discovery order (Terraform/Vault-style):

1. `--config <path>`
2. `RETRO_COLLECTION_TOOL_CONFIG`
3. `./retro-collection-tool.yaml`
4. `./.retro-collection-tool.yaml`
5. `./config/retro-collection-tool.yaml`
6. `$XDG_CONFIG_HOME/retro-collection-tool/config.yaml` (or `~/.config/...` fallback)

- `root` points at the RetroLibrary root.
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
- `nix develop path:. -c make docs-serve`

## Notes

- Tool is safe for automation and exits non-zero on failures.
- `--dry-run` is a plan mode: validates config/input selection and prints Igir commands without executing writes.
- Commands fail fast on unexpected trailing arguments.
- Uses conventional commit style in this repository.
