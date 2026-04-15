# Architecture

This tool follows a small, modular architecture so command behavior stays predictable as workflows grow.

## Core modules

- `internal/app`: command parsing and workflow orchestration.
- `internal/config`: YAML schema and validation.
- `internal/igir`: process execution (igir or npx fallback).
- `internal/fsutil`: filesystem helpers, DAT discovery.
- `internal/platform`: system selection and normalization.

## Execution model

- CLI args are parsed and validated early.
- Config is loaded and merged deterministically.
- A workflow runner resolves selected systems and feature flags.
- Side-effecting operations execute only after validation passes.
- Errors immediately stop execution and return non-zero.

## Error handling

- Commands fail fast with descriptive errors.
- Non-zero exit code on all failures.
- `--dry-run` supported at the global level.

## Safety boundaries

- `--dry-run` avoids writes and prints planned actions.
- Feature flags guard unfinished workflows.
- Command parsers reject unexpected trailing arguments.

## Automation readiness

- Stable command surface.
- Deterministic config file path support.
- Script-friendly stdout/stderr behavior.

## Testing and CI

- Unit tests cover command behavior and config rules.
- CI on push to `master` runs format, lint, test, and build.
- Docs are built and published to GitHub Pages from CI.
