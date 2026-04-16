# Retro Collection Tool

[← docs.vslice.net](https://docs.vslice.net){ .md-button }

CLI that wraps [Igir](https://igir.io/) to turn loose ROM dumps into clean,
[ROMM](https://github.com/rommapp/romm)-ready libraries. Handles retail sync,
hacks patching, BIOS import, arcade DAT workflows, and SD-card export.

## Capabilities

| Workflow | Status |
| --- | --- |
| Retail sync (hardlinks via Igir) | Implemented |
| Hacks patching (`rompatcherjs`) | Implemented |
| BIOS import with catalog matching | Implemented |
| Arcade DAT update, verify, sync | Implemented |
| Export to SD cards | Implemented |
| Cache controls | Implemented |
| Directory bootstrap | Implemented |
| ReDump standalone command | Stubbed (available via `retail_dat_source: redump`) |

## Typical workflow

1. Set your library root in user config or `RETRO_COLLECTION_TOOL_ROOT`.
2. Run `bootstrap` to create the expected directory structure.
3. Use `sync` (or `hacks`) with `--dry-run` first.
4. Run without `--dry-run` after validating command output.
5. Use `export` for target-specific copies.

## Start here

- [Configuration](configuration.md) — config layering and key settings
- [Commands](commands.md) — CLI examples and flags
- [RomVault Arcade Workflow](workflow-romvault-arcade.md) — vault sorting process notes
- [Architecture](architecture.md) — internal design for contributors
