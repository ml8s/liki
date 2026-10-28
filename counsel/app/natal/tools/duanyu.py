"""断语层 — 因子快照 × 断语真值表 → 命理断语。

输入：factors.py 生成的本命/流年因子快照。
输出：{八字: [...], 紫微: [...], 合参: [...]}。
本文件不生成因子；性别作为排盘上下文只参与匹配，不计入因子数。
"""

import time
from concurrent.futures import ThreadPoolExecutor

import app.natal.tools.yearly_eval as yearly_eval
from app.natal.tools.assertion_store import load_rule_table
from app.natal.tools.chart_schema import ChartSchemaError, validate_natal_chart
from app.natal.tools.duanyu_match import assert_snapshot_factors, match_rule, match_table
from app.natal.tools.errors import AssertionRuleError, YearRangeError
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.factor_tables import load_factor_rows, load_liunian_rows
from app.natal.tools.factors import (
    evaluate_liunian_snap_from_natal_chart,
    evaluate_snap_from_natal_chart,
    prepare_natal_context,
)
from app.natal.tools.yearly_eval import MAX_YEARS, resolve_rules

__all__ = [
    "NATAL_RULES", "YEARLY_RULES", "SCENE_ALIASES", "SCENE_DOMAIN_FILTERS",
    "BAZI_ONLY_RULES", "ZIWEI_ONLY_RULES", "CURRENT_LIMIT_RULES",
    "load_rule_table", "load_rule_tables", "required_natal_factors",
    "required_flow_factors", "natal_factors_for_flow", "flow_factor_names",
    "filter_domains", "default_scene_domains", "query", "brief",
    "resolve_current_year", "resolve_current_years", "reset_current_year_cache", "query_yearly",
    "match_table", "match_rule", "yearly_range",
]

def load_rule_tables(rule: str) -> dict[str, list[dict]]:
    """按命理侧配置加载一个域的三侧断言表。"""
    side_config = load_constants()["命理侧"]
    tables = {}
    for side in side_config["断言代码"]:
        if side == "bazi":
            required = rule not in ZIWEI_ONLY_RULES
        else:
            required = rule not in BAZI_ONLY_RULES
        tables[side] = load_rule_table(f"{side}_{rule}.csv", required=required)
    return tables


def _factor_closure(
    groups: dict[str, list[dict]], seeds: set[str]
) -> set[str]:
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


def required_natal_factors(tables: list[list[dict]]) -> set[str]:
    """从断言条件反向求本命因子闭包；query 只计算实际消费的因子。"""
    groups: dict[str, list[dict]] = {}
    for row in load_factor_rows():
        groups.setdefault(row["因子"], []).append(row)
    seeds: set[str] = set()
    for table in tables:
        for row in table:
            for conditions in row.get("约束组", []):
                for factor in conditions:
                    if factor in groups:
                        seeds.add(factor)
    return _factor_closure(groups, seeds)


def required_flow_factors(tables: list[list[dict]]) -> set[str]:
    """从流年断言条件反向求因子闭包；yearly_range 只算实际消费因子。"""
    groups: dict[str, list[dict]] = {}
    for row in load_liunian_rows():
        groups.setdefault(row["因子"], []).append(row)
    seeds: set[str] = set()
    for table in tables:
        for row in table:
            for conditions in row.get("约束组", []):
                for factor in conditions:
                    if factor in groups:
                        seeds.add(factor)
    return _factor_closure(groups, seeds)


def natal_factors_for_flow(flow_factors: set[str]) -> set[str]:
    """返回流年因子显式或隐式消费的本命八字因子闭包。"""
    flow_groups: dict[str, list[dict]] = {}
    for row in load_liunian_rows():
        flow_groups.setdefault(row["因子"], []).append(row)
    groups: dict[str, list[dict]] = {}
    for row in load_factor_rows():
        groups.setdefault(row["因子"], []).append(row)
    seeds: set[str] = set()
    for name in _factor_closure(flow_groups, flow_factors):
        for row in flow_groups.get(name, ()):
            for condition in row.get("conds", {}):
                if condition.startswith("引用本命[") and condition.endswith("]"):
                    seeds.add(condition[len("引用本命["):-1])
    return _factor_closure(groups, seeds)


def flow_factor_names(rules: list[str]) -> set[str]:
    """展开流年场景别名并返回断言消费的流年因子闭包。"""
    resolved_rules = _resolve_rules(rules)
    tables = [table for rule in resolved_rules for table in load_rule_tables(rule).values()]
    return required_flow_factors(tables)

# 命理域全集与场景别名来自 constants.json「命理域」；代码不内置领域清单。
_DOMAIN_CONFIG = load_constants()["命理域"]
NATAL_RULES = frozenset(_DOMAIN_CONFIG["本命规则"])
YEARLY_RULES = frozenset(_DOMAIN_CONFIG["流年规则"])
SCENE_ALIASES = {
    alias: tuple(rules)
    for alias, rules in _DOMAIN_CONFIG["场景别名"].items()
}
SCENE_DOMAIN_FILTERS = {
    alias: tuple(domains)
    for alias, domains in _DOMAIN_CONFIG.get("场景领域过滤", {}).items()
}
# 显式单侧域：八字层次/流年八字域（紫微侧无对应 csv 是设计事实）；紫微宫位/流年宫位域（八字侧无）
# 显式单侧域同样来自 constants.json，不在代码里重复维护。
BAZI_ONLY_RULES = frozenset(_DOMAIN_CONFIG["八字专属域"])
ZIWEI_ONLY_RULES = frozenset(_DOMAIN_CONFIG["紫微专属域"])
# 需 current_year（当前八字大运/紫微大限判断）的本命域；纯本命域不得依赖当前时间。
CURRENT_LIMIT_RULES = frozenset(_DOMAIN_CONFIG["当前限运规则"])


def _assertion_domains(tables: list[list[dict]]) -> set[str]:
    """读取规则表自带的领域闭集；不硬编码领域清单。"""
    return {
        row["领域"]
        for table in tables
        for row in table
        if row.get("领域")
    }


def filter_domains(result: dict, domains: list[str] | None) -> dict:
    """按命理领域过滤命中断语；None 表示不过滤。"""
    if "error" in result:
        return result
    if domains is None:
        return result
    side_labels = load_constants()["命理侧"]["标签"]
    wanted = set(domains)
    return {
        side: [
            row for row in result.get(side, [])
            if not isinstance(row, dict) or row.get("领域") in wanted
        ]
        for side in (side_labels["bazi"], side_labels["ziwei"])
    }


def default_scene_domains(rules: list[str]) -> list[str] | None:
    """场景别名未显式传 domains 时，应用数据表声明的主领域过滤。"""
    if not rules or not all(rule in SCENE_DOMAIN_FILTERS for rule in rules):
        return None
    domains: set[str] = set()
    for rule in rules:
        domains.update(SCENE_DOMAIN_FILTERS[rule])
    return sorted(domains)


def query(rule: str, natal_chart: dict, year: int | None = None,
          domains: list[str] | None = None,
          validate_pan: bool = True, timezone: str = "") -> dict:
    """断语查询：域 + 本命盘 → 该域断语 {八字: [...], 紫微: [...], 合参: [...]}。

    rule ∈ NATAL_RULES（如 "十神"/"旺衰"/"命宫"/"官禄"；流年域走 yearly_range）。
    natal_chart = build_natal_chart 返回的完整本命盘；内部从 natal_chart 生成领域快照。
    year 仅用于「大运/大限」这类限运域：省略用服务端当前年，显式传入则取该年限运。
    流年查询走 yearly_range，本函数不处理流年快照。
    数据层真分开（各查各表）、调用层一次查双盘——内部 load_rule_table + match_table 内嵌。
    """
    # 本命函数不允许查流年域——流年走 yearly_range
    if rule in YEARLY_RULES:
        raise ValueError(
            f"query() 仅支持本命域，收到流年域 '{rule}'。"
            f"流年查询请走 yearly_range(natal_chart, start, end, rules)。"
        )
    if rule not in NATAL_RULES:
        raise AssertionRuleError(f"未知本命域 '{rule}'。有效域: {sorted(NATAL_RULES)}")
    if year is not None and rule not in CURRENT_LIMIT_RULES:
        raise ValueError(
            f"query(year=...) 仅支持限运域 {sorted(CURRENT_LIMIT_RULES)}；"
            f"纯本命域 '{rule}' 不依赖年份。"
        )
    # natal_chart 直通——LLM 传 build_natal_chart 的返回即可，内部从 natal_chart 直读产因子快照。
    # 校验完整排盘结构，杜绝把空 dict/快照/半截盘误当命盘（防兜底断语污染）。
    if validate_pan:
        validate_natal_chart(natal_chart, action="query")
    resolved_current_year = 0
    resolved_lunar_year = 0
    current_year_source = ""
    if year is not None:
        if year <= 0:
            raise ValueError(f"query(year) 必须为正整数，收到: {year}")
        resolved_current_year = year
        resolved_lunar_year = year
        current_year_source = "specified"
    elif rule in CURRENT_LIMIT_RULES:
        if not timezone:
            raise ValueError("query 当前限运域必须提供 timezone（IANA）")
        resolved_current_year, resolved_lunar_year, current_year_source = resolve_current_years(timezone)
    query_tables_by_side = load_rule_tables(rule)
    query_tables = list(query_tables_by_side.values())
    # 单侧域只限制 bazi/ziwei 断言表；bazi/ziwei 各查各表。
    if rule in BAZI_ONLY_RULES:
        requested_sides = {"bazi"}
    elif rule in ZIWEI_ONLY_RULES:
        requested_sides = {"ziwei"}
    else:
        requested_sides = {"bazi", "ziwei"}
    if domains is not None:
        if not isinstance(domains, list) or not domains or \
                any(not isinstance(item, str) or not item for item in domains):
            raise ValueError("domains 必须是非空字符串数组。")
        if len(domains) != len(set(domains)):
            raise ValueError("domains 不能有重复领域。")
        allowed_domains = _assertion_domains(query_tables)
        unknown = [item for item in domains if item not in allowed_domains]
        if unknown:
            raise ValueError(
                f"domains 含无效领域: {unknown}。"
                f"rule={rule} 有效领域: {sorted(allowed_domains)}"
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
        result["fu_yi"] = natal_chart["full"]["fu_yi"]
        result["tiao_hou"] = natal_chart["full"]["tiao_hou"]
        result["ge_ju"] = natal_chart["full"]["ge_ju"]
        result["element_states"] = natal_chart["full"]["element_states"]
        result["ten_god_states"] = natal_chart["full"]["ten_god_states"]
    if resolved_current_year:
        result["current_year"] = resolved_current_year
        result["current_year_source"] = current_year_source
        if resolved_lunar_year and resolved_lunar_year != resolved_current_year:
            result["current_lunar_year"] = resolved_lunar_year
        dayun = _current_step((natal_chart.get("chart") or {}).get("da_yun", {}).get("steps"),
                              resolved_current_year)
        if dayun:
            result["current_dayun"] = dayun
        daxian = _current_step(natal_chart.get("ziwei_daxian"),
                               resolved_lunar_year or resolved_current_year)
        if daxian:
            result["current_daxian"] = daxian
    return result


def _current_step(steps, year: int):
    """当前大运/大限步骤：按公历年段匹配；无年份或不在任何步骤内返回 None。"""
    if not year or not isinstance(steps, list):
        return None
    for index, step in enumerate(steps):
        if not isinstance(step, dict):
            continue
        start, end = step.get("start_year"), step.get("end_year")
        if isinstance(start, int) and isinstance(end, int) and start <= year <= end:
            item = {
                "index": index,
                "name": step.get("name", ""),
                "start_year": start,
                "end_year": end,
                "next_step_in_years": end - year,
            }
            if step.get("gong"):
                item["gong"] = step["gong"]
            return item
    return None

_BRIEF_FIELDS = ("id", "领域", "事件类型", "时间层", "事件", "结论")


def brief(conclusions: list) -> list:
    return [{k: c[k] for k in _BRIEF_FIELDS if k in c}
            for c in conclusions if isinstance(c, dict)]




class _TtlCache:
    """有界 TTL 缓存：显式生命周期，不暴露裸全局变量。"""
    __slots__ = ("_data", "_at", "_ttl")

    def __init__(self, ttl: float):
        self._data = None
        self._at: float | None = None
        self._ttl = ttl

    def get(self, key, now: float):
        """命中返回 value，过期/不匹配返回 None。"""
        if self._at is None or self._data is None:
            return None
        if now - self._at >= self._ttl:
            return None
        if self._data[-1] != key:
            return None
        return self._data[:-1]

    def set(self, key, value: tuple, now: float):
        self._data = (*value, key)
        self._at = now

    def clear(self):
        self._data = None
        self._at = None


# 当前年份事实（time.now + 年界盘）TTL 缓存——当前限运精度到年，单会话内不重复发起 RPC。
# RPC 失败仍抛错（fail-fast，绝不回退本地时钟）；仅成功结果被缓存。
_year_cache = _TtlCache(ttl=60)


def resolve_current_years(timezone: str):
    """返回 (八字年号[立春界], 农历年号[春节界], source)。timezone 必填。"""
    from app.natal.tools.charts import current_years

    if not timezone:
        raise ValueError("timezone is required")
    now = time.monotonic()
    cached = _year_cache.get(timezone, now)
    if cached is not None:
        return cached
    bazi_year, lunar_year = current_years(timezone)
    _year_cache.set(timezone, (bazi_year, lunar_year, "server"), now)
    return bazi_year, lunar_year, "server"


def resolve_current_year(timezone: str):
    """兼容入口：返回 (八字年号, source)。timezone 必填。"""
    bazi_year, _lunar_year, source = resolve_current_years(timezone)
    return bazi_year, source


def reset_current_year_cache():
    _year_cache.clear()


def _resolve_rules(rules):
    return resolve_rules(rules, YEARLY_RULES, SCENE_ALIASES)


def _liunian_for_year(natal_chart, year):
    from app.natal.tools.charts import _bazi_liunian, _ziwei_liunian, liunian
    chart = natal_chart.get("chart") or {}
    if natal_chart.get("ziwei") and "ri" in chart:
        # 完整盘（八字 chart + 紫微）：双盘流年
        return liunian(natal_chart, year, validate=False)
    if "ri" in chart:
        # 正交化八字组合盘（analysis）：只排八字流年，紫微侧留空。
        return {"bazi": _bazi_liunian(chart, year), "ziwei": {}}
    # 正交化紫微组合盘（analysis ziwei）：chart 为紫微盘，只排紫微流年。
    return {"bazi": {}, "ziwei": _ziwei_liunian(chart, year)}


def _batch_liunian_data(natal_chart: dict, start: int, count: int) -> dict[int, dict]:
    """批量预取流年盘数据，返回 year → {bazi, ziwei} 查找表。

    engine 的 series count 上限是 50；这里按上限分块，每个 chunk 内 bazi 与
    ziwei 两次 RPC 并发。单个 chunk 失败只污染该 chunk 覆盖的年份，失败项
    保持 {bazi: {}, ziwei: {}, error: ...} 形状，供 yearly_range 显式降级。
    """
    from app.natal.tools.charts import (
        MAX_LIUNIAN_BATCH,
        RPCError,
        _bazi_liunian_range,
        _ziwei_liunian_range,
    )
    chart = natal_chart.get("chart") or {}
    zw = natal_chart.get("ziwei_chart") or natal_chart.get("ziwei") or {}
    # Keep orthogonal analysis pans on the same side selection as
    # _liunian_for_year: bazi-only has a ri pillar, while a Ziwei analysis pan
    # stores its engine chart in `chart` without ri.
    fetch_bazi = "ri" in chart
    fetch_ziwei = bool(natal_chart.get("ziwei")) if fetch_bazi else bool(zw)
    years = {}

    for chunk_start in range(start, start + count, MAX_LIUNIAN_BATCH):
        chunk_count = min(MAX_LIUNIAN_BATCH, start + count - chunk_start)
        chunk_years = range(chunk_start, chunk_start + chunk_count)
        try:
            with ThreadPoolExecutor(max_workers=2) as executor:
                bazi_future = (
                    executor.submit(_bazi_liunian_range, chart, chunk_start, chunk_count)
                    if fetch_bazi else None
                )
                ziwei_future = (
                    executor.submit(_ziwei_liunian_range, zw, chunk_start, chunk_count)
                    if fetch_ziwei else None
                )
                bazi_items = (
                    bazi_future.result() if bazi_future is not None
                    else [{} for _ in chunk_years]
                )
                ziwei_items = (
                    ziwei_future.result() if ziwei_future is not None
                    else [{} for _ in chunk_years]
                )
            for i, year in enumerate(chunk_years):
                years[year] = {
                    "bazi": bazi_items[i] if i < len(bazi_items) else {},
                    "ziwei": ziwei_items[i] if i < len(ziwei_items) else {},
                }
        except (RPCError, ConnectionError, TimeoutError, OSError) as e:
            error = {"error": f"{type(e).__name__}: {e}"}
            for year in chunk_years:
                years[year] = {"bazi": {}, "ziwei": {}, **error}
    return years


def query_yearly(rule: str, snapshots: dict) -> dict:
    if snapshots.get("_snapshot_type") != "liunian":
        raise AssertionRuleError("query_yearly 仅接受流年快照（含 _snapshot_type='liunian'）")
    if rule not in YEARLY_RULES:
        raise AssertionRuleError(f"未知流年命理域 '{rule}'。有效域: {sorted(YEARLY_RULES)}（场景别名请走 yearly_range 展开）")
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




# ── Re-exports from duanyu_match ──
# Backward compatibility: match_table/match_rule/assert_snapshot_factors
# moved to duanyu_match.py; re-exported here for existing import sites.

def yearly_range(natal_chart: dict, start: int, end: int,
                 rules: list, detail: bool = False,
                 domains: list[str] | None = None,
                 validate_pan: bool = True, timezone: str = "") -> dict:
    """流年断语（年粒度）。

    start/end 为流年年号（直接作为年号使用，不做公历/农历日期换算）：
    - 八字流年：干支年号（公历立春界，年中代表 = 公历年号）
    - 紫微流年：农历年号（春节界，年号级与八字一致，如 2030 → 庚戌）
    年号级下八字与紫微一致（2024 = 甲辰），故同一年号两边通用；不做换算
    以免引入月份信息（用户只按"年"查询）。
    """
    resolved_rules = _resolve_rules(rules)
    if domains is None:
        domains = default_scene_domains(rules)
    if domains is not None:
        domains = list(domains)
    if start > end:
        raise YearRangeError(f"yearly_range start 不能大于 end：{start} > {end}")
    if end - start + 1 > MAX_YEARS:
        raise YearRangeError(
            f"yearly_range 年份跨度过大：{start}-{end} 共 {end - start + 1} 年，"
            f"单次最多 {MAX_YEARS} 年。"
        )
    if validate_pan:
        validate_natal_chart(natal_chart, action="yearly_range")
    else:
        for key in ("chart", "full", "gender"):
            if not natal_chart.get(key):
                raise ChartSchemaError(f"yearly_range natal_chart 缺 {key}，无法流年求值。")
    if not timezone:
        raise ValueError("yearly_range 必须提供 timezone（IANA）")
    cur_year, cur_source = resolve_current_year(timezone)
    flow_factors = flow_factor_names(rules)
    yearly_tables = [
        table
        for rule in resolved_rules
        for table in load_rule_tables(rule).values()
    ]
    if domains is not None:
        if not isinstance(domains, list) or not domains or \
                any(not isinstance(item, str) or not item for item in domains):
            raise ValueError("domains 必须是非空字符串数组。")
        if len(domains) != len(set(domains)):
            raise ValueError("domains 不能有重复领域。")
        allowed_domains = _assertion_domains(yearly_tables)
        unknown = [item for item in domains if item not in allowed_domains]
        if unknown:
            raise ValueError(
                f"domains 含无效领域: {unknown}。"
                f"有效领域: {sorted(allowed_domains)}"
            )
    natal_context = prepare_natal_context(
        natal_chart, factor_names=natal_factors_for_flow(flow_factors)
    )

    total = end - start + 1
    from app.natal.tools.charts import RPCError
    liunian_cache = _batch_liunian_data(natal_chart, start, total)

    def _liunian_cached(natal_chart_: dict, year: int) -> dict:
        """从预取缓存查找流年盘，零 RPC。"""
        return liunian_cache.get(year, {"bazi": {}, "ziwei": {}})

    years = {}
    for year in range(start, end + 1):
        try:
            prefetch = liunian_cache.get(year)
            if prefetch and prefetch.get("error"):
                years[str(year)] = {"error": prefetch["error"]}
                continue
            result = yearly_eval.evaluate_year(
                natal_chart, year, resolved_rules, detail, natal_context,
                query_yearly=query_yearly,
                brief=brief,
                liunian=_liunian_cached,
                evaluate_liunian_snap_from_natal_chart=evaluate_liunian_snap_from_natal_chart,
                factor_names=flow_factors,
            )
            if domains is not None and "error" not in result:
                result = {
                    rule: filter_domains(result, domains)
                    for rule, result in result.items()
                }
            years[str(year)] = result
        except (RPCError, ConnectionError, TimeoutError, OSError) as e:
            years[str(year)] = {"error": f"{type(e).__name__}: {e}"}

    return {"current_year": cur_year, "current_year_source": cur_source,
            "year_basis": dict(load_constants()["流年年界"]),
            "years": years}
