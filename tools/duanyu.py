"""表驱动复合因子求值器（第一层 factor_gen 的核心）。

读 factors.csv（复合因子定义表）+ 内置原子知识（算子字典）→ 对基础因子求值 → 因子快照。
engine 只做机械求值：执行算子 + 查常量表——所有命理知识（五行生克/十神定义/旺衰判断/组合规则）在表。

输入：factors（build_factors 输出的基础因子）+ rpc_data（引擎原始返回：shen_sha/ziwei/da_yun/合会冲）
输出：因子快照 dict（键=因子名，值=0/1 或事实值）
"""
from __future__ import annotations
import os
from typing import Optional

_FACTORS = None
_FACTOR_ERRORS = 0
_FACTOR_DEBUG = __import__("os").environ.get("FACTOR_DEBUG") == "1"

def load_table(name):
    """断语域表（csv——真值表：列=因子取值，行=断语）。csv 标准库——容器无 PyYAML 兼容。
    断语表 csv 在 tools/bazi/ + tools/ziwei/（csv 只有工具读——谁用归谁；domains 留 md 人读知识）。"""
    import csv
    fname = name if name.endswith('.csv') else name + '.csv'
    base = os.path.dirname(os.path.abspath(__file__))   # tools（推理机根）
    if fname.startswith('bazi_'):
        path = os.path.join(base, 'bazi', fname[len('bazi_'):])
    elif fname.startswith('ziwei_'):
        path = os.path.join(base, 'ziwei', fname[len('ziwei_'):])
    else:
        path = os.path.join(base, fname)
    if not os.path.exists(path):
        return {"条目": []}   # 该域无八字表（如 ziwei 域——纯紫微）或无紫微表
    rows = []
    with open(path, encoding='utf-8') as fh:
        for r in csv.DictReader(fh):
            cons = {}
            for k, v in r.items():
                if k in ('id', '事件', '结论', '依据', '经典原文'):
                    continue
                vs = (v or '').strip()
                if not vs:
                    continue  # 无关
                try:
                    cons[k] = int(vs)
                except ValueError:
                    cons[k] = vs
            rows.append({'id': r.get('id', ''), '事件': r.get('事件', ''),
                         '约束': cons, '结论': r.get('结论', ''),
                         '依据': r.get('依据', ''), '经典原文': r.get('经典原文', '')})
    return rows

# ══════════════ 原子符号（单一来源：constants.json——代码不硬编码命理数据）══════════════

# ══════════════ 基础算子实现（机械，无领域知识——知识在常量表）══════════════

# ══════════════ 算子执行 ══════════════

# ══════════════ 特殊算子实现（机械，依赖 chart 原始数据）══════════════

# ══════════════ 主求值入口 ══════════════

_FACTOR_ROWS = None
_FACTOR_COLS = None

def load_factor_rows():
    """因子表（factors.csv 真值表）：列=原子事实实例，行=因子（多行=或）。含术数标记（bazi/ziwei——因子快照分）。"""
    global _FACTOR_ROWS, _FACTOR_COLS
    if _FACTOR_ROWS is not None:
        return _FACTOR_ROWS, _FACTOR_COLS
    import csv as _csv
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "factors", "factors.csv")
    rows, cols = [], []
    with open(path, encoding="utf-8") as fh:
        rd = _csv.DictReader(fh)
        cols = [c for c in rd.fieldnames if c not in ("因子", "原语直通", "结论", "术数", "依据")]
        for r in rd:
            conds = {}
            for c in cols:
                v = (r.get(c) or "").strip()
                if v:
                    conds[c] = int(v)
            rows.append({"因子": r["因子"], "直通": (r.get("原语直通") or "").strip(),
                         "术数": (r.get("术数") or "bazi").strip(), "conds": conds})
    _FACTOR_ROWS, _FACTOR_COLS = rows, cols
    return rows, cols

def _atomic(col: str, factors, gender, chart, ctx: dict = None):
    """原子执行：列名 "op[arg1,arg2]" → 原语（_op 本命 / _liu_op 流年）。
    字符串值算子：列名参数=期望值——比较返回 1/0。"""
    import re as _re
    m = _re.match(r'^([^\[]+)\[(.*)\]$', col)
    if m:
        op, argstr = m.group(1), m.group(2)
        args = [int(a) if a.lstrip('-').isdigit() else a for a in argstr.split(',')] if argstr else []
    else:
        op, args = col, []
    try:
        v = _op(op, args, factors, gender, chart)
    except ValueError:
        v = _liu_op(op, args, factors, gender, chart, ctx)
    if isinstance(v, str):
        return 1 if args and str(args[0]) == v else 0
    return v

def evaluate_factors(factors: dict, gender: str, chart: dict, shushi: Optional[str] = None) -> dict:
    """因子快照（真值表）：factors.csv 行条件匹配 → 因子值；直通因子=原语计算值。
    shushi="bazi"/"ziwei"：只算本术数因子（八字/紫微真分开——各自快照）；None=全算。
    条件列：带 [] = 原子事实（原语执行）；不带 [] = 因子引用（读快照——多遍稳定）。"""
    global _FACTOR_ERRORS
    rows, cols = load_factor_rows()
    if shushi:
        rows = [r for r in rows if r["术数"] == shushi]
    facts = {}
    snapshot = {}
    for _pass in range(6):
        changed = False
        for r in rows:
            if r["因子"] in snapshot:
                continue
            if r["直通"]:
                try:
                    snapshot[r["因子"]] = _atomic(r["直通"], factors, gender, chart)
                except Exception as e:
                    _FACTOR_ERRORS += 1
                    if _FACTOR_DEBUG:
                        raise RuntimeError(f"因子[{r['因子']}] 直通求值失败: {e}") from e
                    snapshot[r["因子"]] = 0
                changed = True
                continue
            ok = True
            for col, expect in r["conds"].items():
                if "[" in col:
                    if col not in facts:
                        try:
                            facts[col] = _atomic(col, factors, gender, chart)
                        except Exception as e:
                            _FACTOR_ERRORS += 1
                            if _FACTOR_DEBUG:
                                raise RuntimeError(f"因子[{r['因子']}] 原子[{col}] 求值失败: {e}") from e
                            facts[col] = 0
                    val = facts[col]
                else:
                    val = snapshot.get(col, 0)   # 因子引用
                if val != expect:
                    ok = False
                    break
            if ok:
                snapshot[r["因子"]] = 1
                changed = True
        if not changed:
            break
    # 补全：所有因子键（未命中=0）——断语表/调用方取值稳定
    for r in rows:
        snapshot.setdefault(r["因子"], 0)
    snapshot["gender"] = gender
    return snapshot

ALL_DUANYU_RULES = ("xueye", "marriage", "shiye", "chushen", "caiyun", "jiankang", "xingge",
                    "liuqin", "waimao", "shensha", "geju", "zuhe", "ziwei", "tiaohou",
                    "dayun", "tianzhai", "qianyi", "zinv", "zhiye")

def load_liunian_rows():
    """流年因子表（factors_liunian.csv 真值表——同 factors.csv）。"""
    import csv as _csv
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "factors", "factors_liunian.csv")
    rows = []
    with open(path, encoding="utf-8") as fh:
        rd = _csv.DictReader(fh)
        cols = [c for c in rd.fieldnames if c not in ("因子", "原语直通", "结论")]
        for r in rd:
            conds = {}
            for c in cols:
                v = (r.get(c) or "").strip()
                if v:
                    conds[c] = int(v)
            rows.append({"因子": r["因子"], "conds": conds})
    return rows

def evaluate_liunian_factors(factors: dict, gender: str, chart: dict, liunian_data: dict,
                             target: str = "配偶星", marriage_bad: int = 0,
                             shi_ke_guan_arg: int = 0,
                             zw_liunian_data: Optional[dict] = None,
                             year: int = 0) -> dict:
    """流年复合因子（表驱动）：读 factors_liunian.csv 逐行求值 → 流年因子快照。

    与 evaluate_factors 同构——流年因子定义在表，engine 纯机械。
    liunian_data: bazi.liunian 返回（调用方预取）；zw_liunian_data: 紫微流年四化（可选）。
    """
    ctx = {
        "liunian": liunian_data or {},
        "zw_liunian": zw_liunian_data or {},
        "target": target,
        "marriage_bad": marriage_bad,
        "shi_ke_guan": shi_ke_guan_arg,
        "year": year,
        "chart": chart,
        "factors": factors,
    }
    rows = load_liunian_rows()
    facts = {}
    snapshot = {}
    for _pass in range(6):
        changed = False
        for r in rows:
            if r["因子"] in snapshot:
                continue
            ok = True
            for col, expect in r["conds"].items():
                if "[" in col:
                    if col not in facts:
                        facts[col] = _atomic(col, factors, gender, chart, ctx)
                    val = facts[col]
                else:
                    val = snapshot.get(col, 0)   # 引用本命（因子引用）
                if val != expect:
                    ok = False
                    break
            if ok:
                snapshot[r["因子"]] = 1
                changed = True
        if not changed:
            break
    for r in rows:
        snapshot.setdefault(r["因子"], 0)
    return snapshot

from aggregate import build_factors
from atoms import _op, _liu_op


# ══════════════ agent 入口（5 工具中的因子生成 2 + 断语查询 1）══════════════

def make_factors(pan: dict, shushi: str) -> dict:
    """因子生成（本命）：full_paipan 返回的本命盘 → 因子快照。

    shushi="bazi"→八字因子（190）；"ziwei"→紫微因子。双盘真分开，分别调两次。
    fac 已在 full_paipan 内嵌（pan["fac"]），此处不再单独 build。
    """
    return evaluate_factors(pan["fac"], pan["gender"], pan, shushi=shushi)


def make_liunian_factors(pan: dict, liunian_pan: dict, target: str = "配偶星", year: int = 0) -> dict:
    """因子生成（流年）：本命盘 + 流年盘 → 28 流年因子快照（应期用）。

    liunian_pan = liunian(pan, 年份) 的返回 {bazi, ziwei}。
    本命婚凶/食伤克官 从本命因子快照取（引用本命），紫微流年四化取自 liunian_pan["ziwei"]。
    """
    bz = evaluate_factors(pan["fac"], pan["gender"], pan, shushi="bazi")
    return evaluate_liunian_factors(
        pan["fac"], pan["gender"], pan, liunian_pan["bazi"],
        target=target,
        marriage_bad=bz.get("本命婚凶", 0),
        shi_ke_guan_arg=bz.get("食伤克官", 0),
        zw_liunian_data=liunian_pan["ziwei"],
        year=year,
    )


def query(rule: str, bz_snapshot: dict, zw_snapshot: Optional[dict] = None) -> dict:
    """断语查询：域 + 因子快照 → 该域断语（{八字: [...], 紫微: [...]}）。

    rule ∈ ALL_DUANYU_RULES（19 域，如 "marriage"/"xueye"/"yingqi"）。
    本命断语：query(域, 八字快照, 紫微快照)——双盘；
    应期断语：query("yingqi", 流年快照)——单快照（bazi/ziwei 两张 yingqi 表都匹配同一流年快照）。
    内部 load_table + match 内嵌（agent 只传域 + 快照，不碰表加载）。
    """
    from engine import match
    bz_e = load_table(f"bazi_{rule}.csv")
    bz_e = bz_e.get("条目", bz_e) if isinstance(bz_e, dict) else bz_e
    zw_e = load_table(f"ziwei_{rule}.csv")
    zw_e = zw_e.get("条目", zw_e) if isinstance(zw_e, dict) else zw_e
    out = {"八字": match(bz_e, bz_snapshot)}
    if zw_e:
        out["紫微"] = match(zw_e, zw_snapshot if zw_snapshot is not None else bz_snapshot)
    return out
