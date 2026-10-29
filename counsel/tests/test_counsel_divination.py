"""counsel 六爻/奇门（divination）MCP wiring 回归测试。

定位：**不是 correctness oracle**。它断言"经 MCP 工具调用 == 直接调用模块函数"，
即 MCP 装配/透传没有引入差异（正交化只改入口 ask→query，不改逻辑）。
领域正确性由 engine 的 `*_oracle_test.go` 与 `tests/fixtures/domain_oracle/` 独立承担。
"""

import asyncio

from app.counsel_mcp import create_counsel_mcp  # noqa: E402

# 动态字段（digest/快照元数据）——1v1 对照时归一化
DYNAMIC = {
    "snapshot_digest", "_meta", "meta", "created_at",
    "captured_at",
}


def _core(d):
    if isinstance(d, dict):
        return {k: _core(v) for k, v in d.items() if k not in DYNAMIC}
    if isinstance(d, list):
        return [_core(v) for v in d]
    return d


LIUYAO_ARGS = {"question": "测试这次面试能不能通过", "casting": {"mode": "yaos", "yaos": [7, 8, 7, 8, 7, 8]},
               "longitude": 116.4, "matter": "general",
               "standard_time": "2026-10-01T10:00:00+08:00"}


def test_liuyao_snapshot_matches_old():
    from app.divination.tools.liuyao_snapshot import create

    srv = create_counsel_mcp("liuyao")
    got = asyncio.run(srv.call_tool("liuyao_snapshot", dict(LIUYAO_ARGS))).structured_content
    want = create(**LIUYAO_ARGS)
    assert _core(got) == _core(want), "liuyao_snapshot 与老实现不一致"
    assert got["board"] == want["board"]
    assert got["evidence"] == want["evidence"]


def test_liuyao_query_matches_old_ask():
    from app.divination.tools.liuyao_ask import ask

    srv = create_counsel_mcp("liuyao")
    snap = asyncio.run(srv.call_tool("liuyao_snapshot", dict(LIUYAO_ARGS))).structured_content
    got = asyncio.run(srv.call_tool(
        "liuyao_query", {"snapshot": snap, "message": "什么时候有结果"},
    )).structured_content
    want = ask(snap, message="什么时候有结果")
    assert _core(got) == _core(want), "liuyao_query 与老 ask 不一致"


QIMEN_ARGS = {"question": "测试谈判时机", "city": "北京",
              "standard_time": "2026-10-01T10:00:00+08:00",
              "method": {"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
              "matter": "career"}


def test_qimen_snapshot_matches_old():
    from app.divination.tools.qimen_snapshot import create

    srv = create_counsel_mcp("qimen")
    got = asyncio.run(srv.call_tool("qimen_snapshot", dict(QIMEN_ARGS))).structured_content
    want = create(**QIMEN_ARGS)
    assert _core(got) == _core(want), "qimen_snapshot 与老实现不一致"


def test_qimen_query_matches_old_ask():
    from app.divination.tools.qimen_ask import ask

    srv = create_counsel_mcp("qimen")
    snap = asyncio.run(srv.call_tool("qimen_snapshot", dict(QIMEN_ARGS))).structured_content
    got = asyncio.run(srv.call_tool(
        "qimen_query", {"snapshot": snap, "message": "哪个方向有利"},
    )).structured_content
    want = ask(snap, message="哪个方向有利")
    assert _core(got) == _core(want), "qimen_query 与老 ask 不一致"
