# Commands

## sync

Runs Igir retail synchronization with hardlinks.

With global `--dry-run`, this command performs validation and prints the exact Igir command without executing it.

Example:
`retro-collection-tool --dry-run sync --systems nes,snes,genesis,sms`

## hacks

Applies curated patches from `roms/Hacks/<system>/<hack-name>/`.

With global `--dry-run`, this command prints planned Igir patch invocations without writing files.

Hacks are now organized under the matched game directory when possible:
`roms/Library/roms/<system>/<game>/hack/<hack-name>.<ext>`

Game matching uses the base ROM name and normalizes region groups (for example `(USA, Europe)`), so translations/hacks can align with retail game folders.

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
