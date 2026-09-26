#!/usr/bin/env bash
# Run the pinned Python linter without requiring a global Ruff installation.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RUFF_VERSION="${RUFF_VERSION:-0.16.0}"

if command -v ruff >/dev/null 2>&1; then
    exec ruff check "$ROOT/counsel/app" "$ROOT/counsel/tests" "$ROOT/scripts"
fi

if command -v uv >/dev/null 2>&1; then
    export UV_CACHE_DIR="${UV_CACHE_DIR:-/tmp/uv-cache}"
    export UV_TOOL_DIR="${UV_TOOL_DIR:-/tmp/uv-tools}"
    exec uv tool run "ruff==${RUFF_VERSION}" check \
        "$ROOT/counsel/app" "$ROOT/counsel/tests" "$ROOT/scripts"
fi

echo "Python lint requires ruff or uv; install https://docs.astral.sh/ruff/ first." >&2
exit 1
