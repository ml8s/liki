"""因子层 — pan 原语算子 + 真值表求值 + 本命/流年因子快照。

分层：
paipan.py 排盘
factors.py pan → snap（本命/流年因子快照：直读 pan + 真值表复合求值）
duanyu.py snap + 断语表 → 断语

本文件只做机械求值；命理成员关系与分组在 constants.json，因子组合在 CSV。
"""
from __future__ import annotations

import re
from typing import Optional

from errors import FactorEvaluateError
from factor_constants import load_constants
from factor_context import FactorContext, NatalContext
from operators_liunian import _LIU_OP_NAMES, _liu_op
from operators_natal import (
    _OP_NAMES,
    _base_ctx_from_pan,
    _op,
)
from domain_snapshot import project_domain_facts
from factor_tables import load_factor_rows, load_liunian_rows

__all__ = [
    "evaluate_factors", "evaluate_liunian_factors",
    "evaluate_snap_from_pan", "evaluate_liunian_snap_from_pan",
    "prepare_natal_context",
]

def _atomic(col: str, gender, chart, ctx: dict = None, current_year: int = 0):
    """原子执行：列名 "op[arg1,arg2]" → 原语（_op 本命 / _liu_op 流年）。
    字符串值算子：列名参数=期望值——比较返回 1/0。"""
    m = re.match(r'^([^\[]+)\[(.*)\]$', col)
    if m:
        op, argstr = m.group(1), m.group(2)
        args = [int(a) if a.lstrip('-').isdigit() else a for a in argstr.split(',')] if argstr else []
    else:
        op, args = col, []
    if op in _OP_NAMES:
        v = _op(op, args, gender, chart, current_year)
    elif op in _LIU_OP_NAMES:
        v = _liu_op(op, args, gender, chart, ctx)
    else:
        raise FactorEvaluateError(f"未知算子: {op}")
    if isinstance(v, str):
        # 「任意」= 取值模式（直读[ri_gan_wx,任意] 返回五行字符串、宫含[..,任意] 等）——
        # 返回字符串原值供断语约束匹配（如 `日主五行: 木`）；否则按期望值比较返回 0/1
        if args and args[-1] == "任意":
            return v
        # args[-1] 是期望值（如 直读[gender,male] 中 male）；与算子返回值 v 比较
        return 1 if args and str(args[-1]) == v else 0
    return v

def _evaluate_truth_table(rows: list, atomic) -> dict:
    """真值表求值核心：直通行取值，条件行 AND，同因子多行 OR。"""
    facts = {}
    result = {}
    # 因子依赖链最长 len(rows) 层，+1 兜底；无变化即提前收敛（多遍稳定）
    for _pass in range(len(rows) + 1):
        changed = False
        for row in rows:
            name = row["因子"]
            if name in result:
                continue
            if row["直通"]:
                result[name] = atomic(row["直通"])
                changed = True
                continue
            ok = True
            for col, expect in row["conds"].items():
                if "[" in col:
                    if col not in facts:
                        facts[col] = atomic(col)
                    val = facts[col]
                else:
                    val = result.get(col, 0)  # 因子引用
                if val != expect:
                    ok = False
                    break
            if ok:
                result[name] = 1
                changed = True
        if not changed:
            break
    # 补全：所有因子键（未命中=0）——断语表/调用方取值稳定
    for row in rows:
        result.setdefault(row["因子"], 0)
    return result

def evaluate_factors(gender: str, chart: dict, shushi: Optional[str] = None,
                     current_year: int = 0,
                     factor_names: Optional[set[str]] = None) -> dict:
    """因子快照（真值表）：factors.csv 行条件匹配 → 因子值；直通因子=原语计算值。
    shushi="bazi"/"ziwei"：只算本术数因子（八字/紫微真分开——各自快照）；None=全算。
    factor_names 可限定因子闭包；None=该侧全量。调用方仍须自行保证引用闭包完整。
    条件列：带 [] = 原子事实（原语执行）；不带 [] = 因子引用（读快照——多遍稳定）。
    chart=排盘 pan，算子从其直读。"""
    rows = load_factor_rows()
    if shushi:
        side_config = load_constants()["命理侧"]
        if shushi not in side_config["快照代码"]:
            raise FactorEvaluateError(f"evaluate_factors shushi 无效: {shushi}")
        common_code = side_config["公共代码"]
        rows = [r for r in rows if r["术数"] in (shushi, common_code)]
    if factor_names is not None:
        rows = [r for r in rows if r["因子"] in factor_names]
    return _evaluate_truth_table(
        rows,
        lambda expression: _atomic(expression, gender, chart, current_year=current_year),
    )


def _factor_context_from_pan(pan: dict) -> FactorContext:
    """构建一次求值复用的 pan 视图；调用方仍持有原始 pan，不做任何写入。"""
    return FactorContext(pan, _base_ctx_from_pan(pan))


def prepare_natal_context(
    pan: dict, factor_names: Optional[set[str]] = None
) -> NatalContext:
    """准备本命求值上下文与八字快照，供多年流年求值复用。

    yearly_range/calibrate 扫多年时，本命盘不变；基础聚合和本命八字因子
    只需算一次。返回值仅作为内部编排句柄，不进入工具输出。
    factor_names 可按流年引用闭包裁剪本命八字因子；None=全量。
    """
    evaluation = _factor_context_from_pan(pan)
    snapshot = evaluate_factors(
        pan.get("gender", ""), evaluation, shushi="bazi",
        factor_names=factor_names,
    )
    return NatalContext(evaluation=evaluation, snapshot=snapshot)


def evaluate_snap_from_pan(
    pan: dict, current_year: int = 0, sides: Optional[set[str]] = None,
    factor_names: Optional[set[str]] = None
) -> dict:
    """从完整 pan 直接产出领域快照。

    返回 {八字: {...}, 紫微: {...}, context: {...}}；
    sides 可内部限定 bazi/ziwei，未请求侧保留空对象；默认双侧求值。
    内部用 _base_ctx_from_pan 从 pan 构建求值上下文，算子仍经 _op 从 pan 直读求值。
    并透传 pan 的稳定领域事实（纳音/三元/旬空/局数/命身主/空宫 等）进对应盘 snap，供断语消费。

    不做全局内容缓存：agent CLI 一次调用只求值一次，长驻场景由调用方决定生命周期。
    同一 pan 的多域查询在进程内可复用外部传入的 NatalContext。
    """
    gender = pan.get("gender", "")
    context = {"性别": gender}
    if pan.get("solar"):
        context["公历出生"] = pan["solar"]
    if pan.get("lunar"):
        context["农历出生"] = pan["lunar"]
    if current_year:
        context["当前年份"] = current_year
    evaluation = _factor_context_from_pan(pan)
    side_config = load_constants()["命理侧"]
    side_labels = side_config["标签"]
    requested = set(side_config["快照代码"]) if sides is None else set(sides)
    if not requested <= set(side_config["快照代码"]):
        raise FactorEvaluateError(
            f"evaluate_snap_from_pan sides 只支持 {side_config['快照代码']}，收到: {sorted(requested)}"
        )
    requested_factors = None if factor_names is None else set(factor_names)
    snap = {
        **{
            side_labels[code]: evaluate_factors(
                gender, evaluation, shushi=code,
                current_year=current_year,
                factor_names=requested_factors,
            ) if code in requested else {}
            for code in side_config["快照代码"]
        },
        "context": context,
    }
    facts = project_domain_facts(pan)
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
    chart=排盘 pan 或内部 FactorContext；基础上下文不写回调用方 chart。
    """
    evidence: dict = {}
    ctx = {
        "evidence": evidence,
        "liunian": liunian_data or {},
        "zw_liunian": zw_liunian_data or {},
        "year": year,
        "chart": chart,
        "base": _base_ctx_from_pan(chart),
        "snapshot": natal_snapshot or {},  # 本命因子快照（流年支受克等算子读本命五行旺衰）
    }
    rows = load_liunian_rows()
    if shushi:
        side_config = load_constants()["命理侧"]
        if shushi not in side_config["快照代码"]:
            raise FactorEvaluateError(f"evaluate_liunian_factors shushi 无效: {shushi}")
        common_code = side_config["公共代码"]
        rows = [r for r in rows if r["术数"] in (shushi, common_code)]
    if factor_names is not None:
        rows = [r for r in rows if r["因子"] in factor_names]
    result = _evaluate_truth_table(
        rows,
        lambda expression: _atomic(expression, gender, chart, ctx),
    )
    if evidence:
        result["_evidence"] = evidence
    return result


# ══════════════ 快照生成入口 ══════════════

def evaluate_liunian_snap_from_pan(pan: dict, liunian_pan: dict, year: int = 0,
                                   natal_context: Optional[NatalContext] = None,
                                   factor_names: Optional[set[str]] = None) -> dict:
    """流年因子快照（从 pan 直读，无中间层）。

    与直读式流年快照同构；可传入 prepare_natal_context 的结果，在多年扫描中复用
    本命基础聚合与本命八字因子。函数只读 pan，不向其写入私有缓存。
    factor_names 可限定流年因子闭包；None=双侧全量。
    """
    gender = pan.get("gender", "")
    if natal_context is None:
        evaluation = _factor_context_from_pan(pan)
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
    evidence = snap[side_labels["bazi"]].pop("_evidence", {})
    if evidence:
        snap["evidence"] = evidence
    return snap
