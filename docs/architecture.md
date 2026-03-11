# Architecture

## Components

- `internal/app`: command parsing and workflow orchestration.
- `internal/config`: YAML schema and validation.
- `internal/igir`: process execution (igir or npx fallback).
- `internal/fsutil`: filesystem helpers, DAT discovery.
- `internal/platform`: system selection and normalization.

## Error Handling

- Commands fail fast with descriptive errors.
- Non-zero exit code on all failures.
- `--dry-run` supported at the global level.

## Automation Readiness

- Stable command surface.
- Deterministic config file path support.
- Script-friendly stdout/stderr behavior.
