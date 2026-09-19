"""本命事实投影契约：reserved domain facts 的结构与只读投影。"""
import pytest

import _helpers  # noqa: F401
from natal_projection import load_contract, project_natal_facts


def _rich_pan() -> dict:
    return {
        "full": {
            pillar: {
                "gan": "甲", "zhi": "子",
                "na_yin": f"{pillar}纳音", "cang_gan": {"main": "甲"},
                "shi_shens": [{"shi_shen": "比肩", "gan": "甲", "source": "gan"}],
                "self_sitting": "沐浴", "shen_sha": [{"name": "天乙贵人"}],
                "day_master_trend": f"{pillar}地势",
                "xun": f"{pillar}旬", "xun_kong": f"{pillar}旬空支",
                "is_void": False, "is_self_he": True,
                "is_kui_gang": False, "self_he_name": "自合",
            } for pillar in ("nian", "yue", "ri", "shi")
        } | {
            "birth_year": 1990,
            "_yong_shen_wrapper": {
                "fu_yi": {"model": "support_control_with_day_master_strength"},
                "tiao_hou": {"model": "qiongtong_primary_secondary_table"},
                "ge_ju": {"model": "ziping_month_command_candidate"},
            },
            "chang_sheng": [{"pillar": "nian", "stage": "长生"}],
            "zodiac": "鼠",
            "zodiac": "鼠",
            "san_yuan": {"胎元": "丙子"}, "tai_xi": "辛丑",
            "gan_chong": [{"gan_a": "甲", "gan_b": "庚", "pillar_a": 0, "pillar_b": 1, "position": "adjacent"}],
            "gan_he": [{"gan_a": "甲", "gan_b": "己", "position": "adjacent"}],
            "zhi_liu_he": [{"zhi_a": "子", "zhi_b": "丑"}],
            "san_he_partial": [{"type": "半合", "name": "申子辰半合"}],
            "san_he": [{"type": "三合", "name": "申子辰水局"}],
            "san_hui": [{"type": "三会", "name": "亥子丑水方"}],
            "liu_chong": [{"zhi_a": "子", "zhi_b": "午"}],
            "liu_hai": [{"zhi_a": "子", "zhi_b": "未"}],
            "liu_xing": [{"zhi_a": "子", "zhi_b": "卯"}],
            "liu_po": [{"zhi_a": "子", "zhi_b": "酉"}],
            "an_he": [{"zhi_a": "子", "zhi_b": "戌"}],
            "shen_sha_school": {
                "dual_reference": ["天乙贵人", "桃花"],
                "policy": "union_of_year_and_day_references",
            },
            "lu_roots": [{"shi_shen": "正印", "gan": "癸", "zhi": "子"}],
            "relation_groups": [
                {"field": "gan_he", "group": "甲己"},
                {"field": "zhi_liu_he", "group": "子丑"},
            ],
            "ten_god_states": [{"shi_shen": "正印", "wuxing": "水", "transparent": True, "hidden": False, "count": 1, "rooted": True, "timely": True, "strength": "strong"}],
            "element_states": [{"wuxing": "水", "season": "旺", "season_strength": "strong", "strength": "strong", "transparent": True, "rooted": True, "controls": "火", "controlled_by": "土", "controller_strength": "weak", "reasons": ["得令"]}],
            "atomic_facts": {
                "day_master_element": "木", "day_master_stem": "甲", "day_branch": "子",
                "month_longevity": "沐浴", "pattern_god_transparent": False,
            },
            "fu_yi": {"qiangruo": "身强", "yong": "木", "xi": "水", "ji": "金", "wuxing_count": {}, "wang_shuai": {}, "model": "test", "basis": {}},
            "tiao_hou": {"primary_wuxing": "火", "season": "春", "model": "qiongtong_primary_secondary_table", "primary": {"stem": "丙", "transparent": True, "hidden": False, "occurrences": [], "relation_facts": []}},
            "ge_ju": {"ge_ju": "正印格", "yong_fa": "顺用", "pattern_god_source": "main_qi", "structure": {}},
            "day_xun": "甲戌旬", "day_xun_kong": "申酉",
            "san_qi_name": "三奇", "gong_jia": ["拱夹"], "nayin_rel": [{"rel": "相生"}],
        },
        "chart": {"birth_year": 1990, "da_yun": {"steps": [], "current_step_index": 0}},
        "ziwei": {
            "school": {"leap_month": "iztro_compatible", "source": "test"},
            "gong_wei": [{"name": "命宫"}], "ju_shu": "火六局", "ju_shu_name": "火六局",
            "ming_zhu": "贪狼", "shen_zhu": "天梁", "ming_gong": "命宫", "shen_gong": "身宫",
            "si_hua": {"贪狼": "禄"}, "palace_facts": [
                {"palace": "命宫", "kind": "star", "target": "紫微主星", "star": "七杀"}
            ],
            "kong_gong": [{"gong_name": "兄弟"}], "patterns": ["七杀朝斗"],
            "san_fang": [{"name": "命宫", "zhu_xing": ["七杀"]}], "nian_gan": "庚", "nian_zhi": "午",
            "shi_zhi": "子", "ziwei_pos": "命宫", "birth_year": 1990,
            "lunar_month": 4, "lunar_day": 26, "birth_lunar_month": 4,
            "birth_is_leap": False, "gender": "male",
        },
        "ziwei_daxian": [{"gong": "命宫", "name": "甲子", "start_year": 1990, "end_year": 1999, "qi_sui": 1, "zhi_sui": 10}],
    }


def test_projected_domain_facts_match_contract():
    facts = project_natal_facts(_rich_pan())
    contract = load_contract()
    assert set(facts["八字"]) == set(contract["八字"])
    assert set(facts["紫微"]) == set(contract["紫微"])


def test_empty_pan_projects_empty_domain_facts():
    assert project_natal_facts({}) == {"八字": {}, "紫微": {}}


def test_domain_projection_does_not_mutate_pan():
    pan = _rich_pan()
    project_natal_facts(pan)
    assert "_snap" not in pan and "_ctx" not in pan


def test_fields_are_declared_when_present():
    facts = project_natal_facts(_rich_pan())
    assert facts["八字"]["大运"]["current_step_index"] == 0
    assert facts["八字"]["tiao_hou"]["model"] == "qiongtong_primary_secondary_table"
    assert facts["八字"]["稳定关系组"][0]["group"] == "甲己"
    assert facts["八字"]["命理原子事实"]["day_master_stem"] == "甲"
    assert facts["八字"]["天干相冲"][0]["gan_a"] == "甲"
    assert facts["紫微"]["局数"] == "火六局"
    assert facts["紫微"]["宫位原子事实"][0]["palace"] == "命宫"
    assert facts["紫微"]["四化"]["贪狼"] == "禄"
    assert facts["紫微"]["格局候选"] == ["七杀朝斗"]
    assert facts["八字"]["十二长生"][0]["pillar"] == "nian"
