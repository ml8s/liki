"""Liuyao internal orchestration and current toolface guardrails."""

import pytest
from app.contracts import DIVINATION_TOOLS
from app.contracts.divination import LiuyaoAskInput
from app.divination.tools.liuyao_board import factors as board_factors
from app.divination.tools.liuyao_casting import qigua
from app.divination.tools.liuyao_factors import project_factors
from app.divination.tools.liuyao_matters import resolve_matter


def test_matter_table_maps_explicit_yong_shen():
    assert resolve_matter("career")["yong_shen"] == "官鬼"
    assert resolve_matter("wealth")["yong_shen"] == "妻财"
    assert resolve_matter("relationship", perspective="male")["yong_shen"] == "妻财"
    assert resolve_matter("relationship", perspective="female")["yong_shen"] == "官鬼"


def test_relationship_requires_perspective():
    with pytest.raises(ValueError, match="perspective"):
        resolve_matter("relationship")


def test_qigua_sends_original_coins_to_engine(monkeypatch):
    calls = []

    def engine(method, params):
        calls.append((method, params))
        return {"casting": {"mode": "coins", "rounds": [["正", "反", "反"]] * 6}}

    monkeypatch.setattr("app.divination.tools.liuyao_casting.engine_data", engine)
    rounds = [["正", "反", "反"]] * 6
    assert qigua(mode="coins", rounds=rounds)["mode"] == "coins"
    assert calls == [("liuyao_qigua", {"casting": {"mode": "coins", "rounds": rounds}})]


def test_qigua_rejects_coin_conversion_inputs():
    with pytest.raises(ValueError, match="exactly 6"):
        qigua(mode="coins", rounds=[["正", "反", "反"]] * 5)
    with pytest.raises(ValueError, match="exactly 3"):
        qigua(mode="coins", rounds=[["正", "反"]] * 6)


def test_board_sends_casting_and_resolved_yongshen_only(monkeypatch):
    calls = []
    raw_chart = {
        "name": "乾为天", "ben_gua": "乾", "gong": "乾", "gong_wuxing": "金",
        "lines": [], "yong_shen": {"name": "官鬼", "position": 1},
        "timing_candidates": [],
    }

    def engine(method, params):
        calls.append((method, params))
        return raw_chart

    monkeypatch.setattr("app.divination.tools.liuyao_board.engine_data", engine)
    casting = {"mode": "coins", "rounds": [["正", "反", "反"]] * 6}
    result = board_factors(
        casting=casting,
        qi_gua_longitude=116.4,
        qi_gua_time="2026-10-01T10:00:00+08:00",
        matter="career",
        question="这次面试能不能通过",
    )
    assert calls == [("liuyao_chart", {
        "yong_shen": "官鬼",
        "qi_gua": {
            "qi_gua_time": "2026-10-01T10:00:00+08:00",
            "longitude": 116.4,
            "casting": casting,
        },
    })]
    assert result["matter"]["matter"] == "career"
    assert result["factors"]["selected_yongshen_context"]["selected_yongshen"]["name"] == "官鬼"


def test_projected_factors_use_current_case_shape():
    casting = {"mode": "yaos", "yaos": [7] * 6, "dong_yao": []}
    chart = {
        "name": "乾为天", "ben_gua": "乾", "gong": "乾", "gong_wuxing": "金",
        "lines": [{
            "position": 1, "type": 7, "gan": "甲", "zhi": "子", "wuxing": "水",
            "liu_qin": "兄弟", "liu_shou": "青龙", "shi_ying": "世",
            "wang_shuai": "旺", "ri_chen_relations": ["扶"],
        }],
        "bian_yao": [{
            "position": 1, "type": 8, "gan": "乙", "zhi": "丑", "wuxing": "土",
            "liu_qin": "官鬼",
        }],
        "ri_chen_gan": "甲", "ri_chen_zhi": "子",
        "yue_jian_gan": "甲", "yue_jian_zhi": "子",
        "xun_kong": ["戌", "亥"],
        "yong_shen": {"name": "父母", "position": 1, "wang_shuai": "旺"},
    }
    result = project_factors(casting, chart, {"text": "测试"})
    assert set(result) == {
        "question", "casting", "liuyao_chart", "selected_yongshen_context",
        "evidence", "chart_facts", "timing_candidates", "policy",
    }
    line = result["liuyao_chart"]["lines"][0]
    assert line["gan_zhi"] == "甲子"
    primary = {item["id"]: item for item in result["evidence"]["primary"]}
    assert primary["yong-shen-state"]["fact"]["wang_shuai"] == "旺"
    assert result["selected_yongshen_context"]["selected_yongshen"]["name"] == "父母"
    assert result["liuyao_chart"]["bian_lines"][0]["liu_qin"] == "官鬼"
    assert (result["liuyao_chart"]["ri_gan"], result["liuyao_chart"]["ri_zhi"]) == ("甲", "子")


def test_divination_tools_are_one_shot_asks():
    assert {contract.name for contract in DIVINATION_TOOLS} == {
        "liuyao_ask", "qimen_ask",
    }


def test_liuyao_ask_input_accepts_primary_call():
    LiuyaoAskInput.model_validate({
        "question": "这次面试能不能通过",
        "casting": {"mode": "yaos", "yaos": [7, 7, 7, 7, 7, 7]},
        "focus": {"kind": "matter", "matter": "career"},
        "qi_gua_longitude": 116.4,
        "qi_gua_time": "2026-10-01T10:00:00+08:00",
    })
