"""流年范围/考时共用的年度求值原语。

依赖通过参数显式注入，避免 calibrate -> duanyu 私有 API 的反向耦合。
"""
from __future__ import annotations

from typing import Callable, Iterable

from errors import YearRangeError
from factor_context import NatalContext

MAX_YEARS = 120


def resolve_rules(
    rules: list[str] | None,
    yearly_rules: Iterable[str],
    scene_aliases: dict[str, tuple[str, ...]],
) -> list[str]:
    """校验并展开流年场景别名；返回去重保序的命理域。"""
    rules = ["yearly_career"] if rules is None else rules
    if not rules:
        raise YearRangeError("yearly_range rules 不能为空")
    if len(rules) != len(set(rules)):
        raise YearRangeError("yearly_range rules 不能有重复域")
    valid = set(yearly_rules) | set(scene_aliases)
    for rule in rules:
        if rule not in valid:
            raise YearRangeError(
                f"yearly_range rules 含无效域 '{rule}'。"
                f"有效域: {sorted(yearly_rules)}，场景别名: {sorted(scene_aliases)}"
            )
    resolved, seen = [], set()
    for rule in rules:
        for expanded in scene_aliases.get(rule, (rule,)):
            if expanded not in seen:
                seen.add(expanded)
                resolved.append(expanded)
    if not resolved:
        raise YearRangeError("yearly_range rules 不能为空")
    return resolved


def query_year_rules(
    snapshot: dict,
    rules: list[str],
    detail: bool,
    *,
    query_yearly: Callable[[str, dict], dict],
    brief: Callable[[list], list],
) -> dict[str, dict]:
    """一个流年快照 × 多个命理域断语。"""
    year_data: dict[str, dict] = {}
    for rule in rules:
        result = query_yearly(rule, snapshot)
        year_data[rule] = result if detail else {
            "八字": brief(result.get("八字", [])),
            "紫微": brief(result.get("紫微", [])),
        }
    return year_data


def yearly_snapshot(
    pan: dict,
    year: int,
    natal_context: NatalContext | None,
    *,
    liunian: Callable[[dict, int], dict],
    evaluate_liunian_snap_from_pan: Callable[..., dict],
) -> dict:
    """排单年流年盘并生成流年快照。"""
    return evaluate_liunian_snap_from_pan(
        pan, liunian(pan, year), year=year, natal_context=natal_context
    )


def evaluate_year(
    pan: dict,
    year: int,
    rules: list[str],
    detail: bool,
    natal_context: NatalContext | None,
    *,
    query_yearly: Callable[[str, dict], dict],
    brief: Callable[[list], list],
    liunian: Callable[[dict, int], dict],
    evaluate_liunian_snap_from_pan: Callable[..., dict],
) -> dict[str, dict]:
    """统一 yearly_range 的年度求值流程。"""
    snapshot = yearly_snapshot(
        pan, year, natal_context,
        liunian=liunian,
        evaluate_liunian_snap_from_pan=evaluate_liunian_snap_from_pan,
    )
    return query_year_rules(
        snapshot, rules, detail,
        query_yearly=query_yearly,
        brief=brief,
    )
