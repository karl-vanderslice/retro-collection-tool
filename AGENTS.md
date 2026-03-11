# AGENTS

## Purpose

This repository contains a production-focused CLI that wraps Igir for curated ROM library workflows targeting ROMM.

## Guardrails

- Never modify `/mnt/media-emulation/RetroLibrary` while developing unless explicitly requested.
- Prefer `--dry-run` for validation in real libraries.
- Keep BIOS, ReDump, and Arcade paths behind explicit feature flags until implemented.

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
