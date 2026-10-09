#!/usr/bin/env bash
# Build a private engine MCP server and run the MCP protocol smoke suite.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
ENGINE_DIR="$(mktemp -d)"
trap 'rm -rf "$ENGINE_DIR"' EXIT
ENGINE_BIN="$ENGINE_DIR/engine-mcp"
go build -o "$ENGINE_BIN" ./cmd/engine-mcp/

LIKI_ENGINE_BIN="$ENGINE_BIN" "$ROOT/scripts/_engine_fixture.py" -- \
    "$(command -v python3)" "$ROOT/scripts/test_engine_mcp.py"