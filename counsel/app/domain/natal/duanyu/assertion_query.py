"""Validate and execute natal/yearly assertion queries."""

from app.natal.tools.chart_schema import validate_natal_chart
from app.natal.tools.duanyu_match import assert_snapshot_factors, match_rule
from app.natal.tools.errors import AssertionRuleError
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.factors import evaluate_snap_from_natal_chart

from .cache import resolve_current_years
from .factor_closure import required_natal_factors
from .projection import brief, current_step
from .rule_loader import (
    BAZI_ONLY_RULES,
    CURRENT_LIMIT_RULES,
    NATAL_RULES,
    YEARLY_RULES,
    ZIWEI_ONLY_RULES,
    assertion_domains,
    load_rule_tables,
)


def filter_domains(result: dict, domains: list[str] | None) -> dict:
    """按命理领域过滤命中断语；None 表示不过滤。"""
    if "error" in result or domains is None:
        return result
    side_labels = load_constants()["命理侧"]["标签"]
    wanted = set(domains)
    return {
        side: [row for row in result.get(side, []) if not isinstance(row, dict) or row.get("领域") in wanted]
        for side in (side_labels["bazi"], side_labels["ziwei"])
    }


def default_scene_domains(rules: list[str]) -> list[str] | None:
    """场景别名未显式传 domains 时，应用数据表声明的主领域过滤。"""
    from .rule_loader import SCENE_DOMAIN_FILTERS

    if not rules or not all(rule in SCENE_DOMAIN_FILTERS for rule in rules):
        return None
    domains: set[str] = set()
    for rule in rules:
        domains.update(SCENE_DOMAIN_FILTERS[rule])
    return sorted(domains)


def validate_domains(domains, tables: list[list[dict]], label: str) -> None:
    if not isinstance(domains, list) or not domains or any(not isinstance(item, str) or not item for item in domains):
        raise ValueError("domains 必须是非空字符串数组。")
    if len(domains) != len(set(domains)):
        raise ValueError("domains 不能有重复领域。")
    allowed = assertion_domains(tables)
    unknown = [item for item in domains if item not in allowed]
    if unknown:
        raise ValueError(f"domains 含无效领域: {unknown}。{label}: {sorted(allowed)}")


def _resolve_current_year(rule: str, year: int | None, timezone: str):
    resolved_current_year = 0
    resolved_lunar_year = 0
    current_year_source = ""
    if year is not None:
        if year <= 0:
            raise ValueError(f"query(year) 必须为正整数，收到: {year}")
        return year, year, "specified"
    if rule in CURRENT_LIMIT_RULES:
        if not timezone:
            raise ValueError("query 当前限运域必须提供 timezone（IANA）")
        resolved_current_year, resolved_lunar_year, current_year_source = resolve_current_years(timezone)
    return resolved_current_year, resolved_lunar_year, current_year_source


def query(
    rule: str,
    natal_chart: dict,
    year: int | None = None,
    domains: list[str] | None = None,
    validate_pan: bool = True,
    timezone: str = "",
) -> dict:
    """断语查询：域 + 本命盘 → 该域断语 {八字: [...], 紫微: [...], 合参: [...]}。"""
    if rule in YEARLY_RULES:
        raise ValueError(
            f"query() 仅支持本命域，收到流年域 '{rule}'。流年查询请走 yearly_range(natal_chart, start, end, rules)。"
        )
    if rule not in NATAL_RULES:
        raise AssertionRuleError(f"未知本命域 '{rule}'。有效域: {sorted(NATAL_RULES)}")
    if year is not None and rule not in CURRENT_LIMIT_RULES:
        raise ValueError(f"query(year=...) 仅支持限运域 {sorted(CURRENT_LIMIT_RULES)}；纯本命域 '{rule}' 不依赖年份。")
    if validate_pan:
        validate_natal_chart(natal_chart, action="query")
    resolved_current_year, resolved_lunar_year, current_year_source = _resolve_current_year(rule, year, timezone)
    query_tables_by_side = load_rule_tables(rule)
    query_tables = list(query_tables_by_side.values())
    if rule in BAZI_ONLY_RULES:
        requested_sides = {"bazi"}
    elif rule in ZIWEI_ONLY_RULES:
        requested_sides = {"ziwei"}
    else:
        requested_sides = {"bazi", "ziwei"}
    if domains is not None:
        validate_domains(
            domains,
            query_tables,
            f"rule={rule} 有效领域",
        )
    snapshots = evaluate_snap_from_natal_chart(
        natal_chart,
        current_year=resolved_current_year,
        current_year_ziwei=resolved_lunar_year or resolved_current_year,
        sides=requested_sides,
        factor_names=required_natal_factors(query_tables),
    )
    side_labels = load_constants()["命理侧"]["标签"]
    for side_code, table in zip(load_constants()["命理侧"]["断言代码"], query_tables):
        if not table:
            continue
        merged = snapshots[side_labels[side_code]]
        assert_snapshot_factors(
            rule,
            table,
            {**merged, **(snapshots.get("context", {}) or {})},
        )
    result = dict(filter_domains(match_rule(rule, snapshots), domains))
    if rule == "用神":
        bazi = natal_chart["bazi"]
        result.update(
            {
                "fu_yi": bazi["fu_yi"],
                "tiao_hou": bazi["tiao_hou"],
                "ge_ju": bazi["ge_ju"],
                "element_states": bazi["element_states"],
                "ten_god_states": bazi["ten_god_states"],
            }
        )
    if resolved_current_year:
        result.update(
            {
                "current_year": resolved_current_year,
                "current_year_source": current_year_source,
            }
        )
        if resolved_lunar_year and resolved_lunar_year != resolved_current_year:
            result["current_lunar_year"] = resolved_lunar_year
        dayun = current_step((natal_chart.get("bazi") or {}).get("da_yun", {}).get("steps"), resolved_current_year)
        if dayun:
            result["current_dayun"] = dayun
        daxian = current_step(
            natal_chart.get("ziwei_daxian"),
            resolved_lunar_year or resolved_current_year,
        )
        if daxian:
            result["current_daxian"] = daxian
    return result


def query_yearly(rule: str, snapshots: dict) -> dict:
    if snapshots.get("_snapshot_type") != "liunian":
        raise AssertionRuleError("query_yearly 仅接受流年快照（含 _snapshot_type='liunian'）")
    if rule not in YEARLY_RULES:
        raise AssertionRuleError(
            f"未知流年命理域 '{rule}'。有效域: {sorted(YEARLY_RULES)}（场景别名请走 yearly_range 展开）"
        )
    side_labels = load_constants()["命理侧"]["标签"]
    for side_code, table in load_rule_tables(rule).items():
        if not table:
            continue
        merged = snapshots[side_labels[side_code]]
        assert_snapshot_factors(
            rule,
            table,
            {**merged, **(snapshots.get("context", {}) or {})},
        )
    return match_rule(rule, snapshots, skip_validation=True)


__all__ = [
    "brief",
    "default_scene_domains",
    "filter_domains",
    "query",
    "query_yearly",
    "validate_domains",
]
