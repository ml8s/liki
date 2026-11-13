"""Natal immutable domain model contracts."""

import pytest
from app.domain.natal.models import Assertion, AssertionSnapshot, Condition, ConditionGroup, Rule


def _rule() -> Rule:
    return Rule(
        id="test_101",
        domain="财运",
        event_type="引动",
        time_layer="流年",
        event="财动",
        conclusion="财动",
        evidence="规则依据",
        classic_evidence="经典依据",
        condition_groups=(ConditionGroup((Condition("财星透", "命中"), Condition("日主旺", "没命中"))),),
    )


def test_condition_group_matches_and_traces_only_declared_factors():
    group = ConditionGroup((Condition("财星透", "命中"), Condition("日主旺", "命中")))
    factors = {"财星透": "命中", "日主旺": "命中", "无关": "命中"}

    assert group.matches(factors)
    assert group.trace(factors, 2) == {
        "condition_group": 2,
        "factors": {
            "财星透": {"expected": "命中", "actual": "命中"},
            "日主旺": {"expected": "命中", "actual": "命中"},
        },
    }


def test_assertion_projects_internal_row_and_trace():
    rule = _rule()
    group = rule.condition_groups[0]
    factors = {"财星透": "命中", "日主旺": "没命中"}
    row = Assertion(rule=rule, matched=((group, factors),)).to_row()

    assert row["id"] == "test_101"
    assert row["约束组"] == [{"财星透": "命中", "日主旺": "没命中"}]
    assert row["trace"] == [
        {
            "condition_group": 1,
            "factors": {
                "财星透": {"expected": "命中", "actual": "命中"},
                "日主旺": {"expected": "没命中", "actual": "没命中"},
            },
        }
    ]


def test_assertion_snapshot_is_read_only():
    factors = {"财星透": "命中"}
    context = {"性别": "male", "时间事实": {"standard_time": "1990-05-20T12:00:00"}}
    snapshot = AssertionSnapshot(factors=factors, context=context)

    factors["财星透"] = "没命中"
    context["性别"] = "female"
    if isinstance(context["时间事实"], dict):
        context["时间事实"]["standard_time"] = "changed"

    assert snapshot.factors == {"财星透": "命中"}
    assert snapshot.context == {"性别": "male", "时间事实": {"standard_time": "1990-05-20T12:00:00"}}
    with pytest.raises(TypeError):
        snapshot.factors["财星透"] = "没命中"


def test_rule_is_immutable():
    rule = _rule()
    with pytest.raises(AttributeError):
        rule.id = "changed"
