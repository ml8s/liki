"""analysis MCP HTTP 传输层正确性测试。

覆盖 in-process 测试不覆盖的维度：真实 HTTP（stateless Streamable HTTP）调用下
的透传一致性、中文编码、长 token 传输、参数边界、并发确定性。

前提：本地引擎可达（LIKI_RPC_URL）。
"""
from __future__ import annotations

import json
import os
import subprocess
import sys
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[2]
NATAL_TOOLS_PATH = Path(__file__).resolve().parents[1] / "liki_analysis" / "natal" / "tools"
VENV_PYTHON = REPO_ROOT / "analysis" / ".venv" / "bin" / "python"
PORT = 8091
BASE = f"http://127.0.0.1:{PORT}/mcp"
LIKI_MCP_URL = os.environ.get("LIKI_MCP_URL", "http://127.0.0.1:18081/mcp")


def _engine_available() -> bool:
    try:
        with urllib.request.urlopen(LIKI_MCP_URL.replace("/mcp", "") + "/health", timeout=3) as r:
            return r.status == 200
    except Exception:
        return False


requires_engine = pytest.mark.skipif(
    not _engine_available(), reason="本地引擎不可达"
)


@pytest.fixture(scope="module")
def http_server():
    proc = subprocess.Popen(
        [str(VENV_PYTHON), "-m", "uvicorn", "liki_analysis.server:app",
         "--host", "127.0.0.1", "--port", str(PORT)],
        cwd=str(REPO_ROOT / "analysis"),
        env={**os.environ, "LIKI_MCP_URL": LIKI_MCP_URL},
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
    for _ in range(20):
        try:
            with urllib.request.urlopen(BASE, timeout=2):
                break
        except Exception:
            time.sleep(0.5)
    yield
    proc.terminate()
    proc.wait(timeout=5)


def http_call(method: str, params: dict, name: str | None = None) -> dict:
    meta = {
        "io.modelcontextprotocol/protocolVersion": "2026-07-28",
        "io.modelcontextprotocol/clientInfo": {"name": "test", "version": "1.0"},
        "io.modelcontextprotocol/clientCapabilities": {},
    }
    body = json.dumps({"jsonrpc": "2.0", "id": 1, "method": method,
                       "params": {**params, "_meta": meta}}).encode("utf-8")
    headers = {
        "Content-Type": "application/json",
        "MCP-Protocol-Version": "2026-07-28",
        "Mcp-Method": method,
        "Accept": "application/json, text/event-stream",
    }
    if name:
        headers["Mcp-Name"] = name
    req = urllib.request.Request(BASE, data=body, headers=headers)
    with urllib.request.urlopen(req, timeout=30) as r:
        return json.loads(r.read().decode())


def direct(fn: str, args: dict, cli: Path) -> dict:
    p = subprocess.run(
        [str(VENV_PYTHON), str(cli)],
        input=json.dumps({"fn": fn, "args": args}, ensure_ascii=False).encode("utf-8"),
        capture_output=True,
        timeout=60,
        cwd=str(cli.parent),
        env={**os.environ, "LIKI_MCP_URL": LIKI_MCP_URL},
    )
    return json.loads(p.stdout.decode("utf-8"))["data"]


NATAL_CLI = NATAL_TOOLS_PATH / "agent_cli.py"
BIRTH = {"gender": "male", "source": {"type": "timestamp", "timestamp": "1990-05-20T12:00:00+08:00",
                                      "precision": "minute", "location": {"city": "北京"}}}


@requires_engine
def test_http_create_parity(http_server) -> None:
    d = http_call("tools/call", {"name": "create_birth_chart", "arguments": BIRTH}, name="create_birth_chart")
    http_data = json.loads(d["result"]["content"][0]["text"])
    direct_data = direct("create_birth_chart", BIRTH, NATAL_CLI)
    assert http_data == direct_data, "HTTP 结果与 agent_cli 不一致"


@requires_engine
def test_http_utf8_and_long_token(http_server) -> None:
    d = http_call("tools/call", {"name": "create_birth_chart", "arguments": BIRTH}, name="create_birth_chart")
    http_data = json.loads(d["result"]["content"][0]["text"])
    # 中文城市无乱码
    assert http_data["chart"]["birth"]["location"]["name"] == "北京市"
    # 长 chart_ref token（自包含压缩盘）HTTP 往返完好 → analyze 可用
    ref = http_data["chart_ref"]
    d2 = http_call("tools/call", {"name": "analyze_natal",
                                  "arguments": {"chart_ref": ref, "topics": ["career"]}}, name="analyze_natal")
    analyze = json.loads(d2["result"]["content"][0]["text"])
    direct_analyze = direct("analyze_natal", {"chart_ref": ref, "topics": ["career"]}, NATAL_CLI)
    assert analyze == direct_analyze
    assert analyze["assertions"], "应命中断语"
    assert analyze["assertions"][0]["conclusion"], "断语结论非空"


@requires_engine
def test_http_argument_boundaries(http_server) -> None:
    cases = [
        {"name": "create_birth_chart", "arguments": {**BIRTH, "gender": "other"}},
        {"name": "create_birth_chart", "arguments": {"gender": "male"}},  # 缺 source
        {"name": "create_birth_chart", "arguments": {**BIRTH, "gender": 123}},  # 类型错误
        {"name": "huangli_days", "arguments": {"question": "搬家", "start_date": "2026-10-01", "days": -3}},
    ]
    for c in cases:
        d = http_call("tools/call", {"name": c["name"], "arguments": c["arguments"]}, name=c["name"])
        is_error = d.get("result", {}).get("isError") or "error" in d
        assert is_error, f"非法参数应被拒绝: {c['name']} {c['arguments']}"


@requires_engine
def test_http_concurrent_deterministic(http_server) -> None:
    def one(_):
        d = http_call("tools/call", {"name": "create_birth_chart", "arguments": BIRTH}, name="create_birth_chart")
        return json.loads(d["result"]["content"][0]["text"])["chart"]["digest"]

    with ThreadPoolExecutor(max_workers=8) as ex:
        digests = list(ex.map(one, range(8)))
    assert len(set(digests)) == 1, "并发调用结果不一致（digest 应稳定）"