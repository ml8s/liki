"""counsel 起名域（naming）1v1 对照测试。

counsel naming 工具（naming_*）复用 qiming 逻辑（字库来自 engine），
输出必须与老实现（基线）完全一致——正交化补集只改入口，不改逻辑。
"""
from __future__ import annotations

import asyncio
import json
from pathlib import Path

import pytest  # noqa: E402
from app.counsel_mcp import create_counsel_mcp  # noqa: E402

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
    r = asyncio.run(srv.call_tool("naming_pick", {"wuxing1": "木"}))
    assert json.loads(r.content[0].text) == baseline["pick"]


def test_naming_char_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_char", {"char": "伟"}))
    assert json.loads(r.content[0].text) == baseline["char"]


def test_naming_compose_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_compose", {"first": ["伟"]}))
    assert json.loads(r.content[0].text) == baseline["compose"]


def test_naming_check_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_check", {
        "given_names": ["伟"], "yongshen": "木", "xishen": ["水"], "jishen": ["金"],
    }))
    assert json.loads(r.content[0].text) == baseline["check"]


def test_naming_surname_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("naming_surname", {"source_surname": "zhang"}))
    assert json.loads(r.content[0].text) == baseline["surname"]
