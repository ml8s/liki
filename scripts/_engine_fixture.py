#!/usr/bin/env python3
"""Shared local-engine fixture: start engine-mcp, wait for /health, run a command, clean up.

Usage (from bash):
    python3 scripts/_engine_fixture.py -- CMD [ARGS...]

The command receives LIKI_MCP_URL pointing at the private engine MCP endpoint.
The engine binary must already be built at /tmp/engine-mcp (see test-*.sh callers).
"""
from __future__ import annotations

import os
import socket
import subprocess
import sys
import time
import urllib.request

ENGINE_BIN = "/tmp/engine-mcp"


def main() -> int:
    if len(sys.argv) < 2 or sys.argv[1] != "--":
        sys.exit("usage: _engine_fixture.py -- CMD [ARGS...]")
    cmd = sys.argv[2:]
    if not cmd:
        sys.exit("_engine_fixture.py: empty command")

    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        port = sock.getsockname()[1]

    proc = subprocess.Popen(
        [ENGINE_BIN, "-addr", f"127.0.0.1:{port}"],
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

        env = dict(os.environ, LIKI_MCP_URL=f"http://127.0.0.1:{port}/mcp")
        cwd = os.environ.get("FIXTURE_CWD")
        return subprocess.call(cmd, cwd=cwd, env=env)
    finally:
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait()


if __name__ == "__main__":
    raise SystemExit(main())
