"""judgment 判断层新工具接口测试（正交化后）。

逻辑输出必须与正交化基线（tests/golden/orthogonal/baseline.json）一致。
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "natal" / "tools"))

GOLDEN = Path(__file__).resolve().parents[2] / "tests" / "golden" / "orthogonal" / "baseline.json"


@pytest.fixture(scope="module")
def baseline() -> dict:
    return json.loads(GOLDEN.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def bazi_chart():
    """engine bazi_chart 排盘结果（公共契约输入）。"""
    import os
    import urllib.request

    url = os.environ.get("LIKI_MCP_URL", "http://127.0.0.1:18081/mcp") + "/bazi"
    body = json.dumps({"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {
        "name": "bazi_chart",
        "arguments": {"solar_time": "1990-05-20T11:49:00+08:00", "gender": "male", "longitude": 116.4},
        "_meta": {"io.modelcontextprotocol/protocolVersion": "2026-07-28",
                  "io.modelcontextprotocol/clientCapabilities": {}},
    }}).encode()
    req = urllib.request.Request(url, data=body, headers={
        "Content-Type": "application/json", "MCP-Protocol-Version": "2026-07-28",
        "Mcp-Method": "tools/call", "Mcp-Name": "bazi_chart", "Accept": "application/json, text/event-stream",
    })
    with urllib.request.urlopen(req, timeout=10) as r:
        d = json.loads(r.read())
    return json.loads(d["result"]["content"][0]["text"])


def test_compute_factors_returns_factors_and_digest(bazi_chart):
    from judgment import compute_factors
    out = compute_factors(bazi_chart)
    assert isinstance(out, dict)
    assert "factors" in out and "factors_digest" in out


def test_compute_factors_matches_baseline(bazi_chart, baseline):
    from judgment import compute_factors
    factors = compute_factors(bazi_chart)["factors"]
    # 八字因子（bazi side）应与基线一致（逻辑不变，仅接口/输入来源变化）
    want = baseline["factors_bazi"]
    for key in want:
        assert factors.get(key) == want[key], f"因子 {key} 与基线不一致"

