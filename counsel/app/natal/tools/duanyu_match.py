"""断语匹配层 — 真值表匹配 + 因子契约校验。

输入：断语表 + 因子快照。
输出：命中的断语条目列表。
本文件不加载表、不生成因子；表由 duanyu.py 加载后传入。
"""

from app.natal.tools.errors import AssertionRuleError
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.factor_semantics import PredicateState, validate_published_factor_value

__all__ = [
    "match_table", "match_rule",
    "assert_snapshot_factors",
]


def _value_matches(cond, actual):
    """单约束匹配：不适用/不可判定不命中；命中/没命中使用四态语义。"""
    if actual in (PredicateState.NOT_APPLICABLE, PredicateState.UNKNOWN):
        return False
    if isinstance(actual, PredicateState):
        return actual == cond
    return actual == cond


def _assert_factor_value(assertion_id: str, factor: str, value) -> None:
    """断语因子必须是稳定 FactorValue；null / boolean / 空串 fail closed。

    PredicateState 是最常见值，inline isinstance fast path 消除函数调用开销。
    """
    if isinstance(value, PredicateState):
        return
    try:
        validate_published_factor_value(value, where=f"[{assertion_id}/{factor}]")
    except ValueError as exc:
        raise AssertionRuleError(str(exc)) from exc


_EXPECTED_FACTOR_VALIDATED: set = set()


def assert_snapshot_factors(
    assertion_id: str, table: list[dict], snapshot: dict,
) -> None:
    """公共 query/yearly 边界：参与断语的因子必须在 snapshot 中存在。

    expected 是表静态数据，首验后通过 _EXPECTED_FACTOR_VALIDATED 跳过；
    snapshot 值每年变化，每次校验。
    """
    for item in table:
        for group_number, conditions in enumerate(item.get("约束组") or [], 1):
            for factor, expected in conditions.items():
                if factor not in snapshot:
                    raise AssertionRuleError(
                        f"[{assertion_id}/{group_number}] 断语因子缺失: {factor}"
                    )
                ekey = (id(item), group_number, factor)
                if ekey not in _EXPECTED_FACTOR_VALIDATED:
                    _assert_factor_value(assertion_id, factor, expected)
                    _EXPECTED_FACTOR_VALIDATED.add(ekey)
                _assert_factor_value(assertion_id, factor, snapshot[factor])


def match_table(table: list, snapshot: dict, exclusive: bool = False) -> list:
    """真值表匹配：因子快照 × 断语条件组 → 命中条目（按表序）。

    同一 condition_group 内的条件为 AND，不同 condition_group 为 OR。
    公共入口 match_rule 会先校验所需因子完整；本低层函数只做机械匹配。
    """
    hits = []
    for item in table:
        condition_groups = item.get("约束组") or []
        traces = []
        for group_number, conditions in enumerate(condition_groups, 1):
            if not all(
                _value_matches(value, snapshot.get(key))
                for key, value in conditions.items()
            ):
                continue
            traces.append({
                "condition_group": group_number,
                "factors": {
                    key: {"expected": value, "actual": snapshot.get(key)}
                    for key, value in conditions.items()
                },
            })
        if not traces:
            continue
        hits.append({**item, "trace": traces})
    return [hits[0]] if exclusive and hits else hits


def _validate_rule_factors(
    assertion_id: str, table: list[dict], snapshot: dict
) -> None:
    """公共断语边界的因子契约：参与条件必须存在且为 FactorValue。

    expected 是表静态数据（同一 table 对象跨年复用），首验后跳过；
    snapshot 值每年变化，每次校验。
    """
    for item in table:
        for group_number, conditions in enumerate(item.get("约束组") or [], 1):
            for factor, expected in conditions.items():
                if factor not in snapshot:
                    continue
                ekey = (id(item), group_number, factor)
                if ekey not in _EXPECTED_FACTOR_VALIDATED:
                    _assert_factor_value(assertion_id, factor, expected)
                    _EXPECTED_FACTOR_VALIDATED.add(ekey)
                _assert_factor_value(assertion_id, factor, snapshot[factor])


def match_rule(rule: str, snapshots: dict, *, skip_validation: bool = False) -> dict:
    """加载断语表 + 匹配。排盘上下文只参与匹配，不计入因子数。

    bazi/ziwei 表分别匹配各自快照；common 表在合并后的双盘快照上匹配，
    专用于显式八紫合参断语，不伪装成单侧事实。
    skip_validation=True 时跳过 _validate_rule_factors（调用方已通过
    assert_snapshot_factors 完成校验，如 yearly_range 年循环内）。
    """
    from app.natal.tools.duanyu import load_rule_tables
    tables = load_rule_tables(rule)
    side_codes = load_constants()["命理侧"]["断言代码"]
    bz_e, zw_e = (tables[side] for side in side_codes)
    context = snapshots.get("context", {}) or {}
    side_labels = load_constants()["命理侧"]["标签"]
    side_snapshots = {
        side_labels["bazi"]: {**snapshots[side_labels["bazi"]], **context},
        side_labels["ziwei"]: {**snapshots[side_labels["ziwei"]], **context},
    }
    if not skip_validation:
        for side_code, table in zip(side_codes, (bz_e, zw_e)):
            for item in table:
                _validate_rule_factors(
                    item["id"], [item], side_snapshots[side_labels[side_code]]
                )
    return {
        side_labels["bazi"]: match_table(
            bz_e, side_snapshots[side_labels["bazi"]]
        ),
        side_labels["ziwei"]: match_table(
            zw_e, side_snapshots[side_labels["ziwei"]]
        ) if zw_e else [],
    }
