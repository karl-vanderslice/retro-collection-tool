#!/usr/bin/env bash
set -euo pipefail

if ! git ls-files --error-unmatch flake.lock >/dev/null 2>&1; then
  echo "flake.lock must be tracked in this repository."
  echo "Run: git add -f flake.lock"
  exit 1
fi

if git check-ignore -q flake.lock; then
  echo "flake.lock is currently ignored. Ensure repo .gitignore has '!flake.lock'."
  exit 1
fi

echo "flake.lock tracking check passed."
