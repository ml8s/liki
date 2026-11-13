"""断语匹配领域 — 真值表匹配 + 因子契约校验。

输入：断语表 + 因子快照。输出：命中的断语条目列表。
本模块只做匹配和 fail-closed 因子校验，不生成因子。
"""

from collections.abc import Mapping
from typing import TypeAlias

from app.domain.natal.models import AssertionHit, AssertionRule, FactorSnapshot
from app.natal.tools.errors import AssertionRuleError
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.factor_semantics import PredicateState, validate_published_factor_value

__all__ = ["assert_snapshot_factors", "match_rule", "match_table"]

SideTable: TypeAlias = list[Mapping]
RuleTable: TypeAlias = list[AssertionRule]


def _value_matches(cond, actual):
    """单约束匹配：不适用/不可判定不命中；命中/没命中使用四态语义。"""
    if actual in (PredicateState.NOT_APPLICABLE, PredicateState.UNKNOWN):
        return False
    if isinstance(actual, PredicateState):
        return actual == cond
    return actual == cond


def _assert_factor_value(assertion_id: str, factor: str, value) -> None:
    """断语因子必须是稳定 FactorValue；null / boolean / 空串 fail closed。"""
    if isinstance(value, PredicateState):
        return
    try:
        validate_published_factor_value(value, where=f"[{assertion_id}/{factor}]")
    except ValueError as exc:
        raise AssertionRuleError(str(exc)) from exc


def _rule(row: Mapping) -> AssertionRule:
    return AssertionRule.from_row(row)


def _snapshot(snapshot: Mapping | FactorSnapshot) -> FactorSnapshot:
    if isinstance(snapshot, FactorSnapshot):
        return snapshot
    factors = {key: value for key, value in snapshot.items() if key != "context"}
    return FactorSnapshot(factors=factors, context=snapshot.get("context") or {})


_EXPECTED_FACTOR_VALIDATED: set = set()


def _validate_expected(rule: AssertionRule, group_number: int, factor: str, expected) -> None:
    # Assertion IDs are the stable domain identity; do not key caches by object identity.
    ekey = (rule.id, group_number, factor)
    if ekey in _EXPECTED_FACTOR_VALIDATED:
        return
    _assert_factor_value(rule.id, factor, expected)
    _EXPECTED_FACTOR_VALIDATED.add(ekey)


def assert_snapshot_factors(assertion_id: str, table: list[dict], snapshot: dict | FactorSnapshot) -> None:
    """公共 query/yearly 边界：参与断语的因子必须在 snapshot 中存在。

    expected 是表静态数据，首验后跳过；snapshot 值每年变化，每次校验。
    """
    factor_snapshot = _snapshot(snapshot)
    for row in table:
        rule = _rule(row)
        for group_number, group in enumerate(rule.condition_groups, 1):
            for condition in group.conditions:
                if condition.factor not in factor_snapshot.factors:
                    raise AssertionRuleError(f"[{assertion_id}/{group_number}] 断语因子缺失: {condition.factor}")
                _validate_expected(rule, group_number, condition.factor, condition.expected)
                _assert_factor_value(assertion_id, condition.factor, factor_snapshot.factors[condition.factor])


def match_table(table: list | RuleTable, snapshot: dict | FactorSnapshot, exclusive: bool = False) -> list:
    """真值表匹配：因子快照 × 断语条件组 → 命中条目（按表序）。

    同一 condition_group 内的条件为 AND，不同 condition_group 为 OR。
    公共入口 match_rule 会先校验所需因子完整；本低层函数只做机械匹配。
    """
    factor_snapshot = _snapshot(snapshot)
    rules = [row if isinstance(row, AssertionRule) else _rule(row) for row in table]
    hits = []
    for rule in rules:
        matched = tuple(
            (group, factor_snapshot.factors)
            for group_number, group in enumerate(rule.condition_groups, 1)
            if group.matches(factor_snapshot.factors)
        )
        if not matched:
            continue
        hits.append(AssertionHit(rule=rule, matched=matched).to_row())
    return hits[:1] if exclusive and hits else hits


def _validate_rule_factors(rules: RuleTable, factor_snapshot: FactorSnapshot) -> None:
    """公共断语边界的因子契约：参与条件必须存在且为 FactorValue。

    expected 是表静态数据（同一 rule 模型跨年复用），首验后跳过。
    """
    for rule in rules:
        for group_number, group in enumerate(rule.condition_groups, 1):
            for condition in group.conditions:
                if condition.factor not in factor_snapshot.factors:
                    continue
                _validate_expected(rule, group_number, condition.factor, condition.expected)
                _assert_factor_value(rule.id, condition.factor, factor_snapshot.factors[condition.factor])


def match_rule(rule: str, snapshots: dict, *, skip_validation: bool = False) -> dict:
    """加载断语表 + 匹配。排盘上下文只参与匹配，不计入因子数。

    bazi/ziwei 表分别匹配各自快照；common 条件在合并后的双盘快照上匹配。
    skip_validation=True 时跳过 _validate_rule_factors（调用方已通过
    assert_snapshot_factors 完成校验，如 yearly_range 年循环内）。
    """
    from .rule_loader import load_rule_tables

    tables = load_rule_tables(rule)
    side_codes = load_constants()["命理侧"]["断言代码"]
    bz_e, zw_e = (tables[side] for side in side_codes)
    side_labels = load_constants()["命理侧"]["标签"]
    side_snapshots = {
        side_labels[side]: _snapshot({**snapshots[side_labels[side]], **(snapshots.get("context", {}) or {})})
        for side in side_codes
    }
    bazi_rules = [_rule(row) for row in bz_e]
    ziwei_rules = [_rule(row) for row in zw_e]
    if not skip_validation:
        _validate_rule_factors(bazi_rules, side_snapshots[side_labels["bazi"]])
        _validate_rule_factors(ziwei_rules, side_snapshots[side_labels["ziwei"]])
    return {
        side_labels["bazi"]: match_table(bazi_rules, side_snapshots[side_labels["bazi"]]),
        side_labels["ziwei"]: match_table(ziwei_rules, side_snapshots[side_labels["ziwei"]]) if zw_e else [],
    }
