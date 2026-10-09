#!/usr/bin/env bash
# Run the 160-question assertion coverage check against a private engine-mcp.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PYTHON="${PYTHON:-python3}"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
ENGINE_DIR="$(mktemp -d)"
trap 'rm -rf "$ENGINE_DIR"' EXIT
ENGINE_BIN="$ENGINE_DIR/engine-mcp"
go build -o "$ENGINE_BIN" ./cmd/engine-mcp/

LIKI_ENGINE_BIN="$ENGINE_BIN" FIXTURE_CWD="$ROOT" "$ROOT/scripts/_engine_fixture.py" -- \
    "$PYTHON" "$ROOT/scripts/eval_hybrid.py"