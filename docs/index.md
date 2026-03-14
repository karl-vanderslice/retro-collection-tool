# retro-collection-tool

retro-collection-tool is a production-focused CLI that wraps Igir for curated ROM library workflows targeting ROMM.

## Why This Exists

Managing ROM libraries manually does not scale. This tool standardizes common workflows so they are:

- repeatable
- scriptable
- safer to run in automation

The CLI emphasizes deterministic behavior, explicit validation, and predictable output structure.

## Current Scope

Implemented:

- retail sync
- hacks workflow
- BIOS import workflow
- export workflow
- cache controls
- directory bootstrap

Explicitly stubbed behind feature flags:

- Arcade

ReDump support is available through retail sync configuration (`retail_dat_source: redump`) rather than a standalone full workflow.

## Typical Workflow

1. Set your library root in user config or `RETRO_COLLECTION_TOOL_ROOT`.
2. Run `bootstrap` once to create expected directory scaffolding.
3. Use `sync` (or `hacks`) with `--dry-run` first.
4. Run without `--dry-run` after validating command output.
5. Use `export` for target-specific copies.

## Start Here

- Read [Configuration](configuration.md) for config layering and key settings.
- Read [Commands](commands.md) for examples and flags.
- Read [Architecture](architecture.md) if you plan to contribute.
