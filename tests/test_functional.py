"""规则引擎功能测试：因子、断语、场景过滤与冲突行为。

本文件只验证规则系统是否按当前契约执行；不声称证明每条命理断语正确。
"""
from __future__ import annotations

import json
from unittest import mock
from pathlib import Path

import pytest

import _helpers  # noqa: F401 —— 注入 liki-bazi tools 路径
import duanyu
from errors import FactorEvaluateError
from duanyu import (
    SCENE_ALIASES,
    YEARLY_RULES,
    default_scene_domains,
    load_rule_table,
    load_rule_tables,
    match_table,
    match_rule,
    resolve_rules,
)
from factors import evaluate_factors, evaluate_operator

ROOT = Path(__file__).resolve().parents[1]
FUNCTIONAL = ROOT / "tests" / "functional"
MANIFEST = json.loads((FUNCTIONAL / "manifest.json").read_text(encoding="utf-8"))

DATA_FILES = {
    "factor": FUNCTIONAL / "factors" / "cases.json",
    "assertion": FUNCTIONAL / "assertions" / "cases.json",
    "scenario": FUNCTIONAL / "scenarios" / "cases.json",
    "conflict": FUNCTIONAL / "conflicts" / "cases.json",
}
DATA = {
    level: json.loads(path.read_text(encoding="utf-8"))
    for level, path in DATA_FILES.items()
}
CASES = {
    level: payload["cases"]
    for level, payload in DATA.items()
}
ALL_CASES = [case for cases in CASES.values() for case in cases]
ASSERTION_TABLE_ROWS = {
    case["id"]: [
        load_rule_table(f"{case['side']}_{domain_rule}.csv")
        for domain_rule in case["domain_rules"]
    ]
    for case in CASES["assertion"]
}
CONFLICT_TABLE_ROWS = {
    case["id"]: [
        load_rule_table(f"{case['side']}_{domain_rule}.csv")
        for domain_rule in case["domain_rules"]
    ]
    for case in CASES["conflict"]
}
ASSERTION_ROWS_BY_ID = {
    row["id"]: row
    for tables in ASSERTION_TABLE_ROWS.values()
    for table in tables
    for row in table
}
ALLOWED_KINDS = {"positive", "negative", "boundary"}
MANIFEST_LEVELS = {
    "factor": "factors",
    "assertion": "assertions",
    "scenario": "scenarios",
    "conflict": "conflicts",
}


def _ids(level: str) -> list[str]:
    return [case["id"] for case in CASES[level]]


def _case(level: str, case_id: str) -> dict:
    return next(case for case in CASES[level] if case["id"] == case_id)


def test_functional_manifest_contract() -> None:
    assert MANIFEST["version"] == 1
    assert MANIFEST["name"] == "liki-rule-functional-suite"
    assert set(MANIFEST["sources"]) == {
        "domain_truth_tables", "factors", "assertions", "scenarios", "conflicts"
    }
    for key, path in MANIFEST["sources"].items():
        assert (FUNCTIONAL / path).resolve().exists(), (key, path)

    for level in DATA_FILES:
        assert DATA[level]["version"] == 1, level
        minimum_key = MANIFEST_LEVELS[level]
        assert len(CASES[level]) >= MANIFEST["minimum_counts"][minimum_key], level


def test_functional_case_ids_are_unique() -> None:
    ids = [case["id"] for case in ALL_CASES]
    assert len(ids) == len(set(ids))


def test_functional_cases_have_behavior_rationale() -> None:
    for case in ALL_CASES:
        assert case["id"], case
        assert case["rule"], case["id"]
        assert case["basis"], case["id"]
        assert case["kind"] in ALLOWED_KINDS, (case["id"], case["kind"])


def test_factor_cases_target_low_level_behavior_only() -> None:
    allowed_modes = {"factor_snapshot", "operator"}
    for case in CASES["factor"]:
        assert case["mode"] in allowed_modes, case["id"]
        if case["mode"] == "factor_snapshot":
            assert case.get("expect_factors"), case["id"]
        else:
            assert case.get("expect") is not None, case["id"]

def _matcher_row(rule_id: str, *groups: dict) -> dict:
    return {"id": rule_id, "约束组": list(groups)}


def test_match_same_group_requires_all_conditions() -> None:
    table = [_matcher_row("both", {"甲": 1, "乙": 1})]
    assert [row["id"] for row in match_table(table, {"甲": 1, "乙": 1})] == ["both"]
    assert [row["id"] for row in match_table(table, {"甲": 1, "乙": 0})] == []
    assert [row["id"] for row in match_table(table, {"甲": 1})] == []


def test_match_different_groups_are_or() -> None:
    table = [
        _matcher_row("first", {"甲": 1}),
        _matcher_row("second", {"乙": 1}),
    ]
    assert [row["id"] for row in match_table(table, {"甲": 1})] == ["first"]
    assert [row["id"] for row in match_table(table, {"乙": 1})] == ["second"]
    assert [row["id"] for row in match_table(table, {})] == []


def test_match_supports_negative_and_string_constraints() -> None:
    table = [
        _matcher_row("not_a", {"甲": 0}),
        _matcher_row("day_is_jia", {"日主": "甲"}),
    ]
    assert [row["id"] for row in match_table(table, {"甲": 0})] == ["not_a"]
    assert [row["id"] for row in match_table(table, {"甲": 1})] == []
    assert [row["id"] for row in match_table(table, {"日主": "甲"})] == ["day_is_jia"]
    assert [row["id"] for row in match_table(table, {"日主": "乙"})] == []


def test_match_exclusive_returns_first_hit_only() -> None:
    table = [
        _matcher_row("first", {"甲": 1}),
        _matcher_row("second", {"甲": 1, "乙": 1}),
    ]
    snapshot = {"甲": 1, "乙": 1}
    assert [row["id"] for row in match_table(table, snapshot)] == ["first", "second"]
    assert [row["id"] for row in match_table(table, snapshot, exclusive=True)] == ["first"]


def test_match_result_contains_condition_trace() -> None:
    table = [
        _matcher_row("either", {"甲": 1}, {"乙": 1}),
    ]
    hits = match_table(table, {"甲": 1, "乙": 1})

    assert hits[0]["trace"] == [
        {
            "condition_group": 1,
            "factors": {
                "甲": {"expected": 1, "actual": 1},
            },
        },
        {
            "condition_group": 2,
            "factors": {
                "乙": {"expected": 1, "actual": 1},
            },
        },
    ]


def test_match_table_accepts_empty_table() -> None:
    assert match_table([], {"任意因子": 1}) == []


def test_evaluate_operator_rejects_unknown_name() -> None:
    with pytest.raises(FactorEvaluateError, match="未知算子"):
        evaluate_operator("不存在", [], "male", {})


def test_duanyu_public_api_has_no_private_symbols() -> None:
    assert duanyu.__all__
    assert all(not name.startswith("_") for name in duanyu.__all__)
    assert all(hasattr(duanyu, name) for name in duanyu.__all__)


def test_query_domain_filter_keeps_only_requested_life_domain() -> None:
    pan = {"gender": "male"}
    matched = {
        "八字": [
            {"id": "hun_101", "领域": "婚姻"},
            {"id": "cai_110", "领域": "财运"},
        ],
        "紫微": [{"id": "zs_301", "领域": "事业"}],
        "合参": [],
    }

    with mock.patch.object(duanyu, "validate_natal_pan"), \
         mock.patch.object(
             duanyu,
             "evaluate_snap_from_pan",
             return_value={"八字": {}, "紫微": {}, "context": {}},
         ), \
         mock.patch.object(duanyu, "match_rule", return_value=matched):
        result = duanyu.query("大运", pan, year=2020, domains=["财运"])

    assert [row["id"] for row in result["八字"]] == ["cai_110"]
    assert result["紫微"] == []
    assert result["合参"] == []


def test_query_domain_filter_rejects_unknown_life_domain() -> None:
    pan = {"gender": "male"}
    with mock.patch.object(duanyu, "validate_natal_pan"), \
         mock.patch.object(duanyu, "evaluate_snap_from_pan"), \
         mock.patch.object(duanyu, "match_rule") as match_rule:
        with pytest.raises(ValueError, match="domains 含无效领域"):
            duanyu.query("大运", pan, year=2020, domains=["不存在"])

    match_rule.assert_not_called()


def test_filter_domains_keeps_error_payload_unchanged() -> None:
    result = {"error": "rpc failed"}
    assert duanyu.filter_domains(result, ["学业"]) == result


def test_filter_domains_omits_snapshot_evidence() -> None:
    result = {
        "八字": [{"id": "yx_101", "领域": "学业"}],
        "紫微": [],
        "合参": [],
        "evidence": {"三刑流年": {"group": "寅巳申"}},
    }
    filtered = duanyu.filter_domains(result, ["学业"])
    assert [row["id"] for row in filtered["八字"]] == ["yx_101"]
    assert "evidence" not in filtered


def test_query_keeps_both_sides_alive_for_common_assertions() -> None:
    """单侧域存在 common 条件时，query 仍必须提供双盘快照。

    cc_106 同时消费八字「身弱」与紫微「疾厄宫化忌」；若按八字专属裁掉
    紫微快照，这类跨术数规则会变成永远不可命中的死规则。
    """
    snapshots = {
        "八字": {"身强弱": "身弱"},
        "紫微": {"疾厄宫化忌": 1},
        "context": {},
    }

    with mock.patch.object(duanyu, "validate_natal_pan"), \
         mock.patch.object(
             duanyu, "evaluate_snap_from_pan", return_value=snapshots
         ) as evaluate:
        result = duanyu.query("十神", {"gender": "male"})

    assert evaluate.call_args.kwargs["sides"] == {"bazi", "ziwei"}
    assert any(row["id"] == "cc_106" for row in result["合参"])


def test_all_single_side_rules_with_common_request_dual_snapshots() -> None:
    """覆盖当前所有单侧域 + common 组合，防止后续新增时回退。"""
    single_side_rules = [
        rule for rule in (
            duanyu.BAZI_ONLY_RULES | duanyu.ZIWEI_ONLY_RULES
        )
        if rule in duanyu.NATAL_RULES
        if load_rule_tables(rule)["common"]
    ]
    assert single_side_rules

    for rule in single_side_rules:
        snapshots = {"八字": {}, "紫微": {}, "context": {}}
        with mock.patch.object(duanyu, "validate_natal_pan"), \
             mock.patch.object(
                 duanyu,
                 "evaluate_snap_from_pan",
                 return_value=snapshots,
             ) as evaluate, \
             mock.patch.object(
                 duanyu,
                 "match_rule",
                 return_value={"八字": [], "紫微": [], "合参": []},
             ):
            duanyu.query(rule, {"gender": "male"})

        assert evaluate.call_args.kwargs["sides"] == {"bazi", "ziwei"}, rule


def test_load_rule_tables_respects_side_scope() -> None:
    bazi_only = load_rule_tables("十神")
    assert bazi_only["bazi"]
    assert bazi_only["ziwei"] == []

    ziwei_only = load_rule_tables("命宫")
    assert ziwei_only["bazi"] == []
    assert ziwei_only["ziwei"]
    assert ziwei_only["common"] == []

    yearly = load_rule_tables("年神煞")
    assert yearly["bazi"]
    assert yearly["ziwei"] == []
    assert yearly["common"]


def test_match_rule_merges_bazi_and_ziwei_for_common_rule() -> None:
    rule = "年神煞"
    snapshots = {
        "八字": {"流年神煞天乙贵人": 1},
        "紫微": {"流年迁移宫禄": 1},
        "context": {"性别": "male"},
    }
    snapshot_copy = {
        side: dict(values)
        for side, values in snapshots.items()
    }

    result = match_rule(rule, snapshots)

    assert "ycai_120" in [row["id"] for row in result["合参"]]
    assert snapshots == snapshot_copy


def test_assertion_cases_define_positive_and_forbidden_space() -> None:
    for case in CASES["assertion"]:
        assert case["domain_rules"], case["id"]
        assert len(case["domain_rules"]) == len(set(case["domain_rules"])), case["id"]
        expected = case.get("expect_assertions")
        forbidden = case.get("forbidden_assertions")
        assert isinstance(expected, list), case["id"]
        assert isinstance(forbidden, list), case["id"]
        assert expected or forbidden, case["id"]
        assert not set(expected) & set(forbidden), case["id"]


def test_scenario_cases_define_rule_and_domain_boundary() -> None:
    for case in CASES["scenario"]:
        assert isinstance(case["expect_rules"], list), case["id"]
        assert isinstance(case["forbidden_rules"], list), case["id"]
        assert isinstance(case["expect_domains"], list), case["id"]
        assert isinstance(case["forbidden_domains"], list), case["id"]
        assert isinstance(case["expect_filter"], bool), case["id"]
        required_assertions = case.get("required_assertions", [])
        assert isinstance(required_assertions, list), case["id"]
        assert all(isinstance(item, str) and item for item in required_assertions), case["id"]
        assert set(case["expect_rules"]).isdisjoint(case["forbidden_rules"]), case["id"]
        assert set(case["expect_domains"]).isdisjoint(case["forbidden_domains"]), case["id"]


def test_conflict_cases_require_explicit_resolution() -> None:
    for case in CASES["conflict"]:
        assert case["domain_rules"], case["id"]
        assert case["expect_assertions"], case["id"]
        assert case["forbidden_assertions"], case["id"]
        assert set(case["expect_assertions"]).isdisjoint(case["forbidden_assertions"])


def test_assertion_and_conflict_expectations_resolve_to_declared_tables() -> None:
    tables = {
        "assertion": ASSERTION_TABLE_ROWS,
        "conflict": CONFLICT_TABLE_ROWS,
    }
    for level, table_rows in tables.items():
        for case in CASES[level]:
            ids = {
                row["id"]
                for table in table_rows[case["id"]]
                for row in table
            }
            expected = case["expect_assertions"]
            forbidden = case["forbidden_assertions"]
            assert set(expected) <= ids, (case["id"], sorted(set(expected) - ids))
            assert set(forbidden) <= ids, (case["id"], sorted(set(forbidden) - ids))


def test_functional_covers_required_factor_categories() -> None:
    covered = {case.get("category") for case in CASES["factor"]}
    missing = set(MANIFEST["coverage"]["factor_categories"]) - covered
    assert not missing, f"factor functional cases 缺少类别: {sorted(missing)}"


def test_functional_covers_required_assertion_domains_and_layers() -> None:
    covered_domains = {
        ASSERTION_ROWS_BY_ID[assertion_id]["领域"]
        for case in CASES["assertion"]
        for assertion_id in case["expect_assertions"]
    }
    missing_domains = set(MANIFEST["coverage"]["assertion_domains"]) - covered_domains
    assert not missing_domains, f"assertion functional cases 缺少领域: {sorted(missing_domains)}"

    covered_layers = {
        "yearly" if any(rule.startswith("年") for rule in case["domain_rules"]) else "natal"
        for case in CASES["assertion"]
    }
    missing_layers = set(MANIFEST["coverage"]["assertion_layers"]) - covered_layers
    assert not missing_layers, f"assertion functional cases 缺少时间层: {sorted(missing_layers)}"


def test_functional_covers_required_assertion_sides_and_rules() -> None:
    covered_sides = {case["side"] for case in CASES["assertion"]}
    missing_sides = set(MANIFEST["coverage"]["assertion_sides"]) - covered_sides
    assert not missing_sides, f"assertion functional cases 缺少命理侧: {sorted(missing_sides)}"

    covered_rules = {
        domain_rule
        for case in CASES["assertion"]
        for domain_rule in case["domain_rules"]
    }
    missing_rules = set(MANIFEST["coverage"]["assertion_rules"]) - covered_rules
    assert not missing_rules, f"assertion functional cases 缺少规则域: {sorted(missing_rules)}"


def test_functional_covers_every_scenario_alias() -> None:
    covered = {
        rule
        for case in CASES["scenario"]
        for rule in case["input"]["rules"]
    }
    missing = set(SCENE_ALIASES) - covered
    assert not missing, f"scenario functional cases 缺少场景别名: {sorted(missing)}"


@pytest.mark.parametrize("case_id", _ids("factor"))
def test_functional_factor(case_id: str) -> None:
    case = _case("factor", case_id)
    input_data = case["input"]
    chart = {**input_data.get("chart", {}), **input_data.get("base", {})}
    input_data = {**input_data, "chart": chart}

    if case["mode"] == "factor_snapshot":
        actual = evaluate_factors(
            input_data["gender"],
            input_data["chart"],
            shushi=case.get("shushi"),
            current_year=input_data.get("current_year", 0),
        )
        for name, expected in case["expect_factors"].items():
            if name == "性别" and expected == 0:
                assert name not in actual, f"{case_id}: 性别不得进入因子快照"
            else:
                assert name in actual, f"{case_id}: 缺因子 {name}"
                assert actual[name] == expected, (
                    f"{case_id}: {name}={actual[name]!r}, want {expected!r}; {case['basis']}"
                )
        return

    op = input_data["op"]
    args = input_data.get("args", [])
    actual = evaluate_operator(
        op,
        args,
        input_data["gender"],
        input_data["chart"],
        ctx=input_data.get("ctx", {}),
        current_year=input_data.get("current_year", 0),
    )

    assert actual == case["expect"], (
        f"{case_id}: {op}={actual!r}, want {case['expect']!r}; {case['basis']}"
    )


@pytest.mark.parametrize("case_id", _ids("assertion"))
def test_functional_assertion(case_id: str) -> None:
    case = _case("assertion", case_id)
    hits = [
        row
        for table in ASSERTION_TABLE_ROWS[case_id]
        for row in match_table(table, case["input"]["snapshot"])
    ]
    actual = [row["id"] for row in hits]
    expected = case["expect_assertions"]
    forbidden = case["forbidden_assertions"]

    missing = [item for item in expected if item not in actual]
    leaked = [item for item in forbidden if item in actual]
    assert not missing, (
        f"{case_id}: 缺少断语 {missing}; actual={actual}; {case['basis']}"
    )
    assert not leaked, (
        f"{case_id}: 禁止断语出现 {leaked}; actual={actual}; {case['basis']}"
    )


@pytest.mark.parametrize("case_id", _ids("scenario"))
def test_functional_scenario(case_id: str) -> None:
    case = _case("scenario", case_id)
    rules = case["input"]["rules"]
    actual_rules = resolve_rules(rules, YEARLY_RULES, SCENE_ALIASES)
    actual_domains = default_scene_domains(rules)

    assert set(actual_rules) == set(case["expect_rules"]), (
        f"{case_id}: 规则展开={actual_rules}, "
        f"want={case['expect_rules']}, "
        f"forbidden={case['forbidden_rules']}; {case['basis']}"
    )
    assert not set(actual_rules) & set(case["forbidden_rules"]), (
        f"{case_id}: 禁止规则出现; actual={actual_rules}; {case['basis']}"
    )

    if case["expect_filter"]:
        assert actual_domains is not None, (
            f"{case_id}: 专题场景缺少默认领域过滤; {case['basis']}"
        )
        assert set(actual_domains or []) == set(case["expect_domains"]), (
            f"{case_id}: 领域过滤={actual_domains}, want={case['expect_domains']}; "
            f"{case['basis']}"
        )
        assert not set(actual_domains or []) & set(case["forbidden_domains"]), (
            f"{case_id}: 禁止领域出现; actual={actual_domains}; {case['basis']}"
        )
    else:
        assert actual_domains is None, (
            f"{case_id}: 总览场景不应默认收窄领域; actual={actual_domains}; {case['basis']}"
        )

    if case.get("required_assertions"):
        available: dict[str, dict] = {}
        for rule in actual_rules:
            for table in load_rule_tables(rule).values():
                for row in table:
                    available[row["id"]] = row

        missing = [
            assertion_id
            for assertion_id in case["required_assertions"]
            if assertion_id not in available
            or available[assertion_id]["领域"] not in case["expect_domains"]
        ]
        assert not missing, (
            f"{case_id}: 场景缺少专属断语 {missing}; "
            f"domains={case['expect_domains']}; {case['basis']}"
        )


@pytest.mark.parametrize("case_id", _ids("conflict"))
def test_functional_conflict(case_id: str) -> None:
    case = _case("conflict", case_id)
    hits = [
        row
        for table in CONFLICT_TABLE_ROWS[case_id]
        for row in match_table(table, case["input"]["snapshot"])
    ]
    actual = [row["id"] for row in hits]

    missing = [item for item in case["expect_assertions"] if item not in actual]
    leaked = [item for item in case["forbidden_assertions"] if item in actual]
    assert not missing, (
        f"{case_id}: 冲突裁决缺少断语 {missing}; actual={actual}; {case['basis']}"
    )
    assert not leaked, (
        f"{case_id}: 冲突裁决禁止断语出现 {leaked}; actual={actual}; {case['basis']}"
    )
