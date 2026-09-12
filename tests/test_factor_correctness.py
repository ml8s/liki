"""因子层契约：引擎字段、关系算子、性别解析与表定义唯一性。"""
from __future__ import annotations

import csv
import json
import re
from collections import defaultdict
from pathlib import Path

from _helpers import mock_base_context
import factors
from operators_liunian import _LIU_OP_NAMES
from operators_natal import _OP_NAMES, _op, _ten_god_states_from_pan

TOOLS = Path(__file__).resolve().parents[1] / "skills" / "liki-bazi" / "tools"


def test_ten_god_state_projection_is_readonly() -> None:
    pan = {"full": {
        "ten_god_states": [
            {"shi_shen": "正官", "wuxing": "金", "transparent": True, "hidden": False, "count": 1},
            {"shi_shen": "正印", "wuxing": "水", "transparent": False, "hidden": True, "count": 1},
        ]
    }}
    states = _ten_god_states_from_pan(pan)
    assert states["正官"]["transparent"] is True
    assert states["正官"]["hidden"] is False
    assert states["正官"]["count"] == 1
    assert states["正印"]["transparent"] is False
    assert states["正印"]["hidden"] is True
    assert states["正印"]["count"] == 1


def test_ge_shen_uses_engine_gan_source() -> None:
    base = mock_base_context()
    base["yongshen"] = {"ge_ju": {"ge_ju": "正官格"}}
    chart = {"full": {"atomic_facts": {"pattern_god_transparent": True}, "yue": {"shi_shens": [
        {"source": "gan", "gan": "辛", "shi_shen": "正官"}
    ]}}}
    assert _op("格神透", [], "male", {**base, **chart}) == 1


def test_relation_operator_reads_engine_results() -> None:
    chart = {
        "full": {"relation_groups": [
            {"field": "gan_he", "group": "甲己"},
            {"field": "zhi_liu_he", "group": "子丑"},
            {"field": "san_he", "group": "申子辰"},
            {"field": "san_hui", "group": "寅卯辰"},
            {"field": "liu_chong", "group": "子午"},
            {"field": "liu_hai", "group": "子未"},
            {"field": "liu_xing", "group": "寅巳申"},
        ]},
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
        grouped: dict[tuple[str, str, str], list[tuple[str, str, str]]] = defaultdict(list)
        with (TOOLS / "factors" / filename).open(encoding="utf-8", newline="") as f:
            for row in csv.DictReader(f):
                key = (row["factor_id"], row["shushi"], row["group_id"])
                grouped[key].append((row["kind"], row["expression"], row["expected"]))

        definitions: dict[tuple, list[str]] = defaultdict(list)
        for (name, side, _group_id), terms in grouped.items():
            signature = (side, tuple(sorted(terms)))
            definitions[signature].append(name)

        duplicates = {
            signature: names
            for signature, names in definitions.items()
            if len(names) > 1
        }
        assert not duplicates, f"{filename}: 因子定义重复: {duplicates}"


def test_operator_arguments_use_constant_closures() -> None:
    const = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    ten_gods = set(const["十神"])
    classes = set(const["十神大类"])
    roles = set(const["六亲角色"])
    valid_ten = ten_gods | classes | roles
    valid_elements = set(const["五行"])
    valid_flow_targets = classes | roles | set(const["干支来源"])
    valid_palaces = set(const["紫微宫位"]) | {"任意"}
    valid_palace_stars = (
        set(const["紫微主星"]) | set(const["紫微煞星"])
        | set(const["紫微六吉星"]) | set(const["紫微文星"])
        | set(const["紫微辅星"])
        | {"任意", "紫微主星", "紫微六吉星", "紫微文星", "煞星"}
        | set(const["紫微星曜特殊值"])
    )
    valid_palace_conditions = (
        {"任意", "禄", "权", "科", "忌", "庙旺", "落陷"}
        | set(const["紫微宫位特殊条件"])
    )
    valid_flow_stars = set(const["紫微流曜"])
    target_ops = {"流年透", "流年值", "流年合", "流年冲", "流年克", "大运窗口流年", "换运流年"}

    ten_arg_ops = {
        "现", "透", "藏", "得令", "有根", "为用", "为忌",
        "克", "克者旺", "禄根",
    }
    element_arg_ops = {"旺", "弱", "缺", "流年支受克"}

    for filename in ("factors.csv", "factors_liunian.csv"):
        with (TOOLS / "factors" / filename).open(
            encoding="utf-8-sig", newline=""
        ) as f:
            rows = list(csv.DictReader(f))
        for row in rows:
            expression = (row.get("expression") or "").strip()
            match = re.match(r"^([^\[]+)\[(.*)\]$", expression)
            if not match:
                continue
            op, args = match.group(1), match.group(2).split(",")
            if op in ten_arg_ops:
                assert all(
                    arg in valid_ten or arg in valid_elements for arg in args
                ), f"{expression} 参数不在十神/五行闭集"
            if op in element_arg_ops:
                assert all(
                    arg in valid_elements for arg in args
                ), f"{expression} 参数不在五行闭集"
            if op == "数量至少" and len(args) > 1:
                assert all(
                    arg in valid_ten for arg in args[1:]
                ), f"{expression} 数量参数不在十神闭集"
            if op == "五行数量至少" and len(args) > 1:
                assert args[1] in valid_elements, f"{expression} 五行参数不在闭集"
            if op == "宫含":
                assert args[0] in valid_palaces, f"{expression} 宫位不在紫微宫位闭集"
                assert args[1] in valid_palace_stars, f"{expression} 星曜不在紫微星曜闭集"
                if len(args) > 2:
                    assert args[2] in valid_palace_conditions, (
                        f"{expression} 宫位条件不在闭集"
                    )
            if op == "流曜入宫":
                assert args[0] in valid_flow_stars, f"{expression} 流曜不在闭集"
                assert args[1] in valid_palaces, f"{expression} 宫位不在紫微宫位闭集"
            if op == "流年宫化":
                assert args[0] in valid_palaces, f"{expression} 宫位不在紫微宫位闭集"
            if op in target_ops:
                assert args[0] in valid_flow_targets, (
                    f"{expression} 流年 target 未显式使用稳定类/角色"
                )


def test_relation_operator_groups_use_constant_closures() -> None:
    const = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    relation_sources = {
        "gan_he": "天干五合",
        "zhi_liu_he": "六合",
        "san_he": "三合",
        "san_hui": "三会",
        "liu_chong": "六冲",
        "liu_hai": "六害",
        "liu_xing": "三刑",
    }
    valid_groups = {field: set() for field in relation_sources}
    for field, source in relation_sources.items():
        for left, right in const[source].items():
            members = [left, right] if isinstance(right, str) else [left, *right]
            valid_groups[field].add(tuple(sorted(members)))

    with (TOOLS / "factors/factors.csv").open(
        encoding="utf-8-sig", newline=""
    ) as source:
        rows = list(csv.DictReader(source))
    checked = 0
    seen_groups = {field: set() for field in relation_sources}
    for row in rows:
        match = re.match(r"^关系\[([^,\]]+),([^\]]+)\]$", row["expression"])
        if not match:
            continue
        field, group = match.groups()
        members = tuple(sorted(group)) if len(set(group)) > 1 else (group[0], group[0])
        assert field in valid_groups, f"{row['expression']} 关系 field 不在闭集"
        assert members in valid_groups[field], (
            f"{row['expression']} 关系 group 不在 constants 闭集"
        )
        seen_groups[field].add(members)
        checked += 1
    assert checked > 30
    assert seen_groups == valid_groups


def test_constant_contract_has_only_stable_ten_god_classes() -> None:
    const = __import__("json").loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    assert "目标星" not in const
    assert "类" not in const
    assert set(const["十神大类"]) == {"官杀", "印星", "财星", "食伤", "比劫"}


def test_ziwei_palace_closure_matches_engine_labels() -> None:
    const = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    assert const["紫微宫位"] == [
        "命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
        "迁移", "仆役", "官禄", "田宅", "福德", "父母",
    ]
    assert set(const["紫微流曜"]) == {
        "流魁", "流钺", "流昌", "流曲", "流禄",
        "流羊", "流陀", "流马", "流鸾", "流喜",
    }


def test_factor_basis_is_present_for_domain_review():
    from pathlib import Path

    root = Path(__file__).resolve().parents[1] / "skills/liki-bazi/tools/factors"
    for name in ("factors.csv", "factors_liunian.csv"):
        with (root / name).open(encoding="utf-8-sig", newline="") as source:
            rows = list(csv.DictReader(source))
        assert rows
        for row in rows:
            assert row["basis"].strip(), f"{name}:{row['factor_id']} 缺少依据"


def test_ten_god_target_arguments_use_closed_vocabulary():
    constants = json.loads((TOOLS / "constants.json").read_text(encoding="utf-8"))
    allowed = (
        set(constants["十神"])
        | set(constants["十神大类"])
        | set(constants["六亲角色"])
        | set(constants["五行"])
        | {"印", "财", "官", "杀", "食", "伤", "比", "劫"}
    )
    target_ops = {"现", "透", "藏", "得令", "有根", "禄根"}

    for name in ("factors.csv", "factors_liunian.csv"):
        with (TOOLS / "factors" / name).open(encoding="utf-8-sig", newline="") as source:
            rows = list(csv.DictReader(source))
        for row in rows:
            match = re.match(r"^([^\[]+)\[(.*)\]$", row["expression"])
            if not match or match.group(1) not in target_ops:
                continue
            for argument in match.group(2).split(","):
                if argument and argument not in allowed:
                    raise AssertionError(
                        f"{name}:{row['factor_id']} 使用未登记十神目标 {argument!r}"
                    )
