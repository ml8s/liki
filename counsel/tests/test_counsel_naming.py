"""counsel 起名域（naming）1v1 对照测试。

counsel naming 工具（naming_*）在本地字库上完成匹配/取字/组名/评估；
输出必须与老实现（基线）完全一致——正交化补集只改入口，不改逻辑。
"""

import asyncio
import json
from pathlib import Path

import pytest  # noqa: E402
from app.counsel_mcp import create_counsel_mcp  # noqa: E402
from mcp import Client  # noqa: E402
from pydantic import ValidationError  # noqa: E402

GOLDEN = Path(__file__).resolve().parents[2] / "tests" / "golden" / "counsel_naming_baseline.json"


@pytest.fixture(scope="module")
def baseline():
    return json.loads(GOLDEN.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def srv():
    return create_counsel_mcp("naming")


def test_naming_tools_registered(srv):
    names = asyncio.run(srv.list_tools())
    assert [t.name for t in names] == [
        "naming_surname", "naming_pick", "naming_char",
        "naming_compose", "naming_check",
    ]


def test_naming_pick_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_pick", {"first": "木"}))
    assert r.structured_content == baseline["pick"]


def test_naming_char_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_char", {"char": "伟"}))
    assert r.structured_content == baseline["char"]


def test_naming_compose_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_compose", {"first": ["伟"]}))
    assert r.structured_content == baseline["compose"]


def test_naming_check_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_check", {
        "given_names": ["伟"], "yongshen": "木", "xishen": ["水"], "jishen": ["金"],
    }))
    assert r.structured_content == baseline["check"]


def _selected_yongshen() -> dict:
    return {
        "status": "resolved",
        "selected": {
            "yong": "木",
            "xi": ["水"],
            "ji": ["金"],
            "source": "fu_yi",
            "confidence": "high",
            "known_fields": ["yong", "xi", "ji"],
        },
        "conflicts": [],
        "factors": {},
        "alternatives": {},
        "ge_ju_context": {},
        "evidence": [],
        "policy_id": "liki.bazi.yongshen",
        "rule_id": "YS-FUYI",
        "chart_digest": "a" * 64,
        "yongshen_digest": "b" * 64,
    }


def _error_call(srv, name: str, arguments: dict):
    async def call():
        async with Client(srv, mode="legacy") as client:
            return await client.call_tool(name, arguments)

    return asyncio.run(call())


def test_naming_check_consumes_selected_yongshen(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_check", {
        "given_names": ["伟"], "selected_yongshen": _selected_yongshen(),
    }))
    out = r.structured_content
    evaluation = out["evaluations"][0]
    assert evaluation["wuxing"] == baseline["check"]["evaluations"][0]["wuxing"]
    assert evaluation["yongshen_source"] == {
        "policy_id": "liki.bazi.yongshen",
        "rule_id": "YS-FUYI",
        "status": "resolved",
        "chart_digest": "a" * 64,
        "yongshen_digest": "b" * 64,
    }


@pytest.mark.parametrize("status", ["needs_review", "unresolved"])
def test_naming_check_rejects_unresolved_selected_yongshen(srv, status):
    selected = {**_selected_yongshen(), "status": status}
    r = _error_call(srv, "naming_check", {
        "given_names": ["伟"], "selected_yongshen": selected,
    })
    assert r.is_error is True
    assert "未形成可消费结论" in r.content[0].text


def test_naming_check_requires_selected_yongshen_provenance(srv):
    selected = _selected_yongshen()
    selected.pop("yongshen_digest")
    r = _error_call(srv, "naming_check", {
        "given_names": ["伟"], "selected_yongshen": selected,
    })
    assert r.is_error is True
    assert "yongshen_digest" in r.content[0].text


def test_naming_check_rejects_two_yongshen_sources():
    from app.contracts.naming import NamingCheckInput
    from pydantic import TypeAdapter

    with pytest.raises(ValidationError):
        # 互斥由 anyOf 两分支的 extra=forbid 表达：混传时两支都拒绝。
        TypeAdapter(NamingCheckInput).validate_python({
            "given_names": ["伟"],
            "yongshen": "木",
            "selected_yongshen": _selected_yongshen(),
        })


def test_naming_surname_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_surname", {"source_surname": "zhang"}))
    assert r.structured_content == baseline["surname"]
