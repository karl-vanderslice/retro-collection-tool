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

## Conventional Commits

Use semantic commit messages such as:

- `feat(cli): add export command`
- `fix(hacks): preserve base rom alongside patched output`
