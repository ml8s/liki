"""counsel 起名域（naming）1v1 对照测试。

counsel naming 工具（qiming_*）复用 qiming 逻辑（字库来自 engine），
输出必须与老实现（基线）完全一致——正交化补集只改入口，不改逻辑。
"""
from __future__ import annotations

import asyncio
import json
import os
from pathlib import Path

os.environ.setdefault("LIKI_COUNSEL_SERVICE_DOMAIN", "naming")

import pytest  # noqa: E402

from app.counsel_server import create_counsel_server  # noqa: E402

GOLDEN = Path(__file__).resolve().parents[2] / "tests" / "golden" / "counsel_qiming_baseline.json"


@pytest.fixture(scope="module")
def baseline():
    return json.loads(GOLDEN.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def srv():
    return create_counsel_server("naming")


def test_naming_tools_registered(srv):
    names = asyncio.run(srv.list_tools())
    assert [t.name for t in names] == [
        "qiming_surname", "qiming_pick", "qiming_char",
        "qiming_compose", "qiming_check",
    ]


def test_qiming_pick_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("qiming_pick", {"wuxing1": "木"}))
    assert json.loads(r.content[0].text) == baseline["pick"]


def test_qiming_char_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("qiming_char", {"char": "伟"}))
    assert json.loads(r.content[0].text) == baseline["char"]


def test_qiming_compose_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("qiming_compose", {"first": ["伟"]}))
    assert json.loads(r.content[0].text) == baseline["compose"]


def test_qiming_check_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("qiming_check", {
        "given_names": ["伟"], "yongshen": "木", "xishen": ["水"], "jishen": ["金"],
    }))
    assert json.loads(r.content[0].text) == baseline["check"]


def test_qiming_surname_matches_baseline(srv, baseline):
    r = asyncio.run(srv.call_tool("qiming_surname", {"source_surname": "zhang"}))
    assert json.loads(r.content[0].text) == baseline["surname"]