"""原子执行器（纯原语——本命/流年算子 + 常量加载——零因子/断语逻辑）。

分层：atoms（原子算子）→ aggregate（聚合）→ evaluate（因子快照）→ duanyu（生成器入口）。
代码只做"读引擎字段 + 查 constants 表 + 机械比较"——命理定义全在表。
"""
from __future__ import annotations
import os
from typing import Optional

__all__ = ["load_constants", "_op", "_liu_op", "_shishen", "_resolve_tens", "_target_stars",
           "_palace_bad", "_dayun_window", "_zw_gong_op"]


_CONST = None
def load_json(name):
    """因子层/字典加载（json——标准库，容器无 PyYAML 兼容）。name 如 "factors.json"/"constants"。"""
    import json as _json
    fname = name if name.endswith('.json') else name + '.json'
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), fname)
    with open(path, encoding="utf-8") as fh:
        return _json.load(fh)
def load_constants() -> dict:
    """复合因子层的字典配置（evaluator 求值 factors.csv 的参数表——目标星/事件宫位/五行等）。"""
    global _CONST
    if _CONST is None:
        _CONST = load_json("constants.json")
    return _CONST
_CONST = load_constants()  # 模块加载即构建（一次读 json——热路径缓存）
_KE = {wx: v["克"] for wx, v in _CONST["五行生克"].items()}
_SHENG = {wx: v["生"] for wx, v in _CONST["五行生克"].items()}
_TIANGAN_WUXING = _CONST["天干五行"]
_DIZHI_WUXING = _CONST["地支五行"]
_TARGET_STARS = _CONST["目标星"]
_TARGET_PALACES = _CONST["事件宫位"]
_TEN_CLASSES = _CONST["类"]
def _shishen(factors, ten: str) -> Optional[dict]:
    """取某十神的状态（factors.py ShishenState）。"""
    return (factors.get("shishen") or {}).get(ten)
def _class_wuxing(factors, ten_class: str) -> str:
    """十神大类的五行（取该类第一个出现的十神之五行）。"""
    classes = load_constants()["类"]
    for ten in classes.get(ten_class, []):
        st = _shishen(factors, ten)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""
def _ten_to_wx(factors, tens: list) -> str:
    """十神列表的五行（聚合）。"""
    for t in tens:
        st = _shishen(factors, t)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""
def _wang_shuai(factors) -> dict:
    return factors.get("wuxing", {}).get("wang_shuai", {}) or {}
def _is_wang(factors, wx: str) -> bool:
    wang_cfg = load_constants()["旺衰"]["旺"]
    return _wang_shuai(factors).get(wx) in wang_cfg
def _is_weak(factors, wx: str) -> bool:
    weak_cfg = load_constants()["旺衰"]["弱"]
    return _wang_shuai(factors).get(wx) in weak_cfg
def _resolve_tens(tens, factors, gender):
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
def _op(op: str, args, factors, gender, chart) -> int:
    """执行单个算子子句，返回 0/1。args 为参数（列表或标量）。"""
    const = load_constants()
    if not isinstance(args, list):
        args = [args]

    if op == "现":
        tens = _resolve_tens(args, factors, gender)
        return 1 if any((_shishen(factors, t) or {}).get("count", 0) >= 1 for t in tens) else 0
    if op == "透":
        tens = _resolve_tens(args, factors, gender)
        return 1 if any((_shishen(factors, t) or {}).get("tou_gan") for t in tens) else 0
    if op == "藏":
        tens = _resolve_tens(args, factors, gender)
        return 1 if any((_shishen(factors, t) or {}).get("cang_zhi") for t in tens) else 0
    if op == "得令":
        tens = _resolve_tens(args, factors, gender)
        return 1 if any((_shishen(factors, t) or {}).get("de_ling") for t in tens) else 0
    if op == "有根":
        tens = _resolve_tens(args, factors, gender)
        return 1 if any((_shishen(factors, t) or {}).get("has_root") for t in tens) else 0
    if op in ("旺", "弱"):
        tens = args
        # 五行名原样（五行旺衰只看月令）；十神/大类/六亲 → 综合旺衰
        if tens and tens[0] in ("木", "火", "土", "金", "水"):
            wx = tens[0]
            return 1 if (_is_wang(factors, wx) if op == "旺" else _is_weak(factors, wx)) else 0
        resolved = _resolve_tens(tens, factors, gender)
        wx = _ten_to_wx(factors, resolved)
        if not wx:
            return 0
        if op == "弱":
            # 十神弱 = 其五行失令 且 不透 且 无根（三者皆弱）
            return 1 if (_is_weak(factors, wx) and not any((_shishen(factors, t) or {}).get("tou_gan") for t in resolved)
                         and not any((_shishen(factors, t) or {}).get("has_root") for t in resolved)) else 0
        # 十神旺 = 得令（其五行月令旺相）或（透干且有根）——《子平真诠》得令为重，失令者透且有根可补
        if _is_wang(factors, wx):
            return 1
        tou = any((_shishen(factors, t) or {}).get("tou_gan") for t in resolved)
        gen = any((_shishen(factors, t) or {}).get("has_root") for t in resolved)
        return 1 if tou and gen else 0
    if op == "缺":
        count = factors.get("wuxing", {}).get("count", {}) or {}
        return 1 if args and count.get(args[0], 0) == 0 else 0
    if op == "克":
        # 克(A,B)：A 五行克 B 五行（A/B 可为十神大类或五行）
        a, b = args[0], args[1]
        wx_a = _resolve_wx(factors, gender, a)
        wx_b = _resolve_wx(factors, gender, b)
        if not wx_a or not wx_b:
            return 0
        return 1 if const["五行生克"].get(wx_a, {}).get("克") == wx_b else 0
    if op == "生":
        a, b = args[0], args[1]
        wx_a = _resolve_wx(factors, gender, a)
        wx_b = _resolve_wx(factors, gender, b)
        if not wx_a or not wx_b:
            return 0
        return 1 if const["五行生克"].get(wx_a, {}).get("生") == wx_b else 0
    if op == "直读":
        path, expect = args[0], args[1]
        if path == "ri_gan_wx":
            gan = factors.get("ri_gan", "")
            val = load_constants()["天干五行"].get(gan)
            if expect == "任意":
                return val if val else 0     # 日主五行返回五行字符串（断语约束 `日主五行: 木` 匹配）
            return 1 if str(val) == str(expect) else 0
        val = _path_get(factors, chart, path)
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
            full = chart.get("full", {}) or {}
            for zhu in ("nian", "yue", "ri", "shi"):
                for it in (full.get(zhu) or {}).get("shen_sha", []) or []:
                    names.append(it.get("name") or it.get("xing"))
        elif field == "patterns":
            # 含(patterns, 府相朝垣) —— 紫微特殊格局（ziwei.patterns）
            zw = chart.get("ziwei", {}) or {}
            names = [p.get("name") for p in (zw.get("patterns", []) or []) if p.get("name")]
        else:
            items = _path_get(factors, chart, field) or []
            names = [it.get("name") or it.get("xing") for it in items] if isinstance(items, list) else []
        if "或" in value:
            return 1 if any(v in names for v in value.replace("或", "|").split("|")) else 0
        return 1 if value in names else 0
    if op == "宫含":
        # 宫含(宫位, 星, 条件) —— ziwei gong_wei
        return _zw_gong_op(factors, chart, args)
    if op == "大运十神":
        return _dayun_op(factors, chart, args)
    if op == "数量至少":
        # 数量至少(N, 十神...)：十神出现总数 ≥ N（事实计数——印杂等"多"的定量）
        n = int(args[0])
        tens = _resolve_tens(args[1:], factors, gender)
        total = sum((_shishen(factors, t) or {}).get("count", 0) for t in tens)
        return 1 if total >= n else 0
    if op == "日主长生月支":
        # 日主长生月支[临官/帝旺/墓/绝]——日主十二长生态落在月支（chang_sheng 列表）
        state = args[0]
        cs = chart.get("full", {}).get("chang_sheng", []) or []
        yue_zhi = (chart.get("chart", {}) or {}).get("yue", {}).get("zhi", "")
        for it in cs:
            if it.get("name") == state and it.get("index") == yue_zhi:
                return 1
        return 0
    if op == "六合存在":
        # 直接用引擎 fullchart.zhi_liu_he（已算六合 + 化气五行）——不重复计算
        full = chart.get("full", {}) or {}
        rels = full.get("zhi_liu_he", []) or []
        return 1 if rels else 0
    if op == "官杀取清":
        # 官杀混杂取清（《子平真诠》）：克我（官杀）所在柱被六合/六冲 → 合杀留官/冲去多余 → 取清
        const = load_constants()
        gan_wx = const["天干五行"]
        ke = const["五行生克"]
        chart = chart.get("chart", {}) or {}
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
        full = chart.get("full", {}) or {}
        rels = (full.get("zhi_liu_he") or []) + (full.get("liu_chong") or [])
        involved = set()
        for r in rels:
            involved.add(r.get("pillar_a"))
            involved.add(r.get("pillar_b"))
        return 1 if any(p in involved for p in guan_pillars) else 0
    if op in ("为用", "为忌"):
        # 为用(十神类)：该十神五行 ∈ {用, 喜}；为忌：== 忌（引擎五神体系 yong/xi/ji）
        fy = factors.get("yongshen", {}).get("fu_yi", {}) or {}
        yong, xi, ji = fy.get("yong", ""), fy.get("xi", ""), fy.get("ji", "")
        tens = _resolve_tens(args, factors, gender)
        wx = _ten_to_wx(factors, tens)
        if not wx:
            return 0
        if op == "为用":
            return 1 if wx in (yong, xi) else 0
        return 1 if wx == ji else 0
    if op == "日主十干":
        # 日主具体天干（甲/乙/丙...）——十干性格基调
        return factors.get("ri_gan", "")
    if op == "月支长生":
        # 日主在月支的长生十二态（引擎 chang_sheng 表：长生在寅/沐浴在卯...）
        full = chart.get("full", {}) or {}
        cs = full.get("chang_sheng", []) or []
        chart = chart.get("chart", {}) or {}
        yue_zhi = (chart.get("yue") or {}).get("zhi", "")
        for item in cs:
            if item.get("index", "") == yue_zhi:
                return item.get("name", "")
        return ""
    if op == "夫妻宫状态":
        # 日支（夫妻宫）被冲/合/刑——引擎 liu_chong/zhi_liu_he/liu_xing
        full = chart.get("full", {}) or {}
        chart = chart.get("chart", {}) or {}
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
        chart = chart.get("chart", {}) or {}
        ri_zhi = (chart.get("ri") or {}).get("zhi", "")
        const = load_constants()
        for cat, zhis in const["日支神煞"].items():
            if ri_zhi in zhis:
                return cat
        return ""
    if op == "财库现":
        # 日干所克=财星五行 → 墓库支（金库丑/木库未/水库辰/火库戌/土库辰）→ 在命局四柱支中
        const = load_constants()
        day_gan = factors.get("ri_gan", "")
        day_wx = const["天干五行"].get(day_gan, "")
        cai_wx = const["五行生克"].get(day_wx, {}).get("克", "")
        ku = const["墓库"].get(cai_wx, "")
        if not ku:
            return 0
        chart = chart.get("chart", {}) or {}
        for zhu in ("nian", "yue", "ri", "shi"):
            if (chart.get(zhu) or {}).get("zhi", "") == ku:
                return 1
        return 0
    if op == "财星入墓":
        # 财星坐墓库支（财藏库中——蓄财之象）
        const = load_constants()
        day_gan = factors.get("ri_gan", "")
        day_wx = const["天干五行"].get(day_gan, "")
        cai_wx = const["五行生克"].get(day_wx, {}).get("克", "")
        ku = const["墓库"].get(cai_wx, "")
        if not ku:
            return 0
        chart = chart.get("chart", {}) or {}
        # 财星（正偏财）所在柱支 = 墓库支
        for k in ("正财", "偏财"):
            st = factors.get("shishen", {}).get(k)
            if st and st.get("count", 0):
                for zhu in ("nian", "yue", "ri", "shi"):
                    c = chart.get(zhu) or {}
                    if c.get("zhi", "") == ku:
                        return 1
        return 0
    if op == "克者旺":
        # A 的五行被 B 克（B=克者），B 得令而旺
        resolved = _resolve_tens(args, factors, gender)
        wx = _ten_to_wx(factors, resolved)
        if not wx:
            return 0
        ke_wx = [k for k, v in load_constants()["五行生克"].items() if v.get("克") == wx]
        return 1 if ke_wx and _is_wang(factors, ke_wx[0]) else 0
    if op == "格神透":
        return _ge_shen_tou(factors, chart)
    if op == "格神根":
        return _ge_shen_gen(factors, chart)
    if op == "月令本气":
        # 月令本气十神（性格主面/格神）——full.yue.cang_gan.main 的十神（《子平真诠》月令为提纲，格神主性）
        full = chart.get("full", {}) or {}
        yue = full.get("yue", {}) or {}
        main = (yue.get("cang_gan") or {}).get("main", "")
        if not main:
            return 0
        for ss in yue.get("shi_shens", []):
            if ss.get("gan") == main:
                return 1 if ss.get("shi_shen") == args[0] else 0
        return 0
    if op == "年柱官杀":
        return _nian_guan(factors, chart)
    if op == "日支冲刑害":
        return _palace_bad(factors, chart)
    if op == "大运窗口":
        return _dayun_window(factors, chart, args)
    # 流年算子（流年透/值/合/冲/克/忌神/财坏印/大运窗口/换运/岁运并临/干合等）由 _liu_op 处理
    raise ValueError(f"未知算子: {op}")
def _resolve_wx(factors, gender, arg):
    """解析任意参数为五行：具体五行名原样；十神/大类 → 五行。"""
    const = load_constants()
    if arg in ("木", "火", "土", "金", "水"):
        return arg
    if arg in const.get("类", {}):
        return _class_wuxing(factors, arg)
    if _shishen(factors, arg):
        return (_shishen(factors, arg) or {}).get("wuxing")
    # 配偶星等
    resolved = _resolve_tens([arg], factors, gender)
    return _ten_to_wx(factors, resolved)
def _path_get(factors, chart, path: str):
    """按路径取值：优先 factors（基础因子），其次 chart 原始数据。"""
    obj = factors if path.startswith(("shishen", "wuxing", "qiangruo", "yongshen", "ri_gan")) else chart
    cur = obj
    for part in path.split("."):
        if isinstance(cur, dict):
            cur = cur.get(part)
        else:
            return None
    return cur
def _zw_gong_op(factors, chart, args):
    """紫微宫位查询：宫含(宫位, 星, 条件)。

    四化（化禄/权/科/忌）用顶层 si_hua（{星:四化}）反推落宫——引擎本命四化无宫位字段，需按星找宫。
    """
    gong_name, star, cond = args[0], args[1], args[2] if len(args) > 2 else "任意"
    zw = chart.get("ziwei", {}) or {}
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
        if cond == "庙旺":
            return 1 if any(st.get("xing") == star and st.get("liang_du") in ("庙", "旺", "得") for st in g.get("xing_yao", [])) else 0
        if cond == "唯一主星":
            main_stars = set(load_constants()["紫微主星"])
            present = [s for s in stars if s in main_stars]
            return 1 if present == [star] else 0
        return 1 if star in stars else 0
    return 0
def _dayun_op(factors, chart, args):
    """大运十神查询：大运十神(当前/年, 星类)。"""
    when, star_class = args[0], args[1]
    dx = (chart.get("full") or chart.get("chart") or {}).get("da_yun", {}) or {}
    steps = dx.get("steps", [])
    idx = dx.get("current_step_index", -1)
    if when == "当前" and 0 <= idx < len(steps):
        shi_shen = (steps[idx].get("shi_shen", "") or "").replace("运", "")
        resolved = _resolve_tens([star_class], factors, chart.get("gender"))
        return 1 if shi_shen in resolved else 0
    return 0
def _dayun_window(factors, chart, args):
    """大运窗口：当前大运十神 ∈ 配偶星。"""
    return _dayun_op(factors, chart, ["当前", args[0] if args else "配偶星"])
def _ge_shen_tou(factors, chart):
    """格神透干：月令本气（格神）透干。"""
    full = chart.get("full", {}) or {}
    yue = full.get("yue", {}) or {}
    ge_ju = (factors.get("yongshen") or {}).get("ge_ju", {}) or {}
    ge_name = ge_ju.get("ge_ju", "")
    if not ge_name:
        return 0
    # 格神 = 月令本气十神；查月柱 shi_shens 是否有 source=stem
    for ss in yue.get("shi_shens", []):
        if ss.get("source") == "stem":
            return 1
    return 0
def _ge_shen_gen(factors, chart):
    full = chart.get("full", {}) or {}
    yue = full.get("yue", {}) or {}
    for ss in yue.get("shi_shens", []):
        if ss.get("source") in ("main_qi", "mid", "minor"):
            return 1
    return 0
def _nian_guan(factors, chart):
    """年柱官杀攻身：年柱 shi_shens 含官杀。"""
    full = chart.get("full", {}) or {}
    nian = full.get("nian", {}) or {}
    for ss in nian.get("shi_shens", []):
        if ss.get("shi_shen") in ("正官", "七杀"):
            return 1
    return 0
def _palace_bad(factors, chart):
    """宫破：日支逢六冲/相刑/六害。"""
    full = chart.get("full", {}) or {}
    ri_zhi = (full.get("ri", {}) or {}).get("zhi", "")
    if not ri_zhi:
        return 0
    for key in ("liu_chong", "liu_xing", "liu_hai"):
        for rel in full.get(key, []) or []:
            if ri_zhi in (rel.get("a"), rel.get("b"), rel.get("zhi1"), rel.get("zhi2")):
                return 1
    return 0
def _liu_op(op: str, args, factors, gender, chart, ctx: dict = None) -> int:
    """流年算子（纯函数——显式 ctx 上下文参数，无全局状态）。"""
    ctx = ctx or {}
    ln = ctx.get("liunian", {})
    target = ctx.get("target", "配偶星")
    year = ctx.get("year", 0)
    const = load_constants()
    star_keys = _target_stars(target, gender, const)
    nz = ln.get("nian_zhi", "")
    nian_gan = ln.get("nian_gan", "")
    ss_year = ln.get("shi_shen", "")

    if op == "流年长生":
        # 日主在流年支的十二长生态（复用本命 chang_sheng 表：长生在寅/帝旺在午...——流年支落哪态）
        state = args[0]
        cs = chart.get("full", {}).get("chang_sheng", []) or []
        nz = ln.get("nian_zhi", "")
        for it in cs:
            if it.get("name") == state and it.get("index") == nz:
                return 1
        return 0
    if op == "流年神煞":
        # 服务端流年神煞（bazi.liunian 返回 shensha[]：红鸾/天喜/劫煞/灾煞/驿马/桃花/羊刃/华盖/天乙贵人）
        ss = ln.get("shensha", []) or []
        return 1 if any((s.get("name") or "") == args[0] for s in ss) else 0
    if op == "流年透":
        return 1 if ss_year in star_keys else 0
    if op in ("流年值", "流年合", "流年冲"):
        palace_key = const.get("事件宫位", {}).get(target, "ri")
        ri_zhi = factors.get("palace_ri", {}).get("zhi", "")
        chart = chart.get("chart", {}) or {}
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
        _DIZHI_WUXING = const["地支五行"]
        target_wx = None
        if target == "日主":
            target_wx = _GANWX.get(factors.get("ri_gan", ""), "")
        else:
            for k in star_keys:
                st = factors.get("shishen", {}).get(k)
                if st and st.get("wuxing"):
                    target_wx = st["wuxing"]
                    break
        if target_wx and nian_gan:
            ke = const["五行生克"].get(_GANWX.get(nian_gan, ""), {}).get("克")
            ke2 = const["五行生克"].get(_DIZHI_WUXING.get(nz, ""), {}).get("克")
            return 1 if ke == target_wx or ke2 == target_wx else 0
        return 0
    if op == "忌神干":
        ji = factors.get("yongshen", {}).get("fu_yi", {}).get("ji", "")
        ji_wx = const["天干五行"].get(ji, "") or ji   # ji 可能是五行名（如"火"）或天干名——转五行
        return 1 if (ji_wx and const["天干五行"].get(nian_gan, "") == ji_wx) else 0
    if op == "忌神支":
        ji = factors.get("yongshen", {}).get("fu_yi", {}).get("ji", "")
        ji_wx = const["天干五行"].get(ji, "") or ji
        return 1 if (ji_wx and const["地支五行"].get(nz, "") == ji_wx) else 0
    if op == "财坏印流年":
        _YIN2 = {"正印", "偏印"}
        if nz and nian_gan and ss_year in _YIN2:
            ke = const["五行生克"].get(const["地支五行"].get(nz, ""), {}).get("克")
            return 1 if ke == const["天干五行"].get(nian_gan, "") else 0
        return 0
    if op == "大运窗口流年":
        # 引擎 2.6.15 起大运步骤带公历年段（start_year/end_year）——直接年份判断，免虚岁换算
        for s in factors.get("dayun_steps", []):
            if any(k in s.get("shi_shen", "") for k in star_keys) and s.get("start_year", 0) <= year <= s.get("end_year", 0):
                return 1
        return 0
    if op == "换运流年":
        # 换运首年 = 该步 start_year（引擎日期段直给）
        for s in factors.get("dayun_steps", []):
            if any(k in s.get("shi_shen", "") for k in star_keys):
                if year in (s.get("start_year", 0), s.get("start_year", 0) + 1):
                    return 1
        return 0
    if op == "流年宫化":
        # 流年宫化[宫, 化]——化 ∈ 禄/权/科/忌；判断流年四化中"该宫"被"该化"
        gong, hua = args[0], args[1]
        sg = ctx.get("zw_liunian", {}).get("si_hua_gong", {}) or {}
        sh = ctx.get("zw_liunian", {}).get("si_hua", {}) or {}
        for star, gname in sg.items():
            if gname == gong and sh.get(star) == hua:
                return 1
        return 0
    if op == "引用本命":
        key = args[0]
        if key == "本命婚凶":
            return ctx.get("marriage_bad", 0)
        if key == "食伤克官":
            return ctx.get("shi_ke_guan", 0)
        if key == "食伤重":
            # 本命食伤重（= 本命快照"食伤旺"——流年因子"食伤重"引用；yingqi ying_h18/19 断语条件）
            return ctx.get("shi_shang_zhong", 0)
        return 0
    # ── 机械原子（查 constants 表/比较——组合定义在表）──
    if op == "干支相等":
        # 干支相等[来源A,来源B]——来源：大运/流年/日柱
        ga, gb = _source_ganzhi(args[0], ctx), _source_ganzhi(args[1], ctx)
        return 1 if (ga and ga == gb) else 0
    if op == "干克":
        # 干克[干A来源,干B来源]——来源：流年干/日干
        g1, g2 = _source_gan(args[0], ctx), _source_gan(args[1], ctx)
        if not g1 or not g2:
            return 0
        ke = const["五行生克"].get(const["天干五行"].get(g1, ""), {}).get("克")
        return 1 if (ke and ke == const["天干五行"].get(g2, "")) else 0
    if op == "支冲":
        # 支冲[支A来源,支B来源]——来源：流年支/日支
        z1, z2 = _source_zhi(args[0], ctx), _source_zhi(args[1], ctx)
        return 1 if (z1 and z2 and const["六冲"].get(z1) == z2) else 0
    if op == "三刑":
        # 三刑[支来源...]——命局四柱支（含流年）凑齐三刑组
        zhis = []
        for a in args:
            zv = _source_zhi(a, ctx)
            # 日支/时支/年支即四柱对应位置，由下方四柱循环计入——跳过，避免同字重复计数
            # （否则日支=午 会被计 2 次，令"午午"自刑在单午时误成立）
            if zv and a not in ("日支", "时支", "年支"):
                zhis.append(zv)
        chart = chart.get("chart", {}) or {}
        for zhu in ("nian", "yue", "ri", "shi"):
            z = (chart.get(zhu) or {}).get("zhi", "")
            if z:
                zhis.append(z)
        # 计数：const["三刑"] 为 dict，key=组内一个地支，value=同组其余地支。
        # 需 k 与其同组其余地支全部在场（自刑组 v=[k]，需同字出现 ≥2 次才算）。
        cnt: dict = {}
        for z in zhis:
            cnt[z] = cnt.get(z, 0) + 1
        for k, v in const["三刑"].items():
            if cnt.get(k, 0) >= 1 and all(cnt.get(g, 0) >= (2 if g == k else 1) for g in v):
                return 1
        return 0
    if op == "旬空":
        # 旬空[日柱干支,流年支]——日柱所在旬的空亡支是否含流年支
        gz = _source_ganzhi(args[0], ctx)
        nz2 = _source_zhi(args[1], ctx)
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
        ln = ctx.get("liunian", {})
        nz = ln.get("nian_zhi", "")
        if not nz:
            return 0
        snap = ctx.get("snapshot", {})
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
        ln = ctx.get("liunian", {})
        chart = ctx.get("chart", {}).get("chart", {}) or {}
        return 1 if (ln.get("nian_gan") and ln.get("nian_gan") == (chart.get("nian") or {}).get("gan", "")) else 0
    if op == "天干合":
        # 天干合[干A来源,干B来源]——查五合表
        g1, g2 = _source_gan(args[0], ctx), _source_gan(args[1], ctx)
        return 1 if (g1 and g2 and const["天干五合"].get(g1) == g2) else 0
    return 0
def _current_dayun_gz(ctx: dict) -> str:
    """当前大运干支（机械——查大运步骤公历年段）。"""
    fac2 = ctx.get("factors", {})
    year = ctx.get("year", 0)
    for s in fac2.get("dayun_steps", []):
        if s.get("start_year", 0) <= year <= s.get("end_year", 0):
            return s.get("name", "")
    return ""
def _source_ganzhi(src: str, ctx: dict) -> str:
    """干支来源解析：大运/流年/日柱 → 干支。"""
    ln = ctx.get("liunian", {})
    chart = ctx.get("chart", {}).get("chart", {}) or {}
    if src == "流年":
        return ln.get("nian_gan", "") + ln.get("nian_zhi", "")
    if src == "大运":
        return _current_dayun_gz(ctx)
    if src == "日柱":
        return (chart.get("ri") or {}).get("gan", "") + (chart.get("ri") or {}).get("zhi", "")
    return ""
def _source_gan(src: str, ctx: dict) -> str:
    """干来源：流年干/日干。"""
    ln = ctx.get("liunian", {})
    fac2 = ctx.get("factors", {})
    if src == "流年干":
        return ln.get("nian_gan", "")
    if src == "日干":
        return fac2.get("ri_gan", "")
    return ""
def _source_zhi(src: str, ctx: dict) -> str:
    """支来源：流年支/日支/时支。"""
    ln = ctx.get("liunian", {})
    chart = ctx.get("chart", {}).get("chart", {}) or {}
    fac2 = ctx.get("factors", {})
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
        # constants.json 性别键为 male/female（英文）——兼容外部传入的中文 男/女（历史漏配：
        # gender="女" 直接 ts.get("女") 取不到 → star_keys 恒空 → 流年透/流年克 算子恒 0）
        g = {"男": "male", "女": "female"}.get(gender, gender)
        return tuple(ts.get(g, ()))
    return tuple(ts) if isinstance(ts, (list, tuple)) else ()
