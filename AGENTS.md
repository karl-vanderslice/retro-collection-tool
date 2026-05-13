# AGENTS

## Snapshot

- Purpose: this repo contains a production-focused CLI that wraps Igir for
  curated ROM library workflows targeting ROMM.
- Load order: load this file first. Repo-local prompt/skill overlays are not
  present yet.
- Current scope: implemented retail sync, hacks workflow, export, cache
  controls, and bootstrap directories. BIOS, ReDump, and Arcade remain stubbed.

## Commands

- `make build`
- `make test`
- `make lint`
- `make docs-serve`

## Guardrails

- Never modify `/mnt/media-emulation/RetroLibrary` unless explicitly
  requested.
- Prefer `--dry-run` for validation in real libraries.
- Keep BIOS, ReDump, and Arcade paths behind explicit feature flags until they
  are implemented.
- Add or update tests for every behavior change.
- Keep `flake.lock` tracked and committed.
- Use Conventional Commits such as `feat(cli): add export command`.

## Docs

- Keep `README.md` as the GitHub entrypoint and `docs/index.md` as the docs
  landing page.
- Do not add duplicate overview pages such as `docs/README.md`.
- Document exported or reused behavior in Go doc comments first, then use
  Markdown for operator workflows and examples.
- If the repo grows machine-readable configuration knobs, define a schema
  instead of maintaining hand-written inventories in multiple places.
