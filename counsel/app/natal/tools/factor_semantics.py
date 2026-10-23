"""因子真值表的封闭语义内核。

四态是唯一公共谓词语义；领域标量因子输出自己的闭集值。
旧 CSV 中的 1/0/unknown 只是加载边界的历史别名，不进入因子快照。
"""

from collections.abc import Iterable
from enum import StrEnum
from typing import TypeAlias

from app.natal.tools.errors import FactorEvaluateError


class PredicateState(StrEnum):
    HIT = "命中"
    MISS = "没命中"
    NOT_APPLICABLE = "不适用"
    UNKNOWN = "不可判定"


# 标量因子返回领域闭集字符串；谓词因子返回 PredicateState。
FactorValue: TypeAlias = PredicateState | str

_ALIASES = {
    1: PredicateState.HIT,
    "1": PredicateState.HIT,
    0: PredicateState.MISS,
    "0": PredicateState.MISS,
    "unknown": PredicateState.UNKNOWN,
}


def factor_value(value: object, *, where: str = "因子") -> FactorValue:
    """把加载/算子边界的旧紧凑值归一到 FactorValue，并拒绝非法形状。"""
    if isinstance(value, PredicateState):
        return value
    if isinstance(value, bool):
        raise FactorEvaluateError(f"{where} 不能返回 boolean: {value!r}")
    if isinstance(value, int) and value in _ALIASES:
        return _ALIASES[value]
    if isinstance(value, str):
        if not value:
            # 旧算子用 "" 表示事实缺失；统一边界归为不可判定。
            return PredicateState.UNKNOWN
        return _ALIASES.get(value, value)
    raise FactorEvaluateError(f"{where} 不是 FactorValue: {value!r}")


def predicate(actual: FactorValue, expected: FactorValue) -> PredicateState:
    """一个可评估条件的四态结果。"""
    if actual in (PredicateState.NOT_APPLICABLE, PredicateState.UNKNOWN):
        return actual  # type: ignore[return-value]
    return PredicateState.HIT if actual == expected else PredicateState.MISS


def all_of(values: Iterable[PredicateState]) -> PredicateState:
    """AND：已知不命中优先；未知次之；不适用最后；全部命中才命中。"""
    states = {value for value in values if isinstance(value, PredicateState)}
    for state in (PredicateState.MISS, PredicateState.UNKNOWN, PredicateState.NOT_APPLICABLE):
        if state in states:
            return state
    return PredicateState.HIT


def any_of(values: Iterable[PredicateState]) -> PredicateState:
    """OR：命中优先；已知不命中次之；不可判定再次；全不适用才不适用。"""
    states = {value for value in values if isinstance(value, PredicateState)}
    for state in (PredicateState.HIT, PredicateState.MISS, PredicateState.UNKNOWN):
        if state in states:
            return state
    return PredicateState.NOT_APPLICABLE


__all__ = [
    "FactorValue", "PredicateState", "all_of", "any_of", "factor_value", "predicate",
]
