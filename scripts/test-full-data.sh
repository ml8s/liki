#!/usr/bin/env bash
# Run the 160-question assertion coverage check against a private engine-mcp.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PYTHON="${PYTHON:-python3}"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
go build -o /tmp/engine-mcp ./cmd/engine-mcp/

FIXTURE_CWD="$ROOT" "$ROOT/scripts/_engine_fixture.py" -- \
    "$PYTHON" "$ROOT/scripts/eval_hybrid.py"