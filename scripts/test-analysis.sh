#!/usr/bin/env bash
# 起本地引擎 + 跑 analysis 测试（用于 make test-analysis）
set -eo pipefail

PORT="${ENGINE_MCP_PORT:-18081}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT/engine"
go build -o /tmp/liki-mcp ./cmd/liki-mcp/

# 清理旧引擎 + 起新引擎
fuser -k "$PORT/tcp" 2>/dev/null || true
python3 - "$PORT" <<'PY'
import subprocess, sys, time, urllib.request
port = int(sys.argv[1])
proc = subprocess.Popen(['/tmp/liki-mcp', '-addr', f'127.0.0.1:{port}'],
                        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
for _ in range(20):
    try:
        with urllib.request.urlopen(f'http://127.0.0.1:{port}/health', timeout=2):
            break
    except Exception:
        time.sleep(0.5)
PY

cd "$ROOT/analysis"
LIKI_MCP_URL="http://127.0.0.1:$PORT/mcp" .venv/bin/python -m pytest tests/ -q