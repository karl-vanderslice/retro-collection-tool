# Commands

## sync

Runs Igir retail synchronization with hardlinks.

Example:
`retro-collection-tool --dry-run sync --systems nes,snes`

## hacks

Applies curated patches from `roms/Hacks/<system>/<hack-name>/`.

Output layout per ROMM style:
`roms/Library/roms/<system>/<hack-name>/hack/<hack-name>.<ext>`

Also attempts to include unaltered base ROM in the same `<hack-name>/` directory.

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
