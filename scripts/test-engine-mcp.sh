#!/usr/bin/env bash
# Build engine-mcp, start it on a private port, and run protocol smoke tests.
set -eo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PYTHON="${PYTHON:-python3}"

cd "$ROOT/engine"
go build -o /tmp/engine-mcp ./cmd/engine-mcp/

"$PYTHON" - "$ROOT" <<'PY'
import socket
import subprocess
import sys
import time
import urllib.request

root = sys.argv[1]
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
        raise SystemExit("engine-mcp failed to become ready")
    raise SystemExit(subprocess.call(
        [sys.executable, f"{root}/scripts/test_engine_mcp.py", f"http://127.0.0.1:{port}"]
    ))
finally:
    proc.terminate()
    try:
        proc.wait(timeout=5)
    except subprocess.TimeoutExpired:
        proc.kill()
        proc.wait()
PY
