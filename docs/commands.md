# Commands

## Global flags

- `--config <path>`: add a highest-precedence config file layer.
- `--dry-run`: validate and print planned actions without writing changes.

Use `--dry-run` first when targeting real libraries.

## sync

Runs the default curation pipeline:

1. Retail sync with Igir hardlinks.
2. Hacks build with `rompatcherjs`.
3. Folder organization per game (`<game>/<retail-file>` and `<game>/hack/<hack-file>`).

Use `--no-hacks` to run retail sync only.

With global `--dry-run`, this command validates selection/config and prints the exact Igir command without executing it.

Example:
`retro-collection-tool --dry-run sync --systems nes,snes,genesis,sms`

## hacks

Applies curated patches from `roms/Hacks/<system>/<hack-name>/` and writes ROMM-compatible outputs.

Safety flag:
`--no-move-retail` keeps retail ROM files in place and only writes hack outputs.

With global `--dry-run`, this command prints the planned sequential patch chain without writing files.

Patching is performed with `rompatcherjs` (`npx --yes rom-patcher`) and applies all patch files in filename order (`.ips`, `.bps`, `.ups`, `.xdelta`, and other supported formats).

Hacks are now organized under the matched game directory when possible:
`roms/Library/roms/<system>/<game>/hack/<hack-name>.<ext>`

Game matching uses the base ROM name and normalizes region groups (for example `(USA, Europe)`), so translations/hacks can align with retail game folders.

Matching retail ROM files in the system root are moved into the matched `<game>/` folder. Source files from `roms/Hacks` are used only as patching inputs.

## clean

Removes generated output directories for selected systems.

Flags:

- `--systems <csv>` or `--all-systems`
- `--include-bios` to include BIOS targets in clean operations

Examples:
`retro-collection-tool clean --systems genesis`

`retro-collection-tool --dry-run clean --all-systems --include-bios`

## bios

Imports BIOS files into ROMM Structure A targets:
`roms/Library/bios/<platform>/...`

The BIOS workflow is feature-gated by `features.enable_bios` and uses catalog matching for known BIOS files.

- Matching always uses filename from the catalog.
- If a catalog source includes one or more signatures (`md5`, `sha1`, `sha256`, `crc32`), every provided signature must match.
- If a catalog source omits MD5, filename-only matching is used.
- Unknown files are skipped and reported.
- Source roots can include raw files and zip packs.

Flags:

- `--systems <csv>` or `--all-systems`
- `--strict` to fail when required BIOS entries are missing

Examples:

`retro-collection-tool bios --systems gba,gbc`

`retro-collection-tool --dry-run bios --all-systems --strict`

## export

Copies selected systems to another destination (for SD cards).

Required flag:

- `--destination <path>`

Example:
`retro-collection-tool export --systems nes --destination /run/media/user/SDCARD/roms`

## curated

Converts pre-curated ROM packs into target firmware layouts.

Currently supported:

- `--set done-set-3`
- `--target nextui`

Required flags:

- `--source <path>`: extracted Done Set 3 root containing `Roms/` and `BIOS/`
- `--destination <path>`: export root where `Roms/` and `Bios/` will be created

Behavior for Done Set 3 -> NextUI:

- Cleans destination `Roms/` and `Bios/` first on every run to avoid duplicate carry-over from prior exports.
- Writes only `Roms/` and `Bios/` into the destination root.
- Also writes `Collections/` with franchise-focused collection lists.
- Flattens ROM subfolders into each system folder for scroll-first browsing.
- Excludes `Translations`, `Unlicensed Homebrew`, and `Hacks` folders by default for a purist baseline set.
- Uses numbered release-order folder naming (for example `06) Nintendo Entertainment System (FC)`) so menu order is deterministic.
- Merges `ARCADE`, `MAME`, `FBNeo`, `CPS3`, and `NEOGEO` sources into one `Arcade` destination folder.
- For `MD`, routes `32X Games (...)` content into dedicated `Roms/10) Sega 32X (32X)/` (with matching artwork in `Roms/10) Sega 32X (32X)/.media/`).
- Copies `map.txt` into the Arcade folder when present in source arcade folders.
- Converts `.7z` archives to `.zip` (keeps all other ROM file extensions unchanged).
- For flatten-mode systems already predominantly using `.zip`, converts remaining raw single-file ROMs (for example `.gba`, `.smc`, `.sfc`) to `.zip` for folder uniformity.
- Copies artwork from `Roms/<system>/Imgs/*.png` into `Roms/<system>/.media/*.png`.
- Recursively copies BIOS from `BIOS/` to `Bios/`.
- Preserves PlayStation `.hidden` content so `.m3u` playlists that reference `.hidden/...` continue to work while disc images stay hidden from normal browsing.
- Preserves directory trees for `DOS`, `SCUMMVM`, and `PORTS` instead of flattening, which matches typical NextUI pak expectations for those systems.
- Generates franchise collections such as Final Fantasy, Castlevania, Metroid, Mario, Donkey Kong, TMNT, Zelda, Mega Man, Sonic, and Pokemon.

Example:
`retro-collection-tool curated convert --set done-set-3 --target nextui --source "/mnt/d/done set/final" --destination "/mnt/d/done set/export"`

## cache

- `cache path`: print active cache path
- `cache clean`: remove cache files

## bootstrap

Creates expected directory structure for configured systems.

## systems

Prints enabled systems.

## arcade

Arcade workflow is feature-gated by `features.enable_arcade` and now supports:

- `arcade dats update`: download/update MAME 2003 Plus and FBNeo DAT files into cache.
- `arcade dats verify`: ensure cached DAT files exist and are non-empty.
- `arcade verify`: run Igir in dry-run mode against arcade DAT + vault inputs.
- `arcade sync`: run Igir with hardlink output into ROMM library targets.

Arcade processing now delegates compatibility/filtering logic to Igir and DAT semantics.

Examples:

`retro-collection-tool arcade dats update`

`retro-collection-tool arcade dats verify`

`retro-collection-tool arcade verify`

`retro-collection-tool --dry-run arcade sync`

## Stub Commands

- `redump` is currently a placeholder command.
