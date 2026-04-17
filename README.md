# retro-collection-tool

Production-oriented CLI wrapper for [Igir](https://github.com/emmercm/igir), built to curate ROM libraries for [ROMM](https://github.com/rommapp/romm).

## Documentation

- Docs site: <https://karl-vanderslice.github.io/retro-collection-tool/>
- Docs source: `docs/`

## What it does

- Runs repeatable, config-driven ROM workflows with strict validation.
- Syncs retail ROM sets using the latest matching DAT (No-Intro or ReDump).
- Builds curated hacks from `roms/Hacks/<system>/<hack-name>/` into ROMM-compatible output.
- Imports BIOS files from configured source roots using catalog matching and optional hash verification.
- Exports selected systems to an external destination (for example, SD card media).

## Quick start

1. Enter the Nix environment:
   - `direnv allow`, or
   - `nix develop path:. --accept-flake-config`
2. Build the binary:
   - `just build`
3. Inspect enabled systems:
   - `bin/retro-collection-tool systems`
4. Create required folder layout safely (create-only):
   - `bin/retro-collection-tool bootstrap`
5. Preview a sync without writes:
   - `bin/retro-collection-tool --dry-run sync --systems nes,snes,genesis,sms`

## Commands at a glance

- `sync --systems <csv> | --all-systems [--compress] [--no-hacks]`
- `hacks --systems <csv> | --all-systems [--no-move-retail]`
- `bios --systems <csv> | --all-systems [--strict]`
- `clean --systems <csv> | --all-systems [--include-bios]`
- `export --systems <csv> | --all-systems --destination <path>`
- `cache clean|path`
- `bootstrap`
- `systems`
- `redump` (stub command)
- `arcade dats update|verify | verify | sync`

See the full command reference in `docs/commands.md`.

## Configuration

Default project config is `config/retro-collection-tool.yaml`.

Config files are merged in this order (low to high precedence):

1. XDG user config (`$XDG_CONFIG_HOME/retro-collection-tool/...`)
2. Project config (`./retro-collection-tool.yaml`, `./.retro-collection-tool.yaml`, `./config/retro-collection-tool.yaml`)
3. `RETRO_COLLECTION_TOOL_CONFIG`
4. `--config <path>`

After merge, `RETRO_COLLECTION_TOOL_ROOT` always overrides `root`.

Recommended practice:

- Keep `root` out of repository-tracked config.
- Set it in XDG user config or via `RETRO_COLLECTION_TOOL_ROOT`.

If `cache_dir` is omitted, cache defaults to `$XDG_CACHE_HOME/retro-collection-tool` (or `~/.cache/retro-collection-tool`).

## Safety and behavior

- `--dry-run` validates input and prints planned operations without writes.
- Commands fail fast and return non-zero on errors.
- CLI output is script-friendly for automation.

## Development

Host dependency model is Nix-only.

The Nix dev shell provides pinned arcade tooling via `flake.lock`, including `igir` and `chdman` (`mame-tools`), so arcade workflows do not need `npx` downloads when running inside the shell.

- `just fmt`
- `just lint`
- `just test`
- `just build`
- `just docs-serve`

## CI and docs publishing

GitHub Actions runs on push to `master`:

1. Format check (`just fmt` + clean diff).
2. Lint (`just lint`).
3. Tests (`just test`).
4. Build (`just build`).

If all steps pass, CI builds MkDocs and deploys to GitHub Pages.
