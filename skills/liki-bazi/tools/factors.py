"""因子层 — pan 原语算子 + 真值表求值 + 本命/流年因子快照。

分层：
paipan.py 排盘
extract.py pan → fac
factors.py fac + factors.csv / factors_liunian.csv → snap
duanyu.py snap + 断语表 → 断语

本文件只做机械求值；命理成员关系与分组在 constants.json，因子组合在 CSV。
"""
from __future__ import annotations

import json
import os
from typing import Optional

__all__ = [
    "load_constants", "evaluate_factors", "evaluate_liunian_factors",
    "make_factors", "make_liunian_factors",
    "_op", "_liu_op", "_OP_NAMES", "_LIU_OP_NAMES",
]

_CONST = None
def load_constants() -> dict:
    """因子层字典配置：基础闭集、十神大类、六亲角色、五行与干支关系。"""
    global _CONST
    if _CONST is None:
        path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "constants.json")
        with open(path, encoding="utf-8") as fh:
            _CONST = json.load(fh)
    return _CONST
_CONST = load_constants()  # 模块加载即构建（一次读 json——热路径缓存）

# 算子名清单（本命/流年）——factors._atomic 显式分派用；新增算子须在此登记
_OP_NAMES = frozenset({
    "现", "透", "藏", "得令", "有根", "旺", "弱", "缺", "克", "直读", "含", "宫含", "关系",
    "大运十神", "数量至少", "五行数量至少", "官杀取清", "为用", "为忌",
    "月支长生", "夫妻宫状态", "日支类型", "财库现", "财星入墓", "克者旺",
    "格神透", "月令本气", "年柱官杀", "日支冲刑害", "柱刑",
})
_LIU_OP_NAMES = frozenset({
    "流年长生", "流年神煞", "流年透", "流年值", "流年合", "流年冲", "流年克",
    "忌神干", "忌神支", "财坏印流年", "大运窗口流年", "换运流年", "流年宫化", "引用本命",
    "干支相等", "干克", "支冲", "三刑", "旬空", "流年支受克", "年柱干伏吟", "天干合",
})
def _shishen(fac, ten: str) -> Optional[dict]:
    """取某十神的提取状态（extract.py ShishenState）。"""
    return (fac.get("shishen") or {}).get(ten)
def _class_wuxing(fac, ten_class: str) -> str:
    """十神大类的五行（取该类第一个出现的十神之五行）。"""
    classes = load_constants()["十神大类"]
    for ten in classes.get(ten_class, []):
        st = _shishen(fac, ten)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""
def _ten_to_wx(fac, tens: list) -> str:
    """十神列表的五行（聚合）。"""
    for t in tens:
        st = _shishen(fac, t)
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
    """解析十神参数：六亲角色与十神大类 → 具体十神列表；配偶星/子女星按性别。"""
    const = load_constants()
    classes = const["十神大类"]
    roles = const["六亲角色"]
    gender_key = {"男": "male", "女": "female"}.get(gender, gender)
    result = []
    for name in tens:
        role = roles.get(name, name)
        if isinstance(role, dict):
            role = role.get(gender_key, "")
        if role in classes:
            result.extend(classes[role])
        else:
            result.append(role)
    return result
def _op(op: str, args, fac: dict, gender: str, chart: dict,
        current_year: int = 0) -> "int | str":
    """执行单个算子子句，返回 0/1（或事实值字符串）。args 为参数（列表或标量）。"""
    const = load_constants()
    if not isinstance(args, list):
        args = [args]

    if op == "现":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_shishen(fac, t) or {}).get("count", 0) >= 1 for t in tens) else 0
    if op == "透":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_shishen(fac, t) or {}).get("tou_gan") for t in tens) else 0
    if op == "藏":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_shishen(fac, t) or {}).get("cang_zhi") for t in tens) else 0
    if op == "得令":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_shishen(fac, t) or {}).get("de_ling") for t in tens) else 0
    if op == "有根":
        tens = _resolve_tens(args, fac, gender)
        return 1 if any((_shishen(fac, t) or {}).get("has_root") for t in tens) else 0
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
            return 1 if (_is_weak(fac, wx) and not any((_shishen(fac, t) or {}).get("tou_gan") for t in resolved)
                         and not any((_shishen(fac, t) or {}).get("has_root") for t in resolved)) else 0
        # 十神旺 = 得令（其五行月令旺相）或（透干且有根）——《子平真诠》得令为重，失令者透且有根可补
        if _is_wang(fac, wx):
            return 1
        tou = any((_shishen(fac, t) or {}).get("tou_gan") for t in resolved)
        gen = any((_shishen(fac, t) or {}).get("has_root") for t in resolved)
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
    if op == "直读":
        path, expect = args[0], args[1]
        if path == "gender":
            # gender 是求值上下文；mock/单元测试可只传 gender 参数而无完整 pan。
            val = gender
            if expect == "任意":
                return val if val else 0
            return 1 if str(val) == str(expect) else 0
        if path == "ri_gan_wx":
            gan = fac.get("ri_gan", "")
            val = load_constants()["天干五行"].get(gan)
            if expect == "任意":
                return val if val else 0     # 日主五行返回五行字符串（断语约束 `日主五行: 木` 匹配）
            return 1 if str(val) == str(expect) else 0
        val = _path_get(fac, chart, path)
        if expect == "任意":
            # 返回字符串原值（月令格=ge_ju.ge_ju 如"正财格"——断语表 `月令格: 正财格` 匹配）；
            # 旧实现返回存在性 1 → 与字符串约束永不匹配 → 月令格断语全灭
            return val if val else 0
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
            items = _path_get(fac, chart, field) or []
            names = [it.get("name") or it.get("xing") for it in items] if isinstance(items, list) else []
        if "或" in value:
            return 1 if any(v in names for v in value.replace("或", "|").split("|")) else 0
        return 1 if value in names else 0
    if op == "宫含":
        # 宫含(宫位, 星, 条件) —— ziwei gong_wei
        return _zw_gong_op(fac, chart, args)
    if op == "大运十神":
        return _dayun_op(fac, chart, args, current_year)
    if op == "数量至少":
        # 数量至少(N, 十神...)：十神出现总数 ≥ N（事实计数——印杂等"多"的定量）
        n = int(args[0])
        tens = _resolve_tens(args[1:], fac, gender)
        total = sum((_shishen(fac, t) or {}).get("count", 0) for t in tens)
        return 1 if total >= n else 0
    if op == "关系":
        # 关系[field, 组名]——读取引擎 fullchart 已判定的合会冲刑。
        field, group = args[0], args[1]
        full = chart.get("full", {}) or {}
        items = full.get(field, []) or []
        if field == "gan_he":
            return 1 if any(
                {item.get("gan_a", ""), item.get("gan_b", "")} == set(group)
                for item in items
            ) else 0
        if field in ("zhi_liu_he", "liu_chong", "liu_hai"):
            return 1 if any(
                {item.get("zhi_a", ""), item.get("zhi_b", "")} == set(group)
                for item in items
            ) else 0
        if field in ("san_he", "san_hui"):
            # 引擎 TripleGroup.name 形如“申子辰水局 / 寅卯辰木方”。
            return 1 if any(group in (item.get("name", "") or "") for item in items) else 0
        if field == "liu_xing":
            # 引擎相刑是成对记录；组名成立要求组内地支在命局中齐备，
            # 且至少有一对实际相刑记录，避免把无关散支拼成三刑组。
            chart_zhi = {
                (chart.get("chart", {}).get(zhu, {}) or {}).get("zhi", "")
                for zhu in ("nian", "yue", "ri", "shi")
            }
            if not set(group).issubset(chart_zhi):
                return 0
            involved = set()
            for item in items:
                pair = {item.get("zhi_a", ""), item.get("zhi_b", "")}
                if pair & set(group):
                    involved.update(pair)
            return 1 if set(group).issubset(involved) else 0
        raise ValueError(f"未知关系字段: {field}")
    if op == "柱刑":
        # 柱刑(年支/月支/时支)：该柱地支是否参与命局相刑（六亲宫位刑伤——星宫分野中"宫"维度）
        # 年柱=父宫 / 月柱=兄弟宫 / 时柱=子女宫（family.md 宫位论；日支由 日支冲刑害/夫妻宫状态 专管）
        pillar_idx = {"年支": 0, "月支": 1, "日支": 2, "时支": 3}.get(args[0] if args else "")
        if pillar_idx is None:
            return 0
        full = chart.get("full", {}) or {}
        for r in full.get("liu_xing", []) or []:
            if pillar_idx in (r.get("pillar_a", -1), r.get("pillar_b", -1)):
                return 1
        return 0
    if op == "五行数量至少":
        # 五行数量至少(N, 五行)：四柱五行计数 ≥ N（数量事实，阈值在因子表）
        n = int(args[0])
        count = fac.get("wuxing", {}).get("count", {}).get(args[1], 0)
        return 1 if count >= n else 0
    if op == "官杀取清":
        # 官杀混杂取清（《子平真诠》）：克我（官杀）所在柱被六合/六冲 → 合杀留官/冲去多余 → 取清
        const = load_constants()
        gan_wx = const["天干五行"]
        ke = const["五行生克"]
        lean_chart = chart.get("chart", {}) or {}
        full = chart.get("full", {}) or {}
        day_gan = (lean_chart.get("ri") or {}).get("gan", "")
        if not day_gan:
            return 0
        day_wx = gan_wx.get(day_gan, "")
        sha_pillars = []
        guan_names = set()
        for i, zhu in enumerate(("nian", "yue", "ri", "shi")):
            g = (lean_chart.get(zhu) or {}).get("gan", "")
            wx = gan_wx.get(g, "")
            # 克我者=官杀（ke[官杀五行] == 日主五行）
            if wx and day_wx and ke.get(wx, {}).get("克", "") == day_wx:
                for ss in (full.get(zhu, {}) or {}).get("shi_shens", []) or []:
                    if ss.get("source") == "gan" and ss.get("gan") == g:
                        guan_names.add(ss.get("shi_shen", ""))
                        if ss.get("shi_shen") == "七杀":
                            sha_pillars.append(i)
        # 未混杂（仅正官或仅七杀）谈不上「取清」；先混后取才是官杀取清。
        if not {"正官", "七杀"} <= guan_names:
            return 0
        if not sha_pillars:
            return 0
        rels = (full.get("zhi_liu_he") or []) + (full.get("liu_chong") or [])
        involved = set()
        for r in rels:
            involved.add(r.get("pillar_a"))
            involved.add(r.get("pillar_b"))
        return 1 if any(p in involved for p in sha_pillars) else 0
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
        # 日支（夫妻宫）被冲/合/刑/害——引擎 liu_chong/zhi_liu_he/liu_xing/liu_hai
        full = chart.get("full", {}) or {}
        chart = chart.get("chart", {}) or {}
        ri_zhi = (chart.get("ri") or {}).get("zhi", "")
        if not ri_zhi:
            return ""
        involved = set()
        for rel_key in ("liu_chong", "zhi_liu_he", "liu_xing", "liu_hai"):
            for r in full.get(rel_key, []) or []:
                involved.add(r.get("zhi_a", ""))
                involved.add(r.get("zhi_b", ""))
        if ri_zhi in involved:
            # 判断具体类型（凶度：冲＞合＞刑＞害；同支多重则取更凶者）
            for r in full.get("liu_chong", []) or []:
                if ri_zhi in (r.get("zhi_a", ""), r.get("zhi_b", "")):
                    return "冲"
            for r in full.get("zhi_liu_he", []) or []:
                if ri_zhi in (r.get("zhi_a", ""), r.get("zhi_b", "")):
                    return "合"
            for r in full.get("liu_xing", []) or []:
                if ri_zhi in (r.get("zhi_a", ""), r.get("zhi_b", "")):
                    return "刑"
            for r in full.get("liu_hai", []) or []:
                if ri_zhi in (r.get("zhi_a", ""), r.get("zhi_b", "")):
                    return "害"
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
        day_gan = fac.get("ri_gan", "")
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
        day_gan = fac.get("ri_gan", "")
        day_wx = const["天干五行"].get(day_gan, "")
        cai_wx = const["五行生克"].get(day_wx, {}).get("克", "")
        ku = const["墓库"].get(cai_wx, "")
        if not ku:
            return 0
        full = chart.get("full", {}) or {}
        # 财星入墓要求财星与墓库同柱；不能把「原局另有财墓支」误作财星坐墓。
        for zhu in ("nian", "yue", "ri", "shi"):
            pillar = full.get(zhu, {}) or {}
            if pillar.get("zhi", "") != ku:
                continue
            if any(
                ss.get("shi_shen") in ("正财", "偏财")
                for ss in pillar.get("shi_shens", []) or []
            ):
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
    if op == "格神透":
        return _ge_shen_tou(fac, chart)
    if op == "月令本气":
        # 月令本气十神（性格主面/格神）——任意模式返回十神标量。
        full = chart.get("full", {}) or {}
        yue = full.get("yue", {}) or {}
        main = (yue.get("cang_gan") or {}).get("main", "")
        for ss in yue.get("shi_shens", []):
            if ss.get("gan") == main:
                name = ss.get("shi_shen", "")
                if args and args[0] == "任意":
                    return name
                return 1 if name == args[0] else 0
        return "" if args and args[0] == "任意" else 0
    if op == "年柱官杀":
        return _nian_guan(fac, chart)
    if op == "日支冲刑害":
        return _palace_bad(fac, chart)
    # 流年算子（流年透/值/合/冲/克/忌神/财坏印/大运窗口/换运/岁运并临/干合等）由 _liu_op 处理
    raise ValueError(f"未知算子: {op}")
def _resolve_wx(fac, gender, arg):
    """解析任意参数为五行：具体五行名原样；十神/大类 → 五行。"""
    const = load_constants()
    if arg in ("木", "火", "土", "金", "水"):
        return arg
    if arg in const.get("十神大类", {}):
        return _class_wuxing(fac, arg)
    if _shishen(fac, arg):
        return (_shishen(fac, arg) or {}).get("wuxing")
    # 配偶星等
    resolved = _resolve_tens([arg], fac, gender)
    return _ten_to_wx(fac, resolved)
def _path_get(fac, chart, path: str):
    """按路径取值：优先 factors（基础因子），其次 chart 原始数据。"""
    obj = fac if path.startswith(("shishen", "wuxing", "yongshen", "ri_gan")) else chart
    cur = obj
    for part in path.split("."):
        if isinstance(cur, dict):
            cur = cur.get(part)
        else:
            return None
    return cur
def _zw_gong_op(fac, chart, args):
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
    const = load_constants()
    group = set(const.get(star, []) or [])
    if star == "任意":
        group = set(const["紫微主星"])
    for g in gw:
        gname = g.get("name", "")
        if gong_name not in gname and gname not in gong_name:
            continue
        star_items = g.get("xing_yao", []) or []
        stars = [st.get("xing", "") for st in star_items]
        selected = group if group else {star}
        if star == "煞星":
            selected = set(const["紫微煞星"])
        if star == "无主星":
            main_stars = set(const["紫微主星"])
            return 1 if not (main_stars & set(stars)) else 0
        if cond == "落陷":
            return 1 if any(st.get("xing") in selected and st.get("liang_du") in ("陷", "平") for st in star_items) else 0
        if cond == "庙旺":
            return 1 if any(st.get("xing") in selected and st.get("liang_du") in ("庙", "旺", "得") for st in star_items) else 0
        if cond == "唯一主星":
            main_stars = set(const["紫微主星"])
            present = [s for s in stars if s in main_stars]
            return 1 if present == [star] else 0
        return 1 if bool(selected & set(stars)) else 0
    return 0
def _ten_class(name: str) -> str:
    """具体十神 → 十神大类；无映射时原样返回。"""
    for class_name, members in load_constants()["十神大类"].items():
        if name in members:
            return class_name
    return name


def _dayun_op(fac, chart, args, current_year: int = 0):
    """大运十神查询：大运十神(当前, 大类/任意)。任意模式返回十神大类标量。"""
    when, star_class = args[0], args[1]
    if when != "当前":
        return "" if star_class == "任意" else 0

    steps = fac.get("dayun_steps", [])
    if current_year:
        selected = next((
            step for step in steps
            if step.get("start_year", 0) <= current_year <= step.get("end_year", 0)
        ), None)
    else:
        dx = (chart.get("full") or chart.get("chart") or {}).get("da_yun", {}) or {}
        raw_steps = dx.get("steps", []) or steps
        idx = dx.get("current_step_index", -1)
        selected = raw_steps[idx] if 0 <= idx < len(raw_steps) else None

    if selected is None:
        return "" if star_class == "任意" else 0
    shi_shen = (selected.get("shi_shen", "") or "").replace("运", "")
    if star_class == "任意":
        return _ten_class(shi_shen)
    resolved = _resolve_tens([star_class], fac, chart.get("gender"))
    return 1 if shi_shen in resolved else 0
    return "" if star_class == "任意" else 0
def _ge_shen_tou(fac, chart):
    """格神透干：月令所定十神格神透出到四柱天干。"""
    full = chart.get("full", {}) or {}
    ge_ju = (fac.get("yongshen") or {}).get("ge_ju", {}) or {}
    ge_name = ge_ju.get("ge_ju", "")
    ge_shen = ge_name[:-1] if ge_name.endswith("格") else ""
    if ge_shen not in load_constants()["十神"]:
        return 0
    # 建禄 / 月刃 / 杂格不是十神格名；十神格以四柱天干十神同气为透。
    for pillar in ("nian", "yue", "ri", "shi"):
        for ss in (full.get(pillar, {}) or {}).get("shi_shens", []) or []:
            if ss.get("source") == "gan" and ss.get("shi_shen") == ge_shen:
                return 1
    return 0
def _nian_guan(fac, chart):
    """年柱官杀攻身：年柱 shi_shens 含官杀。"""
    full = chart.get("full", {}) or {}
    nian = full.get("nian", {}) or {}
    for ss in nian.get("shi_shens", []):
        if ss.get("shi_shen") in ("正官", "七杀"):
            return 1
    return 0
def _palace_bad(fac, chart):
    """夫妻宫破：日支逢六冲/相刑/六害。"""
    full = chart.get("full", {}) or {}
    ri_zhi = (full.get("ri", {}) or {}).get("zhi", "")
    if not ri_zhi:
        return 0
    for key in ("liu_chong", "liu_xing", "liu_hai"):
        for rel in full.get(key, []) or []:
            if ri_zhi in (rel.get("zhi_a", ""), rel.get("zhi_b", "")):
                return 1
    return 0
def _liu_op(op: str, args, fac: dict, gender: str, chart: dict, ctx: Optional[dict] = None) -> int:
    """流年算子（纯函数——显式 ctx 上下文参数，无全局状态）。"""
    ctx = ctx or {}
    ln = ctx.get("liunian", {})
    target_ops = ("流年透", "流年值", "流年合", "流年冲", "流年克", "大运窗口流年", "换运流年")
    if op in target_ops:
        if not args:
            raise ValueError(f"{op} 必须显式传入 target 参数")
        target = args[0]
    else:
        target = ""
    year = ctx.get("year", 0)
    const = load_constants()
    star_keys = _target_stars(target, gender, const)
    nz = ln.get("nian_zhi", "")
    nian_gan = ln.get("nian_gan", "")
    ss_year = ln.get("shi_shen", "")

    if op == "流年长生":
        # 日主在流年支的十二长生态；任意模式返回状态标量。
        cs = chart.get("full", {}).get("chang_sheng", []) or []
        nz = ln.get("nian_zhi", "")
        for it in cs:
            if it.get("index") == nz:
                return it.get("name", "") if args and args[0] == "任意" else (1 if it.get("name") == args[0] else 0)
        return "" if args and args[0] == "任意" else 0
    if op == "流年神煞":
        # 服务端流年神煞（bazi.liunian 返回 shensha[]：红鸾/天喜/劫煞/灾煞/驿马/桃花/羊刃/华盖/天乙贵人）
        ss = ln.get("shensha", []) or []
        return 1 if any((s.get("name") or "") == args[0] for s in ss) else 0
    if op == "流年透":
        return 1 if ss_year in star_keys else 0
    if op in ("流年值", "流年合", "流年冲"):
        palace_key = const.get("事件宫位", {}).get(target, "ri")
        ri_zhi = fac.get("palace_ri", {}).get("zhi", "")
        chart = (chart or {}).get("chart", {}) or {}
        palace_zhi = {"yue": chart.get("yue", {}).get("zhi", ri_zhi),
                      "shi": chart.get("shi", {}).get("zhi", ri_zhi),
                      "nian": chart.get("nian", {}).get("zhi", ri_zhi)}.get(palace_key, ri_zhi)
        if op == "流年值":
            return 1 if nz == palace_zhi else 0
        zhi_he = zhi_chong = 0
        chart_zhis = {
            nz,
            palace_zhi,
            *((chart.get(zhu) or {}).get("zhi", "") for zhu in ("nian", "yue", "ri", "shi")),
        }
        for it in ln.get("natal_interactions", []):
            for zr in it.get("zhi_rels", []):
                za, zb = zr.get("zhi_a", ""), zr.get("zhi_b", "")
                other = zb if za == nz else (za if zb == nz else "")
                if other == palace_zhi:
                    t = zr.get("type", "")
                    if t == "六合":
                        zhi_he = 1
                    elif t in ("三合", "三会"):
                        # 三合/三会必须三方齐备；两支只是半合，不算成局合会。
                        peers = const[t].get(nz, []) or []
                        complete = palace_zhi in peers and set(peers) <= chart_zhis
                        if complete:
                            zhi_he = 1
                    elif t in const["冲类"]:
                        zhi_chong = 1
        return zhi_he if op == "流年合" else zhi_chong
    if op == "流年克":
        _GANWX = const["天干五行"]
        _DIZHI_WUXING = const["地支五行"]
        target_wx = None
        if target == "日主":
            target_wx = _GANWX.get(fac.get("ri_gan", ""), "")
        else:
            for k in star_keys:
                st = fac.get("shishen", {}).get(k)
                if st and st.get("wuxing"):
                    target_wx = st["wuxing"]
                    break
            if not target_wx:
                target_wx = _target_wuxing_from_day_master(
                    fac.get("ri_gan", ""), star_keys, const
                )
        if target_wx and nian_gan:
            ke = const["五行生克"].get(_GANWX.get(nian_gan, ""), {}).get("克")
            ke2 = const["五行生克"].get(_DIZHI_WUXING.get(nz, ""), {}).get("克")
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
        # 引擎 2.6.15 起大运步骤带公历年段（start_year/end_year）——直接年份判断，免虚岁换算
        for s in fac.get("dayun_steps", []):
            if any(k in s.get("shi_shen", "") for k in star_keys) and s.get("start_year", 0) <= year <= s.get("end_year", 0):
                return 1
        return 0
    if op == "换运流年":
        # 换运首年 = 该步 start_year（引擎日期段直给）
        for s in fac.get("dayun_steps", []):
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
        return (ctx.get("snapshot", {}) or {}).get(args[0], 0)
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
        # 旬首地支 = 日支索引 - 日干索引（模 12，旬首天干恒甲）：甲子→子、甲戌→戌、己亥→午（甲午旬）、丙子→戌（甲戌旬）
        # 旧实现 xun="甲"+xun_gan[1:] 恒得"甲"，不在 constants 旬空表（key=甲子/甲戌…）→ 算子恒 0、空亡填实断语永不命中
        xun_zhi_idx = (_ZHI_ORDER.index(day_z) - _GAN_ORDER.index(day_g)) % 12
        xun = "甲" + _ZHI_ORDER[xun_zhi_idx]
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
    fac = ctx.get("fac", {})
    year = ctx.get("year", 0)
    for s in fac.get("dayun_steps", []):
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
    """干来源：流年干/大运干/日干。"""
    ln = ctx.get("liunian", {})
    fac = ctx.get("fac", {})
    if src == "流年干":
        return ln.get("nian_gan", "")
    if src == "大运干":
        gz = _current_dayun_gz(ctx)
        return gz[:1] if gz else ""
    if src == "日干":
        return fac.get("ri_gan", "")
    return ""
def _source_zhi(src: str, ctx: dict) -> str:
    """支来源：流年支/大运支/四柱支。"""
    ln = ctx.get("liunian", {})
    chart = ctx.get("chart", {}).get("chart", {}) or {}
    if src == "流年支":
        return ln.get("nian_zhi", "")
    if src == "大运支":
        gz = _current_dayun_gz(ctx)
        return gz[1:] if gz else ""
    if src == "日支":
        return (chart.get("ri") or {}).get("zhi", "")
    if src == "时支":
        return (chart.get("shi") or {}).get("zhi", "")
    if src == "年支":
        return (chart.get("nian") or {}).get("zhi", "")
    return ""
def _target_stars(target: str, gender: str, const: dict) -> tuple:
    """目标词 → 具体十神：六亲角色 → 十神大类 → 原子十神。"""
    role = const.get("六亲角色", {}).get(target, target)
    if isinstance(role, dict):
        gender_key = {"男": "male", "女": "female"}.get(gender, gender)
        role = role.get(gender_key, "")
    if role in const.get("十神大类", {}):
        return tuple(const["十神大类"][role])
    return (role,)

def _target_wuxing_from_day_master(day_gan: str, star_keys, const: dict) -> str:
    """本命星不现时，按十神与日主的生克关系推导目标五行。"""
    day_wx = const["天干五行"].get(day_gan, "")
    if not day_wx:
        return ""
    shengke = const["五行生克"]
    for star in star_keys:
        if star in ("比肩", "劫财"):
            return day_wx
        if star in ("食神", "伤官"):
            return shengke.get(day_wx, {}).get("生", "")
        if star in ("正财", "偏财"):
            return shengke.get(day_wx, {}).get("克", "")
        if star in ("正官", "七杀"):
            return next((wx for wx, rel in shengke.items() if rel.get("克") == day_wx), "")
        if star in ("正印", "偏印"):
            return next((wx for wx, rel in shengke.items() if rel.get("生") == day_wx), "")
    return ""


# ══════════════ 真值表求值 ══════════════

_FACTOR_ROWS = None

def load_factor_rows():
    """因子表（factors.csv 真值表）：列=原子事实实例，行=因子（多行=或）。含术数标记（bazi/ziwei——因子快照分）。"""
    global _FACTOR_ROWS
    if _FACTOR_ROWS is not None:
        return _FACTOR_ROWS
    import csv as _csv
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "factors", "factors.csv")
    rows = []
    with open(path, encoding="utf-8") as fh:
        rd = _csv.DictReader(fh)
        cols = [c for c in rd.fieldnames if c not in ("因子", "原语直通", "术数", "依据")]
        for r in rd:
            conds = {}
            for c in cols:
                v = (r.get(c) or "").strip()
                if v:
                    try:
                        conds[c] = int(v)
                    except ValueError:
                        conds[c] = v
            rows.append({"因子": r["因子"], "直通": (r.get("原语直通") or "").strip(),
                         "术数": (r.get("术数") or "bazi").strip(), "conds": conds})
    _FACTOR_ROWS = rows
    return rows

def _atomic(col: str, fac, gender, chart, ctx: dict = None, current_year: int = 0):
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
        v = _op(op, args, fac, gender, chart, current_year)
    elif op in _LIU_OP_NAMES:
        v = _liu_op(op, args, fac, gender, chart, ctx)
    else:
        raise ValueError(f"未知算子: {op}")
    if isinstance(v, str):
        # 「任意」= 取值模式（直读[ri_gan_wx,任意] 返回五行字符串、宫含[..,任意] 等）——
        # 返回字符串原值供断语约束匹配（如 `日主五行: 木`）；否则按期望值比较返回 0/1
        if args and args[-1] == "任意":
            return v
        return 1 if args and str(args[0]) == v else 0
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

def evaluate_factors(fac: dict, gender: str, chart: dict, shushi: Optional[str] = None,
                     current_year: int = 0) -> dict:
    """因子快照（真值表）：factors.csv 行条件匹配 → 因子值；直通因子=原语计算值。
    shushi="bazi"/"ziwei"：只算本术数因子（八字/紫微真分开——各自快照）；None=全算。
    条件列：带 [] = 原子事实（原语执行）；不带 [] = 因子引用（读快照——多遍稳定）。"""
    rows = load_factor_rows()
    if shushi:
        rows = [r for r in rows if r["术数"] in (shushi, "common")]
    return _evaluate_truth_table(
        rows,
        lambda expression: _atomic(expression, fac, gender, chart, current_year=current_year),
    )

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
        cols = [c for c in rd.fieldnames if c not in ("因子", "原语直通", "术数", "依据")]
        for r in rd:
            conds = {}
            for c in cols:
                v = (r.get(c) or "").strip()
                if v:
                    try:
                        conds[c] = int(v)
                    except ValueError:
                        conds[c] = v
            rows.append({"因子": r["因子"], "术数": r.get("术数", "bazi"),
                         "直通": (r.get("原语直通") or "").strip(), "conds": conds})
    _LIUNIAN_ROWS = rows
    return rows

def evaluate_liunian_factors(fac: dict, gender: str, chart: dict, liunian_data: dict,
                             zw_liunian_data: Optional[dict] = None,
                             year: int = 0, shushi: Optional[str] = None,
                             natal_snapshot: Optional[dict] = None) -> dict:
    """流年复合因子（表驱动）：读 factors_liunian.csv 逐行求值 → 流年因子快照。

    与 evaluate_factors 同构——流年因子定义在表，本函数只做机械求值。
    liunian_data: bazi.liunian 返回（调用方预取）；zw_liunian_data: 紫微流年四化。
    """
    ctx = {
        "liunian": liunian_data or {},
        "zw_liunian": zw_liunian_data or {},
        "year": year,
        "chart": chart,
        "fac": fac,
        "snapshot": natal_snapshot or {},  # 本命因子快照（流年支受克等算子读本命五行旺衰）
    }
    rows = load_liunian_rows()
    if shushi:
        rows = [r for r in rows if r["术数"] in (shushi, "common")]
    return _evaluate_truth_table(
        rows,
        lambda expression: _atomic(expression, fac, gender, chart, ctx),
    )


# ══════════════ 快照生成入口 ══════════════

def make_factors(pan: dict, current_year: int = 0) -> dict:
    """因子生成（本命）：本命盘 → 双盘因子快照 {八字: {...}, 紫微: {...}}。

    数据层真分开，调用层一次返回双盘。
    fac 已在 full_paipan 内嵌（pan["fac"]），此处不再单独 build。
    """
    context = {"性别": pan["gender"]}
    if current_year:
        context["当前年份"] = current_year
    return {
        "八字": evaluate_factors(
            pan["fac"], pan["gender"], pan, shushi="bazi", current_year=current_year
        ),
        "紫微": evaluate_factors(
            pan["fac"], pan["gender"], pan, shushi="ziwei", current_year=current_year
        ),
        "context": context,
    }


def make_liunian_factors(pan: dict, liunian_pan: dict, year: int = 0) -> dict:
    """因子生成（流年）：本命盘 + 流年盘 → 双盘流年因子快照 {八字: {...}, 紫微: {...}}。

    数据层真分开，调用层一次返回双盘。
    liunian_pan = liunian(pan, 年份) 的返回 {bazi, ziwei}。
    引用本命因子从本命八字快照读取；紫微流年四化取自 liunian_pan["ziwei"]。
    """
    bz = evaluate_factors(pan["fac"], pan["gender"], pan, shushi="bazi")
    base = dict(
        fac=pan["fac"], gender=pan["gender"], chart=pan, liunian_data=liunian_pan["bazi"],
        zw_liunian_data=liunian_pan["ziwei"], year=year,
        natal_snapshot=bz,
    )
    return {
        "_snapshot_type": "liunian",  # 流年快照标记——query 校验 yearly_* 域需此标记（本命快照隔离）
        "八字": evaluate_liunian_factors(**base, shushi="bazi"),
        "紫微": evaluate_liunian_factors(**base, shushi="ziwei"),
        "context": {"性别": pan["gender"]},
    }
