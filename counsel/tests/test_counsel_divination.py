"""counsel 六爻/奇门（divination）1v1 对照测试。

counsel 六爻/奇门复用 analysis 模块（snapshot 创建 + query 追问），
输出必须与老实现（模块函数）一致——正交化补集只改入口（ask→query），不改逻辑。
"""
from __future__ import annotations

import asyncio
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "app" / "divination" / "tools"))


from app.counsel_mcp import create_counsel_mcp  # noqa: E402

# 动态字段（digest/时间戳）——1v1 对照时归一化
DYNAMIC = {
    "snapshot_digest", "_meta", "meta", "timestamp", "created_at",
    "captured_at", "solar_time",
}


def _core(d):
    if isinstance(d, dict):
        return {k: _core(v) for k, v in d.items() if k not in DYNAMIC}
    if isinstance(d, list):
        return [_core(v) for v in d]
    return d


LIUYAO_ARGS = {"question": "测试这次面试能不能通过", "mode": "yaos",
               "yaos": [7, 8, 7, 8, 7, 8], "matter": "general"}


def test_liuyao_snapshot_matches_old():
    from liuyao_snapshot import create

    srv = create_counsel_mcp("liuyao")
    got = json.loads(asyncio.run(srv.call_tool("liuyao_snapshot", dict(LIUYAO_ARGS))).content[0].text)
    want = create(**LIUYAO_ARGS)
    assert _core(got) == _core(want), "liuyao_snapshot 与老实现不一致"
    assert got["board"] == want["board"]
    assert got["evidence"] == want["evidence"]


def test_liuyao_query_matches_old_ask():
    from liuyao_ask import ask

    srv = create_counsel_mcp("liuyao")
    snap = json.loads(asyncio.run(srv.call_tool("liuyao_snapshot", dict(LIUYAO_ARGS))).content[0].text)
    got = json.loads(asyncio.run(srv.call_tool(
        "liuyao_query", {"snapshot": snap, "message": "什么时候有结果"},
    )).content[0].text)
    want = ask(snap, message="什么时候有结果")
    assert _core(got) == _core(want), "liuyao_query 与老 ask 不一致"


QIMEN_ARGS = {"question": "测试谈判时机", "city": "北京",
              "time": "2026-10-01T10:00:00+08:00", "matter": "career"}


def test_qimen_snapshot_matches_old():
    from qimen_snapshot import create

    srv = create_counsel_mcp("qimen")
    got = json.loads(asyncio.run(srv.call_tool("qimen_snapshot", dict(QIMEN_ARGS))).content[0].text)
    want = create(**QIMEN_ARGS)
    assert _core(got) == _core(want), "qimen_snapshot 与老实现不一致"


def test_qimen_query_matches_old_ask():
    from qimen_ask import ask

    srv = create_counsel_mcp("qimen")
    snap = json.loads(asyncio.run(srv.call_tool("qimen_snapshot", dict(QIMEN_ARGS))).content[0].text)
    got = json.loads(asyncio.run(srv.call_tool(
        "qimen_query", {"snapshot": snap, "message": "哪个方向有利"},
    )).content[0].text)
    want = ask(snap, message="哪个方向有利")
    assert _core(got) == _core(want), "qimen_query 与老 ask 不一致"
