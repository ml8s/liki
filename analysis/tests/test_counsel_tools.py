"""analysis 判断层新工具接口测试（正交化后）。

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
    from counsel import compute_factors
    out = compute_factors(bazi_chart)
    assert isinstance(out, dict)
    assert "factors" in out and "factors_digest" in out


def test_compute_factors_matches_baseline(bazi_chart, baseline):
    from counsel import compute_factors
    factors = compute_factors(bazi_chart)["factors"]
    # 八字因子（bazi side）应与基线一致（逻辑不变，仅接口/输入来源变化）
    want = baseline["factors_bazi"]
    for key in want:
        assert factors.get(key) == want[key], f"因子 {key} 与基线不一致"

def test_natal_query_matches_analyze_natal(bazi_chart):
    """natal_query(factors, topics) 断语 == analyze_natal(chart_ref, topics) 八字侧（逻辑一致）。"""
    from app.server import create_server
    import asyncio

    from counsel import compute_factors, natal_query

    out = compute_factors(bazi_chart)
    result = natal_query(out["factors"], ["chart_structure"], context=out["context"])
    got_ids = sorted(a.get("assertion_id") or a["id"] for a in result["assertions"])

    srv = create_server()
    r = asyncio.run(srv.call_tool("create_birth_chart", {
        "gender": "male",
        "source": {"type": "timestamp", "timestamp": "1990-05-20T12:00:00+08:00",
                   "precision": "minute", "location": {"city": "北京"}},
    }))
    chart = json.loads(r.content[0].text)
    r = asyncio.run(srv.call_tool("analyze_natal", {
        "chart_ref": chart["chart_ref"], "topics": ["chart_structure"],
    }))
    data = json.loads(r.content[0].text)
    want_ids = sorted(a["assertion_id"] for a in data["assertions"] if a["side"] == "bazi")

    assert got_ids == want_ids, f"natal_query 八字断语不一致: got={got_ids} want={want_ids}"

def test_period_query_matches_analyze_periods(bazi_chart):
    """period_query(factors, time_scope, topics, chart) 应期断语 == analyze_periods（八字侧）。"""
    import asyncio

    from app.server import create_server
    from counsel import compute_factors, period_query

    out = compute_factors(bazi_chart)
    got = period_query(
        out["factors"],
        {"type": "year_range", "start_year": 2020, "end_year": 2040},
        ["marriage"],
        bazi_chart,
    )
    got_ids = sorted(
        a.get("assertion_id") or a["id"]
        for p in got["periods"]
        for a in p["assertions"]
    )

    srv = create_server()
    r = asyncio.run(srv.call_tool("create_birth_chart", {
        "gender": "male",
        "source": {"type": "timestamp", "timestamp": "1990-05-20T11:49:00+08:00",
                   "precision": "minute", "location": {"city": "北京"}},
    }))
    chart = json.loads(r.content[0].text)
    r = asyncio.run(srv.call_tool("analyze_periods", {
        "chart_ref": chart["chart_ref"], "topics": ["marriage"],
        "time_scope": {"type": "year_range", "start_year": 2020, "end_year": 2040},
    }))
    want = json.loads(r.content[0].text)
    want_ids = sorted(
        a["assertion_id"] for p in want["periods"] for a in p["assertions"]
    )

    assert got_ids == want_ids, (
        f"period_query 应期断语不一致: got={got_ids} want={want_ids}"
    )
