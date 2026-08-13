"""表驱动复合因子求值器（第一层 factor_gen 的核心）。

读 factors.yaml（复合因子定义表）+ 内置原子知识（算子字典）→ 对基础因子求值 → 因子快照。
engine 只做机械求值：执行算子 + 查常量表——所有命理知识（五行生克/十神定义/旺衰判断/组合规则）在表。

输入：fac（build_factors 输出的基础因子）+ rpc_data（引擎原始返回：shen_sha/ziwei/da_yun/合会冲）
输出：因子快照 dict（键=因子名，值=0/1 或事实值）
"""
from __future__ import annotations
import os
from typing import Optional
from dataclasses import dataclass, field, asdict

_CONST = None
_FACT = None
_FACTOR_ERRORS = 0
_FACTOR_DEBUG = __import__("os").environ.get("FACTOR_DEBUG") == "1"



def _load_table(name):
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


def _load_json(name):
    """因子层/字典加载（json——标准库，容器无 PyYAML 兼容）。name 如 "factors.json"/"constants"。"""
    import json as _json
    fname = name if name.endswith('.json') else name + '.json'
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), fname)
    with open(path, encoding="utf-8") as fh:
        return _json.load(fh)


def load_constants() -> dict:
    """复合因子层的字典配置（evaluator 求值 factors.yaml 的参数表——目标星/事件宫位/五行等）。"""
    global _CONST
    if _CONST is None:
        _CONST = _load_json("constants.json")
    return _CONST


# ══════════════ 原子符号（单一来源：constants.json——代码不硬编码命理数据）══════════════
_CONST = load_constants()  # 模块加载即构建（一次读 json——热路径缓存）
_KE = {wx: v["克"] for wx, v in _CONST["五行生克"].items()}
_SHENG = {wx: v["生"] for wx, v in _CONST["五行生克"].items()}
_GAN_WX = _CONST["天干五行"]
_ZHIWX = _CONST["地支五行"]
_TARGET_STARS = _CONST["目标星"]
_TARGET_PALACES = _CONST["事件宫位"]
_TEN_CLASSES = _CONST["类"]
_WANG = set(_CONST["旺衰"]["旺"])
_WEAK = set(_CONST["旺衰"]["弱"])


# ══════════════ 基础算子实现（机械，无领域知识——知识在常量表）══════════════

def _ss(fac, ten: str) -> Optional[dict]:
    """取某十神的状态（factors.py ShishenState）。"""
    return (fac.get("shishen") or {}).get(ten)


def _class_wx(fac, ten_class: str) -> str:
    """十神大类的五行（取该类第一个出现的十神之五行）。"""
    classes = load_constants()["类"]
    for ten in classes.get(ten_class, []):
        st = _ss(fac, ten)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""


def _ten_to_wx(fac, tens: list) -> str:
    """十神列表的五行（聚合）。"""
    for t in tens:
        st = _ss(fac, t)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""


def _wang_shuai(fac) -> dict:
    return fac.get("wuxing", {}).get("wang_shuai", {}) or {}


def _is_wang(fac, wx: str) -> bool:
    wang_cfg = load_constants()["旺衰"]["旺"]
    return _wang_shuai(fac).get(wx) in wang_cfg


def _is_weak(fac, wx: str) -> bool:
    weak_cfg = load_constants()["旺衰"]["弱"]
    return _wang_shuai(fac).get(wx) in weak_cfg


def _resolve_tens(tens, fac, gender):
    """解析十神参数：'配偶星/子女星/父星/母星/官杀/财/印/比劫/食伤' 等大类 → 具体十神列表；'配偶星' 按性别。"""
    const = load_constants()
    classes = const["类"]
    ts = const.get("目标星", {})
    result = []
    for t in tens:
        if t == "配偶星":
            result.extend(ts["配偶星"][gender])
        elif t == "子女星":
            result.extend(ts["子女星"][gender])
        elif t == "父星":
            result.extend(ts["父星"])
        elif t == "母星":
            result.extend(ts["母星"])
        elif t in classes:
            result.extend(classes[t])
        else:
            result.append(t)
    return result


# ══════════════ 算子执行 ══════════════

def _op(op: str, args, fac, gender, rpc) -> int:
    """执行单个算子子句，返回 0/1。args 为参数（列表或标量）。"""
    const = load_constants()
    if not isinstance(args, list):
        args = [args]

    if op == "现":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_ss(fac, t) or {}).get("count", 0) >= 1 for t in tens) else 0
    if op == "透":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_ss(fac, t) or {}).get("tou_gan") for t in tens) else 0
    if op == "藏":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_ss(fac, t) or {}).get("cang_zhi") for t in tens) else 0
    if op == "得地":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_ss(fac, t) or {}).get("has_root") for t in tens) else 0
    if op == "得令":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_ss(fac, t) or {}).get("de_ling") for t in tens) else 0
    if op == "有根":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_ss(fac, t) or {}).get("has_root") for t in tens) else 0
    if op in ("旺", "弱"):
        tens = args
        # 五行名原样（五行旺衰只看月令）；十神/大类/六亲 → 综合旺衰
        if tens and tens[0] in ("木", "火", "土", "金", "水"):
            wx = tens[0]
            return 1 if (_is_wang(fac, wx) if op == "旺" else _is_weak(fac, wx)) else 0
        resolved = _resolve_tens(tens, fac, gender)
        wx = _ten_to_wx(fac, resolved)
        if not wx:
            return 0
        if op == "弱":
            # 十神弱 = 其五行失令 且 不透 且 无根（三者皆弱）
            return 1 if (_is_weak(fac, wx) and not any((_ss(fac, t) or {}).get("tou_gan") for t in resolved)
                         and not any((_ss(fac, t) or {}).get("has_root") for t in resolved)) else 0
        # 十神旺 = 得令（其五行月令旺相）或（透干且有根）——《子平真诠》得令为重，失令者透且有根可补
        if _is_wang(fac, wx):
            return 1
        tou = any((_ss(fac, t) or {}).get("tou_gan") for t in resolved)
        gen = any((_ss(fac, t) or {}).get("has_root") for t in resolved)
        return 1 if tou and gen else 0
    if op == "缺":
        count = fac.get("wuxing", {}).get("count", {}) or {}
        return 1 if args and count.get(args[0], 0) == 0 else 0
    if op == "克":
        # 克(A,B)：A 五行克 B 五行（A/B 可为十神大类或五行）
        a, b = args[0], args[1]
        wx_a = _resolve_wx(fac, gender, a)
        wx_b = _resolve_wx(fac, gender, b)
        if not wx_a or not wx_b:
            return 0
        return 1 if const["五行生克"].get(wx_a, {}).get("克") == wx_b else 0
    if op == "生":
        a, b = args[0], args[1]
        wx_a = _resolve_wx(fac, gender, a)
        wx_b = _resolve_wx(fac, gender, b)
        if not wx_a or not wx_b:
            return 0
        return 1 if const["五行生克"].get(wx_a, {}).get("生") == wx_b else 0
    if op == "直读":
        path, expect = args[0], args[1]
        if path == "ri_gan_wx":
            gan = fac.get("ri_gan", "")
            val = load_constants()["天干五行"].get(gan)
            if expect == "任意":
                return val if val else 0     # 日主五行返回五行字符串（断语约束 `日主五行: 木` 匹配）
            return 1 if str(val) == str(expect) else 0
        val = _path_get(fac, rpc, path)
        if expect == "任意":
            return 1 if val is not None and val != "" else 0
        if expect.startswith("含"):
            target = expect[1:].strip()
            if "或" in target:
                return 1 if val is not None and any(t in str(val) for t in target.split("或")) else 0
            return 1 if val is not None and target in str(val) else 0
        return 1 if str(val) == str(expect) else 0
    if op == "含":
        # 含(shen_sha, 桃花) —— 神煞分布四柱（full.nian/yue/ri/shi），需聚合
        field, value = args[0], args[1]
        if field == "shen_sha":
            names = []
            full = rpc.get("full", {}) or {}
            for zhu in ("nian", "yue", "ri", "shi"):
                for it in (full.get(zhu) or {}).get("shen_sha", []) or []:
                    names.append(it.get("name") or it.get("xing"))
        else:
            items = _path_get(fac, rpc, field) or []
            names = [it.get("name") or it.get("xing") for it in items] if isinstance(items, list) else []
        if "或" in value:
            return 1 if any(v in names for v in value.replace("或", "|").split("|")) else 0
        return 1 if value in names else 0
    if op == "宫含":
        # 宫含(宫位, 星, 条件) —— ziwei gong_wei
        return _zw_gong_op(fac, rpc, args)
    if op == "大运十神":
        return _dayun_op(fac, rpc, args)
    if op == "数量至少":
        # 数量至少(N, 十神...)：十神出现总数 ≥ N（事实计数——印杂等"多"的定量）
        n = int(args[0])
        tens = _resolve_tens(args[1:], fac, gender)
        total = sum((_ss(fac, t) or {}).get("count", 0) for t in tens)
        return 1 if total >= n else 0
    if op == "六合存在":
        # 直接用引擎 fullchart.zhi_liu_he（已算六合 + 化气五行）——不重复计算
        full = rpc.get("full", {}) or {}
        rels = full.get("zhi_liu_he", []) or []
        return 1 if rels else 0
    if op == "合化五行":
        # 本命六合的化气五行（引擎 zhi_liu_he.wuxing）——合化吉凶用
        full = rpc.get("full", {}) or {}
        rels = full.get("zhi_liu_he", []) or []
        wxs = [r.get("wuxing", "") for r in rels if r.get("wuxing")]
        return wxs[0] if wxs else ""
    if op == "官杀取清":
        # 官杀混杂取清（《子平真诠》）：克我（官杀）所在柱被六合/六冲 → 合杀留官/冲去多余 → 取清
        const = load_constants()
        gan_wx = const["天干五行"]
        ke = const["五行生克"]
        chart = rpc.get("chart", {}) or {}
        day_gan = (chart.get("ri") or {}).get("gan", "")
        if not day_gan:
            return 0
        day_wx = gan_wx.get(day_gan, "")
        guan_pillars = []
        for i, zhu in enumerate(("nian", "yue", "ri", "shi")):
            g = (chart.get(zhu) or {}).get("gan", "")
            wx = gan_wx.get(g, "")
            # 克我者=官杀（ke[官杀五行] == 日主五行）
            if wx and day_wx and ke.get(wx, {}).get("克", "") == day_wx:
                guan_pillars.append(i)
        if not guan_pillars:
            return 0
        full = rpc.get("full", {}) or {}
        rels = (full.get("zhi_liu_he") or []) + (full.get("liu_chong") or [])
        involved = set()
        for r in rels:
            involved.add(r.get("pillar_a"))
            involved.add(r.get("pillar_b"))
        return 1 if any(p in involved for p in guan_pillars) else 0
    if op in ("为用", "为忌"):
        # 为用(十神类)：该十神五行 ∈ {用, 喜}；为忌：== 忌（引擎五神体系 yong/xi/ji）
        fy = fac.get("yongshen", {}).get("fu_yi", {}) or {}
        yong, xi, ji = fy.get("yong", ""), fy.get("xi", ""), fy.get("ji", "")
        tens = _resolve_tens(args, fac, gender)
        wx = _ten_to_wx(fac, tens)
        if not wx:
            return 0
        if op == "为用":
            return 1 if wx in (yong, xi) else 0
        return 1 if wx == ji else 0
    if op == "日主十干":
        # 日主具体天干（甲/乙/丙...）——十干性格基调
        return fac.get("ri_gan", "")
    if op == "月支长生":
        # 日主在月支的长生十二态（引擎 chang_sheng 表：长生在寅/沐浴在卯...）
        full = rpc.get("full", {}) or {}
        cs = full.get("chang_sheng", []) or []
        chart = rpc.get("chart", {}) or {}
        yue_zhi = (chart.get("yue") or {}).get("zhi", "")
        for item in cs:
            if item.get("index", "") == yue_zhi:
                return item.get("name", "")
        return ""
    if op == "夫妻宫状态":
        # 日支（夫妻宫）被冲/合/刑——引擎 liu_chong/zhi_liu_he/liu_xing
        full = rpc.get("full", {}) or {}
        chart = rpc.get("chart", {}) or {}
        ri_zhi = (chart.get("ri") or {}).get("zhi", "")
        if not ri_zhi:
            return ""
        involved = set()
        for rel_key in ("liu_chong", "zhi_liu_he", "liu_xing"):
            for r in full.get(rel_key, []) or []:
                involved.add(r.get("zhi_a", ""))
                involved.add(r.get("zhi_b", ""))
        if ri_zhi in involved:
            # 判断具体类型
            for r in full.get("liu_chong", []) or []:
                if ri_zhi in (r.get("zhi_a", ""), r.get("zhi_b", "")):
                    return "冲"
            for r in full.get("zhi_liu_he", []) or []:
                if ri_zhi in (r.get("zhi_a", ""), r.get("zhi_b", "")):
                    return "合"
            return "刑"
        return "静"
    if op == "日支类型":
        # 日支四桃花/四驿马/四墓库——配偶特征（查 constants 日支神煞表）
        chart = rpc.get("chart", {}) or {}
        ri_zhi = (chart.get("ri") or {}).get("zhi", "")
        const = load_constants()
        for cat, zhis in const["日支神煞"].items():
            if ri_zhi in zhis:
                return cat
        return ""
    if op == "财库现":
        # 日干所克=财星五行 → 墓库支（金库丑/木库未/水库辰/火库戌/土库辰）→ 在命局四柱支中
        const = load_constants()
        day_gan = fac.get("ri_gan", "")
        day_wx = const["天干五行"].get(day_gan, "")
        cai_wx = const["五行生克"].get(day_wx, {}).get("克", "")
        ku = const["墓库"].get(cai_wx, "")
        if not ku:
            return 0
        chart = rpc.get("chart", {}) or {}
        for zhu in ("nian", "yue", "ri", "shi"):
            if (chart.get(zhu) or {}).get("zhi", "") == ku:
                return 1
        return 0
    if op == "财星入墓":
        # 财星坐墓库支（财藏库中——蓄财之象）
        const = load_constants()
        day_gan = fac.get("ri_gan", "")
        day_wx = const["天干五行"].get(day_gan, "")
        cai_wx = const["五行生克"].get(day_wx, {}).get("克", "")
        ku = const["墓库"].get(cai_wx, "")
        if not ku:
            return 0
        chart = rpc.get("chart", {}) or {}
        # 财星（正偏财）所在柱支 = 墓库支
        for k in ("正财", "偏财"):
            st = fac.get("shishen", {}).get(k)
            if st and st.get("count", 0):
                for zhu in ("nian", "yue", "ri", "shi"):
                    c = chart.get(zhu) or {}
                    if c.get("zhi", "") == ku:
                        return 1
        return 0
    if op == "克者旺":
        # A 的五行被 B 克（B=克者），B 得令而旺
        resolved = _resolve_tens(args, fac, gender)
        wx = _ten_to_wx(fac, resolved)
        if not wx:
            return 0
        ke_wx = [k for k, v in load_constants()["五行生克"].items() if v.get("克") == wx]
        return 1 if ke_wx and _is_wang(fac, ke_wx[0]) else 0
    if op == "事实计数":
        # 印星三关数：得令/不被财克/有根（0-3）——《渊海子平》印绶三关
        yin = (_ss(fac, "正印") or _ss(fac, "偏印"))
        if not yin or not (yin.get("tou_gan") or yin.get("cang_zhi")):
            return 0
        n = 0
        if yin.get("de_ling"):
            n += 1
        if yin.get("has_root"):
            n += 1
        cai = _ss(fac, "正财") or _ss(fac, "偏财")
        ke = bool(cai and cai.get("tou_gan") and yin.get("wuxing")
                  and load_constants()["五行生克"].get(cai.get("wuxing", ""), {}).get("克") == yin.get("wuxing"))
        if not ke:
            n += 1
        return n
    if op == "格神透":
        return _ge_shen_tou(fac, rpc)
    if op == "格神根":
        return _ge_shen_gen(fac, rpc)
    if op == "年柱官杀":
        return _nian_guan(fac, rpc)
    if op == "日支冲刑害":
        return _palace_bad(fac, rpc)
    if op == "大运窗口":
        return _dayun_window(fac, rpc, args)
    # 流年算子（流年透/值/合/冲/克/忌神/财坏印/大运窗口/换运/岁运并临/干合等）由 _liu_op 处理
    raise ValueError(f"未知算子: {op}")


def _resolve_wx(fac, gender, arg):
    """解析任意参数为五行：具体五行名原样；十神/大类 → 五行。"""
    const = load_constants()
    if arg in ("木", "火", "土", "金", "水"):
        return arg
    if arg in const.get("类", {}):
        return _class_wx(fac, arg)
    if _ss(fac, arg):
        return (_ss(fac, arg) or {}).get("wuxing")
    # 配偶星等
    resolved = _resolve_tens([arg], fac, gender)
    return _ten_to_wx(fac, resolved)


def _path_get(fac, rpc, path: str):
    """按路径取值：优先 fac（基础因子），其次 rpc 原始数据。"""
    obj = fac if path.startswith(("shishen", "wuxing", "qiangruo", "yongshen", "ri_gan")) else rpc
    cur = obj
    for part in path.split("."):
        if isinstance(cur, dict):
            cur = cur.get(part)
        else:
            return None
    return cur


# ══════════════ 特殊算子实现（机械，依赖 rpc 原始数据）══════════════

def _zw_gong_op(fac, rpc, args):
    """紫微宫位查询：宫含(宫位, 星, 条件)。

    四化（化禄/权/科/忌）用顶层 si_hua（{星:四化}）反推落宫——引擎本命四化无宫位字段，需按星找宫。
    """
    gong_name, star, cond = args[0], args[1], args[2] if len(args) > 2 else "任意"
    zw = rpc.get("ziwei", {}) or {}
    gw = zw.get("gong_wei", []) or []
    # 化禄/权/科/忌：顶层 si_hua → 星 → 落宫
    if cond in ("化禄", "化权", "化科", "化忌"):
        sihua_map = {"化禄": "禄", "化权": "权", "化科": "科", "化忌": "忌"}
        target = sihua_map[cond]
        top_sihua = zw.get("si_hua", {}) or {}
        # 找该四化的星
        huaji_stars = [k for k, v in top_sihua.items() if v == target]
        if star != "任意":
            if star not in huaji_stars:
                return 0
        for g in gw:
            gname = g.get("name", "")
            if gong_name not in gname and gname not in gong_name:
                continue
            stars = [st.get("xing", "") for st in g.get("xing_yao", [])]
            if any(hs in stars for hs in huaji_stars):
                return 1
        return 0
    for g in gw:
        gname = g.get("name", "")
        if gong_name not in gname and gname not in gong_name:
            continue
        stars = [st.get("xing", "") for st in g.get("xing_yao", [])]
        if star == "煞星":
            sha = set(load_constants()["紫微煞星"])
            return 1 if sha & set(stars) else 0
        if star == "无主星":
            main_stars = set(load_constants()["紫微主星"])
            return 1 if not (main_stars & set(stars)) else 0
        if cond == "落陷":
            return 1 if any(st.get("xing") == star and st.get("liang_du") in ("陷", "平") for st in g.get("xing_yao", [])) else 0
        if cond == "唯一主星":
            main_stars = set(load_constants()["紫微主星"])
            present = [s for s in stars if s in main_stars]
            return 1 if present == [star] else 0
        return 1 if star in stars else 0
    return 0


def _dayun_op(fac, rpc, args):
    """大运十神查询：大运十神(当前/年, 星类)。"""
    when, star_class = args[0], args[1]
    dx = (rpc.get("full") or rpc.get("chart") or {}).get("da_yun", {}) or {}
    steps = dx.get("steps", [])
    idx = dx.get("current_step_index", -1)
    if when == "当前" and 0 <= idx < len(steps):
        shi_shen = (steps[idx].get("shi_shen", "") or "").replace("运", "")
        resolved = _resolve_tens([star_class], fac, rpc.get("gender"))
        return 1 if shi_shen in resolved else 0
    return 0


def _dayun_window(fac, rpc, args):
    """大运窗口：当前大运十神 ∈ 配偶星。"""
    return _dayun_op(fac, rpc, ["当前", args[0] if args else "配偶星"])


def _ge_shen_tou(fac, rpc):
    """格神透干：月令本气（格神）透干。"""
    full = rpc.get("full", {}) or {}
    yue = full.get("yue", {}) or {}
    ge_ju = (fac.get("yongshen") or {}).get("ge_ju", {}) or {}
    ge_name = ge_ju.get("ge_ju", "")
    if not ge_name:
        return 0
    # 格神 = 月令本气十神；查月柱 shi_shens 是否有 source=stem
    for ss in yue.get("shi_shens", []):
        if ss.get("source") == "stem":
            return 1
    return 0


def _ge_shen_gen(fac, rpc):
    full = rpc.get("full", {}) or {}
    yue = full.get("yue", {}) or {}
    for ss in yue.get("shi_shens", []):
        if ss.get("source") in ("main_qi", "mid", "minor"):
            return 1
    return 0


def _nian_guan(fac, rpc):
    """年柱官杀攻身：年柱 shi_shens 含官杀。"""
    full = rpc.get("full", {}) or {}
    nian = full.get("nian", {}) or {}
    for ss in nian.get("shi_shens", []):
        if ss.get("shi_shen") in ("正官", "七杀"):
            return 1
    return 0


def _palace_bad(fac, rpc):
    """宫破：日支逢六冲/相刑/六害。"""
    full = rpc.get("full", {}) or {}
    ri_zhi = (full.get("ri", {}) or {}).get("zhi", "")
    if not ri_zhi:
        return 0
    for key in ("liu_chong", "liu_xing", "liu_hai"):
        for rel in full.get(key, []) or []:
            if ri_zhi in (rel.get("a"), rel.get("b"), rel.get("zhi1"), rel.get("zhi2")):
                return 1
    return 0


# ══════════════ 主求值入口 ══════════════

_FACTOR_ROWS = None
_FACTOR_COLS = None


def _load_factor_rows():
    """因子表（factors.csv 真值表）：列=原子事实实例，行=因子（多行=或）。含术数标记（bazi/ziwei——因子快照分）。"""
    global _FACTOR_ROWS, _FACTOR_COLS
    if _FACTOR_ROWS is not None:
        return _FACTOR_ROWS, _FACTOR_COLS
    import csv as _csv
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "factors.csv")
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


def _atomic(col: str, fac, gender, rpc):
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
        v = _op(op, args, fac, gender, rpc)
    except ValueError:
        v = _liu_op(op, args, fac, gender, rpc)
    if isinstance(v, str):
        return 1 if args and str(args[0]) == v else 0
    return v


def evaluate_factors(fac: dict, gender: str, rpc: dict, shushi: Optional[str] = None) -> dict:
    """因子快照（真值表）：factors.csv 行条件匹配 → 因子值；直通因子=原语计算值。
    shushi="bazi"/"ziwei"：只算本术数因子（八字/紫微真分开——各自快照）；None=全算。
    条件列：带 [] = 原子事实（原语执行）；不带 [] = 因子引用（读快照——多遍稳定）。"""
    global _FACTOR_ERRORS
    rows, cols = _load_factor_rows()
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
                    snapshot[r["因子"]] = _atomic(r["直通"], fac, gender, rpc)
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
                            facts[col] = _atomic(col, fac, gender, rpc)
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


def evaluate_from_rpc(data: dict, liunian_years: Optional[list] = None, gender: Optional[str] = None,
                       domain_filter: Optional[list] = None) -> dict:
    """生成器唯一入口（纯——rpc 排盘数据 → 断语；不依赖 RPC 客户端）。

    data: 排盘结果（full_panchang 输出——rpc_data；由调用方排盘后传入）。
    gender: 可省略（data 内含）。

    返回：
    - gender: 性别
    - snapshot: 复合因子快照（如需自定义查表）
    - domains: {断语域: [命中断语条目]}——全部断语域
    - liunian: {选项年份: {候选断语}}——应期题（婚动/婚变/凶事——命中即候选）
    """
    from engine import match
    from client import bazi_liunian  # 应期流年排盘（client 为排盘工具——tests/ 承载）
    if gender is None:
        # 从排盘数据取性别（full_panchang 返回含 gender 或 chart 推断）
        gender = data.get("gender") or _infer_gender(data)
    fac = build_factors(data)
    # 真分开：八字因子快照 + 紫微因子快照（各自独立——各表只查本术数快照）
    bz_snapshot = evaluate_factors(fac, gender, data, shushi="bazi")
    zw_snapshot = evaluate_factors(fac, gender, data, shushi="ziwei")
    domains = {}
    rules = ALL_DUANYU_RULES if not domain_filter else [d for d in ALL_DUANYU_RULES if d in domain_filter]
    for rule in rules:
        bz = _load_table(f"bazi_{rule}.csv")
        bz_entries = bz.get("条目", []) if isinstance(bz, dict) else bz
        zw = _load_table(f"ziwei_{rule}.csv")
        zw_entries = zw.get("条目", []) if isinstance(zw, dict) else zw
        entry = {"八字": match(bz_entries, bz_snapshot)}
        if zw_entries:
            entry["紫微"] = match(zw_entries, zw_snapshot)
        domains[rule] = entry
    out = {"gender": gender, "snapshot": bz_snapshot, "snapshot_ziwei": zw_snapshot, "domains": domains,
           "因子错误数": _FACTOR_ERRORS}
    if liunian_years:
        chart = data.get("chart", {})
        yt = _load_table("bazi_yingqi.csv")
        yt_entries = yt.get("条目", []) if isinstance(yt, dict) else yt
        out["liunian"] = {}
        zwyt = _load_table("ziwei_yingqi.csv")
        zwyt_entries = zwyt.get("条目", []) if isinstance(zwyt, dict) else zwyt
        for y in liunian_years:
            ln = bazi_liunian(chart, y)
            fl = evaluate_liunian_factors(fac, gender, data, ln, target="配偶星", year=y)
            hits = [h for h in match(yt_entries, fl) if h.get("事件") in ("婚动", "婚变", "凶事")]
            entry = {"断语": [h["结论"] for h in hits]} if hits else {}
            # 紫微流年四化事件（agent 辅助参考——不程序候选）
            if zwyt_entries:
                zh = [h for h in match(zwyt_entries, fl) if h.get("事件")]
                if zh:
                    entry["紫微流年"] = [h["结论"] for h in zh]
            if entry:
                out["liunian"][y] = entry
    return out



def _infer_gender(data: dict) -> str:
    """从排盘数据推断性别（full_panchang 的 chart 里或外层）。"""
    for key in ("gender", "sex", "xingbie"):
        v = data.get(key)
        if v:
            return v
    chart = data.get("chart", {}) or {}
    for key in ("gender", "sex"):
        v = chart.get(key)
        if v:
            return v
    return "male"



def _load_liunian_rows():
    """流年因子表（factors_liunian.csv 真值表——同 factors.csv）。"""
    import csv as _csv
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "factors_liunian.csv")
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


def evaluate_liunian_factors(fac: dict, gender: str, rpc: dict, liunian_data: dict,
                             target: str = "配偶星", marriage_bad: int = 0,
                             shi_ke_guan_arg: int = 0,
                             zw_liunian_data: Optional[dict] = None,
                             year: int = 0) -> dict:
    """流年复合因子（表驱动）：读 factors_liunian.yaml 逐行求值 → 流年因子快照。

    与 evaluate_factors 同构——流年因子定义在表，engine 纯机械。
    liunian_data: bazi.liunian 返回（调用方预取）；zw_liunian_data: 紫微流年四化（可选）。
    """
    global _LIU_CTX
    _LIU_CTX = {
        "liunian": liunian_data or {},
        "zw_liunian": zw_liunian_data or {},
        "target": target,
        "marriage_bad": marriage_bad,
        "shi_ke_guan": shi_ke_guan_arg,
        "year": year,
        "rpc": rpc,
        "fac": fac,
    }
    rows = _load_liunian_rows()
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
                        facts[col] = _atomic(col, fac, gender, rpc)
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


def _liu_op(op: str, args, fac, gender, rpc) -> int:
    """流年算子（读 _LIU_CTX 上下文）。"""
    ctx = _LIU_CTX
    ln = ctx.get("liunian", {})
    target = ctx.get("target", "配偶星")
    year = ctx.get("year", 0)
    const = load_constants()
    star_keys = _target_stars(target, gender, const)
    nz = ln.get("nian_zhi", "")
    nian_gan = ln.get("nian_gan", "")
    ss_year = ln.get("shi_shen", "")

    if op == "流年透":
        return 1 if ss_year in star_keys else 0
    if op in ("流年值", "流年合", "流年冲"):
        palace_key = const.get("事件宫位", {}).get(target, "ri")
        ri_zhi = fac.get("palace_ri", {}).get("zhi", "")
        chart = rpc.get("chart", {}) or {}
        palace_zhi = {"yue": chart.get("yue", {}).get("zhi", ri_zhi),
                      "shi": chart.get("shi", {}).get("zhi", ri_zhi),
                      "nian": chart.get("nian", {}).get("zhi", ri_zhi)}.get(palace_key, ri_zhi)
        if op == "流年值":
            return 1 if nz == palace_zhi else 0
        zhi_he = zhi_chong = 0
        for it in ln.get("natal_interactions", []):
            for zr in it.get("zhi_rels", []):
                za, zb = zr.get("zhi_a", ""), zr.get("zhi_b", "")
                other = zb if za == nz else (za if zb == nz else "")
                if other == palace_zhi:
                    t = zr.get("type", "")
                    if t in const["合类"]:
                        zhi_he = 1
                    elif t in const["冲类"]:
                        zhi_chong = 1
        return zhi_he if op == "流年合" else zhi_chong
    if op == "流年克":
        _GANWX = const["天干五行"]
        _ZHIWX = const["地支五行"]
        target_wx = None
        if target == "日主":
            target_wx = _GANWX.get(fac.get("ri_gan", ""), "")
        else:
            for k in star_keys:
                st = fac.get("shishen", {}).get(k)
                if st and st.get("wuxing"):
                    target_wx = st["wuxing"]
                    break
        if target_wx and nian_gan:
            ke = const["五行生克"].get(_GANWX.get(nian_gan, ""), {}).get("克")
            ke2 = const["五行生克"].get(_ZHIWX.get(nz, ""), {}).get("克")
            return 1 if ke == target_wx or ke2 == target_wx else 0
        return 0
    if op == "忌神干":
        ji = fac.get("yongshen", {}).get("fu_yi", {}).get("ji", "")
        ji_wx = const["天干五行"].get(ji, "") or ji   # ji 可能是五行名（如"火"）或天干名——转五行
        return 1 if (ji_wx and const["天干五行"].get(nian_gan, "") == ji_wx) else 0
    if op == "忌神支":
        ji = fac.get("yongshen", {}).get("fu_yi", {}).get("ji", "")
        ji_wx = const["天干五行"].get(ji, "") or ji
        return 1 if (ji_wx and const["地支五行"].get(nz, "") == ji_wx) else 0
    if op == "财坏印流年":
        _YIN2 = {"正印", "偏印"}
        if nz and nian_gan and ss_year in _YIN2:
            ke = const["五行生克"].get(const["地支五行"].get(nz, ""), {}).get("克")
            return 1 if ke == const["天干五行"].get(nian_gan, "") else 0
        return 0
    if op == "大运窗口流年":
        birth_y = int(fac.get("_birth_year") or 0) or 0
        age = (year - birth_y + 1) if birth_y else 0
        for s in fac.get("dayun_steps", []):
            if any(k in s.get("shi_shen", "") for k in star_keys) and s["qi_sui"] <= age < s["zhi_sui"]:
                return 1
        return 0
    if op == "换运流年":
        birth_y = int(fac.get("_birth_year") or 0) or 0
        age = (year - birth_y + 1) if birth_y else 0
        for s in fac.get("dayun_steps", []):
            if any(k in s.get("shi_shen", "") for k in star_keys):
                if age in (s["qi_sui"], s["qi_sui"] + 1) and birth_y and year >= birth_y:
                    return 1
        return 0
    if op == "流年宫忌":
        gong = args[0]
        sg = ctx.get("zw_liunian", {}).get("si_hua_gong", {}) or {}
        sh = ctx.get("zw_liunian", {}).get("si_hua", {}) or {}
        for star, gname in sg.items():
            if gname == gong and sh.get(star) == "忌":
                return 1
        return 0
    if op == "引用本命":
        key = args[0]
        if key == "本命婚凶":
            return ctx.get("marriage_bad", 0)
        if key == "食伤克官":
            return ctx.get("shi_ke_guan", 0)
        return 0
    # ── 机械原子（查 constants 表/比较——组合定义在表）──
    if op == "干支相等":
        # 干支相等[来源A,来源B]——来源：大运/流年/日柱
        ga, gb = _src_gz(args[0]), _src_gz(args[1])
        return 1 if (ga and ga == gb) else 0
    if op == "干克":
        # 干克[干A来源,干B来源]——来源：流年干/日干
        g1, g2 = _src_gan(args[0]), _src_gan(args[1])
        if not g1 or not g2:
            return 0
        ke = const["五行生克"].get(const["天干五行"].get(g1, ""), {}).get("克")
        return 1 if (ke and ke == const["天干五行"].get(g2, "")) else 0
    if op == "支冲":
        # 支冲[支A来源,支B来源]——来源：流年支/日支
        z1, z2 = _src_zhi(args[0]), _src_zhi(args[1])
        return 1 if (z1 and z2 and const["六冲"].get(z1) == z2) else 0
    if op == "三刑":
        # 三刑[支来源...]——命局四柱支（含流年）凑齐三刑组
        zhis = set()
        for a in args:
            zv = _src_zhi(a)
            if zv:
                zhis.add(zv)
        chart = rpc.get("chart", {}) or {}
        for zhu in ("nian", "yue", "ri", "shi"):
            z = (chart.get(zhu) or {}).get("zhi", "")
            if z:
                zhis.add(z)
        for grp in const["三刑"]:
            if grp and all(g in zhis for g in grp[0]):
                return 1
        return 0
    if op == "旬空":
        # 旬空[日柱干支,流年支]——日柱所在旬的空亡支是否含流年支
        gz = _src_gz(args[0])
        nz2 = _src_zhi(args[1])
        if not gz or len(gz) < 2 or not nz2:
            return 0
        day_g, day_z = gz[0], gz[1]
        _GAN_ORDER = list(const["天干五行"].keys())
        _ZHI_ORDER = list(const["地支五行"].keys())
        if day_g not in _GAN_ORDER or day_z not in _ZHI_ORDER:
            return 0
        diff = (_ZHI_ORDER.index(day_z) - _GAN_ORDER.index(day_g)) % 12
        xun_gan = _GAN_ORDER[_GAN_ORDER.index(day_g) - diff]
        xun = "甲" + xun_gan[1:]
        return 1 if (xun in const["旬空"] and nz2 in const["旬空"][xun]) else 0
    if op == "流年支受克":
        # 流年支受克：流年支五行被本命旺相五行（木旺/火旺/土旺/金旺/水旺）所克——本命亢+流年受克→健康凶年
        ln = _LIU_CTX.get("liunian", {})
        nz = ln.get("nian_zhi", "")
        if not nz:
            return 0
        snap = _LIU_CTX.get("snapshot", {})
        const = load_constants()
        zhi_wx = const["地支五行"].get(nz, "")
        if not zhi_wx:
            return 0
        for wx in ("木", "火", "土", "金", "水"):
            if snap.get(f"{wx}旺") and const["五行生克"].get(wx, {}).get("克") == zhi_wx:
                return 1
        return 0
    if op == "年柱干伏吟":
        # 年柱干伏吟：流年天干 == 年柱天干（父母宫/祖上宫伏吟——主家变/父母变动）
        ln = _LIU_CTX.get("liunian", {})
        chart = _LIU_CTX.get("rpc", {}).get("chart", {}) or {}
        return 1 if (ln.get("nian_gan") and ln.get("nian_gan") == (chart.get("nian") or {}).get("gan", "")) else 0
    if op == "天干合":
        # 天干合[干A来源,干B来源]——查五合表
        g1, g2 = _src_gan(args[0]), _src_gan(args[1])
        return 1 if (g1 and g2 and const["天干五合"].get(g1) == g2) else 0
    return 0


def _current_dayun_gz() -> str:
    """当前大运干支（机械——查大运步骤+虚岁）。"""
    fac2 = _LIU_CTX.get("fac", {})
    year = _LIU_CTX.get("year", 0)
    birth_y = int(fac2.get("_birth_year") or 0) or 0
    age = (year - birth_y + 1) if birth_y else 0
    for s in fac2.get("dayun_steps", []):
        if s["qi_sui"] <= age < s["zhi_sui"]:
            return s.get("name", "")
    return ""


def _src_gz(src: str) -> str:
    """干支来源解析：大运/流年/日柱 → 干支。"""
    ln = _LIU_CTX.get("liunian", {})
    chart = _LIU_CTX.get("rpc", {}).get("chart", {}) or {}
    if src == "流年":
        return ln.get("nian_gan", "") + ln.get("nian_zhi", "")
    if src == "大运":
        return _current_dayun_gz()
    if src == "日柱":
        return (chart.get("ri") or {}).get("gan", "") + (chart.get("ri") or {}).get("zhi", "")
    return ""


def _src_gan(src: str) -> str:
    """干来源：流年干/日干。"""
    ln = _LIU_CTX.get("liunian", {})
    fac2 = _LIU_CTX.get("fac", {})
    if src == "流年干":
        return ln.get("nian_gan", "")
    if src == "日干":
        return fac2.get("ri_gan", "")
    return ""


def _src_zhi(src: str) -> str:
    """支来源：流年支/日支/时支。"""
    ln = _LIU_CTX.get("liunian", {})
    chart = _LIU_CTX.get("rpc", {}).get("chart", {}) or {}
    fac2 = _LIU_CTX.get("fac", {})
    if src == "流年支":
        return ln.get("nian_zhi", "")
    if src == "日支":
        return fac2.get("palace_ri", {}).get("zhi", "") or (chart.get("ri") or {}).get("zhi", "")
    if src == "时支":
        return (chart.get("shi") or {}).get("zhi", "")
    if src == "年支":
        return (chart.get("nian") or {}).get("zhi", "")
    return ""


def _target_stars(target: str, gender: str, const: dict) -> tuple:
    ts = const.get("目标星", {}).get(target, ())
    if isinstance(ts, dict):
        return tuple(ts.get(gender, ()))
    return tuple(ts) if isinstance(ts, (list, tuple)) else ()


WUXING = {"木": "木", "火": "火", "土": "土", "金": "金", "水": "水"}

# 月令旺相（得令）：引擎 fu_yi.wang_shuai 的取值
DE_LING_STATES = {"旺", "相"}


@dataclass
class ShishenState:
    """某十神在原局的总体状态（聚合自四柱 shi_shens + cang_gan）。"""
    name: str                    # 十神名（如 正财/偏印/食神）
    wuxing: str                  # 该十神对应天干的五行（取第一处出现的 gan）
    tou_gan: bool                # 透干（天干出现）
    cang_zhi: bool               # 藏支（地支藏干出现）
    has_root: bool               # 有根（该五行在地支藏干出现，不限本十神）
    de_ling: bool                # 得令（该五行在月令旺相）
    count: int = 0               # 出现次数（透干+藏干各计）

    def present(self) -> bool:
        return self.tou_gan or self.cang_zhi


@dataclass
class ThreeChecks:
    """十神三关：得令 / 不被克 / 有根。"""
    de_ling: bool
    not_ke: bool
    has_root: bool

    def passed(self) -> int:
        return sum([self.de_ling, self.not_ke, self.has_root])


@dataclass
class PalaceFlags:
    """宫位状态（筛自 fullchart 的合会冲刑全局表）。"""
    pillar: str            # 宫位名（日支/年支/月支/时支）
    zhi: str
    chong: list = field(default_factory=list)   # 六冲：与谁冲
    liu_he: list = field(default_factory=list)  # 六合：与谁合
    san_he: list = field(default_factory=list)  # 三合：成局
    void: bool = False     # 空亡
    kui_gang: bool = False


@dataclass
class WuxingProfile:
    """五行分布（直接读引擎 fu_yi，零计算）。"""
    count: dict                     # wuxing_count
    wang_shuai: dict                # wang_shuai
    qiangruo: str                   # 身强弱
    yongshen: dict                  # 三派用神 {fu_yi/tiao_hou/ge_ju: {yong,xi,ji}}

    def over_strong(self, threshold: int = 4) -> list:
        """过旺五行（出现 ≥ threshold 次 或 旺）。"""
        strong = [wx for wx, n in self.count.items() if n >= threshold]
        strong += [wx for wx, st in self.wang_shuai.items() if st == "旺" and wx not in strong]
        return strong

    def over_weak(self) -> list:
        """过弱五行（死/囚，且出现 ≤ 1 次）。"""
        return [wx for wx, st in self.wang_shuai.items() if st in ("死", "囚") and self.count.get(wx, 0) <= 1]


# ---------------------------------------------------------------------------
# 聚合函数
# ---------------------------------------------------------------------------

def _all_shishens(full: dict) -> list[dict]:
    """四柱所有十神条目（含干支来源）。"""
    out = []
    for key in ("nian", "yue", "ri", "shi"):
        zhu = full.get(key, {})
        for ss in zhu.get("shi_shens", []):
            out.append({"pillar": key, **ss})
    return out


def aggregate_shishen(full: dict, wang_shuai: Optional[dict] = None) -> dict[str, ShishenState]:
    """聚合每十神：透干/藏支/有根/得令/次数。

    - 透干：source == 'stem'（天干）
    - 藏支：source == 'main_qi'/'mid_qi'/'minor_qi'（地支藏干）
    - 有根：该十神对应天干五行，在四柱地支藏干（任意）中出现
    - 得令：该五行在月令旺相（wang_shuai，由 yongshen.fu_yi 提供）
    """
    states: dict[str, ShishenState] = {}
    # 先统计透干/藏支
    for item in _all_shishens(full):
        name = item["shi_shen"]
        if name not in states:
            states[name] = ShishenState(
                name=name, wuxing="", tou_gan=False, cang_zhi=False,
                has_root=False, de_ling=False, count=0,
            )
        st = states[name]
        src = item.get("source", "")
        gan = item.get("gan", "")
        if not st.wuxing and gan:
            # 十神对应天干五行（日主五行已知，取十神对应干支五行）
            st.wuxing = _wuxing_of_gan(gan)
        if src == "stem":
            st.tou_gan = True
            st.count += 1
        elif src and src.endswith("qi"):
            st.cang_zhi = True
            st.count += 1
    # 有根：该五行出现在任一地支藏干
    roots: set[str] = set()
    for key in ("nian", "yue", "ri", "shi"):
        cg = full.get(key, {}).get("cang_gan", {}) or {}
        for v in (cg.get("main"), cg.get("mid"), cg.get("minor")):
            if v:
                roots.add(_wuxing_of_gan(v))
    for st in states.values():
        if st.wuxing:
            st.has_root = st.wuxing in roots
            # 得令：月令旺相
            if wang_shuai:
                st.de_ling = st.wuxing in wang_shuai and wang_shuai[st.wuxing] in DE_LING_STATES
    return states


def _wuxing_of_gan(gan: str) -> str:
    """天干 → 五行。"""
    table = {"甲": "木", "乙": "木", "丙": "火", "丁": "火", "戊": "土",
             "己": "土", "庚": "金", "辛": "金", "壬": "水", "癸": "水"}
    return table.get(gan, "")


def three_checks(st: ShishenState, wang_shuai: dict, gan_ke_list: list[dict]) -> ThreeChecks:
    """十神三关：得令（月令旺相）/ 不被克（无天干克它）/ 有根。

    不被克：gan_ke_list 为 fullchart.gan_he 中克该十神天干五行者——
    此处简化为「该十神透干且未被天干合绊/克」；完全版见 rules。
    """
    de_ling = st.de_ling if st.de_ling else (st.wuxing in wang_shuai and wang_shuai[st.wuxing] in DE_LING_STATES)
    # 有根已在聚合时算好；若未算则查 has_root
    return ThreeChecks(de_ling=de_ling, not_ke=True, has_root=st.has_root)


def palace_flags(full: dict, pillar: str = "ri") -> PalaceFlags:
    """筛某柱（默认日支）的冲/刑/合/空亡/魁罡。

    引擎 fullchart 的 liu_chong/zhi_liu_he/san_he 含 pillar 索引：
    nian=0, yue=1, ri=2, shi=3。
    """
    idx = {"nian": 0, "yue": 1, "ri": 2, "shi": 3}[pillar]
    zhu = full.get(pillar, {})
    pf = PalaceFlags(
        pillar=pillar,
        zhi=zhu.get("zhi", ""),
        void=bool(zhu.get("is_void")),
        kui_gang=bool(zhu.get("is_kui_gang")),
    )
    for c in full.get("liu_chong", []):
        if c.get("zhi_a_idx") == idx or c.get("zhi_b_idx") == idx or c.get("zhi_a") == pf.zhi or c.get("zhi_b") == pf.zhi:
            other = c.get("zhi_b") if c.get("zhi_a") == pf.zhi else c.get("zhi_a")
            pf.chong.append(other or c.get("detail", ""))
    for h in full.get("zhi_liu_he", []):
        if h.get("zhi_a") == pf.zhi:
            pf.liu_he.append(h.get("zhi_b", ""))
        elif h.get("zhi_b") == pf.zhi:
            pf.liu_he.append(h.get("zhi_a", ""))
    for s in full.get("san_he", []):
        for z in (s.get("zhi_a"), s.get("zhi_b"), s.get("zhi_c")):
            if z == pf.zhi:
                pf.san_he.append(s.get("he_element", "") or s.get("detail", ""))
                break
    return pf


def wuxing_profile(yongshen: dict) -> WuxingProfile:
    """直接读引擎 fu_yi（零计算）。"""
    fu = yongshen.get("fu_yi", {})
    return WuxingProfile(
        count=fu.get("wuxing_count", {}),
        wang_shuai=fu.get("wang_shuai", {}),
        qiangruo=fu.get("qiangruo", ""),
        yongshen=yongshen,
    )


def build_factors(data: dict) -> dict:
    """把 Layer 0（引擎数据）组织成 Layer 1 因子快照（纯聚合）。"""
    full = data["full"]
    ys = data["yongshen"]
    wp = wuxing_profile(ys)
    ss = aggregate_shishen(full, wp.wang_shuai)
    # 大运配偶星窗口（依据：大运透配偶星则婚缘实——原局不现可大运得）
    dy_steps = data.get("chart", {}).get("da_yun", {}).get("steps", [])
    return {
        "shishen": {k: asdict(v) for k, v in ss.items()},
        "wuxing": asdict(wp),
        "qiangruo": wp.qiangruo,
        "palace_ri": asdict(palace_flags(full, "ri")),
        "palace_nian": asdict(palace_flags(full, "nian")),
        "nian_gan": full.get("nian", {}).get("gan", ""),
        "nian_zhi": full.get("nian", {}).get("zhi", ""),
        "ziwei": data.get("ziwei", {}),   # 紫微本命盘（四化/夫妻宫交叉用）
        "yongshen": data.get("yongshen", {}),  # 三派用神（扶抑/调候/格局，喜忌五行）
        "ri_gan": full.get("ri", {}).get("gan", ""),
        "yue_zhi": full.get("yue", {}).get("zhi", ""),
        "shi_zhi": full.get("shi", {}).get("zhi", ""),
        "_birth_year": int(data.get("solar", "")[:4]) if str(data.get("solar", ""))[:4].isdigit() else (data.get("lunar") or {}).get("year", 0) or 0,
        # 大运每步：{name, qi_sui, zhi_sui, shi_shen}（shi_shen 如"正财运"/"七杀运"）
        "dayun_steps": [
            {"name": s.get("name", ""), "qi_sui": s.get("qi_sui"), "zhi_sui": s.get("zhi_sui"),
             "shi_shen": s.get("shi_shen", "")}
            for s in dy_steps
        ],
    }
