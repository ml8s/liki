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
    states = list(values)
    if any(not isinstance(state, PredicateState) for state in states):
        raise FactorEvaluateError("AND 只接受四态状态，不接受领域标量")

    for state in (PredicateState.MISS, PredicateState.UNKNOWN, PredicateState.NOT_APPLICABLE):
        if state in states:
            return state
    return PredicateState.HIT


def any_of(values: Iterable[PredicateState]) -> PredicateState:
    """OR：命中优先；未知能覆盖已知没命中；全部不适用才不适用。"""
    states = list(values)
    if any(not isinstance(state, PredicateState) for state in states):
        raise FactorEvaluateError("OR 只接受四态状态，不接受领域标量")

    for state in (PredicateState.HIT, PredicateState.UNKNOWN, PredicateState.MISS):
        if state in states:
            return state
    return PredicateState.NOT_APPLICABLE


_LEGACY_PUBLISHED_VALUES = {"0", "1", "unknown"}


def validate_published_factor_value(value: object, *, where: str = "因子") -> FactorValue:
    """严格校验已发布因子：不归一旧别名，空串/旧别名/非标量 fail closed。"""
    if isinstance(value, PredicateState):
        return value
    if (
        isinstance(value, int)
        and not isinstance(value, bool)
        and value in (0, 1)
    ):
        raise FactorEvaluateError(f"{where} 不能使用旧整数语义: {value!r}")
    if isinstance(value, str) and value:
        if value in _LEGACY_PUBLISHED_VALUES:
            raise FactorEvaluateError(f"{where} 不能使用旧哨兵: {value!r}")
        return value
    raise FactorEvaluateError(f"{where} 不是已发布 FactorValue: {value!r}")


__all__ = [
    "FactorValue", "PredicateState", "all_of", "any_of", "factor_value", "predicate",
    "validate_published_factor_value",
]
