"""liuyao Python 编排层契约：不换算硬币，不把 matter 传给 engine。"""
from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest
from jsonschema import validate


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_casting as cast_tool  # noqa: E402
import liuyao_matters as matters  # noqa: E402
import liuyao_paipan as paipan  # noqa: E402


def test_matter_table_maps_explicit_yong_shen():
    assert matters.resolve_matter("career")["yong_shen"] == "官鬼"
    assert matters.resolve_matter("wealth")["yong_shen"] == "妻财"
    assert matters.resolve_matter("study")["yong_shen"] == "父母"
    assert matters.resolve_matter("relationship", perspective="male")["yong_shen"] == "妻财"
    assert matters.resolve_matter("relationship", perspective="female")["yong_shen"] == "官鬼"


def test_relationship_requires_perspective():
    with pytest.raises(ValueError, match="perspective"):
        matters.resolve_matter("relationship")


def test_qigua_sends_original_coins_to_engine(monkeypatch):
    calls = []

    def fake_engine_data(method, params):
        calls.append((method, params))
        return {
            "yaos": [8, 7, 9, 6, 7, 8],
            "dong_yao": [3, 4],
            "casting": {
                "schema_version": "liuyao-cast-v1",
                "mode": "coins",
                "yaos": [8, 7, 9, 6, 7, 8],
                "dong_yao": [3, 4],
            },
        }

    monkeypatch.setattr(cast_tool, "engine_data", fake_engine_data)
    rounds = [["正", "反", "反"]] * 6
    result = cast_tool.qigua(mode="coins", rounds=rounds)
    assert calls == [("liuyao.qigua", {"mode": "coins", "rounds": rounds})]
    assert result["casting"]["mode"] == "coins"


def test_qigua_rejects_coin_conversion_inputs(monkeypatch):
    monkeypatch.setattr(
        cast_tool,
        "engine_data",
        lambda *_: (_ for _ in ()).throw(AssertionError("engine called")),
    )
    with pytest.raises(ValueError, match="exactly 6"):
        cast_tool.qigua(mode="coins", rounds=[["正", "反", "反"]] * 5)
    with pytest.raises(ValueError, match="exactly 3"):
        cast_tool.qigua(mode="coins", rounds=[["正", "反"]] * 6)


def test_chart_maps_matter_and_builds_snapshot(monkeypatch):
    calls = []
    casting = {
        "schema_version": "liuyao-cast-v1",
        "mode": "coins",
        "yaos": [7, 7, 7, 7, 7, 7],
        "dong_yao": [],
    }
    raw_chart = {
        "name": "乾为天",
        "ben_gua": "乾",
        "gong": "乾",
        "gong_wuxing": "金",
        "lines": [{
            "position": 1,
            "type": 7,
            "gan": "甲",
            "zhi": "子",
            "wuxing": "水",
            "liu_qin": "子孙",
            "liu_shou": "青龙",
            "shi_ying": "世",
            "chang_sheng_yue": "长生",
            "dong_self": False,
            "mu_ku_branch": "戌",
            "mu_ku_element": "火",
        }],
        "yong_shen": {"name": "官鬼", "position": 1, "wang_shuai": "休", "chang_sheng": "长生"},
        "ying_qi": {"assessment": "candidate only"},
        "day_clash_facts": [{"position": 1, "kind": "暗动"}],
        "moving_transformations": [{"position": 1, "from_branch": "子", "to_branch": "丑"}],
        "san_he_candidates": [],
        "hidden_lines": [{"position": 1, "liu_qin": "妻财"}],
        "branch_relation_facts": [{"left_position": 1, "right_position": 2, "relations": ["六冲"]}],
        "force_chain": {"yong_shen": "官鬼", "yong_element": "火", "position": 1, "entries": []},
        "yong_shen_candidates": [{"position": 1, "selected": True}],
        "timing_candidates": [{
            "id": "moving-value-1-子",
            "mechanism": "动爻逢值",
            "position": 1,
            "trigger_branch": "子",
            "confidence": "candidate",
        }],
    }

    def fake_engine_data(method, params):
        calls.append((method, params))
        if method == "time.now":
            return {"cst": "2026-09-08T12:00:00+08:00"}
        assert method == "liuyao.chart"
        return raw_chart

    monkeypatch.setattr(paipan, "engine_data", fake_engine_data)
    result = paipan.chart(
        casting=casting,
        matter="career",
        question="这次面试能不能通过",
    )
    assert calls[0] == ("time.now", {})
    assert calls[1][0] == "liuyao.chart"
    assert calls[1][1] == {
        "solar_time": "2026-09-08T12:00:00+08:00",
        "yong_shen": "官鬼",
        "casting": casting,
    }
    assert "matter" not in calls[1][1]
    assert result["matter"]["matter"] == "career"
    assert result["snapshot"]["schema_version"] == "liuyao-snapshot-v2"
    assert result["snapshot"]["focus"]["yong_shen"]["name"] == "官鬼"
    assert result["snapshot"]["focus"]["world"]["position"] == 1
    assert result["snapshot"]["board"]["lines"][0]["chang_sheng_yue"] == "长生"
    assert result["snapshot"]["board"]["lines"][0]["flags"]["mu_ku_branch"] == "戌"
    assert result["snapshot"]["focus"]["yong_shen"]["chang_sheng"] == "长生"
    assert result["snapshot"]["facts"]["hidden_lines"][0]["liu_qin"] == "妻财"
    assert result["snapshot"]["facts"]["branch_relation_facts"][0]["relations"] == ["六冲"]
    assert result["snapshot"]["facts"]["day_clash_facts"][0]["kind"] == "暗动"
    assert result["snapshot"]["facts"]["moving_transformations"][0]["to_branch"] == "丑"
    assert result["snapshot"]["facts"]["force_chain"]["yong_element"] == "火"
    assert result["snapshot"]["followup"]["locked"] is True
    assert result["snapshot"]["followup"]["original_solar_time"] == "2026-09-08T12:00:00+08:00"
    assert result["snapshot"]["timing_candidates"][0]["mechanism"] == "动爻逢值"


def test_chart_requires_exclusive_yong_shen_source():
    with pytest.raises(ValueError, match="互斥"):
        paipan.chart(matter="career", yong_shen="官鬼", casting={"yaos": [7] * 6})


def test_tool_schema_matches_python_surface():
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    names = {item["function"]["name"] for item in schema["tools"]}
    assert {"liuyao_snapshot", "liuyao_ask"} <= names
    assert {"liuyao_chart", "liuyao_qigua", "liuyao_read", "liuyao_report"} .isdisjoint(names)


def test_liuyao_snapshot_input_against_tool_schema():
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    tool = next(
        item["function"]
        for item in schema["tools"]
        if item["function"]["name"] == "liuyao_snapshot"
    )
    validate(
        {
            "question": "这次面试能不能通过",
            "mode": "yaos",
            "yaos": [7, 7, 7, 7, 7, 7],
            "matter": "career",
        },
        tool["parameters"],
    )
