#!/usr/bin/env bash
# Build and start a private engine MCP server, then run the counsel suite.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PYTHON="${PYTHON:-python3}"
COUNSEL_PYTHON="${COUNSEL_PYTHON:-}"
COUNSEL_TEST_TARGETS="${COUNSEL_TEST_TARGETS:-tests/}"
export UV_CACHE_DIR="${UV_CACHE_DIR:-/tmp/uv-cache}"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
ENGINE_DIR="$(mktemp -d)"
trap 'rm -rf "$ENGINE_DIR"' EXIT
ENGINE_BIN="$ENGINE_DIR/engine-mcp"
go build -o "$ENGINE_BIN" ./cmd/engine-mcp/

if [ -z "$COUNSEL_PYTHON" ] && [ ! -x "$ROOT/counsel/.venv/bin/python" ]; then
    if command -v uv >/dev/null 2>&1; then
        make --no-print-directory -C "$ROOT" python-env
    else
        "$PYTHON" -m venv "$ROOT/counsel/.venv"
        "$ROOT/counsel/.venv/bin/python" -m pip install \
            --retries 10 --timeout 60 \
            --require-hashes --requirement "$ROOT/counsel/requirements.lock"
        "$ROOT/counsel/.venv/bin/python" -m pip install \
            --retries 10 --timeout 60 \
            --requirement "$ROOT/counsel/requirements-dev.txt"
    fi
fi

COUNSEL_PYTHON="${COUNSEL_PYTHON:-$ROOT/counsel/.venv/bin/python}"
if [ "$COUNSEL_PYTHON" = "$ROOT/counsel/.venv/bin/python" ] && \
    ! "$COUNSEL_PYTHON" -c 'import jsonschema, pytest, yaml' >/dev/null 2>&1; then
    # Older venvs may predate requirements-dev.txt. Refresh only the missing
    # dev layer instead of forcing a network sync on every test run.
    if command -v uv >/dev/null 2>&1; then
        uv pip install --python "$COUNSEL_PYTHON" \
            --requirement "$ROOT/counsel/requirements-dev.txt"
    else
        "$COUNSEL_PYTHON" -m pip install \
            --retries 10 --timeout 60 \
            --requirement "$ROOT/counsel/requirements-dev.txt"
    fi
fi

read -r -a counsel_targets <<< "$COUNSEL_TEST_TARGETS"
if [ "${COUNSEL_COVERAGE:-0}" = "1" ]; then
    export COVERAGE_FILE="$ROOT/.coverage"
    LIKI_ENGINE_BIN="$ENGINE_BIN" FIXTURE_CWD="$ROOT/counsel" \
        "$ROOT/scripts/_engine_fixture.py" -- \
        "$COUNSEL_PYTHON" -m coverage run --append \
        --source="$ROOT/counsel/app" \
        -m pytest "${counsel_targets[@]}" -q
else
    LIKI_ENGINE_BIN="$ENGINE_BIN" FIXTURE_CWD="$ROOT/counsel" \
        "$ROOT/scripts/_engine_fixture.py" -- \
        "$COUNSEL_PYTHON" -m pytest "${counsel_targets[@]}" -q
fi
