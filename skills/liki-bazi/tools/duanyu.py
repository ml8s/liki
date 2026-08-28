"""表驱动复合因子求值器（第一层 factor_gen 的核心）。

读 factors.csv（复合因子定义表）+ 内置原子知识（算子字典）→ 对基础因子求值 → 因子快照。
分层：基础算子原语（现/透/旺/弱等稳定语义）在 atoms.py 代码；命理组合知识（五行生克/十神定义/旺衰判断/断语组合规则）在 constants.json + factors.csv + 断语 csv 表。

输入：factors（build_factors 输出的基础因子）+ rpc_data（引擎原始返回：shen_sha/ziwei/da_yun/合会冲）
输出：因子快照 dict（键=因子名，值=0/1 或事实值）
"""
from __future__ import annotations
import os
import sys
from typing import Optional

from atoms import _op, _liu_op, _OP_NAMES, _LIU_OP_NAMES

_FACTORS = None
_FACTOR_DEBUG = os.environ.get("FACTOR_DEBUG") == "1"

_TABLE_CACHE = {}

def load_table(name):
    """断语域表（csv——真值表：列=因子取值，行=断语）。csv 标准库——容器无 PyYAML 兼容。
    断语表 csv 在 tools/bazi/ + tools/ziwei/（csv 只有工具读——谁用归谁；domains 留 md 人读知识）。
    返回 list（无表时返回空 list，与有表同型——调用方无需 isinstance 判断）；模块级缓存。"""
    if name in _TABLE_CACHE:
        return _TABLE_CACHE[name]
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
        _TABLE_CACHE[name] = []   # 该域无表（如纯紫微域无八字表）——空列表，与有表同型
        return []
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
    _TABLE_CACHE[name] = rows
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
    if op in _OP_NAMES:
        v = _op(op, args, factors, gender, chart)
    elif op in _LIU_OP_NAMES:
        v = _liu_op(op, args, factors, gender, chart, ctx)
    else:
        raise ValueError(f"未知算子: {op}")
    if isinstance(v, str):
        # 「任意」= 取值模式（直读[ri_gan_wx,任意] 返回五行字符串、宫含[..,任意] 等）——
        # 返回字符串原值供断语约束匹配（如 `日主五行: 木`）；否则按期望值比较返回 0/1
        if args and args[-1] == "任意":
            return v
        return 1 if args and str(args[0]) == v else 0
    return v

def evaluate_factors(factors: dict, gender: str, chart: dict, shushi: Optional[str] = None) -> dict:
    """因子快照（真值表）：factors.csv 行条件匹配 → 因子值；直通因子=原语计算值。
    shushi="bazi"/"ziwei"：只算本术数因子（八字/紫微真分开——各自快照）；None=全算。
    条件列：带 [] = 原子事实（原语执行）；不带 [] = 因子引用（读快照——多遍稳定）。"""
    rows, cols = load_factor_rows()
    if shushi:
        rows = [r for r in rows if r["术数"] == shushi]
    facts = {}
    snapshot = {}
    errors = []  # 求值失败的因子/原子（键 → 异常描述）——不再静默吞，FACTOR_DEBUG 时打印
    # 因子依赖链最长 len(rows) 层，+1 兜底；无变化即提前收敛（多遍稳定）
    for _pass in range(len(rows) + 1):
        changed = False
        for r in rows:
            if r["因子"] in snapshot:
                continue
            if r["直通"]:
                try:
                    snapshot[r["因子"]] = _atomic(r["直通"], factors, gender, chart)
                except Exception as e:
                    errors.append((r["因子"], f"{type(e).__name__}: {e}"))
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
                            errors.append((col, f"{type(e).__name__}: {e}"))
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
    if errors and _FACTOR_DEBUG:
        for k, msg in errors:
            print(f"[FACTOR_DEBUG] 因子求值失败 {k}: {msg}", file=sys.stderr)
    return snapshot

ALL_DUANYU_RULES = ("study", "marriage", "career", "chushen", "wealth", "health", "personality",
                    "family", "waimao", "shensha", "geju", "zuhe", "ziwei", "tiaohou",
                    "dayun", "tianzhai", "qianyi", "zinv", "zhiye",
                    # 流年断语域（yearly_*——列=流年因子名，流年快照查询专用；本命快照查询不命中）
                    "yearly_marriage", "yearly_family", "yearly_wealth",
                    "yearly_career", "yearly_health", "yearly_study", "yearly_zinv")

# 有效域白名单（含 yingqi——有 CSV 但不在 ALL_DUANYU_RULES 中）
_NATAL_RULES = frozenset(r for r in ALL_DUANYU_RULES if not r.startswith("yearly_")) | {"yingqi"}
_YEARLY_RULES = frozenset(r for r in ALL_DUANYU_RULES if r.startswith("yearly_")) | {"yingqi"}

_LIUNIAN_ROWS = None

def load_liunian_rows():
    """流年因子表（factors_liunian.csv 真值表——同 factors.csv）。模块级缓存。"""
    global _LIUNIAN_ROWS
    if _LIUNIAN_ROWS is not None:
        return _LIUNIAN_ROWS
    import csv as _csv
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "factors", "factors_liunian.csv")
    rows = []
    with open(path, encoding="utf-8") as fh:
        rd = _csv.DictReader(fh)
        cols = [c for c in rd.fieldnames if c not in ("因子", "原语直通", "结论", "术数")]
        for r in rd:
            conds = {}
            for c in cols:
                v = (r.get(c) or "").strip()
                if v:
                    conds[c] = int(v)
            rows.append({"因子": r["因子"], "术数": r.get("术数", "bazi"), "conds": conds})
    _LIUNIAN_ROWS = rows
    return rows

def evaluate_liunian_factors(factors: dict, gender: str, chart: dict, liunian_data: dict,
                             target: str = "配偶星", marriage_bad: int = 0,
                             shi_ke_guan_arg: int = 0, shi_shang_zhong_arg: int = 0,
                             zw_liunian_data: Optional[dict] = None,
                             year: int = 0, shushi: Optional[str] = None,
                             snapshot: Optional[dict] = None) -> dict:
    """流年复合因子（表驱动）：读 factors_liunian.csv 逐行求值 → 流年因子快照。

    与 evaluate_factors 同构——流年因子定义在表，engine 纯机械。
    liunian_data: bazi.liunian 返回（调用方预取）；zw_liunian_data: 紫微流年四化（可选）。
    shi_shang_zhong_arg: 本命"食伤旺"快照值——"引用本命[食伤重]"算子读取（yingqi 损胎/婚变断语）。
    """
    ctx = {
        "liunian": liunian_data or {},
        "zw_liunian": zw_liunian_data or {},
        "target": target,
        "marriage_bad": marriage_bad,
        "shi_ke_guan": shi_ke_guan_arg,
        "shi_shang_zhong": shi_shang_zhong_arg,
        "year": year,
        "chart": chart,
        "factors": factors,
        "snapshot": snapshot or {},  # 本命因子快照（流年支受克等算子读本命五行旺衰——曾缺此键恒 0）
    }
    rows = load_liunian_rows()
    if shushi:
        rows = [r for r in rows if r["术数"] == shushi]
    facts = {}
    snapshot = {}
    errors = []  # 求值失败的因子/原子——不再静默吞
    for _pass in range(len(rows) + 1):
        changed = False
        for r in rows:
            if r["因子"] in snapshot:
                continue
            ok = True
            for col, expect in r["conds"].items():
                if "[" in col:
                    if col not in facts:
                        try:
                            facts[col] = _atomic(col, factors, gender, chart, ctx)
                        except Exception as e:
                            errors.append((col, f"{type(e).__name__}: {e}"))
                            facts[col] = 0
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
    if errors and _FACTOR_DEBUG:
        for k, msg in errors:
            print(f"[FACTOR_DEBUG] 流年因子求值失败 {k}: {msg}", file=sys.stderr)
    return snapshot


# ══════════════ agent 入口（5 工具中的因子生成 2 + 断语查询 1）══════════════

def make_factors(pan: dict) -> dict:
    """因子生成（本命）：本命盘 → 双盘因子快照 {八字: {...}, 紫微: {...}}。

    数据层真分开（八字 190 / 紫微 5 因子各算）、调用层一次返回双盘。
    fac 已在 full_paipan 内嵌（pan["fac"]），此处不再单独 build。
    """
    return {
        "八字": evaluate_factors(pan["fac"], pan["gender"], pan, shushi="bazi"),
        "紫微": evaluate_factors(pan["fac"], pan["gender"], pan, shushi="ziwei"),
    }


def make_liunian_factors(pan: dict, liunian_pan: dict, target: str = "配偶星", year: int = 0) -> dict:
    """因子生成（流年）：本命盘 + 流年盘 → 双盘流年因子快照 {八字: {...}, 紫微: {...}}。

    数据层真分开（八字 23 / 紫微 5 流年因子各算）、调用层一次返回双盘。
    liunian_pan = liunian(pan, 年份) 的返回 {bazi, ziwei}。
    本命婚凶/食伤克官 从本命因子快照取（引用本命），紫微流年四化取自 liunian_pan["ziwei"]。
    """
    bz = evaluate_factors(pan["fac"], pan["gender"], pan, shushi="bazi")
    base = dict(
        factors=pan["fac"], gender=pan["gender"], chart=pan, liunian_data=liunian_pan["bazi"],
        target=target, marriage_bad=bz.get("本命婚凶", 0), shi_ke_guan_arg=bz.get("食伤克官", 0),
        shi_shang_zhong_arg=bz.get("食伤旺", 0),
        zw_liunian_data=liunian_pan["ziwei"], year=year,
        snapshot=bz,
    )
    return {
        "_snapshot_type": "liunian",  # 流年快照标记——query 校验 yearly_* 域需此标记（本命快照隔离）
        "八字": evaluate_liunian_factors(**base, shushi="bazi"),
        "紫微": evaluate_liunian_factors(**base, shushi="ziwei"),
    }


def query(rule: str, pan: dict) -> dict:
    """断语查询：域 + 本命盘 → 该域断语 {八字: [...], 紫微: [...]}。

    rule ∈ _NATAL_RULES（如 "marriage"/"study"/"yingqi"；流年域走 yearly_range）。
    pan = full_paipan 返回的本命盘（含 fac 字段）——内部自动生成因子快照。
    流年查询走 yearly_range，本函数不处理流年快照。
    数据层真分开（各查各表）、调用层一次查双盘——内部 load_table + match 内嵌。
    """
    # 本命函数不允许查流年域——流年走 yearly_range
    if rule.startswith("yearly_"):
        raise ValueError(
            f"query() 仅支持本命域，收到流年域 '{rule}'。"
            f"流年查询请走 yearly_range(pan, start, end, rules)。"
        )
    if rule not in _NATAL_RULES:
        raise ValueError(f"未知本命域 '{rule}'。有效域: {sorted(_NATAL_RULES)}")
    # pan 直通——LLM 传 full_paipan 的返回即可，无需先调 make_factors
    if "fac" in pan:
        snapshots = make_factors(pan)
    elif "八字" in pan and "紫微" in pan:
        snapshots = pan  # 兼容：直接传快照也可
    else:
        raise ValueError(
            "pan 必须是 full_paipan 返回的本命盘（含 fac）。"
            "流年查询请走 yearly_range。"
        )
    return _match_rule(rule, snapshots)


# ── 聚合编排层 ──

_RULE_TARGET_MAP = {
    "yearly_marriage": "配偶星",
    "yearly_career": "官杀",
    "yearly_wealth": "财星",
    "yearly_study": "母星",
    "yearly_health": "日主",
    "yearly_family": "配偶星",
    "yearly_liuqin": "配偶星",
    "yearly_zinv": "子女星",
}

_BRIEF_FIELDS = ("事件", "结论")


def _brief(conclusions: list) -> list:
    return [{k: c[k] for k in _BRIEF_FIELDS if k in c}
            for c in conclusions if isinstance(c, dict)]


def _current_year():
    try:
        from paipan import call
        r = call("time.now", {})
        return int(r["data"]["cst"][:4]), "server"
    except Exception:
        from datetime import datetime
        return datetime.now().year, "local"


def query_yearly(rule: str, snapshots: dict) -> dict:
    if snapshots.get("_snapshot_type") != "liunian":
        raise ValueError("query_yearly 仅接受流年快照（含 _snapshot_type='liunian'）")
    if not rule.startswith("yearly_") and rule != "yingqi":
        raise ValueError(f"query_yearly 仅支持流年域，收到: '{rule}'")
    if rule not in _YEARLY_RULES:
        raise ValueError(f"未知流年域 '{rule}'。有效域: {sorted(_YEARLY_RULES)}")
    return _match_rule(rule, snapshots)


def _match_rule(rule: str, snapshots: dict) -> dict:
    """加载断语表 + 匹配（query / query_yearly 共用）。"""
    from engine import match
    bz_e = load_table(f"bazi_{rule}.csv")
    zw_e = load_table(f"ziwei_{rule}.csv")
    return {"八字": match(bz_e, snapshots["八字"]),
            "紫微": match(zw_e, snapshots["紫微"]) if zw_e else []}


def yearly_range(pan: dict, start: int, end: int,
                 rules: list = None, detail: bool = False) -> dict:
    if rules is None:
        rules = ["yearly_career", "yingqi"]
    for rule in rules:
        if rule not in _YEARLY_RULES:
            raise ValueError(
                f"yearly_range rules 含无效域 '{rule}'。"
                f"有效域: {sorted(_YEARLY_RULES)}")
    from paipan import liunian, RPCError
    cur_year, cur_source = _current_year()
    years = {}
    for year in range(start, end + 1):
        try:
            lnp = liunian(pan, year)
            year_data = {}
            for rule in rules:
                target = _RULE_TARGET_MAP.get(rule, "配偶星")
                snap = make_liunian_factors(pan, lnp, target=target, year=year)
                r = query_yearly(rule, snap)
                if not detail:
                    r = {"八字": _brief(r.get("八字", [])),
                         "紫微": _brief(r.get("紫微", []))}
                year_data[rule] = r
            years[str(year)] = year_data
        except (RPCError, ConnectionError, TimeoutError, OSError) as e:
            years[str(year)] = {"error": f"{type(e).__name__}: {e}"}
    return {"current_year": cur_year, "current_year_source": cur_source,
            "years": years}


def calibrate(candidates: list, events: list, detail: bool = False) -> dict:
    for e in events:
        if not e.get("rule", "").startswith("yearly_"):
            raise ValueError(
                f"calibrate events.rule 必须以 yearly_ 开头，收到: '{e.get('rule')}'")
        if "year" not in e or "label" not in e:
            raise ValueError("calibrate events 每项必须含 year、rule、label")
    labels = [c.get("label", "") for c in candidates]
    if len(labels) != len(set(labels)):
        dupes = [l for l in labels if labels.count(l) > 1]
        raise ValueError(
            f"calibrate candidates label 必须唯一，重复: {set(dupes)}。"
            f"重复 label 会静默覆盖前一个候选的结果。")
    if not candidates:
        raise ValueError("calibrate candidates 不能为空")
    if not events:
        raise ValueError("calibrate events 不能为空")
    from paipan import full_paipan as _fp, liunian as _ln
    results = {}
    for c in candidates:
        label = c.get("label", "")
        if not label:
            raise ValueError("calibrate candidates 每项必须含 label")
        if "longitude" not in c or c.get("longitude") is None:
            raise ValueError(
                f"candidate '{label}' 缺少 longitude，禁止静默降级")
        if "gregorian" not in c or not c.get("gregorian"):
            raise ValueError(f"candidate '{label}' 缺少 gregorian（出生公历时间）")
        if "gender" not in c or c.get("gender") not in ("male", "female"):
            raise ValueError(f"candidate '{label}' 缺少 gender 或值不是 male/female")
        pan = _fp(c["gregorian"], c["gender"],
                  longitude=c["longitude"], correct=c.get("correct", True))
        event_results = []
        for e in events:
            lnp = _ln(pan, e["year"])
            target = _RULE_TARGET_MAP.get(e["rule"], "配偶星")
            snap = make_liunian_factors(pan, lnp, target=target, year=e["year"])
            r = query_yearly(e["rule"], snap)
            event_results.append({
                "year": e["year"], "label": e["label"], "rule": e["rule"],
                "八字": r.get("八字", []), "紫微": r.get("紫微", []),
            })
        results[label] = event_results
    return results
