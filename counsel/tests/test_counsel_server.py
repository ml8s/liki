"""counsel 判断层 MCP 端点测试（分域 bazi/ziwei，3 工具）。"""
from __future__ import annotations

import asyncio
import json
import os

os.environ.setdefault("LIKI_COUNSEL_SERVICE_DOMAIN", "bazi")

import pytest  # noqa: E402

from app.counsel_server import create_counsel_server  # noqa: E402


def _engine_bazi_chart():
    import urllib.request

    url = os.environ["LIKI_MCP_URL"] + "/bazi"
    body = json.dumps({
        "jsonrpc": "2.0", "id": 1, "method": "tools/call",
        "params": {"name": "bazi_chart", "arguments": {
            "solar_time": "1990-05-20T11:49:00+08:00", "gender": "male",
            "longitude": 116.4,
        }, "_meta": {
            "io.modelcontextprotocol/protocolVersion": "2026-07-28",
            "io.modelcontextprotocol/clientCapabilities": {},
        }},
    }).encode()
    req = urllib.request.Request(
        url, data=body,
        headers={
            "Content-Type": "application/json",
            "MCP-Protocol-Version": "2026-07-28",
            "Mcp-Method": "tools/call", "Mcp-Name": "bazi_chart",
            "Accept": "application/json, text/event-stream",
        },
    )
    with urllib.request.urlopen(req, timeout=8) as r:
        return json.loads(json.loads(r.read())["result"]["content"][0]["text"])


@pytest.fixture(scope="module")
def bazi_chart():
    return _engine_bazi_chart()


def test_counsel_server_tools():
    srv = create_counsel_server("bazi")
    names = asyncio.run(srv.list_tools())
    assert [t.name for t in names] == ["compute_factors", "natal_query", "period_query"]


def test_compute_factors_over_mcp(bazi_chart):
    srv = create_counsel_server("bazi")
    r = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    out = json.loads(r.content[0].text)
    assert set(out) == {"factors", "factors_digest", "context"}
    assert out["context"]["性别"] == "male"
    assert len(out["factors_digest"]) == 64  # sha256 hex


def test_natal_query_over_mcp(bazi_chart):
    srv = create_counsel_server("bazi")
    r = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    factors = json.loads(r.content[0].text)["factors"]
    r = asyncio.run(srv.call_tool(
        "natal_query",
        {"factors": factors, "topics": ["chart_structure"], "context": {"性别": "male"}},
    ))
    out = json.loads(r.content[0].text)
    assert out["assertions"] and all(a.get("side") == "bazi" for a in out["assertions"])


def test_period_query_over_mcp(bazi_chart):
    srv = create_counsel_server("bazi")
    r = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    factors = json.loads(r.content[0].text)["factors"]
    r = asyncio.run(srv.call_tool(
        "period_query",
        {
            "factors": factors,
            "time_scope": {"type": "year_range", "start_year": 2020, "end_year": 2040},
            "topics": ["marriage"],
            "chart": bazi_chart,
        },
    ))
    out = json.loads(r.content[0].text)
    periods = out["periods"]
    assert len(periods) == 21  # 2020-2040 每年一段
    assert sum(len(p["assertions"]) for p in periods) == 20
    assert all(a.get("side") == "bazi" for p in periods for a in p["assertions"])


def test_domain_rejects_wrong_chart():
    srv = create_counsel_server("bazi")
    with pytest.raises(Exception, match="Error executing tool"):
        asyncio.run(srv.call_tool("compute_factors", {"chart": {"ziwei": {}}}))