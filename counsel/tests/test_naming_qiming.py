"""起名（qiming）对照测试：内联期望值来自 engine Go qiming 实现验证输出。

起名是"命理之上的应用"，纯函数实现在 Python 层；期望值逐字段取自 Go 引擎
（engine/internal/engine/qiming）对相同输入的输出，保证移植一致。
"""
from __future__ import annotations

import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from app.naming import qiming  # noqa: E402


def test_pick_mu_huo_2():
    r = qiming.pick_chars("木", "火", 2)
    assert r["wuxing1"] == "木" and r["wuxing2"] == "火"
    assert [p["slot"] for p in r["pools"]] == ["first", "second"]
    assert len(r["pools"][0]["chars"]) == 1969
    assert len(r["pools"][1]["chars"]) == 1472
    assert r["pools"][0]["chars"][0] == {"char": "㭎", "frequency": "rare"}


def test_pick_shui_1():
    r = qiming.pick_chars("水", "", 1)
    assert r["wuxing1"] == "水" and "wuxing2" not in r
    assert len(r["pools"]) == 1 and len(r["pools"][0]["chars"]) == 1673


def test_pick_jin_shui_2():
    r = qiming.pick_chars("金", "水", 2)
    assert r["wuxing1"] == "金" and r["wuxing2"] == "水"
    assert [len(p["chars"]) for p in r["pools"]] == [1522, 1673]


def test_pick_errors():
    with pytest.raises(ValueError, match="count must be 1 or 2"):
        qiming.pick_chars("木", "", 3)
    with pytest.raises(ValueError, match="wuxing2 is not allowed"):
        qiming.pick_chars("木", "火", 1)
    with pytest.raises(ValueError, match="invalid wuxing1"):
        qiming.pick_chars("金木", "", 2)


def test_compose_double():
    r = qiming.compose_names(["子", "轩"], ["然", "宇"], 5)
    assert r == {"total_possible": 4, "names": ["子然", "子宇", "轩然", "轩宇"]}


def test_compose_single():
    r = qiming.compose_names(["安", "宁", "静"], [], 3)
    assert r == {"total_possible": 3, "names": ["安", "宁", "静"]}


def test_compose_errors():
    with pytest.raises(ValueError, match="first must contain 1 to 256"):
        qiming.compose_names([], [], 3)
    with pytest.raises(ValueError, match="duplicate character"):
        qiming.compose_names(["子", "子"], [], 3)


def test_evaluate_zixuan():
    e = qiming.evaluate_names(["子轩", "浩然"], "木", ["火"], ["金"])
    assert e[0]["given_name"] == "子轩" and e[0]["valid"] is True
    assert e[0]["phonetic"]["tones"] == "5-1"
    assert e[0]["wuxing"] == {"yong": False, "xi": False, "ji": False}
    assert e[0]["characters"][0]["char"] == "子"
    assert e[0]["characters"][0]["wuxing"] == "水"
    assert e[1]["given_name"] == "浩然" and e[1]["phonetic"]["tones"] == "4-2"
    assert e[1]["wuxing"]["ji"] is True


def test_evaluate_invalid():
    e = qiming.evaluate_names(["子轩", "死", "x", "子"], "木", ["火"], ["金"])
    assert e[0]["valid"] is True
    assert e[1] == {
        "given_name": "死", "valid": False,
        "errors": [{"code": "negative_character_forbidden", "char": "死"}],
    }
    assert e[2]["errors"] == [{"code": "character_not_found", "char": "x"}]
    assert e[3]["valid"] is True


def test_evaluate_no_constraint():
    e = qiming.evaluate_names(["云"], "", [], [])
    assert e[0]["valid"] is True and "wuxing" not in e[0]


def test_evaluate_errors():
    with pytest.raises(ValueError, match="invalid yongshen"):
        qiming.evaluate_names(["子"], "金木", [], [])
    with pytest.raises(ValueError, match="yongshen must be disjoint"):
        qiming.evaluate_names(["子"], "木", ["木"], [])
    with pytest.raises(ValueError, match="xishen and jishen must be disjoint"):
        qiming.evaluate_names(["子"], "木", ["火"], ["火"])


def test_surname_zhang():
    r = qiming.match_surnames("zhang", 3)
    assert r["strategy"] == "phonetic"
    assert [c["surname"] for c in r["candidates"]] == ["张", "章", "赵"]
    assert [c["match_level"] for c in r["candidates"]] == [
        "pinyin_exact", "pinyin_exact", "phonetic_close",
    ]
    assert [c["tone"] for c in r["candidates"]] == [1, 1, 4]


def test_surname_liu():
    r = qiming.match_surnames("liu", 4)
    assert [c["surname"] for c in r["candidates"]] == ["柳", "刘", "廖", "李"]
    assert [c["match_level"] for c in r["candidates"]] == [
        "pinyin_exact", "pinyin_exact", "romanization_exact", "phonetic_close",
    ]


def test_surname_fallback():
    r = qiming.match_surnames("zyxwv", 3)
    assert r["strategy"] == "baijiaxing_fallback"
    assert [c["surname"] for c in r["candidates"]] == ["赵", "钱", "孙"]
    assert all(c["match_level"] == "fallback_baijiaxing" for c in r["candidates"])


def test_lookup_char():
    c = qiming.lookup_char("轩")
    assert c and c["char"] == "轩" and c["wuxing"] == "土" and c["stroke"] == 7
    assert qiming.lookup_char("x") is None
    assert qiming.lookup_char("轩轩") is None