# Commands

## sync

Runs Igir retail synchronization with hardlinks.

With global `--dry-run`, this command performs validation and prints the exact Igir command without executing it.

Example:
`retro-collection-tool --dry-run sync --systems nes,snes,genesis,sms`

## hacks

Applies curated patches from `roms/Hacks/<system>/<hack-name>/`.

Safety flag:
`--no-move-retail` keeps retail ROM files in place and only writes hack outputs.

With global `--dry-run`, this command prints the planned sequential patch chain without writing files.

Patching is performed with `rompatcherjs` (`npx --yes rom-patcher`) and applies all patch files in filename order (`.ips`, `.bps`, `.ups`, `.xdelta`, and other supported formats).

Hacks are now organized under the matched game directory when possible:
`roms/Library/roms/<system>/<game>/hack/<hack-name>.<ext>`

Game matching uses the base ROM name and normalizes region groups (for example `(USA, Europe)`), so translations/hacks can align with retail game folders.

Matching retail ROM files in the system root are moved into the matched `<game>/` folder. Source files from `roms/Hacks` are only used for patching inputs.

## clean

Removes target output directories for selected systems.

Examples:
`retro-collection-tool clean --systems genesis`

`retro-collection-tool --dry-run clean --all-systems --include-bios`

## export

Copies selected systems to another destination (for SD cards).

Example:
`retro-collection-tool export --systems nes --destination /run/media/user/SDCARD/roms`

## cache

- `cache path`
- `cache clean`

## bootstrap

Creates required directory structure for configured systems.

## systems

Prints enabled systems.
