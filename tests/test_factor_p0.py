"""P0 因子回归：恒真因子、格神透干与格局值域。"""
import csv
import json
from pathlib import Path

from _helpers import mock_base_context
from duanyu import load_table
from factors import _op, evaluate_factors

ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills" / "liki-bazi" / "tools"
CONST = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))


def _minimal_fac() -> dict:
    base = mock_base_context()
    base["wuxing"] = {"wang_shuai": {}, "count": {}}
    base["yongshen"] = {}
    return base


def test_career_palace_main_star_prosperity_is_not_always_true() -> None:
    empty = evaluate_factors("male", _minimal_fac(), shushi="ziwei")
    assert empty["官禄宫主星庙旺"] == 0

    prosperous = evaluate_factors(
        "male",
        {**_minimal_fac(),
         "ziwei": {
            "gong_wei": [
                {
                    "name": "官禄宫",
                    "xing_yao": [{"xing": "紫微", "liang_du": "庙"}],
                }
            ]
         }},
        shushi="ziwei",
    )
    assert prosperous["官禄宫主星庙旺"] == 1


def test_ge_shen_tou_requires_pattern_ten_god_on_stem() -> None:
    base = _minimal_fac()
    base["yongshen"] = {"ge_ju": {"ge_ju": "正官格"}}

    matching = {
        "full": {
            "yue": {
                "shi_shens": [
                    {"shi_shen": "正官", "gan": "辛", "source": "gan"}
                ]
            }
        }
    }
    assert _op("格神透", [], "male", {**base, **matching}) == 1

    non_matching = {
        "full": {
            "yue": {
                "shi_shens": [
                    {"shi_shen": "正财", "gan": "己", "source": "gan"}
                ]
            }
        }
    }
    assert _op("格神透", [], "male", {**base, **non_matching}) == 0


def test_geju_closures_are_explicit() -> None:
    assert set(CONST["月令格局"]) == {
        "正官格", "七杀格", "正财格", "偏财格",
        "正印格", "偏印格", "食神格", "伤官格",
        "建禄格", "月刃格", "杂格",
    }
    assert set(CONST["扶抑从格"]) == {
        "从旺格", "从杀格", "从财格", "从儿格",
        "假从旺格", "假从杀格", "假从财格", "假从儿格",
    }


def test_fuyi_congge_is_a_scalar_factor() -> None:
    base = _minimal_fac()
    base["yongshen"] = {"fu_yi": {"pattern": "从杀格"}}

    snap = evaluate_factors("male", base, shushi="bazi")
    assert snap["扶抑从格"] == "从杀格"


def test_geju_table_separates_month_pattern_from_fuyi_congge() -> None:
    rows = load_table("bazi_格局.csv")

    month_values = {row["约束"]["月令格"] for row in rows if row["约束"].get("月令格")}
    congge_values = {row["约束"]["扶抑从格"] for row in rows if row["约束"].get("扶抑从格")}

    assert month_values == set(CONST["月令格局"])
    assert congge_values <= set(CONST["扶抑从格"])
    assert not month_values & set(CONST["扶抑从格"])
