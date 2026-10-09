#!/usr/bin/env bash
# Build and start a private engine MCP server, then run the counsel suite.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PYTHON="${PYTHON:-python3}"
COUNSEL_PYTHON="${COUNSEL_PYTHON:-}"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
ENGINE_DIR="$(mktemp -d)"
trap 'rm -rf "$ENGINE_DIR"' EXIT
ENGINE_BIN="$ENGINE_DIR/engine-mcp"
go build -o "$ENGINE_BIN" ./cmd/engine-mcp/

if [ -z "$COUNSEL_PYTHON" ] && [ ! -x "$ROOT/counsel/.venv/bin/python" ]; then
    "$PYTHON" -m venv "$ROOT/counsel/.venv"
    "$ROOT/counsel/.venv/bin/python" -m pip install \
        --retries 10 --timeout 60 \
        --require-hashes --requirement "$ROOT/counsel/requirements.lock"
    "$ROOT/counsel/.venv/bin/python" -m pip install pytest
fi

COUNSEL_PYTHON="${COUNSEL_PYTHON:-$ROOT/counsel/.venv/bin/python}"
LIKI_ENGINE_BIN="$ENGINE_BIN" FIXTURE_CWD="$ROOT/counsel" "$ROOT/scripts/_engine_fixture.py" -- \
    "$COUNSEL_PYTHON" -m pytest tests/ -q