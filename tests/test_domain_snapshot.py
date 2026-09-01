"""领域快照契约：reserved domain facts 的结构与只读投影。"""
import pytest

import _helpers  # noqa: F401
from domain_snapshot import load_contract, project_domain_facts


def _rich_pan() -> dict:
    return {
        "full": {
            pillar: {
                "na_yin": f"{pillar}纳音", "cang_gan": {"main": "甲"},
                "is_void": False, "is_self_he": True,
                "is_kui_gang": False, "self_he_name": "自合",
            } for pillar in ("nian", "yue", "ri", "shi")
        } | {
            "san_yuan": {"胎元": "丙子"}, "xun_kong": "甲子旬空戌亥",
            "san_qi_name": "三奇", "gong_jia": ["拱夹"], "nayin_rel": [{"rel": "相生"}],
        },
        "chart": {"da_yun": {"steps": [], "current_step_index": 0}},
        "ziwei": {
            "gong_wei": [{"name": "命宫"}], "ju_shu": "火六局", "ming_zhu": "贪狼",
            "shen_zhu": "天梁", "ming_gong": "命宫", "shen_gong": "身宫",
            "kong_gong": [{"gong_name": "兄弟"}], "nian_gan": "庚", "nian_zhi": "午",
            "shi_zhi": "子", "ziwei_pos": "命宫",
        },
    }


def test_projected_domain_facts_match_contract():
    facts = project_domain_facts(_rich_pan())
    contract = load_contract()
    assert set(facts["八字"]) == set(contract["八字"])
    assert set(facts["紫微"]) == set(contract["紫微"])


def test_empty_pan_projects_empty_domain_facts():
    assert project_domain_facts({}) == {"八字": {}, "紫微": {}}


def test_domain_projection_does_not_mutate_pan():
    pan = _rich_pan()
    project_domain_facts(pan)
    assert "_snap" not in pan and "_ctx" not in pan


def test_fields_are_declared_when_present():
    facts = project_domain_facts(_rich_pan())
    assert facts["八字"]["大运"]["current_step_index"] == 0
    assert facts["紫微"]["局数"] == "火六局"
