"""qimen counsel 投影表的「通用术数常识」一致性。

方位 / 时干阴阳 / 五行生克属各层通用的术数常识（跨 MCP 服务边界，py 侧独立持有），
这里把 counsel 的投影表绑定到权威基础（luoshu / 干支阴阳 / 五行生克），
防止跨服务边界静默漂移。奇门盘面事实（内外盘/远近）与解读性表（门星吉凶/神将/宫象）
不在本测试范围。
"""
from __future__ import annotations

import csv
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
QDATA = ROOT / "counsel" / "app" / "divination" / "tools" / "data"
LUOSHU = ROOT / "engine" / "internal" / "engine" / "luoshu" / "data" / "luoshu.json"


def _rows(name: str) -> list[dict]:
    with (QDATA / name).open(encoding="utf-8-sig", newline="") as fh:
        return list(csv.DictReader(fh))


def test_palace_directions_match_luoshu():
    luoshu = {
        p["name"]: ("center" if p["direction"] == "中" else p["direction"])
        for p in json.loads(LUOSHU.read_text(encoding="utf-8"))
    }
    got = {r["gong"]: r["direction"] for r in _rows("qimen_palace_directions.csv")}
    assert got == luoshu, "方位表与 luoshu 不一致"


def test_hour_polarities_match_gan_yinyang():
    yang = set("甲丙戊庚壬")
    expected = {g: ("yang" if g in yang else "yin") for g in "甲乙丙丁戊己庚辛壬癸"}
    got = {r["gan"]: r["polarity"] for r in _rows("qimen_hour_polarities.csv")}
    assert got == expected


def test_wuxing_relations_match_universal():
    generates = {("木", "火"), ("火", "土"), ("土", "金"), ("金", "水"), ("水", "木")}
    controls = {("木", "土"), ("土", "水"), ("水", "火"), ("火", "金"), ("金", "木")}
    expected = {}
    for a in "木火土金水":
        for b in "木火土金水":
            if a == b:
                rel = "same"
            elif (a, b) in generates:
                rel = "generates"
            elif (a, b) in controls:
                rel = "controls"
            elif (b, a) in generates:
                rel = "generated_by"
            elif (b, a) in controls:
                rel = "controlled_by"
            else:
                raise AssertionError(f"no relation for {a}{b}")
            expected[(a, b)] = rel
    got = {
        (r["source"], r["target"]): r["relation"]
        for r in _rows("qimen_wuxing_relations.csv")
    }
    assert got == expected


PLATE = ROOT / "engine" / "internal" / "engine" / "qimen" / "data" / "plate.json"
JI_XIONG = QDATA / "qimen_men_star_ji_xiong.json"


def _assert_shared_match(engine: dict, counsel: dict, label: str) -> None:
    shared = engine.keys() & counsel.keys()
    mismatch = {k: (engine[k], counsel[k]) for k in shared if engine[k] != counsel[k]}
    assert not mismatch, f"{label}五行与 engine 不一致: {mismatch}"


def test_men_star_wuxing_matches_engine_plate():
    # engine plate 定义门/星五行；counsel 表为投影用，两集合可各自有流派扩展
    # （engine 有「中门」而 counsel 无），只要求共享键的五行一致。
    plate = json.loads(PLATE.read_text(encoding="utf-8"))
    engine_men = {e["door"]: e["wuxing"] for e in plate["door_wuxing"]}
    engine_star = {e["star"]: e["wuxing"] for e in plate["star_wuxing"]}
    doc = json.loads(JI_XIONG.read_text(encoding="utf-8"))
    counsel_men = {k: v["wuxing"] for k, v in doc["men"].items()}
    counsel_star = {k: v["wuxing"] for k, v in doc["xing"].items()}
    _assert_shared_match(engine_men, counsel_men, "门")
    _assert_shared_match(engine_star, counsel_star, "星")
