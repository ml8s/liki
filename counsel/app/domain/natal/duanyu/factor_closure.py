"""Reverse factor dependency closure for assertion tables."""

from app.natal.tools.factor_tables import load_factor_rows, load_liunian_rows

_NATIVE_REF = "引用本命["


def factor_closure(groups: dict[str, list[dict]], seeds: set[str]) -> set[str]:
    """从种子因子沿 factor_ref 求闭包。"""
    required = set(seeds)
    queue = list(required)
    while queue:
        name = queue.pop()
        for row in groups.get(name, ()):
            for condition in row.get("conds", {}):
                if "[" not in condition and condition in groups and condition not in required:
                    required.add(condition)
                    queue.append(condition)
    return required


def _seed_names(tables: list[list[dict]], groups: dict[str, list[dict]]) -> set[str]:
    seeds: set[str] = set()
    for table in tables:
        for row in table:
            for conditions in row.get("约束组", []):
                for factor in conditions:
                    if factor in groups:
                        seeds.add(factor)
    return seeds


def _grouped_rows(rows: list[dict]) -> dict[str, list[dict]]:
    groups: dict[str, list[dict]] = {}
    for row in rows:
        groups.setdefault(row["因子"], []).append(row)
    return groups


def required_natal_factors(tables: list[list[dict]]) -> set[str]:
    """从断言条件反向求本命因子闭包；query 只计算实际消费的因子。"""
    groups = _grouped_rows(load_factor_rows())
    return factor_closure(groups, _seed_names(tables, groups))


def required_flow_factors(tables: list[list[dict]]) -> set[str]:
    """从流年断言条件反向求因子闭包；yearly_range 只算实际消费因子。"""
    groups = _grouped_rows(load_liunian_rows())
    return factor_closure(groups, _seed_names(tables, groups))


def natal_factors_for_flow(flow_factors: set[str]) -> set[str]:
    """返回流年因子显式或隐式消费的本命八字因子闭包。"""
    flow_groups = _grouped_rows(load_liunian_rows())
    groups = _grouped_rows(load_factor_rows())
    seeds: set[str] = set()
    for name in factor_closure(flow_groups, flow_factors):
        for row in flow_groups.get(name, ()):
            for condition in row.get("conds", {}):
                if condition.startswith(_NATIVE_REF) and condition.endswith("]"):
                    seeds.add(condition[len(_NATIVE_REF) : -1])
    return factor_closure(groups, seeds)


def flow_factor_names(rules: list[str]) -> set[str]:
    """展开流年场景别名并返回断言消费的流年因子闭包。"""
    from .rule_loader import load_rule_tables, resolve_rule_selection

    resolved_rules = resolve_rule_selection(rules)
    tables = [table for rule in resolved_rules for table in load_rule_tables(rule).values()]
    return required_flow_factors(tables)
