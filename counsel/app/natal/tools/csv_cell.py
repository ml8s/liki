"""CSV 单元格解析。"""

from app.natal.tools.factor_semantics import PredicateState


def typed_expected(value: str | None):
    """历史 CSV 的 1/0 是命中/没命中的紧凑别名。"""
    value = (value or "").strip()
    if value == "1":
        return PredicateState.HIT
    if value == "0":
        return PredicateState.MISS
    return value
