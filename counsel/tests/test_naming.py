"""起名（qiming）契约测试。

起名是"命理之上的应用"，纯函数实现在 counsel 层；期望值锁定当前字库
（data/naming_characters.csv）的输出，防止字库或算法漂移。
"""


import pytest
from app.naming import naming  # noqa: E402


def selected_yongshen(yong="木", xi=None, ji=None):
    return {
        "selected": {"yong": yong, "xi": xi or [], "ji": ji or []},
        "status": "resolved",
    }


def test_pick_yong_yong():
    plan = selected_yongshen("木")
    r = naming.pick_chars_by_selected_yongshen(plan, "yong", "yong")
    assert [(p["slot"], p["role"], p["wuxing"]) for p in r["pools"]] == [
        ("first", "yong", "木"), ("second", "yong", "木"),
    ]
    assert len(r["pools"][0]["chars"]) == 1969
    assert r["pools"][0]["chars"][0] == {"char": "㭎", "frequency": "rare"}


def test_pick_single_yong():
    r = naming.pick_chars_by_selected_yongshen(selected_yongshen("水"), "yong")
    assert len(r["pools"]) == 1
    assert (r["pools"][0]["slot"], r["pools"][0]["role"]) == ("first", "yong")
    assert len(r["pools"][0]["chars"]) == 1673


def test_pick_multiple_xi_pools():
    r = naming.pick_chars_by_selected_yongshen(
        selected_yongshen("金", ["水", "木"]), "yong", "xi"
    )
    assert [(p["slot"], p["role"], p["wuxing"]) for p in r["pools"]] == [
        ("first", "yong", "金"),
        ("second", "xi", "水"),
        ("second", "xi", "木"),
    ]
    assert [len(p["chars"]) for p in r["pools"]] == [1522, 1673, 1969]


def test_pick_errors():
    with pytest.raises(ValueError, match="只支持 yong/xi"):
        naming.pick_chars_by_selected_yongshen(
            selected_yongshen("木", [], ["金"]), "ji"
        )
    with pytest.raises(ValueError, match="单名 first_role 只能是 yong"):
        naming.pick_chars_by_selected_yongshen(
            selected_yongshen("木", ["水"]), "xi"
        )
    with pytest.raises(ValueError, match="至少一个字位必须是 yong"):
        naming.pick_chars_by_selected_yongshen(
            selected_yongshen("木", ["水", "火"]), "xi", "xi"
        )


def test_compose_double():
    r = naming.compose_names(["子", "轩"], ["然", "宇"], 5)
    assert r == {"total_possible": 4, "names": ["子然", "子宇", "轩然", "轩宇"]}


def test_compose_single():
    r = naming.compose_names(["安", "宁", "静"], [], 3)
    assert r == {"total_possible": 3, "names": ["安", "宁", "静"]}


def test_compose_errors():
    with pytest.raises(ValueError, match="first must contain 1 to 256"):
        naming.compose_names([], [], 3)
    with pytest.raises(ValueError, match="duplicate character"):
        naming.compose_names(["子", "子"], [], 3)


def test_evaluate_zixuan():
    plan = selected_yongshen("木")
    e = naming.evaluate_names(["子轩", "浩然"], plan)
    assert e[0]["given_name"] == "子轩" and e[0]["valid"] is True
    assert e[0]["phonetic"]["tones"] == "5-1"
    assert e[0]["characters"][0]["wuxing_role"] == "neutral"
    assert e[0]["characters"][0]["char"] == "子"
    assert e[0]["characters"][0]["wuxing"] == "水"
    assert e[1]["given_name"] == "浩然" and e[1]["phonetic"]["tones"] == "4-2"
    assert e[1]["characters"][0]["wuxing_role"] == "neutral"


def test_evaluate_invalid():
    e = naming.evaluate_names(["子轩", "死", "x", "子"], selected_yongshen("木"))
    assert e[0]["valid"] is True
    assert e[1] == {
        "given_name": "死", "valid": False,
        "errors": [{"code": "negative_character_forbidden", "char": "死"}],
    }
    assert e[2]["errors"] == [{"code": "character_not_found", "char": "x"}]
    assert e[3]["valid"] is True


def test_evaluate_no_constraint():
    e = naming.evaluate_names(["云"])
    assert e[0]["valid"] is True
    assert all("wuxing_role" not in character for character in e[0]["characters"])


def test_evaluate_errors():
    with pytest.raises(ValueError, match="同时属于"):
        naming.evaluate_names(["子"], selected_yongshen("木", ["木"]))


def test_surname_zhang():
    r = naming.match_surnames("zhang", 3)
    assert r["strategy"] == "phonetic"
    assert [c["surname"] for c in r["candidates"]] == ["张", "章", "赵"]
    assert [c["match_kind"] for c in r["candidates"]] == [
        "pinyin_exact", "pinyin_exact", "phonetic_close",
    ]
    assert [c["tone"] for c in r["candidates"]] == [1, 1, 4]


def test_surname_liu():
    r = naming.match_surnames("liu", 4)
    assert [c["surname"] for c in r["candidates"]] == ["柳", "刘", "廖", "李"]
    assert [c["match_kind"] for c in r["candidates"]] == [
        "pinyin_exact", "pinyin_exact", "romanization_exact", "phonetic_close",
    ]


def test_surname_fallback():
    r = naming.match_surnames("zyxwv", 3)
    assert r["strategy"] == "baijiaxing_fallback"
    assert [c["surname"] for c in r["candidates"]] == ["赵", "钱", "孙"]
    assert all(c["match_kind"] == "fallback_baijiaxing" for c in r["candidates"])


def test_lookup_char():
    c = naming.lookup_char("轩")
    assert c and c["char"] == "轩" and c["wuxing"] == "土" and c["stroke"] == 7
    assert naming.lookup_char("x") is None
    assert naming.lookup_char("轩轩") is None
