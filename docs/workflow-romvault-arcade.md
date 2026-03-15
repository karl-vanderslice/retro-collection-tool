# RomVault Arcade Workflow (Stub)

This page is a placeholder for documenting the full RomVault-based arcade workflow used with this project.

Status: TODO (draft scaffold only)

## Goals

- Document how Vault sources are prepared and maintained.
- Document how curated outputs are linked into a ROMM-compatible `arcade` folder.
- Document frontend launch configuration expectations (ES-DE and others).

## Current Working Inputs

- MAME 2003 Plus Vault source:
  - `/mnt/media-emulation/RetroLibrary/roms/Vault/Arcade/mame-2003-plus-reference-set/roms`
- FBNeo Vault source:
  - `/mnt/media-emulation/RetroLibrary/roms/Vault/Arcade/fbneo_1003_bestset/fbneo_1_0_0_3_best/games`

## Planned Steps (To Be Expanded)

1. Refresh DATs into cache:
   - `retro-collection-tool arcade dats update`
2. Validate DAT parsing/filter rules:
   - `retro-collection-tool arcade dats verify`
3. Verify vault coverage report:
   - `retro-collection-tool arcade verify`
4. Perform hardlink sync into ROMM arcade folder:
   - `retro-collection-tool arcade sync`

## TODO: RomVault Sorting Notes

- Add exact RomVault profile/layout assumptions.
- Add naming/normalization rules expected before running sync.
- Add reconciliation flow for missing or mismatched sets.
- Add collision review process for archives present in both FBNeo and MAME sets.

## TODO: Frontend Configuration (ES-DE and Others)

- Add ES-DE platform/core mapping examples for single `arcade` folder layout.
- Add FBNeo-first, MAME-2003 fallback strategy.
- Add examples for RetroArch core assignment and launch command conventions.
- Add notes for alternative frontends with similar core-priority behavior.

## Notes

- Current implementation links both filtered game archives and DAT-marked BIOS archives.
- If both sets target one output folder, later-linked files with the same archive name replace earlier ones.
