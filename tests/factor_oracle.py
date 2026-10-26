"""因子真值表 oracle 解释器 — 独立测试基础设施。

本模块是编译求值器（factors._evaluate_compiled）的等价验证 oracle。
放在 tests/ 目录以保证与生产代码隔离：
- 生产代码变更不会意外修改 oracle
- 共享语义原语（predicate/all_of/any_of）来自 factor_semantics 契约层
- oracle 自身不做 regex 缓存、不做拓扑排序、不做编译 — 与实现策略正交
"""

import re

from app.natal.tools.factor_semantics import (
    FactorValue,
    PredicateState,
    all_of,
    any_of,
    factor_value,
    predicate,
)
from app.natal.tools.factor_tokens import FACTOR_WILDCARD

__all__ = ["oracle_atomic", "oracle_evaluate_truth_table"]


def oracle_atomic(col: str, gender, chart, ctx: dict | None = None,
                  current_year: int = 0,
                  evaluate_operator=None) -> FactorValue:
    """独立原子执行：每次调用重新 regex 解析，不使用编译缓存。"""
    from app.natal.tools.factors import evaluate_operator as _eo
    eo = evaluate_operator or _eo
    atoms = re.findall(r"([^\[\],]+)\[([^\]]*)\]", col)
    if not atoms:
        return eo(col, [], gender, chart, ctx=ctx, current_year=current_year)

    states: list[FactorValue] = []
    for op, argstr in atoms:
        raw_args = argstr.split(",") if argstr else []
        args = [int(a) if a.lstrip("-").isdigit() else a for a in raw_args]
        value = eo(op, args, gender, chart, ctx=ctx, current_year=current_year)
        if isinstance(value, PredicateState):
            states.append(value)
        elif args and args[-1] == FACTOR_WILDCARD:
            states.append(value)
        else:
            expected = factor_value(args[-1], where=f"原子 {col} 期望值") if args else PredicateState.HIT
            states.append(predicate(value, expected))

    if len(states) == 1:
        return states[0]
    return all_of(states)


def oracle_evaluate_truth_table(rows: list, atomic) -> dict:
    """独立真值表解释器：多轮 fixed-point，不做编译。

    与 _evaluate_compiled 的差异（测试验证目标）：
    - 逐 pass 遍历 vs 单 pass 拓扑序
    - regex 逐次解析 vs 预解析 AtomicPlan
    - 分组逐次构建 vs 预编译 FactorPlan
    """
    facts = {}
    grouped: dict[str, list[dict]] = {}
    for row in rows:
        grouped.setdefault(row["因子"], []).append(row)

    def condition_value(col: str, expected: FactorValue) -> PredicateState:
        if "[" in col:
            if col not in facts:
                facts[col] = atomic(col)
            actual = facts[col]
        else:
            actual = result.get(col, PredicateState.MISS)
        return predicate(actual, expected)

    result: dict[str, FactorValue] = {}
    for _pass in range(len(rows) + 1):
        changed = False
        for name, groups in grouped.items():
            if result.get(name) == PredicateState.HIT:
                continue
            direct_values = [atomic(row["直通"]) for row in groups if row["直通"]]
            if direct_values:
                value = direct_values[0]
            else:
                value = any_of(
                    all_of(condition_value(col, expect) for col, expect in row["conds"].items())
                    for row in groups
                )
            if result.get(name) != value:
                result[name] = value
                changed = True
        if not changed:
            break
    for name in grouped:
        result.setdefault(name, PredicateState.MISS)
    return result
