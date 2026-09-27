#!/usr/bin/env bash
# Build a private engine MCP server and run the MCP protocol smoke suite.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
go build -o /tmp/engine-mcp ./cmd/engine-mcp/

"$ROOT/scripts/_engine_fixture.py" -- \
    "$(command -v python3)" "$ROOT/scripts/test_engine_mcp.py"