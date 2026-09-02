"""断语层 — 因子快照 × 断语真值表 → 命理断语。

输入：factors.py 生成的本命/流年因子快照。
输出：{八字: [...], 紫微: [...]}。
本文件不生成因子；性别作为排盘上下文只参与匹配，不计入因子数。
"""
from __future__ import annotations

import os
import time

from assertion_store import load_rule_table
from pan_schema import validate_natal_pan
from errors import AssertionRuleError, YearRangeError
from factors import (

    evaluate_snap_from_pan,
    evaluate_liunian_snap_from_pan,
    prepare_natal_context,
)

import yearly_eval
from yearly_eval import MAX_YEARS, query_year_rules, resolve_rules, yearly_snapshot



def load_table(name, required: bool = True):
    """加载 `{side}_{rule}` 对应的断语长表行。"""
    return load_rule_table(name, required=required)

# 命理域规则全集：紫微按宫位，八字按命理层次。
ZW_PALACE_RULES = ("命宫", "官禄", "财帛", "疾厄", "夫妻", "子女", "迁移",
                    "田宅", "父母", "兄弟", "仆役", "福德")
ZW_COMMON_RULE = ("格局",)   # 八字也有格局域（双体系共用 rule 名，双侧都有表）
BZ_LAYER_RULES = ("十神", "旺衰", "用神", "大运", "合会", "神煞", "调候",
                   "五行", "六亲", "出身", "外貌")
# 流年命理域（八字层次 + 紫微宫位，前缀"年"）
YEAR_PALACE_RULES = ("年命宫", "年官禄", "年财帛", "年疾厄", "年福德",
                      "年夫妻", "年父母", "年子女", "年田宅", "年迁移",
                      "年兄弟", "年仆役")
YEAR_BZ_RULES = ("年十神", "年六亲", "年神煞", "年用神", "年旺衰",
                  "年合会", "年大运", "年五行")

# 本命命理域
NATAL_RULES = frozenset(ZW_PALACE_RULES + ZW_COMMON_RULE + BZ_LAYER_RULES)
# 流年命理域
YEARLY_RULES = frozenset(YEAR_PALACE_RULES + YEAR_BZ_RULES)
# 流年场景别名 → 命理域；yearly_range 与 calibrate 使用同一展开规则
SCENE_ALIASES = {
    "yearly_marriage": ("年六亲", "年神煞", "年合会", "年用神"),
    "yearly_career":    ("年十神", "年合会", "年用神", "年大运", "年神煞", "年旺衰"),
    "yearly_wealth":    ("年十神", "年用神", "年合会", "年神煞"),
    "yearly_health":    ("年神煞", "年旺衰", "年合会", "年五行", "年用神", "年十神"),
    "yearly_family":    ("年六亲", "年合会", "年十神"),
    "yearly_study":     ("年六亲", "年十神", "年神煞", "年用神"),
    "yearly_zinv":      ("年六亲", "年合会", "年十神"),
    "yingqi":           ("年合会", "年用神", "年神煞", "年十神", "年五行",
                         "年命宫", "年官禄", "年财帛", "年疾厄", "年福德", "年夫妻", "年子女",
                         "年田宅", "年迁移", "年兄弟", "年仆役"),
}
# 显式单侧域：八字层次/流年八字域（紫微侧无对应 csv 是设计事实）；紫微宫位/流年宫位域（八字侧无）
BAZI_ONLY_RULES = frozenset(BZ_LAYER_RULES + YEAR_BZ_RULES)
ZIWEI_ONLY_RULES = frozenset(ZW_PALACE_RULES + YEAR_PALACE_RULES)
# 需 current_year（当前大运判断）的本命域——仅断语表约束直接用大运因子的规则
CURRENT_DAYUN_RULES = frozenset({"大运"})


def query(rule: str, pan: dict) -> dict:
    """断语查询：域 + 本命盘 → 该域断语 {八字: [...], 紫微: [...]}。

    rule ∈ NATAL_RULES（如 "十神"/"旺衰"/"命宫"/"官禄"；流年域走 yearly_range）。
    pan = full_paipan 返回的完整本命盘；内部从 pan 生成领域快照。
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
    # pan 直通——LLM 传 full_paipan 的返回即可，内部从 pan 直读产因子快照。
    # 校验完整排盘结构，杜绝把空 dict/快照/半截盘误当命盘（防兜底断语污染）。
    validate_natal_pan(pan, action="query")
    current_year = 0
    if rule in CURRENT_DAYUN_RULES:
        current_year, _ = _current_year()
    snapshots = evaluate_snap_from_pan(pan, current_year=current_year)
    return _match_rule(rule, snapshots)

_BRIEF_FIELDS = ("事件", "结论")


def brief(conclusions: list) -> list:
    return [{k: c[k] for k in _BRIEF_FIELDS if k in c}
            for c in conclusions if isinstance(c, dict)]




# time.now RPC 结果 TTL 缓存（秒）——当前大运精度到年，单会话内重复查询不必反复发起 RPC。
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


def _query_year_rules(snapshot, rules, detail=False):
    return query_year_rules(
        snapshot, rules, detail,
        query_yearly=query_yearly,
        brief=brief,
    )


def _yearly_snapshot(pan, year, natal_context=None):
    return yearly_snapshot(
        pan, year, natal_context,
        liunian=_liunian_for_year,
        evaluate_liunian_snap_from_pan=evaluate_liunian_snap_from_pan,
    )


def _evaluate_year(pan, year, rules, detail=False, natal_context=None):
    return yearly_eval.evaluate_year(
        pan, year, rules, detail, natal_context,
        query_yearly=query_yearly,
        brief=brief,
        liunian=_liunian_for_year,
        evaluate_liunian_snap_from_pan=evaluate_liunian_snap_from_pan,
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
    """真值表匹配：因子快照 × 断语表 → 命中条目（按表序）。"""
    hits = []
    for item in table:
        conditions = item.get("约束", {}) or {}
        if not all(_val_match(value, snapshot.get(key)) for key, value in conditions.items()):
            continue
        hits.append(item)
    return [hits[0]] if exclusive and hits else hits


def _match_rule(rule: str, snapshots: dict) -> dict:
    """加载断语表 + 匹配。排盘上下文只参与匹配，不计入因子。"""
    bz_e = load_table(f"bazi_{rule}.csv", required=rule not in ZIWEI_ONLY_RULES)
    zw_e = load_table(f"ziwei_{rule}.csv", required=rule not in BAZI_ONLY_RULES)
    context = snapshots.get("context", {}) or {}
    return {"八字": match(bz_e, {**snapshots["八字"], **context}),
            "紫微": match(zw_e, {**snapshots["紫微"], **context}) if zw_e else []}


def yearly_range(pan: dict, start: int, end: int,
                 rules: list = None, detail: bool = False) -> dict:
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
    natal_context = prepare_natal_context(pan)
    years = {}
    for year in range(start, end + 1):
        try:
            years[str(year)] = _evaluate_year(
                pan, year, resolved_rules, detail, natal_context
            )
        except (RPCError, ConnectionError, TimeoutError, OSError) as e:
            years[str(year)] = {"error": f"{type(e).__name__}: {e}"}
    return {"current_year": cur_year, "current_year_source": cur_source,
            "years": years}
