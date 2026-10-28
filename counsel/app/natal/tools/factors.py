"""因子层 — natal_chart 原语算子 + 真值表求值 + 本命/流年因子快照。

分层：
charts.py 引擎排盘调用
factors.py natal_chart → snap（本命/流年因子快照：直读 natal_chart + 真值表复合求值）
duanyu.py snap + 断语表 → 断语

本文件只做机械求值；命理成员关系与分组在 constants.json，因子组合在 CSV。
"""

import re
from collections import OrderedDict, deque
from typing import NamedTuple, Optional

from app.natal.tools.errors import FactorEvaluateError
from app.natal.tools.factor_constants import load_constants
from app.natal.tools.factor_context import FactorContext, NatalContext
from app.natal.tools.factor_semantics import FactorValue, PredicateState, all_of, any_of, factor_value, predicate
from app.natal.tools.factor_tables import load_factor_rows, load_liunian_rows
from app.natal.tools.factor_tokens import FACTOR_WILDCARD
from app.natal.tools.natal_projection import project_natal_facts
from app.natal.tools.operators_liunian import _LIU_OP_NAMES, _liu_op
from app.natal.tools.operators_natal import (
    _OP_NAMES,
    _base_ctx_from_natal_chart,
    _op,
)
from app.natal.tools.yongshen import resolve_yongshen

__all__ = [
    "evaluate_operator",
    "evaluate_factors", "evaluate_liunian_factors",
    "evaluate_snap_from_natal_chart", "evaluate_liunian_snap_from_natal_chart",
    "prepare_natal_context",
]


# ══════════════ 真值表编译层 ═══════════════════════════════════
# 编译一次，求值多次。编译只依赖表结构（rows），不依赖用户数据。


class _AtomicPlan(NamedTuple):
    """预解析的原子表达式：bare_operator 或 ops 二选一。"""
    expression: str
    bare_operator: Optional[str]
    ops: tuple  # ((op_name, args_tuple), ...)


class _GroupPlan(NamedTuple):
    """一个 AND 条件组：conditions 内 AND。"""
    conditions: tuple  # ((_AtomicPlan|str, expected), ...)


class _FactorPlan(NamedTuple):
    """一个因子：direct 优先，否则 OR(groups)。"""
    factor_id: str
    direct: Optional[_AtomicPlan]
    groups: tuple  # (_GroupPlan, ...)
    dependencies: frozenset


class _TruthTablePlan:
    """编译后真值表：factors=拓扑序（求值），order=首次出现序（输出）。"""
    __slots__ = ("factors", "by_id", "order")

    def __init__(self, factors: tuple, order: tuple):
        self.factors = factors
        self.order = order
        self.by_id = {f.factor_id: f for f in factors}


_ATOMIC_RE = re.compile(r"([^\[\],]+)\[([^\]]*)\]")


def _compile_atomic(expression: str) -> _AtomicPlan:
    """编译一个表达式：预解析 regex 和参数类型。"""
    atoms = _ATOMIC_RE.findall(expression)
    if not atoms:
        return _AtomicPlan(expression=expression, bare_operator=expression, ops=())
    parsed = []
    for op, argstr in atoms:
        raw_args = argstr.split(",") if argstr else []
        args = tuple(int(a) if a.lstrip("-").isdigit() else a for a in raw_args)
        parsed.append((op, args))
    return _AtomicPlan(expression=expression, bare_operator=None, ops=tuple(parsed))


def _compile_truth_table(rows: list) -> _TruthTablePlan:
    """编译真值表：分组 + 预解析 + 稳定拓扑排序。"""
    grouped: dict[str, list[dict]] = {}
    for row in rows:
        grouped.setdefault(row["因子"], []).append(row)

    plans: dict[str, _FactorPlan] = {}
    for name, group_rows in grouped.items():
        direct_plan = None
        group_plans: list[_GroupPlan] = []
        deps: set[str] = set()
        for row in group_rows:
            if row["直通"]:
                if direct_plan is None:
                    direct_plan = _compile_atomic(row["直通"])
            else:
                conditions = []
                for col, expected in row["conds"].items():
                    if "[" in col:
                        conditions.append((_compile_atomic(col), expected))
                    else:
                        conditions.append((col, expected))
                        deps.add(col)
                group_plans.append(_GroupPlan(conditions=tuple(conditions)))
        plans[name] = _FactorPlan(
            factor_id=name, direct=direct_plan,
            groups=tuple(group_plans), dependencies=frozenset(deps),
        )

    # 稳定拓扑排序（Kahn，FIFO 保持首次出现序的 tiebreak）
    in_degree = {name: 0 for name in grouped}
    dependents: dict[str, list[str]] = {}
    for name, plan in plans.items():
        for dep in plan.dependencies:
            if dep in plans:
                in_degree[name] += 1
                dependents.setdefault(dep, []).append(name)
    queue = deque(name for name in grouped if in_degree[name] == 0)
    topo: list[str] = []
    while queue:
        name = queue.popleft()
        topo.append(name)
        for dep_name in dependents.get(name, ()):
            in_degree[dep_name] -= 1
            if in_degree[dep_name] == 0:
                queue.append(dep_name)
    # 环已在 factor_tables 层校验；此处兜底（不影响已校验表）
    placed = set(topo)
    topo.extend(name for name in grouped if name not in placed)

    return _TruthTablePlan(
        factors=tuple(plans[name] for name in topo),
        order=tuple(grouped),
    )


_PLAN_CACHE: dict[tuple[str, Optional[str]], _TruthTablePlan] = {}

_NATAL_CONTEXT_CACHE: OrderedDict[str, NatalContext] = OrderedDict()
_NATAL_CONTEXT_CACHE_MAX = 8


def _get_plan(source: str, rows: list, shushi: Optional[str]) -> _TruthTablePlan:
    """获取编译后真值表（缓存按 source+shushi）。"""
    key = (source, shushi)
    plan = _PLAN_CACHE.get(key)
    if plan is None:
        if shushi:
            rows = [r for r in rows if r["术数"] == shushi]
        compiled = _compile_truth_table(rows)
        plan = _PLAN_CACHE.setdefault(key, compiled)
    return plan


def _execute_atomic(
    ap: _AtomicPlan, gender: str, chart, ctx: dict | None,
    current_year: int, facts: dict,
) -> FactorValue:
    """求值预解析的原子表达式（facts 按 expression 去重）。"""
    cached = facts.get(ap.expression)
    if cached is not None:
        return cached
    if ap.bare_operator is not None:
        value = evaluate_operator(ap.bare_operator, [], gender, chart,
                                  ctx=ctx, current_year=current_year)
    else:
        states: list[FactorValue] = []
        for op, args in ap.ops:
            value = evaluate_operator(op, list(args), gender, chart,
                                      ctx=ctx, current_year=current_year)
            if isinstance(value, PredicateState):
                states.append(value)
            elif args and args[-1] == FACTOR_WILDCARD:
                states.append(value)
            else:
                expected = (factor_value(args[-1], where=f"原子 {ap.expression} 期望值")
                            if args else PredicateState.HIT)
                states.append(predicate(value, expected))
        value = states[0] if len(states) == 1 else all_of(states)
    facts[ap.expression] = value
    return value


def _evaluate_compiled(
    plan: _TruthTablePlan, gender: str, chart,
    ctx: dict | None = None, current_year: int = 0,
    factor_names: Optional[set[str]] = None,
) -> dict:
    """编译后求值：单 pass 拓扑序，输出按首次出现序。"""
    facts: dict[str, FactorValue] = {}
    values: dict[str, FactorValue] = {}
    for fp in plan.factors:
        if factor_names is not None and fp.factor_id not in factor_names:
            continue
        if fp.direct is not None:
            value = _execute_atomic(fp.direct, gender, chart, ctx, current_year, facts)
        else:
            group_states = []
            for gp in fp.groups:
                cond_states = []
                for parsed, expected in gp.conditions:
                    if isinstance(parsed, _AtomicPlan):
                        actual = _execute_atomic(parsed, gender, chart, ctx, current_year, facts)
                    else:
                        actual = values.get(parsed, PredicateState.MISS)
                    cond_states.append(predicate(actual, expected))
                group_states.append(all_of(cond_states))
            value = any_of(group_states)
        values[fp.factor_id] = value
    return {name: values[name] for name in plan.order if name in values}


def evaluate_operator(op: str, args, gender: str, chart: dict, *,
                      ctx: dict | None = None, current_year: int = 0):
    """执行单个命理算子，并在边界归一到 FactorValue。"""
    if op in _OP_NAMES:
        raw = _op(op, args, gender, chart, current_year)
    elif op in _LIU_OP_NAMES:
        if current_year:
            flow_ctx = dict(ctx or {})
            flow_ctx["year"] = current_year
        else:
            flow_ctx = ctx or {}
        raw = _liu_op(op, args, gender, chart, flow_ctx)
    else:
        raise FactorEvaluateError(f"未知算子: {op}")
    return factor_value(raw, where=f"算子 {op}")





def evaluate_factors(gender: str, chart: dict, shushi: Optional[str] = None,
                     current_year: int = 0,
                     factor_names: Optional[set[str]] = None) -> dict:
    """因子快照（真值表）：factors.csv 行条件匹配 → 因子值；直通因子=原语计算值。"""
    if shushi:
        side_config = load_constants()["命理侧"]
        if shushi not in side_config["快照代码"]:
            raise FactorEvaluateError(f"evaluate_factors shushi 无效: {shushi}")
    plan = _get_plan("natal", load_factor_rows(), shushi)
    evaluation = chart if isinstance(chart, FactorContext) else _factor_context_from_natal_chart(chart)
    return _evaluate_compiled(
        plan, gender, evaluation, current_year=current_year,
        factor_names=factor_names,
    )

def _factor_context_from_natal_chart(
    natal_chart: dict, selected_yongshen: Optional[dict] = None
) -> FactorContext:
    """构建一次求值复用的只读 natal_chart 视图；不修改调用方 natal_chart。"""
    if isinstance(natal_chart, FactorContext):
        return natal_chart
    evaluated_pan = natal_chart
    if selected_yongshen is not None:
        evaluated_pan = {**natal_chart, "_selected_yongshen": selected_yongshen}
    return FactorContext(natal_chart, _base_ctx_from_natal_chart(evaluated_pan))


def _selected_for_pan(natal_chart: dict) -> dict:
    """入口处完成唯一一次用神裁决；下游只消费显式 selected。"""
    if isinstance(natal_chart, FactorContext):
        return natal_chart.base.get("selected_yongshen") or {}
    natal_chart = natal_chart or {}
    selected = natal_chart.get("_selected_yongshen")
    if isinstance(selected, dict):
        return selected
    return resolve_yongshen(natal_chart.get("full") or {})


def prepare_natal_context(
    natal_chart: dict, factor_names: Optional[set[str]] = None
) -> NatalContext:
    """准备本命求值上下文与八字快照，供多年流年求值复用。

    yearly_range/calibrate 扫多年时，本命盘不变；基础聚合和本命八字因子
    只需算一次。返回值仅作为内部编排句柄，不进入工具输出。
    factor_names 可按流年引用闭包裁剪本命八字因子；None=全量。
    """
    from app.natal.tools.chart_integrity import build_natal_chart_digest
    digest = build_natal_chart_digest(natal_chart)
    cache_key = (digest, frozenset(factor_names) if factor_names else None)
    cached = _NATAL_CONTEXT_CACHE.get(cache_key)
    if cached is not None:
        _NATAL_CONTEXT_CACHE.move_to_end(cache_key)
        return cached
    selected = _selected_for_pan(natal_chart)
    evaluation = _factor_context_from_natal_chart(natal_chart, selected)
    snapshot = evaluate_factors(
        natal_chart.get("gender", ""), evaluation, shushi="bazi",
        factor_names=factor_names,
    )
    result = NatalContext(evaluation=evaluation, snapshot=snapshot)
    _NATAL_CONTEXT_CACHE[cache_key] = result
    while len(_NATAL_CONTEXT_CACHE) > _NATAL_CONTEXT_CACHE_MAX:
        _NATAL_CONTEXT_CACHE.popitem(last=False)
    return result


def evaluate_snap_from_natal_chart(
    natal_chart: dict, current_year: int = 0, sides: Optional[set[str]] = None,
    factor_names: Optional[set[str]] = None, current_year_ziwei: Optional[int] = None
) -> dict:
    """从完整 natal_chart 直接产出领域快照。

    返回 {八字: {...}, 紫微: {...}, context: {...}}；
    sides 可内部限定 bazi/ziwei，未请求侧保留空对象；默认双侧求值。
    内部用 _base_ctx_from_natal_chart 从 natal_chart 构建求值上下文，算子仍经 _op 从 natal_chart 直读求值。
    并透传 natal_chart 的稳定领域事实（纳音/三元/旬空/局数/命身主/空宫 等）进对应盘 snap，供断语消费。

    不做全局内容缓存：agent CLI 一次调用只求值一次，长驻场景由调用方决定生命周期。
    同一 natal_chart 的多域查询在进程内可复用外部传入的 NatalContext。
    """
    gender = natal_chart.get("gender", "")
    ziwei_year = current_year if current_year_ziwei is None else current_year_ziwei
    context = {"性别": gender}
    if natal_chart.get("standard_time"):
        context["时间事实"] = {"standard_time": natal_chart["standard_time"]}
    if current_year:
        context["当前年份"] = current_year
    if ziwei_year and ziwei_year != current_year:
        context["当前农历年"] = ziwei_year
    side_config = load_constants()["命理侧"]
    side_labels = side_config["标签"]
    requested = set(side_config["快照代码"]) if sides is None else set(sides)
    if not requested <= set(side_config["快照代码"]):
        raise FactorEvaluateError(
            f"evaluate_snap_from_natal_chart sides 只支持 {side_config['快照代码']}，收到: {sorted(requested)}"
        )
    selected = _selected_for_pan(natal_chart) if "bazi" in requested else None
    evaluation = _factor_context_from_natal_chart(natal_chart, selected)
    requested_factors = None if factor_names is None else set(factor_names)
    snap = {
        **{
            side_labels[code]: evaluate_factors(
                gender, evaluation, shushi=code,
                current_year=(ziwei_year if code == "ziwei" else current_year),
                factor_names=requested_factors,
            ) if code in requested else {}
            for code in side_config["快照代码"]
        },
        "context": context,
    }
    facts = project_natal_facts(natal_chart)
    for code in side_config["快照代码"]:
        snap[side_labels[code]].update(facts[side_labels[code]])
    return snap

def evaluate_liunian_factors(gender: str, chart: dict, liunian_data: dict,
                             zw_liunian_data: Optional[dict] = None,
                             year: int = 0, shushi: Optional[str] = None,
                             natal_snapshot: Optional[dict] = None,
                             factor_names: Optional[set[str]] = None) -> dict:
    """流年复合因子（表驱动）：读 factors_liunian.csv 逐行求值 → 流年因子快照。

    与 evaluate_factors 同构——流年因子定义在表，本函数只做机械求值。
    factor_names 可限定流年因子闭包；None=该侧全量。
    liunian_data: bazi.liunian 返回（调用方预取）；zw_liunian_data: 紫微流年四化。
    chart=排盘 natal_chart 或内部 FactorContext；基础上下文不写回调用方 chart。
    """
    ctx = {
        "liunian": liunian_data or {},
        "zw_liunian": zw_liunian_data or {},
        "year": year,
        "chart": chart,
        "base": _base_ctx_from_natal_chart(chart),
        "snapshot": natal_snapshot or {},  # 本命因子快照（流年支受克等算子读本命五行旺衰）
    }
    if shushi:
        side_config = load_constants()["命理侧"]
        if shushi not in side_config["快照代码"]:
            raise FactorEvaluateError(f"evaluate_liunian_factors shushi 无效: {shushi}")
    plan = _get_plan("liunian", load_liunian_rows(), shushi)
    return _evaluate_compiled(
        plan, gender, chart, ctx=ctx,
        factor_names=factor_names,
    )


# ══════════════ 快照生成入口 ══════════════

def evaluate_liunian_snap_from_natal_chart(natal_chart: dict, liunian_pan: dict, year: int = 0,
                                   natal_context: Optional[NatalContext] = None,
                                   factor_names: Optional[set[str]] = None) -> dict:
    """流年因子快照（从 natal_chart 直读，无中间层）。

    与直读式流年快照同构；可传入 prepare_natal_context 的结果，在多年扫描中复用
    本命基础聚合与本命八字因子。函数只读 natal_chart，不向其写入私有缓存。
    factor_names 可限定流年因子闭包；None=双侧全量。
    """
    gender = natal_chart.get("gender", "")
    if natal_context is None:
        evaluation = _factor_context_from_natal_chart(natal_chart, _selected_for_pan(natal_chart))
        bz = evaluate_factors(gender, evaluation, shushi="bazi")
    else:
        evaluation = natal_context.evaluation
        bz = natal_context.snapshot
    base = dict(
        gender=gender, chart=evaluation, liunian_data=liunian_pan["bazi"],
        zw_liunian_data=liunian_pan["ziwei"], year=year,
        natal_snapshot=bz,
    )
    side_config = load_constants()["命理侧"]
    side_labels = side_config["标签"]
    snap = {
        "_snapshot_type": "liunian",
        **{
            side_labels[code]: evaluate_liunian_factors(
                **base, shushi=code, factor_names=factor_names
            )
            for code in side_config["快照代码"]
        },
        "context": {"性别": gender},
    }
    return snap
