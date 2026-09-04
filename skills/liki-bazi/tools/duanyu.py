"""断语层 — 因子快照 × 断语真值表 → 命理断语。

输入：factors.py 生成的本命/流年因子快照。
输出：{八字: [...], 紫微: [...], 合参: [...]}。
本文件不生成因子；性别作为排盘上下文只参与匹配，不计入因子数。
"""
from __future__ import annotations

import time

from assertion_store import load_rule_table
from factor_constants import load_constants
from pan_schema import validate_natal_pan
from errors import AssertionRuleError, YearRangeError
from factor_tables import load_factor_rows, load_liunian_rows
from factors import (

    evaluate_snap_from_pan,
    evaluate_liunian_snap_from_pan,
    prepare_natal_context,
)

import yearly_eval
from yearly_eval import MAX_YEARS, resolve_rules



def load_table(name, required: bool = True):
    """加载 `{side}_{rule}` 对应的断语长表行。"""
    return load_rule_table(name, required=required)


def _load_rule_tables(rule: str) -> dict[str, list[dict]]:
    """按命理侧配置加载一个域的三侧断言表。"""
    side_config = load_constants()["命理侧"]
    tables = {}
    for side in side_config["断言代码"]:
        if side == side_config["公共代码"]:
            required = False
        elif side == "bazi":
            required = rule not in ZIWEI_ONLY_RULES
        else:
            required = rule not in BAZI_ONLY_RULES
        tables[side] = load_table(f"{side}_{rule}.csv", required=required)
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


def _required_natal_factors(tables: list[list[dict]]) -> set[str]:
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


def _required_flow_factors(tables: list[list[dict]]) -> set[str]:
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
    tables = [table for rule in resolved_rules for table in _load_rule_tables(rule).values()]
    return _required_flow_factors(tables)

# 命理域全集与场景别名来自 constants.json「命理域」；代码不内置领域清单。
_DOMAIN_CONFIG = load_constants()["命理域"]
NATAL_RULES = frozenset(_DOMAIN_CONFIG["本命规则"])
YEARLY_RULES = frozenset(_DOMAIN_CONFIG["流年规则"])
SCENE_ALIASES = {
    alias: tuple(rules)
    for alias, rules in _DOMAIN_CONFIG["场景别名"].items()
}
# 显式单侧域：八字层次/流年八字域（紫微侧无对应 csv 是设计事实）；紫微宫位/流年宫位域（八字侧无）
# 显式单侧域同样来自 constants.json，不在代码里重复维护。
BAZI_ONLY_RULES = frozenset(_DOMAIN_CONFIG["八字专属域"])
ZIWEI_ONLY_RULES = frozenset(_DOMAIN_CONFIG["紫微专属域"])
# 需 current_year（当前八字大运/紫微大限判断）的本命域；纯本命域不得依赖当前时间。
CURRENT_LIMIT_RULES = frozenset(_DOMAIN_CONFIG["当前限运规则"])


def query(rule: str, pan: dict, year: int | None = None) -> dict:
    """断语查询：域 + 本命盘 → 该域断语 {八字: [...], 紫微: [...], 合参: [...]}。

    rule ∈ NATAL_RULES（如 "十神"/"旺衰"/"命宫"/"官禄"；流年域走 yearly_range）。
    pan = full_paipan 返回的完整本命盘；内部从 pan 生成领域快照。
    year 仅用于「大运/大限」这类限运域：省略用服务端当前年，显式传入则取该年限运。
    流年查询走 yearly_range，本函数不处理流年快照。
    数据层真分开（各查各表）、调用层一次查双盘——内部 load_table + match 内嵌。
    """
    # 本命函数不允许查流年域——流年走 yearly_range
    if rule in YEARLY_RULES:
        raise ValueError(
            f"query() 仅支持本命域，收到流年域 '{rule}'。"
            f"流年查询请走 yearly_range(pan, start, end, rules)。"
        )
    if rule not in NATAL_RULES:
        raise AssertionRuleError(f"未知本命域 '{rule}'。有效域: {sorted(NATAL_RULES)}")
    if year is not None and rule not in CURRENT_LIMIT_RULES:
        raise ValueError(
            f"query(year=...) 仅支持限运域 {sorted(CURRENT_LIMIT_RULES)}；"
            f"纯本命域 '{rule}' 不依赖年份。"
        )
    # pan 直通——LLM 传 full_paipan 的返回即可，内部从 pan 直读产因子快照。
    # 校验完整排盘结构，杜绝把空 dict/快照/半截盘误当命盘（防兜底断语污染）。
    validate_natal_pan(pan, action="query")
    current_year = 0
    current_year_source = ""
    if year is not None:
        if year <= 0:
            raise ValueError(f"query(year) 必须为正整数，收到: {year}")
        current_year = year
        current_year_source = "specified"
    elif rule in CURRENT_LIMIT_RULES:
        current_year, current_year_source = _current_year()
    if rule in BAZI_ONLY_RULES:
        requested_sides = {"bazi"}
    elif rule in ZIWEI_ONLY_RULES:
        requested_sides = {"ziwei"}
    else:
        requested_sides = {"bazi", "ziwei"}
    query_tables = list(_load_rule_tables(rule).values())
    snapshots = evaluate_snap_from_pan(
        pan,
        current_year=current_year,
        sides=requested_sides,
        factor_names=_required_natal_factors(query_tables),
    )
    result = _match_rule(rule, snapshots)
    if current_year:
        result["current_year"] = current_year
        result["current_year_source"] = current_year_source
    return result

_BRIEF_FIELDS = ("id", "领域", "事件类型", "时间层", "事件", "结论")


def brief(conclusions: list) -> list:
    return [{k: c[k] for k in _BRIEF_FIELDS if k in c}
            for c in conclusions if isinstance(c, dict)]




# time.now RPC 结果 TTL 缓存（秒）——当前限运精度到年，单会话内重复查询不必反复发起 RPC。
# RPC 失败仍抛错（fail-fast，绝不回退本地时钟）；仅成功结果被缓存。
_CURRENT_YEAR_TTL = 60
_current_year_cached = None
_current_year_cached_at = None


def _current_year():
    global _current_year_cached, _current_year_cached_at
    from paipan import call

    now = time.monotonic()
    if (_current_year_cached_at is not None
            and now - _current_year_cached_at < _CURRENT_YEAR_TTL):
        return _current_year_cached
    payload = call("time.now", {})
    cst = (payload.get("data") or {}).get("cst")
    if not cst:
        raise ValueError("time.now 返回缺少 cst")
    result = (int(cst[:4]), "server")
    _current_year_cached = result
    _current_year_cached_at = now
    return result


def _reset_current_year_cache():
    global _current_year_cached, _current_year_cached_at
    _current_year_cached = None
    _current_year_cached_at = None


def _resolve_rules(rules):
    return resolve_rules(rules, YEARLY_RULES, SCENE_ALIASES)


def _evaluate_year(
    pan, year, rules, detail=False, natal_context=None, factor_names=None
):
    return yearly_eval.evaluate_year(
        pan, year, rules, detail, natal_context,
        query_yearly=query_yearly,
        brief=brief,
        liunian=_liunian_for_year,
        evaluate_liunian_snap_from_pan=evaluate_liunian_snap_from_pan,
        factor_names=factor_names,
    )


def _liunian_for_year(pan, year):
    from paipan import liunian
    return liunian(pan, year)


def query_yearly(rule: str, snapshots: dict) -> dict:
    if snapshots.get("_snapshot_type") != "liunian":
        raise AssertionRuleError("query_yearly 仅接受流年快照（含 _snapshot_type='liunian'）")
    if rule not in YEARLY_RULES:
        raise AssertionRuleError(f"未知流年命理域 '{rule}'。有效域: {sorted(YEARLY_RULES)}（场景别名请走 yearly_range 展开）")
    return _match_rule(rule, snapshots)


def _val_match(cond, actual):
    """单约束匹配：条件值与因子实际值相等即命中。"""
    return actual == cond


def match(table: list, snapshot: dict, exclusive: bool = False) -> list:
    """真值表匹配：因子快照 × 断言条件组 → 命中条目（按表序）。

    同一 condition_group 内的条件为 AND，不同 condition_group 为 OR。
    """
    hits = []
    for item in table:
        condition_groups = item.get("约束组") or []
        matched = any(
            all(_val_match(value, snapshot.get(key)) for key, value in conditions.items())
            for conditions in condition_groups
        )
        if not matched:
            continue
        hits.append(item)
    return [hits[0]] if exclusive and hits else hits


def _match_rule(rule: str, snapshots: dict) -> dict:
    """加载断语表 + 匹配。排盘上下文只参与匹配，不计入因子数。

    bazi/ziwei 表分别匹配各自快照；common 表在合并后的双盘快照上匹配，
    专用于显式八紫合参断语，不伪装成单侧事实。
    """
    tables = _load_rule_tables(rule)
    side_codes = load_constants()["命理侧"]["断言代码"]
    bz_e, zw_e, common_e = (tables[side] for side in side_codes)
    context = snapshots.get("context", {}) or {}
    side_labels = load_constants()["命理侧"]["标签"]
    return {
        side_labels["bazi"]: match(
            bz_e, {**snapshots[side_labels["bazi"]], **context}
        ),
        side_labels["ziwei"]: match(
            zw_e, {**snapshots[side_labels["ziwei"]], **context}
        ) if zw_e else [],
        side_labels["common"]: match(common_e, {
            **snapshots[side_labels["bazi"]],
            **snapshots[side_labels["ziwei"]],
            **context,
        }),
    }


def yearly_range(pan: dict, start: int, end: int,
                 rules: list, detail: bool = False) -> dict:
    resolved_rules = _resolve_rules(rules)
    if start > end:
        raise YearRangeError(f"yearly_range start 不能大于 end：{start} > {end}")
    if end - start + 1 > MAX_YEARS:
        raise YearRangeError(
            f"yearly_range 年份跨度过大：{start}-{end} 共 {end - start + 1} 年，"
            f"单次最多 {MAX_YEARS} 年。"
        )
    validate_natal_pan(pan, action="yearly_range")
    from paipan import RPCError
    cur_year, cur_source = _current_year()
    flow_factors = flow_factor_names(rules)
    natal_context = prepare_natal_context(
        pan, factor_names=natal_factors_for_flow(flow_factors)
    )
    years = {}
    for year in range(start, end + 1):
        try:
            years[str(year)] = _evaluate_year(
                pan, year, resolved_rules, detail, natal_context,
                factor_names=flow_factors,
            )
        except (RPCError, ConnectionError, TimeoutError, OSError) as e:
            years[str(year)] = {"error": f"{type(e).__name__}: {e}"}
    return {"current_year": cur_year, "current_year_source": cur_source,
            "year_basis": dict(load_constants()["流年年界"]),
            "years": years}
