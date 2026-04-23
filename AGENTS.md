# AGENTS

## Purpose

This repository contains a production-focused CLI that wraps Igir for curated ROM library workflows targeting ROMM.

## Guardrails

- Never modify `/mnt/media-emulation/RetroLibrary` while developing unless explicitly requested.
- Prefer `--dry-run` for validation in real libraries.
- Keep BIOS, ReDump, and Arcade paths behind explicit feature flags until implemented.
- Always commit completed work with a Conventional Commit message.
- Add or update tests for every behavior change before committing.
- Keep `flake.lock` tracked and committed (repo policy overrides global gitignore patterns).

## Current Scope

- Implemented: retail sync, hacks workflow, export, cache controls, bootstrap directories.
- Stubbed: BIOS, ReDump, Arcade.

## Development Commands

- `make build`
- `make test`
- `make lint`
- `make docs-serve`

## Documentation Standards

- Keep `README.md` as the GitHub entrypoint and `docs/index.md` as the docs
  landing page.
- Do not add duplicate overview pages such as `docs/README.md`.
- Keep command and package behavior documented first in Go doc comments when
  the surface is exported or reused, then use Markdown for operator workflows,
  configuration, and examples.
- If the repo grows machine-readable configuration knobs, define a schema
  instead of maintaining hand-written field inventories in multiple places.
- Use the docs hub flow from `terraform-cloudflare-docs-sites` for published
  docs instead of inventing per-repo deploy secrets.

## External References

- EmuDeck Nintendo Cheat Sheet: <https://emudeck.github.io/cheat-sheet/#nintendo-cheat-sheet>
- RetroDECK BIOS/Firmware Directory Structure: <https://retrodeck.readthedocs.io/en/latest/wiki_management/bios-firmware/#directory-structure>
- RomM Folder Structure (Structure A): <https://docs.romm.app/latest/Getting-Started/Folder-Structure/>

## Conventional Commits

Use semantic commit messages such as:

- `feat(cli): add export command`
- `fix(hacks): preserve base rom alongside patched output`
