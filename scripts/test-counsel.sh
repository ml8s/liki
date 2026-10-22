#!/usr/bin/env bash
# Build and start a private engine MCP server, then run the counsel suite.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PYTHON="${PYTHON:-python3}"
COUNSEL_PYTHON="${COUNSEL_PYTHON:-}"
COUNSEL_TEST_TARGETS="${COUNSEL_TEST_TARGETS:-tests/}"

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
        # 与 make python-env 的开发依赖保持一致：gate 的 check/deployment
        # 目标会通过 venv 的 python3 运行，需要 pyyaml/jsonschema。
        "$ROOT/counsel/.venv/bin/python" -m pip install pytest pyyaml jsonschema
    fi
fi

COUNSEL_PYTHON="${COUNSEL_PYTHON:-$ROOT/counsel/.venv/bin/python}"
read -r -a counsel_targets <<< "$COUNSEL_TEST_TARGETS"
LIKI_ENGINE_BIN="$ENGINE_BIN" FIXTURE_CWD="$ROOT/counsel" "$ROOT/scripts/_engine_fixture.py" -- \
    "$COUNSEL_PYTHON" -m pytest "${counsel_targets[@]}" -q
