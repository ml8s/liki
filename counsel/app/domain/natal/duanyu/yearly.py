"""Batched yearly assertion projection."""

from concurrent.futures import ThreadPoolExecutor

import app.natal.tools.yearly_eval as yearly_eval
from app.natal.tools.chart_schema import ChartSchemaError, validate_natal_chart
from app.natal.tools.errors import YearRangeError
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.factors import evaluate_liunian_snap_from_natal_chart, prepare_natal_context
from app.natal.tools.yearly_eval import MAX_YEARS

from .assertion_query import brief, default_scene_domains, filter_domains, query_yearly, validate_domains
from .factor_closure import flow_factor_names, natal_factors_for_flow
from .rule_loader import load_rule_tables, resolve_rule_selection


def batch_liunian_data(natal_chart: dict, start: int, count: int) -> dict[int, dict]:
    """单次预取流年盘数据，返回 year → {bazi, ziwei} 查找表。"""
    from app.natal.tools.charts import (
        RPCError,
        _bazi_liunian_range,
        _ziwei_liunian_range,
    )

    subject = natal_chart.get("subject") or {}
    bazi = natal_chart.get("bazi") or {}
    ziwei = natal_chart.get("ziwei") or {}
    fetch_bazi = "ri" in bazi
    fetch_ziwei = bool(ziwei)
    years = {}
    year_range = range(start, start + count)
    try:
        with ThreadPoolExecutor(max_workers=2) as executor:
            bazi_future = executor.submit(_bazi_liunian_range, subject, start, count) if fetch_bazi else None
            ziwei_future = executor.submit(_ziwei_liunian_range, subject, start, count) if fetch_ziwei else None
            bazi_items = bazi_future.result() if bazi_future is not None else [{} for _ in year_range]
            ziwei_items = ziwei_future.result() if ziwei_future is not None else [{} for _ in year_range]
        for index, year in enumerate(year_range):
            years[year] = {
                "bazi": bazi_items[index] if index < len(bazi_items) else {},
                "ziwei": ziwei_items[index] if index < len(ziwei_items) else {},
            }
    except (RPCError, ConnectionError, TimeoutError, OSError) as exc:
        error = {"error": f"{type(exc).__name__}: {exc}"}
        for year in year_range:
            years[year] = {"bazi": {}, "ziwei": {}, **error}
    return years


def yearly_range(
    natal_chart: dict,
    start: int,
    year_count: int,
    rules: list,
    detail: bool = False,
    domains: list[str] | None = None,
    validate_pan: bool = True,
) -> dict:
    """流年断语（年粒度）；输入是 engine series 同款 start/count 口径。"""
    resolved_rules = resolve_rule_selection(rules)
    if domains is None:
        domains = default_scene_domains(rules)
    if domains is not None:
        domains = list(domains)
    if year_count <= 0:
        raise YearRangeError(f"yearly_range year_count 必须是正整数：{year_count}")
    if year_count > MAX_YEARS:
        raise YearRangeError(f"yearly_range 年份跨度过大：{year_count} 年，单次最多 {MAX_YEARS} 年。")
    if validate_pan:
        validate_natal_chart(natal_chart, action="yearly_range")
    else:
        for key in ("subject", "bazi", "ziwei"):
            if not natal_chart.get(key):
                raise ChartSchemaError(f"yearly_range natal_chart 缺 {key}，无法流年求值。")
    flow_factors = flow_factor_names(rules)
    yearly_tables = [table for rule in resolved_rules for table in load_rule_tables(rule).values()]
    if domains is not None:
        validate_domains(domains, yearly_tables, "有效领域")
    natal_context = prepare_natal_context(natal_chart, factor_names=natal_factors_for_flow(flow_factors))

    from app.natal.tools.charts import RPCError

    liunian_cache = batch_liunian_data(natal_chart, start, year_count)

    def _liunian_cached(_natal_chart: dict, year: int) -> dict:
        """从预取缓存查找流年盘，零 RPC。"""
        return liunian_cache.get(year, {"bazi": {}, "ziwei": {}})

    years = {}
    for year in range(start, start + year_count):
        try:
            prefetch = liunian_cache.get(year)
            if prefetch and prefetch.get("error"):
                years[str(year)] = {"error": prefetch["error"]}
                continue
            result = yearly_eval.evaluate_year(
                natal_chart,
                year,
                resolved_rules,
                detail,
                natal_context,
                query_yearly=query_yearly,
                brief=brief,
                liunian=_liunian_cached,
                evaluate_liunian_snap_from_natal_chart=evaluate_liunian_snap_from_natal_chart,
                factor_names=flow_factors,
            )
            if domains is not None and "error" not in result:
                result = {rule: filter_domains(result, domains) for rule, result in result.items()}
            years[str(year)] = result
        except (RPCError, ConnectionError, TimeoutError, OSError) as exc:
            years[str(year)] = {"error": f"{type(exc).__name__}: {exc}"}

    return {
        "year_basis": dict(load_constants()["流年年界"]),
        "years": years,
    }
