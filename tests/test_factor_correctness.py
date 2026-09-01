"""因子层契约：引擎字段、关系算子、性别解析与表定义唯一性。"""
from __future__ import annotations

import csv
from collections import defaultdict
from pathlib import Path

from _helpers import mock_base_context
from factors import _op, _OP_NAMES, _LIU_OP_NAMES, _shishen_from_pan
import factors

TOOLS = Path(__file__).resolve().parents[1] / "skills" / "liki-bazi" / "tools"
META = {"因子", "术数", "原语直通", "依据"}


def test_engine_source_gan_means_stem_and_counts() -> None:
    pan = {"full": {
        "nian": {"shi_shens": [{"shi_shen": "正官", "gan": "辛", "source": "gan"}]},
        "yue": {"shi_shens": [{"shi_shen": "正印", "gan": "壬", "source": "main_qi"}]},
        "ri": {},
        "shi": {},
    }}
    states = _shishen_from_pan(pan)
    assert states["正官"]["tou_gan"] is True
    assert states["正官"]["cang_zhi"] is False
    assert states["正官"]["count"] == 1
    assert states["正印"]["tou_gan"] is False
    assert states["正印"]["cang_zhi"] is True
    assert states["正印"]["count"] == 1


def test_ge_shen_uses_engine_gan_source() -> None:
    base = mock_base_context()
    base["yongshen"] = {"ge_ju": {"ge_ju": "正官格"}}
    chart = {"full": {"yue": {"shi_shens": [
        {"source": "gan", "gan": "辛", "shi_shen": "正官"}
    ]}}}
    assert _op("格神透", [], "male", {**base, **chart}) == 1


def test_relation_operator_reads_engine_results() -> None:
    full = {
        "gan_he": [{"gan_a": "甲", "gan_b": "己"}],
        "zhi_liu_he": [{"zhi_a": "子", "zhi_b": "丑"}],
        "san_he": [{"name": "申子辰水局"}],
        "san_hui": [{"name": "寅卯辰木方"}],
        "liu_chong": [{"zhi_a": "子", "zhi_b": "午"}],
        "liu_hai": [{"zhi_a": "子", "zhi_b": "未"}],
        "liu_xing": [{"zhi_a": "寅", "zhi_b": "巳"}, {"zhi_a": "巳", "zhi_b": "申"}],
    }
    chart = {
        "full": full,
        "chart": {
            "nian": {"zhi": "寅"},
            "yue": {"zhi": "巳"},
            "ri": {"zhi": "申"},
            "shi": {"zhi": "子"},
        },
    }
    base = mock_base_context()
    assert _op("关系", ["gan_he", "甲己"], "male", chart) == 1
    assert _op("关系", ["gan_he", "乙庚"], "male", chart) == 0
    assert _op("关系", ["zhi_liu_he", "子丑"], "male", chart) == 1
    assert _op("关系", ["san_he", "申子辰"], "male", chart) == 1
    assert _op("关系", ["san_hui", "寅卯辰"], "male", chart) == 1
    assert _op("关系", ["liu_chong", "子午"], "male", chart) == 1
    assert _op("关系", ["liu_hai", "子未"], "male", chart) == 1
    assert _op("关系", ["liu_xing", "寅巳申"], "male", chart) == 1
    assert _op("关系", ["liu_xing", "丑戌未"], "male", chart) == 0


def test_factor_table_has_no_unknown_operator_or_self_reference() -> None:
    for filename in ("factors.csv", "factors_liunian.csv"):
        with (TOOLS / "factors" / filename).open(encoding="utf-8", newline="") as f:
            rows = list(csv.DictReader(f))
        for row in rows:
            factor_id = row["factor_id"]
            expression = row["expression"]
            if row["kind"] == "direct":
                name = expression.split("[", 1)[0]
                assert name in _OP_NAMES | _LIU_OP_NAMES, f"{filename}: {factor_id} 未知算子 {name}"
                continue
            assert expression != factor_id, f"{filename}: {factor_id} 自引用"
            if row["kind"] == "condition":
                name = expression.split("[", 1)[0]
                assert name in _OP_NAMES | _LIU_OP_NAMES, f"{filename}: {expression} 未知算子 {name}"


def test_spouse_star_mixed_uses_gender_specific_stars() -> None:
    male = mock_base_context(
        正财={"count": 1, "wuxing": "土"},
        偏财={"count": 1, "wuxing": "土"},
    )
    snap = factors.evaluate_factors("male", male, shushi="bazi")
    assert snap["配偶星混杂"] == 1

    female = mock_base_context(
        正官={"count": 1, "wuxing": "金"},
        七杀={"count": 1, "wuxing": "金"},
    )
    snap = factors.evaluate_factors("female", female, shushi="bazi")
    assert snap["配偶星混杂"] == 1

    female_looking_at_wealth = mock_base_context(
        正财={"count": 1, "wuxing": "土"},
        偏财={"count": 1, "wuxing": "土"},
    )
    snap = factors.evaluate_factors("female", female_looking_at_wealth, shushi="bazi")
    assert snap["配偶星混杂"] == 0


def test_factor_definitions_have_no_duplicate_signature() -> None:
    for filename in ("factors.csv", "factors_liunian.csv"):
        grouped: dict[str, list[dict[str, str]]] = defaultdict(list)
        with (TOOLS / "factors" / filename).open(encoding="utf-8", newline="") as f:
            for row in csv.DictReader(f):
                if row.get("因子", "").strip():
                    grouped[row["因子"]].append(row)
        signatures: dict[tuple, str] = {}
        for name, rows in grouped.items():
            if any((row.get("原语直通") or "").strip() for row in rows):
                continue
            variants = tuple(sorted(
                tuple(sorted((key, value.strip()) for key, value in row.items()
                             if key not in META and value.strip()))
                for row in rows
            ))
            assert variants not in signatures, f"{filename}: {name} 与 {signatures[variants]} 重复"
            signatures[variants] = name


def test_operator_arguments_use_constant_closures() -> None:
    const = __import__("json").loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    ten_gods = set(const["十神"])
    classes = set(const["十神大类"])
    roles = set(const["六亲角色"])
    valid_ten = ten_gods | classes | roles
    target_ops = {"流年透", "流年值", "流年合", "流年冲", "流年克", "大运窗口流年", "换运流年"}
    import re

    for filename in ("factors.csv", "factors_liunian.csv"):
        with (TOOLS / "factors" / filename).open(encoding="utf-8", newline="") as f:
            rows = list(csv.DictReader(f))
        for row in rows:
            expressions = []
            direct = (row.get("原语直通") or "").strip()
            if direct:
                expressions.append(direct)
            expressions += [key for key, value in row.items() if key not in META and value.strip() and "[" in key]
            for expression in expressions:
                match = re.match(r"^([^\[]+)\[(.*)\]$", expression)
                if not match:
                    continue
                op, args = match.group(1), match.group(2).split(",")
                if op in {"现", "透", "藏", "得令", "有根", "为用", "为忌"}:
                    assert all(arg in valid_ten for arg in args), f"{expression} 参数不在十神闭集"
                if op == "数量至少" and len(args) > 1:
                    assert all(arg in valid_ten for arg in args[1:]), f"{expression} 数量参数不在十神闭集"
                if op in target_ops:
                    assert args[0] in classes | roles, f"{expression} 流年 target 未显式使用稳定类/角色"


def test_legacy_constant_aliases_do_not_return() -> None:
    const = __import__("json").loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    assert "目标星" not in const
    assert "类" not in const
    assert set(const["十神大类"]) == {"官杀", "印星", "财星", "食伤", "比劫"}
