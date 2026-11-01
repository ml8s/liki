"""Counsel naming MCP contract and baseline tests."""

import asyncio
import json
from pathlib import Path

import pytest
from app.counsel_mcp import create_counsel_root_mcp  # noqa: E402
from mcp import Client

GOLDEN = (
    Path(__file__).resolve().parents[2]
    / "tests"
    / "golden"
    / "counsel_naming_baseline.json"
)


def _selected_yongshen(yong="木", xi=None, ji=None, status="resolved"):
    return {
        "selected": {
            "yong": yong,
            "xi": xi or [],
            "ji": ji or [],
        },
        "status": status,
        "conflicts": [],
        "rule_id": "TEST-RULE",
        "policy_id": "liki.bazi.yongshen",
        "chart_digest": "a" * 64,
        "yongshen_digest": "b" * 64,
    }


@pytest.fixture(scope="module")
def baseline():
    return json.loads(GOLDEN.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def srv():
    return create_counsel_root_mcp()


def test_naming_tools_registered(srv):
    names = {tool.name for tool in asyncio.run(srv.list_tools())}
    assert names >= {
        "naming_surname",
        "naming_pick",
        "naming_char",
        "naming_compose",
        "naming_check",
    }


def _call(server, name, arguments):
    async def call():
        async with Client(server, mode="legacy") as client:
            return await client.call_tool(name, arguments)

    return asyncio.run(call())


def test_naming_pick_uses_yongshen_role(srv, baseline):
    result = _call(
        srv,
        "naming_pick",
        {
            "selected_yongshen": _selected_yongshen("木"),
            "first_role": "yong",
        },
    )
    assert result.structured_content == baseline["pick"]


def test_naming_pick_supports_yong_xi(srv):
    result = _call(
        srv,
        "naming_pick",
        {
            "selected_yongshen": _selected_yongshen("木", ["水"]),
            "first_role": "yong",
            "second_role": "xi",
        },
    )
    pools = result.structured_content["pools"]
    assert [(pool["slot"], pool["role"], pool["wuxing"]) for pool in pools] == [
        ("first", "yong", "木"),
        ("second", "xi", "水"),
    ]


def test_naming_pick_rejects_ji_role(srv):
    result = _call(
        srv,
        "naming_pick",
        {
            "selected_yongshen": _selected_yongshen("木", [], ["金"]),
            "first_role": "ji",
        },
    )
    assert result.is_error is True


def test_naming_char_matches_baseline(srv, baseline):
    result = _call(srv, "naming_char", {"char": "伟"})
    assert result.structured_content == baseline["char"]


def test_naming_compose_matches_baseline(srv, baseline):
    result = _call(srv, "naming_compose", {"first_chars": ["伟"]})
    assert result.structured_content == baseline["compose"]


def test_naming_check_classifies_each_character(srv, baseline):
    result = _call(
        srv,
        "naming_check",
        {
            "given_names": ["伟"],
            "selected_yongshen": _selected_yongshen("木"),
        },
    )
    assert result.structured_content == baseline["check"]


def test_naming_check_without_selected_yongshen_only_reports_facts(srv):
    result = _call(srv, "naming_check", {"given_names": ["伟"]})
    evaluation = result.structured_content["evaluations"][0]
    assert evaluation["valid"] is True
    assert all("wuxing_role" not in character for character in evaluation["characters"])


def test_naming_check_rejects_unresolved_yongshen(srv):
    result = _call(
        srv,
        "naming_check",
        {
            "given_names": ["伟"],
            "selected_yongshen": _selected_yongshen(status="needs_review"),
        },
    )
    assert result.is_error is True


def test_naming_check_rejects_forbidden_character(srv):
    result = _call(
        srv,
        "naming_check",
        {
            "given_names": ["伟"],
            "selected_yongshen": _selected_yongshen("木"),
            "forbidden_chars": ["伟"],
        },
    )
    evaluation = result.structured_content["evaluations"][0]
    assert evaluation["valid"] is False
    assert evaluation["errors"] == [
        {"code": "forbidden_character", "char": "伟"}
    ]


def test_naming_check_requires_required_character(srv):
    result = _call(
        srv,
        "naming_check",
        {
            "given_names": ["伟"],
            "selected_yongshen": _selected_yongshen("木"),
            "required_chars": ["伟"],
        },
    )
    assert result.structured_content["evaluations"][0]["valid"] is True

    result = _call(
        srv,
        "naming_check",
        {
            "given_names": ["伟"],
            "selected_yongshen": _selected_yongshen("木"),
            "required_chars": ["林"],
        },
    )
    evaluation = result.structured_content["evaluations"][0]
    assert evaluation["valid"] is False
    assert evaluation["errors"] == [
        {"code": "required_character_missing", "char": "林"}
    ]


def test_naming_check_rejects_conflicting_constraints(srv):
    result = _call(
        srv,
        "naming_check",
        {
            "given_names": ["伟"],
            "selected_yongshen": _selected_yongshen("木"),
            "forbidden_chars": ["伟"],
            "required_chars": ["伟"],
        },
    )
    assert result.is_error is True


def test_naming_surname_matches_baseline(srv, baseline):
    result = _call(srv, "naming_surname", {"source_surname": "zhang"})
    assert result.structured_content == baseline["surname"]
