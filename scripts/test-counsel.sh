#!/usr/bin/env bash
# Build and start a private engine MCP server, then run the counsel suite.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PYTHON="${PYTHON:-python3}"
COUNSEL_PYTHON="${COUNSEL_PYTHON:-}"

"$ROOT/scripts/check-go-version.sh"

cd "$ROOT/engine"
go build -o /tmp/engine-mcp ./cmd/engine-mcp/

if [ -z "$COUNSEL_PYTHON" ] && [ ! -x "$ROOT/counsel/.venv/bin/python" ]; then
    "$PYTHON" -m venv "$ROOT/counsel/.venv"
    "$ROOT/counsel/.venv/bin/python" -m pip install \
        --retries 10 --timeout 60 \
        --require-hashes --requirement "$ROOT/counsel/requirements.lock"
    "$ROOT/counsel/.venv/bin/python" -m pip install pytest
fi

"$PYTHON" - "$ROOT/counsel" <<'PY'
import os
import socket
import subprocess
import sys
import time
import urllib.request

counsel_root = sys.argv[1]
with socket.socket() as sock:
    sock.bind(("127.0.0.1", 0))
    port = sock.getsockname()[1]
proc = subprocess.Popen(
    ["/tmp/engine-mcp", "-addr", f"127.0.0.1:{port}"],
    stdout=subprocess.DEVNULL,
    stderr=subprocess.PIPE,
)
try:
    for _ in range(40):
        try:
            with urllib.request.urlopen(f"http://127.0.0.1:{port}/health", timeout=1):
                break
        except Exception:
            time.sleep(0.25)
    else:
        raise SystemExit("local engine failed to become ready")
    env = dict(os.environ, LIKI_MCP_URL=f"http://127.0.0.1:{port}/mcp")
    raise SystemExit(subprocess.call(
        [os.environ.get("COUNSEL_PYTHON", f"{counsel_root}/.venv/bin/python"),
         "-m", "pytest", "tests/", "-q"],
        cwd=counsel_root,
        env=env,
    ))
finally:
    proc.terminate()
    try:
        proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        proc.kill()
        proc.wait()
PY
