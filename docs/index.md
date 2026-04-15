# retro-collection-tool

retro-collection-tool is a production-focused CLI that wraps Igir for curated ROM library workflows targeting ROMM.

## Current scope

Implemented:

- retail sync
- hacks workflow
- BIOS import workflow
- arcade dat update/verify workflow
- arcade vault verify and hardlink sync
- export workflow
- cache controls
- directory bootstrap

Explicitly stubbed behind feature flags:

- ReDump standalone workflow command

ReDump support is available through retail sync configuration (`retail_dat_source: redump`) rather than a standalone full workflow.

## Typical workflow

1. Set your library root in user config or `RETRO_COLLECTION_TOOL_ROOT`.
2. Run `bootstrap` once to create expected directory scaffolding.
3. Use `sync` (or `hacks`) with `--dry-run` first.
4. Run without `--dry-run` after validating command output.
5. Use `export` for target-specific copies.

## Start here

- Read [Configuration](configuration.md) for config layering and key settings.
- Read [Commands](commands.md) for examples and flags.
- Read [RomVault Arcade Workflow (Stub)](workflow-romvault-arcade.md) for the planned vault sorting process notes.
- Read [Architecture](architecture.md) if you plan to contribute.
