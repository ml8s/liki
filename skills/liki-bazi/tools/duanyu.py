"""断语层 — 因子快照 × 断语真值表 → 命理断语。

输入：factors.py 生成的本命/流年因子快照。
输出：{八字: [...], 紫微: [...]}。
本文件不生成因子；性别作为排盘上下文只参与匹配，不计入因子数。
"""
from __future__ import annotations

import os

from factors import make_factors, make_liunian_factors

_TABLE_CACHE = {}

def load_table(name, required: bool = True):
    """断语域表（csv——真值表：列=因子取值，行=断语）。csv 标准库——容器无 PyYAML 兼容。
    断语表 csv 在 tools/bazi/ + tools/ziwei/（csv 只有工具读——谁用归谁；domains 留 md 人读知识）。
    返回 list；只有显式声明的单侧域可 required=False（缺表为空，不是静默降级）。"""
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
        if required:
            raise FileNotFoundError(f"断语表不存在: {path}")
        return []  # 显式单侧域不缓存缺文件结果，避免后续 required 调用被缓存屏蔽
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

# 命理域规则全集（激进重构：紫微按宫位、八字按命理层次；yly_* 流年保留）
_ZW_PALACE_RULES = ("命宫", "官禄", "财帛", "疾厄", "夫妻", "子女", "迁移",
                    "田宅", "父母", "兄弟", "仆役", "福德")
_ZW_COMMON_RULE = ("格局",)   # 八字也有格局域（双体系共用 rule 名，双侧都有表）
_BZ_LAYER_RULES = ("十神", "旺衰", "用神", "大运", "合会", "神煞", "调候",
                   "五行", "六亲", "出身", "外貌")
# 流年命理域（八字层次 + 紫微宫位，前缀"年"）
_YEAR_PALACE_RULES = ("年命宫", "年官禄", "年财帛", "年疾厄", "年福德",
                      "年夫妻", "年父母", "年子女")
_YEAR_BZ_RULES = ("年十神", "年六亲", "年神煞", "年用神", "年旺衰",
                  "年合会", "年大运", "年五行")

ALL_DUANYU_RULES = _ZW_PALACE_RULES + _ZW_COMMON_RULE + _BZ_LAYER_RULES + \
    _YEAR_PALACE_RULES + _YEAR_BZ_RULES

# 本命命理域
_NATAL_RULES = frozenset(_ZW_PALACE_RULES + _ZW_COMMON_RULE + _BZ_LAYER_RULES)
# 流年命理域
_YEARLY_RULES = frozenset(_YEAR_PALACE_RULES + _YEAR_BZ_RULES)
# 场景别名 → 流年命理域（yearly_range rules 参数兼容旧生活场景名，自动展开为多命理域）
_SCENE_ALIASES = {
    "yearly_marriage": ("年六亲", "年神煞", "年合会", "年用神"),
    "yearly_career":    ("年十神", "年合会", "年用神", "年大运", "年神煞"),
    "yearly_wealth":    ("年十神", "年用神", "年合会", "年神煞"),
    "yearly_health":    ("年神煞", "年旺衰", "年合会", "年五行", "年用神", "年十神"),
    "yearly_family":    ("年六亲", "年合会", "年十神"),
    "yearly_study":     ("年六亲", "年十神", "年神煞", "年用神"),
    "yearly_zinv":      ("年六亲", "年合会", "年十神"),
    "yingqi":           ("年合会", "年用神", "年神煞", "年十神", "年五行"),
}
# 显式单侧域：八字层次/流年八字域（紫微侧无对应 csv 是设计事实）；紫微宫位/流年宫位域（八字侧无）
_BAZI_ONLY_RULES = frozenset(_BZ_LAYER_RULES + _YEAR_BZ_RULES)
_ZIWEI_ONLY_RULES = frozenset(_ZW_PALACE_RULES + _YEAR_PALACE_RULES)
# 需 current_year（当前大运判断）的本命域——仅断语表约束直接用大运因子的规则
_CURRENT_DAYUN_RULES = frozenset({"大运"})


def query(rule: str, pan: dict) -> dict:
    """断语查询：域 + 本命盘 → 该域断语 {八字: [...], 紫微: [...]}。

    rule ∈ _NATAL_RULES（如 "marriage"/"study"；流年域走 yearly_range）。
    pan = full_paipan 返回的本命盘（含 fac 字段）——内部自动生成因子快照。
    流年查询走 yearly_range，本函数不处理流年快照。
    数据层真分开（各查各表）、调用层一次查双盘——内部 load_table + match 内嵌。
    """
    # 本命函数不允许查流年域——流年走 yearly_range
    if rule in _YEARLY_RULES:
        raise ValueError(
            f"query() 仅支持本命域，收到流年域 '{rule}'。"
            f"流年查询请走 yearly_range(pan, start, end, rules)。"
        )
    if rule not in _NATAL_RULES:
        raise ValueError(f"未知本命域 '{rule}'。有效域: {sorted(_NATAL_RULES)}")
    # pan 直通——LLM 传 full_paipan 的返回即可，无需先调 make_factors
    if "fac" in pan:
        current_year = 0
        if rule in _CURRENT_DAYUN_RULES:
            current_year, _ = _current_year()
        snapshots = make_factors(pan, current_year=current_year)
    else:
        raise ValueError(
            "pan 必须是 full_paipan 返回的本命盘（含 fac）。"
            "流年查询请走 yearly_range。"
        )
    return _match_rule(rule, snapshots)

_BRIEF_FIELDS = ("事件", "结论")


def _brief(conclusions: list) -> list:
    return [{k: c[k] for k in _BRIEF_FIELDS if k in c}
            for c in conclusions if isinstance(c, dict)]


def _current_year():
    from paipan import call
    r = call("time.now", {})
    return int(r["data"]["cst"][:4]), "server"


def query_yearly(rule: str, snapshots: dict) -> dict:
    if snapshots.get("_snapshot_type") != "liunian":
        raise ValueError("query_yearly 仅接受流年快照（含 _snapshot_type='liunian'）")
    if rule not in _YEARLY_RULES:
        raise ValueError(f"未知流年命理域 '{rule}'。有效域: {sorted(_YEARLY_RULES)}（场景别名请走 yearly_range 展开）")
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
    bz_e = load_table(f"bazi_{rule}.csv", required=rule not in _ZIWEI_ONLY_RULES)
    zw_e = load_table(f"ziwei_{rule}.csv", required=rule not in _BAZI_ONLY_RULES)
    context = snapshots.get("context", {}) or {}
    return {"八字": match(bz_e, {**snapshots["八字"], **context}),
            "紫微": match(zw_e, {**snapshots["紫微"], **context}) if zw_e else []}


def yearly_range(pan: dict, start: int, end: int,
                 rules: list = None, detail: bool = False) -> dict:
    if rules is None:
        rules = ["yearly_career"]   # 场景别名 → 命理域（默认查流年事业）
    # 先在原始 rules 上校验：非空 + 无重复 + 有效（命理域或场景别名）
    if not rules:
        raise ValueError("yearly_range rules 不能为空")
    if len(rules) != len(set(rules)):
        raise ValueError("yearly_range rules 不能有重复域")
    valid = set(_YEARLY_RULES) | set(_SCENE_ALIASES)
    for rule in rules:
        if rule not in valid:
            raise ValueError(
                f"yearly_range rules 含无效域 '{rule}'。"
                f"有效域: {sorted(_YEARLY_RULES)}，场景别名: {sorted(_SCENE_ALIASES)}")
    # 场景别名展开为命理域（yearly_marriage→年六亲/年神煞/…），去重保序
    resolved = []
    seen = set()
    for r in rules:
        expanded = _SCENE_ALIASES.get(r, (r,))
        for er in expanded:
            if er not in seen:
                seen.add(er); resolved.append(er)
    rules = resolved
    if not rules:
        raise ValueError("yearly_range rules 不能为空")
    if start > end:
        raise ValueError(f"yearly_range start 不能大于 end：{start} > {end}")
    for rule in rules:
        if rule not in _YEARLY_RULES:
            raise ValueError(
                f"yearly_range rules 含无效域 '{rule}'。"
                f"有效域: {sorted(_YEARLY_RULES)}，场景别名: {sorted(_SCENE_ALIASES)}")
    from paipan import liunian, RPCError
    cur_year, cur_source = _current_year()
    years = {}
    for year in range(start, end + 1):
        try:
            lnp = liunian(pan, year)
            year_data = {}
            snap = make_liunian_factors(pan, lnp, year=year)
            for rule in rules:
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
