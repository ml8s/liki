"""liuyao Python 编排层契约：不换算硬币，不把 matter 传给 engine。"""

import sys
from pathlib import Path

import pytest
from app.contracts import DIVINATION_TOOLS
from app.contracts.divination import LiuyaoSnapshotInput

ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "counsel/app/divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import app.divination.tools.liuyao_board as liuyao_board  # noqa: E402
import app.divination.tools.liuyao_casting as cast_tool  # noqa: E402
import app.divination.tools.liuyao_factors as projection  # noqa: E402
import app.divination.tools.liuyao_matters as matters  # noqa: E402


def test_matter_table_maps_explicit_yong_shen():
    assert matters.resolve_matter("career")["yong_shen"] == "官鬼"
    assert matters.resolve_matter("wealth")["yong_shen"] == "妻财"
    assert matters.resolve_matter("study")["yong_shen"] == "父母"
    assert matters.resolve_matter("self")["yong_shen"] == "世爻"
    assert matters.resolve_matter("other")["yong_shen"] == "应爻"
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
                "mode": "coins",
                "yaos": [8, 7, 9, 6, 7, 8],
                "dong_yao": [3, 4],
            },
        }

    monkeypatch.setattr(cast_tool, "engine_data", fake_engine_data)
    rounds = [["正", "反", "反"]] * 6
    casting = cast_tool.qigua(mode="coins", rounds=rounds)
    assert calls == [("liuyao.qigua", {"mode": "coins", "rounds": rounds})]
    assert casting["mode"] == "coins"


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


def test_factors_maps_matter_and_projection(monkeypatch):
    calls = []
    casting = {
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

    monkeypatch.setattr(liuyao_board, "server_time", lambda tz: {"instant": "2026-09-08T04:00:00Z", "timezone": tz, "local_time": "2026-09-08T12:00:00+08:00"})

    def fake_engine_data(method, params):
        calls.append((method, params))
        assert method == "liuyao.chart"
        return raw_chart

    monkeypatch.setattr(liuyao_board, "engine_data", fake_engine_data)
    result = liuyao_board.factors(
        casting=casting,
        matter="career",
        question="这次面试能不能通过",
        longitude=116.4,
        timezone="Asia/Shanghai",
    )
    assert len(calls) == 1
    assert calls[0][0] == "liuyao.chart"
    assert calls[0][1] == {
        "instant": "2026-09-08T04:00:00Z",
        "timezone": "Asia/Shanghai",
        "longitude": 116.4,
        "yong_shen": "官鬼",
        "casting": casting,
    }
    assert "matter" not in calls[0][1]
    assert result["matter"]["matter"] == "career"
    assert "schema_version" not in result["factors"]
    assert result["factors"]["focus"]["yong_shen"]["name"] == "官鬼"
    assert result["factors"]["focus"]["world"]["position"] == 1
    assert result["factors"]["board"]["lines"][0]["chang_sheng_yue"] == "长生"
    assert result["factors"]["board"]["lines"][0]["flags"]["mu_ku_branch"] == "戌"




def test_projected_board_preserves_core_time_and_line_states():
    casting = {"mode": "yaos", "yaos": [7] * 6, "dong_yao": []}
    chart = {
        "name": "乾为天",
        "ben_gua": "乾",
        "gong": "乾",
        "gong_wuxing": "金",
        "lines": [
            {
                "position": index + 1,
                "type": 7,
                "gan": "甲",
                "zhi": "子",
                "wuxing": "水",
                "liu_qin": "兄弟",
                "liu_shou": "青龙",
                "shi_ying": "世" if index == 0 else "",
            }
            for index in range(6)
        ],
        "bian_yao": [
            {"position": 1, "type": 8, "gan": "乙", "zhi": "丑", "wuxing": "土", "liu_qin": "官鬼"}
        ],
        "ri_chen_gan": "甲",
        "ri_chen_zhi": "子",
        "yue_jian_gan": "甲",
        "yue_jian_zhi": "子",
        "xun_kong": ["戌", "亥"],
        "wang_shuai": ["旺", "相", "休", "囚", "死", "旺"],
        "ri_chen_relations": [["扶"], ["生"], None, None, None, None],
        "yong_shen": {"name": "父母", "position": 1, "wang_shuai": "旺"},
    }
    result = projection.project_factors(casting, chart, {"text": "测试"})
    board = result["board"]
    assert board["lines"][0]["wang_shuai"] == "旺"
    assert board["lines"][0]["ri_chen_relations"] == ["扶"]
    assert board["bian_lines"][0]["liu_qin"] == "官鬼"
    assert (board["ri_gan"], board["ri_zhi"]) == ("甲", "子")
    assert (board["yue_gan"], board["yue_zhi"]) == ("甲", "子")
    assert board["xun_kong"] == ["戌", "亥"]


def test_factors_requires_exclusive_yong_shen_source():
    with pytest.raises(ValueError, match="互斥"):
        liuyao_board.factors(matter="career", yong_shen="官鬼", casting={"yaos": [7] * 6}, longitude=116.4)


def test_factors_projects_hidden_yong_shen_without_flying_line_facts():
    casting = {"mode": "yaos", "yaos": [8, 7, 9, 9, 7, 7], "dong_yao": [3, 4]}
    chart = {
        "name": "天风姤",
        "ben_gua": "姤",
        "bian_gua": "涣",
        "gong": "乾",
        "gong_wuxing": "金",
        "lines": [],
        "yong_shen": {
            "name": "妻财",
            "position": 0,
            "is_hidden": True,
            "wang_shuai": "死",
            "fu_shen": {"position": 2, "liu_qin": "妻财", "zhi": "寅"},
        },
        "force_chain": {
            "yong_shen": "妻财",
            "yong_element": "木",
            "position": 2,
            "is_hidden": True,
            "entries": [],
        },
        "hidden_lines": [],
        "timing_candidates": [],
    }
    factors = projection.project_factors(casting, chart, {
        "text": "这次能否成？", "matter": "general", "explicit_yong_shen": None, "perspective": None,
    })
    focus = factors["focus"]
    assert focus["yong_shen"]["is_hidden"] is True
    assert focus["yong_line"] is None
    primary_ids = {item["id"] for item in factors["evidence"]["primary"]}
    assert "yong-shen-hidden-state" in primary_ids
    hidden_state = next(item for item in factors["evidence"]["primary"] if item["id"] == "yong-shen-hidden-state")
    assert hidden_state["fact"]["fu_shen"]["zhi"] == "寅"
    assert hidden_state["fact"]["wang_shuai"] == "死"


def test_tool_schema_matches_python_surface():
    names = {contract.name for contract in DIVINATION_TOOLS}
    assert {"liuyao_snapshot", "liuyao_query"} <= names
    assert {"liuyao_chart", "liuyao_qigua", "liuyao_read", "liuyao_report"} .isdisjoint(names)


def test_matter_topics_and_tool_schema_use_one_closed_set():
    import app.divination.tools.liuyao_snapshot as liuyao_snapshot
    from app.divination.tools.liuyao_topic_guidance import load_topic_table

    contract = next(c for c in DIVINATION_TOOLS if c.name == "liuyao_snapshot")
    schema_topics = set(contract.input_schema["properties"]["topic"]["enum"])
    topics = set(load_topic_table()["topics"])
    mapped = set(liuyao_snapshot.MATTER_TO_TOPIC.values())

    assert schema_topics == topics
    assert liuyao_snapshot.ALLOWED_TOPICS == topics
    assert mapped <= topics
    assert {"relationship", "legal"} <= schema_topics


def test_liuyao_snapshot_input_accepts_primary_call():
    LiuyaoSnapshotInput.model_validate(
        {
            "question": "这次面试能不能通过",
            "casting": {"mode": "yaos", "yaos": [7, 7, 7, 7, 7, 7]},
            "matter": "career",
            "longitude": 116.4,
            "timezone": "Asia/Shanghai",
        }
    )
