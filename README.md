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
- Converts pre-curated set layouts (Done Set 3) into NextUI-ready `Roms` + `Bios` structure with flattened ROM folders and copied artwork.
  - Validates source system folder names against known NextUI emulator tags before conversion.
  - Rebuilds destination `Roms/` and `Bios/` from scratch on each run (clean-first) to avoid duplicate carry-over.
  - Excludes `Translations`, `Unlicensed Homebrew`, and `Hacks` folders by default for a purist baseline set.
  - Applies release-order numbered system folders (for example `06) Nintendo Entertainment System (FC)`).
  - Uses a single Arcade destination by merging `ARCADE`, `MAME`, `FBNeo`, `CPS3`, and `NEOGEO` sources.
  - Prefers FBNeo `map.txt` metadata for arcade display names and filters merged arcade ROMs to names present in the selected arcade map set when available.
  - Splits `MD` `32X Games (...)` content into dedicated `Roms/10) Sega 32X (32X)/` with artwork routed to `Roms/10) Sega 32X (32X)/.media/`.
  - Copies `map.txt` to the Arcade folder when present in source folders and emits tab-delimited `map.txt` files for NextUI compatibility.
  - `.7z` ROM archives are converted to `.zip` for NextUI compatibility.
  - PlayStation `.m3u` + `.hidden` multi-disc layouts are preserved.
  - `SEGACD`, `PCECD`, `DOS`, `SCUMMVM`, and `PORTS` retain folder trees instead of flattening.
  - Copies `.cht` cheat files from `Cheats/<system>/` into `Cheats/<NextUI_TAG>/`.
  - Auto-generates `Collections/*.txt` for major series (Final Fantasy, Castlevania, Metroid, Mario, Donkey Kong, TMNT, Zelda, Mega Man, Sonic, Pokemon).

## Quick start

1. Enter the Nix environment with `direnv allow` or `nix develop path:. --accept-flake-config`.
1. Build the binary with `just build`.
1. Inspect enabled systems with `bin/retro-collection-tool systems`.
1. Create required folder layout safely (create-only) with `bin/retro-collection-tool bootstrap`.
1. Preview a sync without writes with `bin/retro-collection-tool --dry-run sync --systems nes,snes,genesis,sms`.

## Commands at a glance

- `sync --systems <csv> | --all-systems [--compress] [--no-hacks]`
- `hacks --systems <csv> | --all-systems [--no-move-retail]`
- `bios --systems <csv> | --all-systems [--strict]`
- `clean --systems <csv> | --all-systems [--include-bios]`
- `export --systems <csv> | --all-systems --destination <path>`
- `curated convert --set done-set-3 --target nextui --source <path> --destination <path>`
- `cache clean|path`
- `bootstrap`
- `systems`
- `redump` (stub command)
- `arcade dats update|verify | verify | sync`

See the full command reference in `docs/commands.md`.

Done Set 3 to NextUI example:

- `bin/retro-collection-tool curated convert --set done-set-3 --target nextui --source "/mnt/d/done set/final" --destination "/mnt/d/done set/export"`
- `bin/retro-collection-tool --dry-run curated convert --set done-set-3 --target nextui --source "/mnt/d/done set/final" --destination "/mnt/d/done set/export"`

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

## Quality and docs publishing

Run quality checks locally through the repo entrypoints:

1. `just fmt`
2. `just lint`
3. `just test`
4. `just build`

Documentation for this repository is published through the centralized
`docs.vslice.net` hub workflow.
