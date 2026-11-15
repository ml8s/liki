"""Liuyao internal orchestration and current toolface guardrails."""

import json
from pathlib import Path

from app.contracts import DIVINATION_TOOLS
from app.contracts.divination import LiuyaoAskInput
from app.contracts.divination.liuyao import LiuyaoCastingEvent, LiuyaoChart
from app.divination.tools.liuyao_board import factors as board_factors
from app.divination.tools.liuyao_factors import project_factors


def test_board_sends_immutable_case_to_chart(monkeypatch):
    calls = []
    chart_view = {
        "gua_name": "乾为天", "palace": "乾", "palace_wuxing": "金",
        "lines": [], "bian_lines": [], "patterns": [],
        "yong_shen": {"name": "官鬼", "position": 1, "wang_shuai": "旺"},
    }
    analysis = {
        "moving_relations": [], "hidden_lines": [], "branch_relation_facts": [],
        "day_clash_facts": [], "moving_transformations": [], "san_he_candidates": [],
        "timing_candidates": [], "yong_shen_force_chain": None,
        "yong_shen_candidates": [], "conflicts": [],
        "tomb_school": {"earth_branch": "辰", "school": "engine_default_chen"},
        "reference_only_facts": [],
    }

    chart_view["analysis"] = analysis

    def engine(method, params):
        calls.append((method, params))
        return {
            "event": event,
            "chart": chart_view,
        }

    monkeypatch.setattr("app.divination.tools.liuyao_board.engine_data", engine)
    casting = {
        "mode": "coins", "order": "bottom_up",
        "coin_convention": {"正": 3, "反": 2},
        "rounds": [
            {"position": index, "coins": ["正", "反", "反"], "yao_value": 7,
             "yao_label": "少阳", "is_dong_yao": False}
            for index in range(1, 7)
        ],
        "yaos": [7] * 6, "dong_yao": [],
    }
    context = {
        "standard_time": "2026-10-01T10:00:00+08:00",
        "longitude": 116.4,
        "yong_shen": "官鬼",
    }
    event = {
        "casting": casting,
        "casting_context": context,
    }
    result = board_factors(event=event)
    assert calls == [("liuyao_chart", {"event": event})]
    assert "matter" not in result
    assert result["factors"]["casting"] == casting
    assert result["factors"]["liuyao_chart"] == chart_view
    assert result["factors"]["liuyao_chart"]["yong_shen"]["name"] == "官鬼"


def test_projected_factors_use_current_case_shape():
    casting = {
        "mode": "yaos", "order": "bottom_up",
        "yaos": [7] * 6, "dong_yao": [],
    }
    chart = {
        "gua_name": "乾为天", "bian_gua": None, "palace": "乾", "palace_wuxing": "金",
        "lines": [{
            "position": 1, "yao_value": 7, "gan_zhi": "甲子", "wuxing": "水",
            "liu_qin": "兄弟", "liu_shou": "青龙", "shi_ying": "世",
            "wang_shuai": "旺", "ri_chen_relations": {"relations": ["扶"]},
            "flags": {
                "moving": False, "yue_po": False, "xun_kong": False, "mu_ku": False,
                "mu_ku_branch": None, "mu_ku_element": None,
                "dong_sheng": False, "dong_ke": False,
            },
        }],
        "bian_lines": [],
        "ri_gan": "甲", "ri_zhi": "子", "yue_gan": "甲", "yue_zhi": "子",
        "xun_kong": ["戌", "亥"],
        "patterns": [],
        "yong_shen": {"name": "父母", "position": 1, "wang_shuai": "旺"},
    }
    analysis = {
        "moving_relations": [], "hidden_lines": [], "branch_relation_facts": [],
        "day_clash_facts": [], "moving_transformations": [], "san_he_candidates": [],
        "timing_candidates": [], "yong_shen_force_chain": None,
        "yong_shen_candidates": [], "conflicts": [],
        "tomb_school": {"earth_branch": "辰", "school": "engine_default_chen"},
        "reference_only_facts": [],
    }
    result = project_factors(
        casting=casting,
        chart=chart,
        analysis=analysis,
    )
    assert set(result) == {
        "casting", "liuyao_chart", "evidence", "analysis", "policy",
    }
    line = result["liuyao_chart"]["lines"][0]
    assert line["gan_zhi"] == "甲子"
    primary = {item["id"]: item for item in result["evidence"]["primary"]}
    assert primary["yong-shen-state"]["fact"]["wang_shuai"] == "旺"
    assert result["liuyao_chart"]["yong_shen"]["name"] == "父母"
    assert result["analysis"]["tomb_school"]["school"] == "engine_default_chen"
    assert (result["liuyao_chart"]["ri_gan"], result["liuyao_chart"]["ri_zhi"]) == ("甲", "子")


def test_divination_tools_have_single_ask_surface():
    assert {contract.name for contract in DIVINATION_TOOLS} == {
        "liuyao_ask", "qimen_ask",
    }


def test_liuyao_ask_input_accepts_primary_call():
    LiuyaoAskInput.model_validate({
        "event": {
            "casting": {
                "mode": "yaos", "order": "bottom_up",
                "yaos": [7, 7, 7, 7, 7, 7], "dong_yao": [],
            },
            "casting_context": {
                "standard_time": "2026-10-01T10:00:00+08:00",
                "longitude": 116.4,
                "yong_shen": "官鬼",
            },
        },
    })


def test_ask_never_creates_a_casting(monkeypatch):
    from app.divination.tools import liuyao_ask

    def forbidden_qigua(**kwargs):
        raise AssertionError("liuyao_ask must not call liuyao_qigua")

    chart = {
        "gua_name": "乾为天", "bian_gua": None, "palace": "乾", "palace_wuxing": "金",
        "lines": [], "yong_shen": {"name": "官鬼", "position": 1},
        "patterns": [],
    }
    casting = {
        "mode": "yaos", "order": "bottom_up",
        "yaos": [7, 7, 7, 7, 7, 7], "dong_yao": [],
    }
    context = {
        "standard_time": "2026-10-01T10:00:00+08:00",
        "longitude": 116.4,
        "yong_shen": "官鬼",
    }
    monkeypatch.setattr(
        "app.divination.tools.liuyao_case.build_liuyao_factors",
        lambda **kwargs: {
            "context": {
                "standard_time": "2026-10-01T10:00:00+08:00",
                "longitude": 116.4,
                "focus": {
                    "kind": "matter", "matter": "career",
                },
            },
            "factors": {
                "casting": casting,
                "liuyao_chart": chart,
                "evidence": {
                    "primary": [{
                        "id": "yong-shen-state", "source_ref": "liuyao_chart.lines[0]",
                        "fact": {}, "conclusion_scope": "conditional_candidate_not_outcome",
                    }],
                    "secondary": [], "reference": [], "conflicts": [],
                },
                "analysis": {
                    "moving_relations": [], "hidden_lines": [], "branch_relation_facts": [],
                    "day_clash_facts": [], "moving_transformations": [],
                    "san_he_candidates": [], "timing_candidates": [],
                    "yong_shen_force_chain": None, "yong_shen_candidates": [],
                    "conflicts": [], "tomb_school": {"earth_branch": "辰", "school": "engine_default_chen"},
                    "reference_only_facts": [],
                },
                "policy": {"may_output": ["tendency"], "must_cover_conflicts": True, "forbidden": []},
            },
        },
    )
    monkeypatch.setattr(
        "app.divination.tools.liuyao_case.validate_document",
        lambda name, document: None,
    )
    event = {
        "casting": casting,
        "casting_context": context,
    }
    reading = liuyao_ask.ask(event=event)
    assert set(reading) == {
        "technique", "evidence", "timing_candidates",
        "condition_candidates", "interpretation", "safety_advisory",
    }
    assert "event" not in reading
    assert "chart" not in reading
    assert reading["technique"] == "liuyao"


def test_go_chart_result_matches_python_contract():
    raw = json.loads(
        (Path(__file__).parent / "fixtures/liuyao/chart_result.json").read_text(
            encoding="utf-8"
        )
    )
    event = LiuyaoCastingEvent.model_validate(raw["event"])
    chart = LiuyaoChart.model_validate(raw["chart"])

    casting = event.casting
    assert casting.mode == "yaos"
    assert chart.yong_shen.name == "官鬼"
    assert len(chart.lines) == 6
    if casting.dong_yao == []:
        assert chart.bian_lines == []

    factors = project_factors(
        casting=casting.model_dump(by_alias=True),
        chart=chart.model_dump(),
        analysis=raw["chart"]["analysis"],
    )
    assert factors["liuyao_chart"] == chart.model_dump()
    assert factors["analysis"]["conflicts"] == raw["chart"]["analysis"].get("conflicts", [])
