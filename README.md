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
3. Inspect systems: `bin/retro-collection-tool --config config/retro-collection-tool.yaml systems`.
4. Bootstrap directories (safe create-only):
   `bin/retro-collection-tool --config config/retro-collection-tool.yaml bootstrap`
5. Dry-run sync:
   `bin/retro-collection-tool --config config/retro-collection-tool.yaml --dry-run sync --systems nes,snes`

## Commands

- `sync --systems <csv> | --all-systems [--compress]`
- `hacks --systems <csv> | --all-systems`
- `export --systems <csv> | --all-systems --destination <path>`
- `cache clean|path`
- `bootstrap`
- `systems`
- `bios` (stub)
- `redump` (stub)
- `arcade` (stub)

## Configuration

Default config: `config/retro-collection-tool.yaml`

- `root` points at the RetroLibrary root.
- `systems` maps platform slugs to DAT matching patterns and ROMM slugs.
- `features` gates unfinished workflows.

See `docs/configuration.md` for details.

## Development

- Host requirement is `nix` only.
- `nix develop path:. -c make fmt`
- `nix develop path:. -c make lint`
- `nix develop path:. -c make test`
- `nix develop path:. -c make docs-serve`

## Notes

- Tool is safe for automation and exits non-zero on failures.
- Prefer `--dry-run` before writes in production libraries.
- Uses conventional commit style in this repository.
