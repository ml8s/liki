"""analysis 断语引擎新工具接口契约测试（digest 绑定与来源校验）。"""

import json

import pytest


@pytest.fixture(scope="module")
def bazi_chart():
    """engine bazi_chart 排盘结果（公共契约输入）。"""
    import os
    import urllib.request

    url = os.environ.get("LIKI_ENGINE_MCP_URL", "http://127.0.0.1:18081/mcp") + "/bazi"
    body = json.dumps({"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {
        "name": "bazi_chart",
        "arguments": {"standard_time": "1990-05-20T11:49:00+08:00", "longitude": 116.4, "gender": "male"},
        "_meta": {"io.modelcontextprotocol/protocolVersion": "2026-07-28",
                  "io.modelcontextprotocol/clientCapabilities": {}},
    }}).encode()
    req = urllib.request.Request(url, data=body, headers={
        "Content-Type": "application/json", "MCP-Protocol-Version": "2026-07-28",
        "Mcp-Method": "tools/call", "Mcp-Name": "bazi_chart", "Accept": "application/json, text/event-stream",
    })
    with urllib.request.urlopen(req, timeout=10) as r:
        d = json.loads(r.read())
    payload = d["result"].get("structuredContent")
    if payload is None:
        payload = json.loads(d["result"]["content"][0]["text"])
    return payload


def test_compute_factors_returns_factors_and_digest(bazi_chart):
    from app.natal.tools.counsel import compute_factors
    out = compute_factors(bazi_chart)
    assert isinstance(out, dict)
    assert "factors" in out and "factors_digest" in out


def test_natal_query_returns_bazi_assertions(bazi_chart):
    """natal_query 校验因子摘要后返回八字断语。"""
    from app.natal.tools.counsel import compute_factors, natal_query

    out = compute_factors(bazi_chart)
    result = natal_query(
        out["factors"],
        ["chart_structure"],
        out["factors_digest"],
    )
    assert result["assertions"]
    assert all(a.get("side") == "bazi" for a in result["assertions"])
    assert all(a.get("assertion_id") for a in result["assertions"])


def test_period_query_returns_bazi_assertions(bazi_chart):
    """period_query 校验因子摘要后返回八字应期断语。"""
    from app.natal.tools.counsel import compute_factors, period_query

    out = compute_factors(bazi_chart)
    got = period_query(
        out["factors"],
        out["factors_digest"],
        {"type": "year_range", "start_year": 2020, "end_year": 2040, "timezone": "Asia/Shanghai"},
        ["marriage"],
        bazi_chart,
    )
    assert got["periods"]
    assert any(p.get("assertions") for p in got["periods"])
    assert all(
        a.get("side") == "bazi"
        for p in got["periods"]
        for a in p["assertions"]
    )


def test_natal_query_rejects_modified_factors(bazi_chart):
    from app.natal.tools.counsel import compute_factors, natal_query

    out = compute_factors(bazi_chart)
    factors = dict(out["factors"])
    factors[next(iter(factors))] = "tampered"
    with pytest.raises(ValueError, match="factors_digest mismatch"):
        natal_query(
            factors,
            ["chart_structure"],
            out["factors_digest"],
        )


def test_period_query_rejects_chart_from_another_snapshot(bazi_chart):
    """chart 与 factors 经 _provenance.chart_digest 绑定，换盘即拒。"""
    from app.natal.tools.counsel import compute_factors, period_query

    out = compute_factors(bazi_chart)
    other = dict(bazi_chart)
    other["shi"] = {"gan": "甲", "zhi": "子", "na_yin": "海中金"}
    with pytest.raises(ValueError, match="chart 与 compute_factors 因子来源不一致"):
        period_query(
            out["factors"],
            out["factors_digest"],
            {"type": "year_range", "start_year": 2020, "end_year": 2040},
            ["marriage"],
            other,
        )
