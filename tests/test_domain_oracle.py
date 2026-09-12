"""独立领域 oracle 测试：外部口径数据驱动，不复制实现内部表。

这些用例先由经典规则写成 JSON fixture，再打到实现边界上；fixture 与实现
表分开维护，任何一侧行为变化都必须显式核对领域口径。
"""
from __future__ import annotations

import json
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[1]
ORACLE = ROOT / "tests" / "fixtures" / "domain_oracle"


def _load(name: str) -> dict:
    return json.loads((ORACLE / name).read_text(encoding="utf-8"))


def _lu_chart(case: dict) -> dict:
    """把 oracle 用例机械映射为因子算子输入；不添加命理判断。"""
    full: dict[str, dict] = {key: {"shi_shens": [], "zhi": ""} for key in ("nian", "yue", "ri", "shi")}
    for pillar, zhi in zip(("nian", "yue", "ri", "shi"), case.get("branches", [])):
        full[pillar]["zhi"] = zhi
    for item in case.get("visible", []):
        full[item["pillar"]]["shi_shens"].append({
            "shi_shen": item["shi_shen"], "gan": item["gan"], "source": "gan",
        })
    for item in case.get("hidden", []):
        full[item["pillar"]]["shi_shens"].append({
            "shi_shen": item["shi_shen"], "gan": item["gan"], "source": "main_qi",
        })
    full["lu_roots"] = case.get("lu_roots", [])
    return {"full": full}


def test_domain_oracle_datasets_have_expected_coverage() -> None:
    bazi = _load("bazi_core.json")
    ziwei = _load("ziwei_core.json")
    liuyao = _load("liuyao_core.json")
    qimen = _load("qimen_core.json")
    fengshui = _load("fengshui_core.json")
    atomic = _load("bazi_atomic_facts.json")
    ten_god_states = _load("bazi_ten_god_states.json")
    relation_groups = _load("bazi_relation_groups.json")
    liunian_atomic = _load("bazi_liunian_atomic.json")
    lu_roots = _load("bazi_lu_roots.json")
    ziwei_pattern = _load("ziwei_pattern_semantics.json")
    liuyao_pattern = _load("liuyao_pattern_semantics.json")
    liuyao_dongyao = _load("liuyao_dong_yao_priority.json")
    gender_shen_sha = _load("bazi_gender_shen_sha.json")
    annual_shen_sha = _load("bazi_annual_shen_sha.json")
    nayin_shen_sha = _load("bazi_nayin_shen_sha.json")
    bazi_classic_corrections = _load("bazi_classic_corrections.json")
    bazi_yongshen_structure = _load("bazi_yongshen_structure.json")
    huangli = _load("huangli_core.json")
    qimen_specialized = _load("qimen_specialized.json")
    assert len(bazi["ten_gods_all_100"]) == 10 * 10
    assert len(bazi["hidden_stems_all_12"]) == 12
    assert len(bazi["chang_sheng_all_120"]) == 10 * 12
    assert len(bazi["nayin_all_30_pairs"]) == 30
    assert len(ziwei["si_hua_all_10"]) == 10
    assert len(ziwei["tian_kui_all_10"]) == 10
    assert len(ziwei["tian_yue_all_10"]) == 10
    assert len(ziwei["soul_star_all_12"]) == 12
    assert len(ziwei["body_star_all_12"]) == 12
    assert len(liuyao["na_jia_all_8"]) == 8
    assert len(liuyao["world_by_palace_sequence"]) == 8
    assert len(liuyao["conflict_cases"]) == 6
    assert len(qimen["solar_term_ju_all_72"]) == 24 * 3
    assert len(qimen["yingqi_cases"]) == 2
    assert len(qimen_specialized["geng_ge_cases"]) == 2
    assert len(qimen_specialized["lost_cases"]) == 1
    assert len(qimen_specialized["thief_cases"]) == 3
    assert len(qimen_specialized["tianwang_cases"]) == 1
    assert len(fengshui["mountains_24"]) == 24
    assert len(fengshui["eight_mansion_patterns"]) == 8
    assert len(fengshui["san_yuan_yun_periods"]) == 10
    assert len(fengshui["xuankong_four_situation_anchors"]) == 4
    assert len(huangli["jianchu_sequence"]) == 12
    assert len(huangli["qing_long_start_all_12"]) == 12
    assert len(huangli["event_rules"]) == 15
    assert len(atomic["cases"]) == 4
    assert len(ten_god_states["cases"]) == 4
    assert len(ten_god_states["dayun_root_cases"]) == 4
    assert len(relation_groups["cases"]) == 5
    assert len(liunian_atomic["cases"]) == 6
    assert len(lu_roots["cases"]) == 2
    assert len(gender_shen_sha["cases"]) == 4
    assert len(annual_shen_sha["positions"]) == 4
    assert len(nayin_shen_sha["cases"]) == 6
    assert len(bazi_classic_corrections["yang_ren"]) == 10
    assert len(bazi_classic_corrections["yue_ren_ge"]) == 10
    assert len(bazi_classic_corrections["feedback_charts"]) == 1
    assert len(bazi_classic_corrections["season_shen_sha"]) == 12
    assert len(bazi_classic_corrections["tian_luo_di_wang"]) == 6
    assert len(bazi_yongshen_structure["ge_ju_cases"]) == 4
    assert len(bazi_yongshen_structure["fu_yi_cases"]) == 2
    assert len(bazi_yongshen_structure["tiao_hou_cases"]) == 1
    assert len(ziwei_pattern["san_fang_cases"]) == 2
    assert len(ziwei_pattern["palace_fact_cases"]) == 2
    assert len(ziwei_pattern["pattern_cases"]) == 17
    assert len(liuyao_pattern["tomb_cases"]) == 4
    assert len(liuyao_pattern["selection_cases"]) == 4
    assert len(liuyao_pattern["pattern_cases"]) == 3
    assert len(liuyao_pattern["computed_cases"]) == 1
    assert len(liuyao_pattern["whole_gua_cases"]) == 5
    assert len(liuyao_pattern["whole_gua_closure"]["six_clash"]) == 10
    assert len(liuyao_pattern["whole_gua_closure"]["six_combination"]) == 8
    assert len(liuyao_pattern["sanhe_cases"]) == 2
    assert len(liuyao_dongyao["cases"]) == 6


def test_lu_root_operator_uses_ten_stem_lu_not_any_same_element() -> None:
    from _helpers import TOOLS  # noqa: F402? helper is module under tests
    import sys
    if str(TOOLS) not in sys.path:
        sys.path.insert(0, str(TOOLS))
    from factor_constants import load_constants
    from operators_natal import _op

    oracle = _load("bazi_operator_lu_root.json")
    for case in oracle["cases"]:
        got = _op(case["op"], case["args"], "male", _lu_chart(case))
        assert got == case["expect"], f"{case['id']}: {case['why']}"
