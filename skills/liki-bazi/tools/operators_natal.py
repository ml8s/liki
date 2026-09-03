"""本命因子算子与 pan 基础上下文聚合。"""
from __future__ import annotations

from typing import Optional

from errors import FactorEvaluateError
from factor_constants import load_constants
from factor_context import FactorContext

__all__ = [
    "_OP_NAMES", "_op", "_base_ctx_from_pan",
    "_shishen_from_pan",
]

# 本命算子名清单：_atomic 显式分派；新增算子必须同步登记与测试。
_OP_NAMES = frozenset({
    "现", "透", "藏", "得令", "有根", "旺", "弱", "缺", "克", "直读", "含", "宫含", "关系",
    "大运十神", "数量至少", "五行数量至少", "官杀取清", "为用", "为忌",
    "月支长生", "夫妻宫状态", "日支类型", "财库现", "财星入墓", "克者旺",
    "格神透", "月令本气", "时柱十神", "年柱官杀", "柱刑",
})


def _shishen_from_pan(chart: dict) -> dict:
    """直读 pan 聚合四柱十神 → {十神名: {wuxing,tou_gan,cang_zhi,has_root,de_ling,count}}。

    直接从 pan 构建 shishen 状态（算子不依赖外部中间结构）。
    仅做命理判定的机械聚合。
    """
    full = chart.get("full", {}) or {}
    gan_wx = load_constants()["天干五行"]
    de_ling_states = {"旺", "相"}
    states: dict = {}
    roots: set = set()
    for pillar in ("nian", "yue", "ri", "shi"):
        for item in (full.get(pillar, {}) or {}).get("shi_shens", []) or []:
            name = item.get("shi_shen", "")
            if not name:
                continue
            state = states.setdefault(name, {
                "wuxing": "", "tou_gan": False, "cang_zhi": False,
                "has_root": False, "de_ling": False, "count": 0,
            })
            gan = item.get("gan", "")
            if not state["wuxing"] and gan:
                state["wuxing"] = gan_wx.get(gan, "")
            src = item.get("source", "")
            if src == "gan":
                state["tou_gan"] = True
                state["count"] += 1
            elif src.endswith("qi"):
                state["cang_zhi"] = True
                state["count"] += 1
        hidden = (full.get(pillar, {}) or {}).get("cang_gan", {}) or {}
        for v in (hidden.get("main"), hidden.get("mid"), hidden.get("minor")):
            if v:
                roots.add(gan_wx.get(v, ""))
    wang_shuai = (chart.get("yongshen", {}) or {}).get("fu_yi", {}).get("wang_shuai", {}) or {}
    for name, st in states.items():
        if not st["wuxing"]:
            continue
        st["has_root"] = st["wuxing"] in roots
        st["de_ling"] = bool(wang_shuai.get(st["wuxing"]) in de_ling_states)
    return states


def _shishen(base, ten: str) -> Optional[dict]:
    """取某十神的聚合状态（factors._shishen_from_pan 产物）。"""
    return (base.get("shishen") or {}).get(ten)
def _class_wuxing(base, ten_class: str) -> str:
    """十神大类的五行（取该类第一个出现的十神之五行）。"""
    classes = load_constants()["十神大类"]
    for ten in classes.get(ten_class, []):
        st = _shishen(base, ten)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""
def _ten_to_wx(base, tens: list) -> str:
    """十神列表的五行（聚合）。"""
    for t in tens:
        st = _shishen(base, t)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""
def _wang_shuai(base) -> dict:
    return base.get("wuxing", {}).get("wang_shuai", {}) or {}
def _is_wang(base, wx: str) -> bool:
    wang_cfg = load_constants()["旺衰"]["旺"]
    return _wang_shuai(base).get(wx) in wang_cfg
def _is_weak(base, wx: str) -> bool:
    weak_cfg = load_constants()["旺衰"]["弱"]
    return _wang_shuai(base).get(wx) in weak_cfg
def _resolve_tens(tens, gender):
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


def _eval_natal_op(op: str, args, base: dict, gender: str, chart: dict,
                      current_year: int = 0) -> "int | str":
    """执行单个算子子句，返回 0/1（或事实值字符串）。args 为参数（列表或标量）。"""
    const = load_constants()
    if not isinstance(args, list):
        args = [args]

    if op == "现":
        tens = _resolve_tens(args, gender)
        return 1 if any((_shishen(base, t) or {}).get("count", 0) >= 1 for t in tens) else 0
    if op == "透":
        tens = _resolve_tens(args, gender)
        return 1 if any((_shishen(base, t) or {}).get("tou_gan") for t in tens) else 0
    if op == "藏":
        tens = _resolve_tens(args, gender)
        return 1 if any((_shishen(base, t) or {}).get("cang_zhi") for t in tens) else 0
    if op == "得令":
        tens = _resolve_tens(args, gender)
        return 1 if any((_shishen(base, t) or {}).get("de_ling") for t in tens) else 0
    if op == "有根":
        tens = _resolve_tens(args, gender)
        return 1 if any((_shishen(base, t) or {}).get("has_root") for t in tens) else 0
    if op in ("旺", "弱"):
        tens = args
        # 五行名原样（五行旺衰只看月令）；十神/大类/六亲 → 综合旺衰
        if tens and tens[0] in ("木", "火", "土", "金", "水"):
            wx = tens[0]
            return 1 if (_is_wang(base, wx) if op == "旺" else _is_weak(base, wx)) else 0
        resolved = _resolve_tens(tens, gender)
        wx = _ten_to_wx(base, resolved)
        if not wx:
            return 0
        if op == "弱":
            # 十神弱 = 其五行失令 且 不透 且 无根（三者皆弱）
            return 1 if (_is_weak(base, wx) and not any((_shishen(base, t) or {}).get("tou_gan") for t in resolved)
                         and not any((_shishen(base, t) or {}).get("has_root") for t in resolved)) else 0
        # 十神旺 = 得令（其五行月令旺相）或（透干且有根）或（五行为某地支本气=禄根）
        # ——《子平真诠》得令为重，失令者透且有根可补；禄根本气亦有力（《渊海子平》禄根论）
        if _is_wang(base, wx):
            return 1
        tou = any((_shishen(base, t) or {}).get("tou_gan") for t in resolved)
        gen = any((_shishen(base, t) or {}).get("has_root") for t in resolved)
        if tou and gen:
            return 1
        # 禄根：五行为某地支本气藏干（main qi）→ 有强根，作旺论
        full = (chart or {}).get("full", {}) or {}
        for pillar in ("nian", "yue", "ri", "shi"):
            pi = full.get(pillar, {})
            cang = pi.get("cang_gan", {}) or {}
            if cang.get("main"):
                main_wx = load_constants()["天干五行"].get(cang["main"], "")
                if main_wx == wx:
                    return 1
        return 0
    if op == "缺":
        count = base.get("wuxing", {}).get("count", {}) or {}
        return 1 if args and count.get(args[0], 0) == 0 else 0
    if op == "克":
        # 克(A,B)：A 五行克 B 五行（A/B 可为十神大类或五行）
        a, b = args[0], args[1]
        wx_a = _resolve_wx(base, gender, a)
        wx_b = _resolve_wx(base, gender, b)
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
            gan = base.get("ri_gan", "")
            val = load_constants()["天干五行"].get(gan)
            if expect == "任意":
                return val if val else 0     # 日主五行返回五行字符串（断语约束 `日主五行: 木` 匹配）
            return 1 if str(val) == str(expect) else 0
        val = _path_get(base, chart, path)
        if expect == "任意":
            # 返回字符串原值，供断语表做标量等值匹配（如 `月令格: 正财格`）。
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
            items = _path_get(base, chart, field) or []
            names = [it.get("name") or it.get("xing") for it in items] if isinstance(items, list) else []
        if "或" in value:
            return 1 if any(v in names for v in value.replace("或", "|").split("|")) else 0
        return 1 if value in names else 0
    if op == "宫含":
        # 宫含(宫位, 星, 条件) —— ziwei gong_wei
        return _zw_gong_op(base, chart, args)
    if op == "大运十神":
        return _dayun_op(base, chart, args, current_year)
    if op == "数量至少":
        # 数量至少(N, 十神...)：十神出现总数 ≥ N（事实计数——印杂等"多"的定量）
        n = int(args[0])
        tens = _resolve_tens(args[1:], gender)
        total = sum((_shishen(base, t) or {}).get("count", 0) for t in tens)
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
        count = base.get("wuxing", {}).get("count", {}).get(args[1], 0)
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
        fy = base.get("yongshen", {}).get("fu_yi", {}) or {}
        yong, xi, ji = fy.get("yong", ""), fy.get("xi", ""), fy.get("ji", "")
        tens = _resolve_tens(args, gender)
        wx = _ten_to_wx(base, tens)
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
        day_gan = base.get("ri_gan", "")
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
        day_gan = base.get("ri_gan", "")
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
        resolved = _resolve_tens(args, gender)
        wx = _ten_to_wx(base, resolved)
        if not wx:
            return 0
        ke_wx = [k for k, v in load_constants()["五行生克"].items() if v.get("克") == wx]
        return 1 if ke_wx and _is_wang(base, ke_wx[0]) else 0
    if op == "格神透":
        return _ge_shen_tou(base, chart)
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
    if op == "时柱十神":
        # 时柱天干十神（子女性别判断）——任意模式返回十神标量。
        full = chart.get("full", {}) or {}
        shi = full.get("shi", {}) or {}
        for ss in shi.get("shi_shens", []):
            if ss.get("source") == "gan":
                name = ss.get("shi_shen", "")
                if args and args[0] == "任意":
                    return name
                return 1 if name == args[0] else 0
        return "" if args and args[0] == "任意" else 0
    if op == "年柱官杀":
        return _nian_guan(base, chart)
        return _palace_bad(base, chart)
    # 流年算子（流年透/值/合/冲/克/忌神/财坏印/大运窗口/换运/岁运并临/干合等）由 _liu_op 处理
    raise FactorEvaluateError(f"未知算子: {op}")

def _resolve_wx(base, gender, arg):
    """解析任意参数为五行：具体五行名原样；十神/大类 → 五行。"""
    const = load_constants()
    if arg in ("木", "火", "土", "金", "水"):
        return arg
    if arg in const.get("十神大类", {}):
        return _class_wuxing(base, arg)
    if _shishen(base, arg):
        return (_shishen(base, arg) or {}).get("wuxing")
    # 配偶星等
    resolved = _resolve_tens([arg], gender)
    return _ten_to_wx(base, resolved)
def _path_get(base, chart, path: str):
    """按路径取值：优先 factors（基础因子），其次 chart 原始数据。"""
    obj = base if path.startswith(("shishen", "wuxing", "yongshen", "ri_gan")) else chart
    cur = obj
    for part in path.split("."):
        if isinstance(cur, dict):
            cur = cur.get(part)
        else:
            return None
    return cur
def _zw_gong_op(base, chart, args):
    """紫微宫位查询：宫含(宫位, 星, 条件)。

    四化（化禄/权/科/忌）用顶层 si_hua（{星:四化}）反推落宫——引擎本命四化无宫位字段，需按星找宫。
    """
    gong_name, star, cond = args[0], args[1], args[2] if len(args) > 2 else "任意"
    # 读紫微宫位/四化（chart["ziwei"]）
    zw = chart.get("ziwei", {}) or {}
    gw, top_sihua = zw.get("gong_wei", []) or [], zw.get("si_hua", {}) or {}
    # 化禄/权/科/忌：顶层 si_hua → 星 → 落宫
    if cond in ("化禄", "化权", "化科", "化忌"):
        sihua_map = {"化禄": "禄", "化权": "权", "化科": "科", "化忌": "忌"}
        target = sihua_map[cond]
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


def _dayun_op(base, chart, args, current_year: int = 0):
    """大运十神查询：大运十神(当前, 大类/任意)。任意模式返回十神大类标量。"""
    when, star_class = args[0], args[1]
    if when != "当前":
        return "" if star_class == "任意" else 0

    steps = base.get("dayun_steps", [])
    if current_year:
        selected = next((
            step for step in steps
            if step.get("start_year", 0) <= current_year <= step.get("end_year", 0)
        ), None)
    else:
        # 排盘时索引（原子层已提取 current_step_index），缺则回退 da_yun 原始表
        idx = base.get("dayun_current_index", -1)
        if idx < 0:
            dx = (chart.get("full") or chart.get("chart") or {}).get("da_yun", {}) or {}
            raw_steps = dx.get("steps", []) or steps
            idx = dx.get("current_step_index", -1)
            raw_steps = raw_steps or steps
            selected = raw_steps[idx] if 0 <= idx < len(raw_steps) else None
        else:
            selected = steps[idx] if 0 <= idx < len(steps) else None

    if selected is None:
        return "" if star_class == "任意" else 0
    shi_shen = (selected.get("shi_shen", "") or "").replace("运", "")
    if star_class == "任意":
        return _ten_class(shi_shen)
    resolved = _resolve_tens([star_class], chart.get("gender"))
    return 1 if shi_shen in resolved else 0
def _ge_shen_tou(base, chart):
    """格神透干：月令所定十神格神透出到四柱天干。"""
    ge_ju = (base.get("yongshen") or {}).get("ge_ju", {}) or {}
    ge_name = ge_ju.get("ge_ju", "")
    ge_shen = ge_name[:-1] if ge_name.endswith("格") else ""
    if ge_shen not in load_constants()["十神"]:
        return 0
    # 建禄 / 月刃 / 杂格不是十神格名；十神格以四柱天干十神同气为透。
    full = chart.get("full", {}) or {}
    for pillar in ("nian", "yue", "ri", "shi"):
        ss_list = (full.get(pillar, {}) or {}).get("shi_shens", []) or []
        for ss in ss_list:
            if ss.get("source") == "gan" and ss.get("shi_shen") == ge_shen:
                return 1
    return 0
def _nian_guan(base, chart):
    """年柱官杀攻身：年柱 shi_shens 含官杀。"""
    full = chart.get("full", {}) or {}
    nian_ss = (full.get("nian", {}) or {}).get("shi_shens", []) or []
    for ss in nian_ss:
        if ss.get("shi_shen") in ("正官", "七杀"):
            return 1
    return 0
def _base_ctx_from_pan(chart: dict) -> dict:
    """从 pan 直接构建算子求值所需的基础上下文（从 pan 直读，无中间层）。

    只读 pan、只做机械汇集且不修改调用方 pan；shishen 用 _shishen_from_pan，
    五行/日干/日支/大运/用神从 pan 取。
    输入为基础求值上下文（含 shishen 等聚合键）时原样返回。
    """
    if isinstance(chart, FactorContext):
        return chart.base
    if chart and "shishen" in chart:
        return chart
    chart = chart or {}
    full = chart.get("full", {}) or {}
    bazi_chart = chart.get("chart", {}) or {}
    da_yun = bazi_chart.get("da_yun") or chart.get("da_yun") or {}
    steps = chart.get("dayun_steps") or da_yun.get("steps", [])
    ri_gan = (full.get("ri", {}) or {}).get("gan", "") or chart.get("ri_gan", "")
    ri_zhi = (full.get("ri", {}) or {}).get("zhi", "") or chart.get("palace_ri", {}).get("zhi", "")
    fu_yi = (chart.get("yongshen", {}) or {}).get("fu_yi", {}) or {}
    ctx = {
        "shishen": _shishen_from_pan(chart),
        "wuxing": {
            "count": fu_yi.get("wuxing_count", {}) or {},
            "wang_shuai": fu_yi.get("wang_shuai", {}) or {},
        },
        "yongshen": chart.get("yongshen", {}) or {},
        "ri_gan": ri_gan,
        "palace_ri": {"zhi": ri_zhi},
        "dayun_steps": [
            {"name": s.get("name", ""), "start_year": s.get("start_year", 0),
             "end_year": s.get("end_year", 0), "shi_shen": s.get("shi_shen", "")}
            for s in steps
        ],
        "dayun_current_index": da_yun.get("current_step_index", -1),
    }
    return ctx


def _op(op: str, args, gender: str, chart: dict,
        current_year: int = 0) -> "int | str":
    """执行本命算子；base context 由 pan/FactorContext 内部构建。"""
    return _eval_natal_op(
        op, args, _base_ctx_from_pan(chart) or {}, gender, chart, current_year
    )
