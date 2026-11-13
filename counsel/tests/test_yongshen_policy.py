
import copy
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TOOLS = Path(__file__).resolve().parents[1] / "app" / "natal" / "tools"

from app.natal.tools.factor_semantics import PredicateState  # noqa: E402
from app.natal.tools.factors import evaluate_factors  # noqa: E402
from app.natal.tools.natal_judgment import natal_liunian_query  # noqa: E402
from app.natal.tools.operators_liunian import _flow_element_in_selected  # noqa: E402
from app.natal.tools.yongshen import load_policy, resolve_yongshen, validate_policy  # noqa: E402
from app.natal.tools.yongshen.evaluator import _MISSING  # noqa: E402

WIRE_INPUTS = ROOT / "engine" / "internal" / "mcpcontract" / "testdata" / "wire_inputs"


def _bazi_chart_fixture(name: str) -> dict:
    from app.natal.tools.charts import _bazi_chart

    subject = json.loads((WIRE_INPUTS / name).read_text(encoding="utf-8"))
    return _bazi_chart(subject["input"]["subject"])


def _pan(full: dict, gender: str = "male") -> dict:
    return {
        "subject": {"standard_time": "", "longitude": 0, "gender": gender},
        "bazi": full,
        "ziwei": {"shi_er_gong": []},
        "ziwei_daxian": [],
    }


def test_policy_is_structurally_valid():
    validate_policy(load_policy())


def test_selected_yongshen_never_leaks_internal_missing_sentinel():
    full = _bazi_chart_fixture("bazi_chart.json")
    full["tiao_hou"] = {}

    result = resolve_yongshen(full)

    json.dumps(result, ensure_ascii=False)

    def assert_materialized(value: object) -> None:
        assert value is not _MISSING
        if isinstance(value, dict):
            for item in value.values():
                assert_materialized(item)
        elif isinstance(value, list):
            for item in value:
                assert_materialized(item)

    assert_materialized(result)
def test_neutral_chart_uses_tiaohou_merged_neutral_fallback():
    result = resolve_yongshen(_bazi_chart_fixture("bazi_chart_1984.json"))
    assert result["status"] == "resolved"
    assert result["rule_id"] == "YS-NEUTRAL-TH"
    assert result["selected"] == {
        "yong": "火",
        "xi": ["木"],
        "ji": ["水", "土"],
    }
    assert result["ge_ju_context"]["is_clear"] is True


def test_normal_chart_uses_fuyi_and_exposes_geju_context():
    result = resolve_yongshen(_bazi_chart_fixture("bazi_chart_1990.json"))
    assert result["status"] == "resolved"
    assert result["rule_id"] == "YS-FUYI"
    assert result["selected"]["yong"] == "水"
    assert result["selected"]["xi"] == ["金"]
    assert result["selected"]["ji"] == ["木"]
    assert result["ge_ju_context"]["theme"] == "表达、技艺、流通"


def test_summer_tiaohou_conflict_is_explicit_and_not_forced():
    result = resolve_yongshen(_bazi_chart_fixture("bazi_chart.json"))
    assert result["status"] == "needs_review"
    assert result["rule_id"] == "YS-TH"
    assert result["selected"]["yong"] == "水"
    assert result["selected"]["xi"] == []
    assert result["selected"]["ji"] == []
    assert {item["code"] for item in result["conflicts"]} == {
        "tiaohou_evidence_hidden", "xi_ji_overlap",
    }


def test_tiaohou_primary_missing_uses_secondary_when_condition_ok():
    full = _bazi_chart_fixture("bazi_chart.json")
    th = copy.deepcopy(full["tiao_hou"])
    th["primary"] = {
        **th["primary"],
        "transparent": False,
        "hidden": False,
        "occurrences": [],
    }
    th["secondary"] = {
        **th["secondary"],
        "transparent": True,
        "hidden": False,
        "occurrences": [{"stem": "庚", "pillar": "nian", "branch": "辰", "source": "gan"}],
    }
    th["secondary_condition"] = "庚金发水之源为佐"
    full["tiao_hou"] = th
    full["fu_yi"] = {**copy.deepcopy(full["fu_yi"]), "ji": "土"}

    result = resolve_yongshen(full)
    assert result["status"] == "partial"
    assert result["rule_id"] == "YS-TH"
    assert result["selected"]["yong"] == "金"
    assert result["selected"]["xi"] == []
def test_tiaohou_secondary_condition_not_met_keeps_needed_primary_partial():
    full = _bazi_chart_fixture("bazi_chart.json")
    th = copy.deepcopy(full["tiao_hou"])
    th["primary"] = {
        **th["primary"],
        "transparent": False,
        "hidden": False,
        "occurrences": [],
    }
    th["secondary_wuxing"] = "土"
    th["secondary"] = {
        **th["secondary"],
        "stem": "戊",
        "transparent": True,
        "hidden": False,
        "occurrences": [{"stem": "戊", "pillar": "nian", "branch": "辰", "source": "gan"}],
    }
    th["secondary_condition"] = "水多以戊佐"
    full["tiao_hou"] = th
    full["fu_yi"] = {
        **copy.deepcopy(full["fu_yi"]),
        "ji": "土",
        "wuxing_count": {**(full["fu_yi"].get("wuxing_count") or {}), "水": 1},
    }

    result = resolve_yongshen(full)
    assert result["status"] == "partial"
    assert result["rule_id"] == "YS-TH"
    assert result["selected"]["yong"] == "水"
    assert result["conflicts"] == [{
        "code": "tiaohou_evidence_missing",
        "message": "调候主辅神均未在原局落实，仅保留所需五行，不作强断",
    }]


def test_tiaohou_controlling_auxiliary_is_not_promoted_to_xi():
    full = _bazi_chart_fixture("bazi_chart.json")
    th = copy.deepcopy(full["tiao_hou"])
    th["secondary_wuxing"] = "土"
    th["secondary"] = {
        **th["secondary"],
        "stem": "戊",
        "transparent": True,
        "hidden": False,
        "occurrences": [{"stem": "戊", "pillar": "nian", "branch": "辰", "source": "gan"}],
    }
    th["secondary_condition"] = "水多以戊佐"
    full["tiao_hou"] = th

    result = resolve_yongshen(full)

    assert result["rule_id"] == "YS-TH"
    assert result["selected"]["xi"] == []
    assert result["selected"]["yong"] == "水"


def test_hidden_rooted_officer_breaks_cong_wang_verification():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    full["fu_yi"] = {
        **copy.deepcopy(full["fu_yi"]),
        "pattern": "从旺格",
        "yong": "火",
        "xi": "土",
        "ji": "金",
        "wang_shuai": {"火": "旺"},
        "basis": {
            **(full["fu_yi"].get("basis") or {}),
            "yin_bi_shu_liang": 3,
        },
    }
    full["ten_god_states"] = [
        {"shi_shen": "正印", "transparent": False, "rooted": True, "strength": "strong"},
        {"shi_shen": "比肩", "transparent": False, "rooted": True, "strength": "strong"},
        {"shi_shen": "劫财", "transparent": False, "rooted": True, "strength": "strong"},
        {"shi_shen": "正官", "transparent": False, "rooted": True, "strength": "strong"},
    ]

    result = resolve_yongshen(full)

    assert result["status"] == "needs_review"
    assert result["rule_id"] == "YS-CG-UNVERIFIED"
    assert any(item["source"] == "cong_ge" for item in result["evidence"])


def test_hidden_rooted_resource_breaks_cong_sha_verification():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    full["fu_yi"] = {
        **copy.deepcopy(full["fu_yi"]),
        "pattern": "从杀格",
        "yong": "水",
        "xi": "金",
        "ji": "木",
        "wang_shuai": {"火": "死"},
        "basis": {
            **(full["fu_yi"].get("basis") or {}),
            "ri_gan_gen": "none",
        },
    }
    full["ten_god_states"] = [
        {"shi_shen": "正官", "transparent": True, "rooted": True, "strength": "strong"},
        {"shi_shen": "七杀", "transparent": True, "rooted": True, "strength": "strong"},
        {"shi_shen": "正印", "transparent": False, "rooted": True, "strength": "neutral"},
    ]

    result = resolve_yongshen(full)

    assert result["status"] == "needs_review"
    assert result["rule_id"] == "YS-CG-UNVERIFIED"
def test_geju_clear_changes_context_without_yongshen():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    baseline = resolve_yongshen(full)
    full["ge_ju"] = {**copy.deepcopy(full["ge_ju"]), "pattern_god_source": "main_qi"}
    result = resolve_yongshen(full)
    assert result["rule_id"] == baseline["rule_id"] == "YS-FUYI"
    assert result["selected"] == baseline["selected"]
    assert result["ge_ju_context"]["is_clear"] is True
    assert baseline["ge_ju_context"]["is_clear"] is False


def test_congge_label_requires_matching_season_and_stem_evidence():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    full["fu_yi"] = {
        **copy.deepcopy(full["fu_yi"]),
        "pattern": "从财格",
        "yong": "土",
        "xi": "火",
        "ji": "木",
        "wang_shuai": {"火": "旺"},
        "basis": {
            **(full["fu_yi"].get("basis") or {}),
            "ri_gan_gen": "none",
        },
    }
    full["ten_god_states"] = [
        {"shi_shen": "正财", "transparent": True},
        {"shi_shen": "偏财", "transparent": True},
        {"shi_shen": "比肩", "transparent": False},
        {"shi_shen": "劫财", "transparent": False},
        {"shi_shen": "正印", "transparent": False},
        {"shi_shen": "偏印", "transparent": False},
    ]
    result = resolve_yongshen(full)
    assert result["status"] == "needs_review"
    assert result["rule_id"] == "YS-CG-UNVERIFIED"
    assert result["selected"]["yong"] == "无候选"


def test_verified_congge_selects_fuyi_cong_candidate():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    full["fu_yi"] = {
        **copy.deepcopy(full["fu_yi"]),
        "pattern": "从财格",
        "yong": "金",
        "xi": "土",
        "ji": "木",
            "wang_shuai": {"火": "囚"},
        "basis": {
                **(full["fu_yi"].get("basis") or {}),
                "ri_gan_gen": "none",
            },
    }
    full["ten_god_states"] = [
        {"shi_shen": "正财", "transparent": True},
        {"shi_shen": "偏财", "transparent": True},
        {"shi_shen": "比肩", "transparent": False},
        {"shi_shen": "劫财", "transparent": False},
        {"shi_shen": "正印", "transparent": False},
        {"shi_shen": "偏印", "transparent": False},
    ]
    result = resolve_yongshen(full)
    assert result["status"] == "resolved"
    assert result["rule_id"] == "YS-CG-VERIFIED"
    assert result["selected"]["yong"] == "金"
    assert "source" not in result["selected"]


def test_fake_congge_requires_review():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    full["fu_yi"] = {**copy.deepcopy(full["fu_yi"]), "pattern": "假从财格"}
    result = resolve_yongshen(full)
    assert result["status"] == "needs_review"
    assert result["rule_id"] == "YS-CG-FAKE"
    assert result["selected"]["yong"] == "无候选"


def test_neutral_tie_requires_review():
    full = _bazi_chart_fixture("bazi_chart_1984.json")
    full["element_states"] = [
        {**item, "season": "旺", "strength": "strong"}
        for item in full["element_states"]
        if item["wuxing"] in ("木", "火")
    ]
    result = resolve_yongshen(full)
    assert result["status"] == "needs_review"
    assert result["rule_id"] == "YS-NEUTRAL-REVIEW"


def test_unknown_yongshen_blocks_positive_and_negative_assertions():
    full = _bazi_chart_fixture("bazi_chart.json")
    factors = evaluate_factors("male", _pan(full), shushi="bazi")
    assert factors["官杀为用"] == PredicateState.UNKNOWN
    assert factors["官杀为忌"] == PredicateState.UNKNOWN
    assert factors["财星为忌"] == PredicateState.UNKNOWN


def test_known_yongshen_still_evaluates_factors():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    natal_chart = _pan(full)
    natal_chart["_selected_yongshen"] = resolve_yongshen(full)
    factors = evaluate_factors("male", natal_chart, shushi="bazi")
    assert factors["官杀为用"] == PredicateState.HIT
    assert factors["印为忌"] == PredicateState.HIT
    assert factors["食伤为用"] == PredicateState.MISS


def test_factor_evaluation_never_resolves_yongshen_implicitly():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    factors = evaluate_factors("male", _pan(full), shushi="bazi")
    assert factors["官杀为用"] == PredicateState.UNKNOWN
    assert factors["官杀为忌"] == PredicateState.UNKNOWN


def test_liunian_jishen_uses_selected_not_engine_fuyi():
    full = _bazi_chart_fixture("bazi_chart_1990.json")
    selected = resolve_yongshen(full)
    base = {"selected_yongshen": selected}
    const = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    assert _flow_element_in_selected(base, "甲", "gan", const) == 1
    assert _flow_element_in_selected(base, "寅", "zhi", const) == 1
    assert _flow_element_in_selected(base, "子", "zhi", const) == 0
    assert _flow_element_in_selected(base, "丙", "gan", const) == 0


def test_natal_liunian_query_computes_selected_yongshen_internally():
    subject = {
        "standard_time": "1990-05-20T11:49:00+08:00",
        "longitude": 116.4,
        "gender": "male",
    }
    result = natal_liunian_query(
        subject,
        2026, 1,
        ["career"],
    )
    assert result["natal_chart_digest"]
    assert result["years"]
    assert result["years"][0]["bazi"]["assertions"] == []
    assert isinstance(result["years"][0]["ziwei"]["assertions"], list)
