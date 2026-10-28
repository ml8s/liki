#!/usr/bin/env bash
# Combine the root Python contract suite with counsel integration coverage.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FLOOR="${COVERAGE_FLOOR:-85}"
COVERAGE_PYTHON="${COUNSEL_PYTHON:-$ROOT/counsel/.venv/bin/python}"

make --no-print-directory -C "$ROOT" python-env

cd "$ROOT"
"$COVERAGE_PYTHON" -m coverage erase
"$COVERAGE_PYTHON" -m coverage run \
    --source="$ROOT/counsel/app" \
    -m pytest tests -q

COUNSEL_COVERAGE=1 COUNSEL_PYTHON="$COVERAGE_PYTHON" \
    "$ROOT/scripts/test-counsel.sh"

"$COVERAGE_PYTHON" -m coverage report --fail-under="$FLOOR"
"$COVERAGE_PYTHON" -m coverage xml
"$COVERAGE_PYTHON" -m coverage html
