"""六爻 / 奇门 one-shot ask MCP wiring tests."""

import asyncio

from app.counsel_mcp import create_counsel_root_mcp  # noqa: E402

LIUYAO_ARGS = {
    "question": "测试这次面试能不能通过",
    "casting": {"mode": "yaos", "yaos": [7, 8, 7, 8, 7, 8]},
    "longitude": 116.4,
    "focus": {"kind": "matter", "matter": "general"},
    "standard_time": "2026-10-01T10:00:00+08:00",
}

LIUYAO_COINS_ARGS = {
    **LIUYAO_ARGS,
    "casting": {
        "mode": "coins",
        "rounds": [
            ["正", "正", "反"],
            ["正", "反", "反"],
            ["反", "反", "反"],
            ["正", "正", "正"],
            ["正", "反", "反"],
            ["正", "正", "反"],
        ],
    },
}

QIMEN_ARGS = {
    "question": "测试谈判时机",
    "longitude": 116.4,
    "standard_time": "2026-10-01T10:00:00+08:00",
    "method": {"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
    "focus": {"kind": "matter", "matter": "career"},
}


def test_liuyao_ask_returns_self_contained_reading():
    srv = create_counsel_root_mcp()
    got = asyncio.run(srv.call_tool("liuyao_ask", dict(LIUYAO_ARGS))).structured_content
    assert got["technique"] == "liuyao"
    assert "reading_digest" not in got
    assert got["scope"]["focus_value"] == "general"
    assert got["interpretation"]["headline"]
    assert any(item["layer"] == "primary" for item in got["evidence"])
    assert isinstance(got["timing_candidates"], list)
    assert all(item["source_kind"] for item in got["evidence"])
    assert all("pattern_kind" in pattern for pattern in got["chart_view"]["patterns"])
    if got["chart_view"]["gua_ci"] is not None:
        assert {"name", "text", "yao_ci"} <= set(got["chart_view"]["gua_ci"])
    policy = got["scope"]["followup_policy"]
    assert set(policy["same_matter"]) == {"clarification", "timing", "obstacle"}
    assert "used_evidence_ids" not in got["interpretation"]
    assert all("applies" not in item for item in got["condition_candidates"])


def test_liuyao_ask_coins_uses_raw_outcomes_and_engine_normalization():
    """Guard the full counsel -> engine coins contract, not mocked projections."""
    srv = create_counsel_root_mcp()
    got = asyncio.run(
        srv.call_tool("liuyao_ask", dict(LIUYAO_COINS_ARGS))
    ).structured_content

    assert got["casting"]["mode"] == "coins"
    assert got["casting"]["coin_convention"] == {"正": 3, "反": 2}
    assert got["casting"]["yaos"] == [8, 7, 6, 9, 7, 8]
    assert got["casting"]["dong_yao"] == [3, 4]
    for position, round_value in enumerate(got["casting"]["rounds"], 1):
        assert round_value["position"] == position
        assert round_value["yao_value"] in {6, 7, 8, 9}
        assert round_value["is_dong_yao"] == (round_value["yao_value"] in {6, 9})
    assert got["interpretation"]["headline"]


def test_qimen_ask_returns_self_contained_reading():
    srv = create_counsel_root_mcp()
    got = asyncio.run(srv.call_tool("qimen_ask", dict(QIMEN_ARGS))).structured_content
    assert got["technique"] == "qimen"
    assert "reading_digest" not in got
    assert got["scope"]["focus_value"] == "career"
    assert got["interpretation"]["headline"]
    assert got["evidence"]
    assert isinstance(got["timing_candidates"], list)
    assert all(item["fact_ref"] for item in got["evidence"])
    assert all("ying_qi_kind" in item for item in got["timing_candidates"])
    assert "used_evidence_ids" not in got["interpretation"]
    assert "used_timing_ids" not in got["interpretation"]
    assert all("applies" not in item for item in got["judgment_assertions"])


def test_qimen_ask_compacts_relation_evidence_without_dropping_pairs():
    srv = create_counsel_root_mcp()
    got = asyncio.run(srv.call_tool("qimen_ask", dict(QIMEN_ARGS))).structured_content
    relation = next(
        item for item in got["evidence"]
        if item["fact_ref"] == "factors.spirit_star_relations"
    )

    assert isinstance(relation["fact"], list)
    assert relation["fact"]
    pair_count = 0
    for left in relation["fact"]:
        assert {"spirit", "spirit_name", "spirit_palace", "relations"} <= set(left)
        for targets in left["relations"].values():
            pair_count += len(targets)
            for target in targets:
                assert {"star", "palace", "same_palace"} <= set(target)
                assert "gong" not in target
    assert pair_count == 72
