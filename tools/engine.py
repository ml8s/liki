"""断语库查表器：派生因子计算 + 条目匹配。
用法：
    from duanyu.engine import derive_factors, match
    fac = build_factors(data)                # 原始因子
    df = derive_factors(fac, gender)          # 派生因子（命理语言）
    hits = match(table, df)                   # 命中条目（多行=多面性，按约束数排序）
匹配语义：
- 行内 AND：所有约束命中 → 行成立
- 约束值支持：0/1、字符串枚举、">=N"、"<=N"、any_of（列表内任一）
- 行间：多行成立 = 多面性（不冲突），返回全部并按"约束命中数"降序（具体度优先）
"""
from __future__ import annotations
from typing import Optional
def zw_da_xian(zw: dict, year: int) -> dict:
    """紫微大限：起岁=五行局数、阳男阴女顺行（数组逆时针=-1）、每宫十年、宫干起四化。
    方向与起岁已按 liki-engine ziwei.daxian 校准（0020 火六局6岁起、顺行命→父母→福德）。"""
    gw = zw.get("gong_wei") or []
    if not gw:
        return {}
    order = [pg["name"] for pg in gw]
    yang = (zw.get("nian_gan") or "甲") in "甲丙戊庚壬"
    is_female = zw.get("gender") == "female"
    forward = (is_female == (not yang))  # 阳男阴女顺
    direction = -1 if forward else 1      # 顺行=数组逆时针（-1）
    _JU_CN = {"二": 2, "三": 3, "四": 4, "五": 5, "六": 6, "七": 7, "八": 8, "九": 9, "十": 10, "12": 12, "2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7, "8": 8, "9": 9, "10": 10}
    ju = zw.get("ju_shu") or zw.get("ju_shu_name") or ""
    qi_sui = 0
    for ch in str(ju):
        if ch in _JU_CN:
            qi_sui = _JU_CN[ch]; break
    age = year - int(zw.get("birth_year") or year)
    idx = ((age - qi_sui) // 10) * direction % 12
    gong = order[idx]
    gan = gw[idx]["gan"]
    sh = _ZW_SIHUA.get(gan, ["","","",""])
    lu_gong = ji_gong = None
    for pg in gw:
        stars = [st.get("xing","") for st in pg.get("xing_yao",[])]
        if sh[0] and sh[0] in stars: lu_gong = pg["name"]
        if sh[3] and sh[3] in stars: ji_gong = pg["name"]
    return {"大限宫位": gong, "大限宫干": gan, "大限四化": sh, "禄落": lu_gong, "忌落": ji_gong, "大限年龄": age, "大限起岁": qi_sui}
def _val_match(cond, actual):
    """单约束匹配：支持 0/1、枚举、>=N、<=N、any_of 列表。"""
    if isinstance(cond, list):
        return any(_val_match(c, actual) for c in cond)
    if isinstance(cond, str) and cond.startswith(">="):
        return actual >= int(cond[2:])
    if isinstance(cond, str) and cond.startswith("<="):
        return actual <= int(cond[2:])
    return actual == cond
def match(table: dict, df: dict, exclusive: bool = False) -> list[dict]:
    """查表：返回命中条目（按优先级+约束命中数降序）。table = yaml 加载的 {领域: [条目]} 或 [条目]。
    exclusive=True 时只返回最高优先级的一条（婚姻状态等单一事实域用——命理上状态互斥，不可多面）。"""
    entries = table if isinstance(table, list) else list(table.values())[0]
    hits = []
    for e in entries:
        cons = dict(e.get("约束", {}) or {})   # copy，避免 pop 污染原始条目
        any_groups = cons.pop("any_of", None)
        if any_groups is not None:
            if not any(all(_val_match(cond.get(k), df.get(k)) for k in cond) for cond in any_groups):
                continue
        if not all(_val_match(v, df.get(k)) for k, v in cons.items()):
            continue
        hits.append(e)
    # 权重 = 断语表设计（条目顺序 + 约束内容）——程序不做权重定义/排序
    # 命理：断语表由命理师按确定性排序（铁断/具体断语在前），match 按表顺序返回
    if exclusive and hits:
        # 取表序最前的命中（单一事实域——婚姻状态互斥）
        return [hits[0]]
    return hits
# ================= 应期域：流年派生因子 =================
# 目标星定义（通用事件应期：婚姻/六亲/子女/健康/官非/财运）
