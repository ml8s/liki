"""流年因子算子：目标星、宫位、干支、神煞、大运与紫微四化判定。"""
from __future__ import annotations

from errors import FactorEvaluateError
from factor_constants import load_constants
from operators_natal import _base_ctx_from_pan

# 流年算子名清单：_atomic 显式分派；新增算子必须同步登记与测试。
_LIU_OP_NAMES = frozenset({
    "流年长生", "流年神煞", "流年透", "流年值", "流年合", "流年冲", "流年克",
    "忌神干", "忌神支", "财坏印流年", "大运窗口流年", "换运流年", "流年宫化", "引用本命",
    "干支相等", "干克", "支冲", "三刑", "旬空", "流年支受克", "年柱干伏吟", "天干合",
    "半合", "流曜入宫",
})


def _liu_handler_longevity(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    year = current_year
    if op == '流年长生':
        cs = (chart.get('full', {}) or {}).get('chang_sheng', []) or []
        nz = ln.get('nian_zhi', '')
        for it in cs:
            if it.get('index') == nz:
                return it.get('name', '') if args and args[0] == '任意' else 1 if it.get('name') == args[0] else 0
        return '' if args and args[0] == '任意' else 0

def _liu_handler_shensha(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    year = current_year
    if op == '流年神煞':
        ss = ln.get('shensha', []) or []
        return 1 if any(((s.get('name') or '') == args[0] for s in ss)) else 0

def _liu_handler_target_star(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    year = current_year
    if op == '流年透':
        return 1 if ss_year in star_keys else 0

    if op in ('流年值', '流年合', '流年冲'):
        palace_key = const.get('事件宫位', {}).get(target, const['事件宫位默认'])
        ri_zhi = base.get('palace_ri', {}).get('zhi', '')
        pan_chart = (chart or {}).get('chart', {}) or {}
        pillars_zhi = {
            zhu: (pan_chart.get(zhu, {}) or {}).get('zhi', ri_zhi)
            for zhu in const['四柱']
        }
        palace_zhi = pillars_zhi.get(palace_key, ri_zhi)
        if target in const['四柱序号']:
            palace_zhi = _source_zhi(target, ctx)
        chong_map = const.get(const['关系取冲类型'], {})
        if op == '流年冲' and palace_zhi:
            if chong_map.get(nz) == palace_zhi:
                return 1
        if op == '流年值':
            return 1 if nz == palace_zhi else 0
        zhi_he = zhi_chong = 0
        chart_zhis = {nz, palace_zhi, *pillars_zhi.values()}
        for it in ln.get('natal_interactions', []):
            for zr in it.get('zhi_rels', []):
                za, zb = (zr.get('zhi_a', ''), zr.get('zhi_b', ''))
                other = zb if za == nz else za if zb == nz else ''
                if other == palace_zhi:
                    t = zr.get('type', '')
                    he_mode = const.get('关系取合类型', {}).get(t)
                    if he_mode == '两支':
                        zhi_he = 1
                    elif he_mode == '全组':
                        peers = const[t].get(nz, []) or []
                        complete = palace_zhi in peers and set(peers) <= chart_zhis
                        if complete:
                            zhi_he = 1
                    elif t == const['关系取冲类型']:
                        zhi_chong = 1
        return zhi_he if op == '流年合' else zhi_chong

    if op == '流年克':
        _GANWX = const['天干五行']
        _DIZHI_WUXING = const['地支五行']
        target_wx = None
        if target == const['日主目标']:
            target_wx = _GANWX.get(base.get('ri_gan', ''), '')
        else:
            for k in star_keys:
                st = base.get('shishen', {}).get(k)
                if st and st.get('wuxing'):
                    target_wx = st['wuxing']
                    break
            if not target_wx:
                target_wx = _target_wuxing_from_day_master(base.get('ri_gan', ''), star_keys, const)
        if target_wx and nian_gan:
            ke = const['五行生克'].get(_GANWX.get(nian_gan, ''), {}).get('克')
            ke2 = const['五行生克'].get(_DIZHI_WUXING.get(nz, ''), {}).get('克')
            return 1 if ke == target_wx or ke2 == target_wx else 0
        return 0

def _liu_handler_yongshen(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    year = current_year
    if op == '忌神干':
        ji = base.get('yongshen', {}).get('fu_yi', {}).get('ji', '')
        ji_wx = const['天干五行'].get(ji, '') or ji
        return 1 if ji_wx and const['天干五行'].get(nian_gan, '') == ji_wx else 0

    if op == '忌神支':
        ji = base.get('yongshen', {}).get('fu_yi', {}).get('ji', '')
        ji_wx = const['天干五行'].get(ji, '') or ji
        return 1 if ji_wx and const['地支五行'].get(nz, '') == ji_wx else 0

    if op == '财坏印流年':
        _YIN2 = set(const['十神大类'][const['印星十神大类']])
        if nz and nian_gan and (ss_year in _YIN2):
            ke = const['五行生克'].get(const['地支五行'].get(nz, ''), {}).get('克')
            return 1 if ke == const['天干五行'].get(nian_gan, '') else 0
        return 0

def _liu_handler_dayun(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    year = current_year
    if op == '大运窗口流年':
        for s in base.get('dayun_steps', []):
            if any((k in s.get('shi_shen', '') for k in star_keys)) and s.get('start_year', 0) <= year <= s.get('end_year', 0):
                return 1
        return 0

    if op == '换运流年':
        for s in base.get('dayun_steps', []):
            if any((k in s.get('shi_shen', '') for k in star_keys)):
                if year in (s.get('start_year', 0), s.get('start_year', 0) + 1):
                    return 1
        return 0


def _liu_handler_ziwei(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    year = current_year
    if op == '流年宫化':
        gong, hua = (args[0], args[1])
        sg = ctx.get('zw_liunian', {}).get('si_hua_gong', {}) or {}
        sh = ctx.get('zw_liunian', {}).get('si_hua', {}) or {}
        for star, gname in sg.items():
            if gname == gong and sh.get(star) == hua:
                return 1
        return 0

    if op == '引用本命':
        return (ctx.get('snapshot', {}) or {}).get(args[0], 0)


def _liu_handler_flow_star(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                           current_year: int, const: dict, ln: dict, nz: str,
                           nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> int:
    """机械检查流曜是否入指定流年宫位；星曜含义由断语表表达。"""
    if len(args) < 2:
        return 0
    star, palace = str(args[0]), str(args[1])
    for item in (ctx.get('zw_liunian', {}).get('gong_wei', []) or []):
        engine_name = item.get('name')
        suffix = const['流年宫名后缀']
        if engine_name in const['流年宫名不加后缀']:
            aliases = {engine_name}
        else:
            aliases = {engine_name, engine_name + suffix}
        if palace not in aliases:
            continue
        if star in (item.get('xing_yao', []) or []):
            return 1
    return 0


def _liu_handler_banhe(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                       current_year: int, const: dict, ln: dict, nz: str,
                       nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> int:
    """机械检查：三合组旺支是否参与两支组合。

    查 constants.json「三合半合」：子/午/卯/酉为各局旺支；缺旺支的
    生墓两支为拱合，不属于本算子的半合事实。
    """
    if len(args) < 2:
        return 0
    branches = set(const.get("地支", []))

    def branch(value: str) -> str:
        return value if value in branches else _source_zhi(value, ctx)

    a, b = branch(str(args[0])), branch(str(args[1]))
    half = const.get("三合半合", {})
    return 1 if b in half.get(a, []) or a in half.get(b, []) else 0


def _liu_handler_mechanical(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    year = current_year
    if op == '干支相等':
        ga, gb = (_source_ganzhi(args[0], ctx), _source_ganzhi(args[1], ctx))
        return 1 if ga and ga == gb else 0

    if op == '干克':
        g1, g2 = (_source_gan(args[0], ctx), _source_gan(args[1], ctx))
        if not g1 or not g2:
            return 0
        ke = const['五行生克'].get(const['天干五行'].get(g1, ''), {}).get('克')
        return 1 if ke and ke == const['天干五行'].get(g2, '') else 0

    if op == '支冲':
        z1, z2 = (_source_zhi(args[0], ctx), _source_zhi(args[1], ctx))
        chong_map = const.get(const['关系取冲类型'], {})
        return 1 if z1 and z2 and (chong_map.get(z1) == z2) else 0

    if op == '三刑':
        available = {}
        for a in args:
            zv = _source_zhi(a, ctx)
            if zv:
                available[a] = zv
        pan = chart
        pan_chart = (pan or {}).get('chart', {}) or {}
        pillar_zhis = []
        for zhu in const['四柱']:
            z = (pan_chart.get(zhu) or {}).get('zhi', '')
            if z:
                pillar_zhis.append(z)
        zhis = list(available.values()) + pillar_zhis
        cnt: dict = {}
        for z in zhis:
            cnt[z] = cnt.get(z, 0) + 1
        for k, v in const['三刑'].items():
            members = (k, *v)
            if cnt.get(k, 0) >= 1 and all((cnt.get(g, 0) >= (2 if g == k else 1) for g in v)):
                ctx.setdefault('evidence', {})['三刑流年'] = {
                    'group': k + ''.join(v),
                    'members': list(dict.fromkeys(members)),
                    'sources': available,
                    'pillars': pillar_zhis,
                }
                return 1
        return 0

    if op == '旬空':
        gz = _source_ganzhi(args[0], ctx)
        nz2 = _source_zhi(args[1], ctx)
        if not gz or len(gz) < 2 or (not nz2):
            return 0
        day_g, day_z = (gz[0], gz[1])
        _GAN_ORDER = const['天干']
        _ZHI_ORDER = const['地支']
        if day_g not in _GAN_ORDER or day_z not in _ZHI_ORDER:
            return 0
        xun_zhi_idx = (_ZHI_ORDER.index(day_z) - _GAN_ORDER.index(day_g)) % 12
        xun = const['旬空起点'] + _ZHI_ORDER[xun_zhi_idx]
        return 1 if xun in const['旬空'] and nz2 in const['旬空'][xun] else 0

    if op == '流年支受克':
        if not args:
            return 0
        wx = str(args[0])
        ln = ctx.get('liunian', {})
        nz = ln.get('nian_zhi', '')
        zhi_wx = const.get('地支五行', {}).get(nz, '')
        if not zhi_wx:
            return 0
        snap = ctx.get('snapshot', {})
        natal_factor = wx + const['五行旺因子后缀']
        return 1 if snap.get(natal_factor) and const.get('五行生克', {}).get(wx, {}).get('克') == zhi_wx else 0

    if op == '年柱干伏吟':
        ln = ctx.get('liunian', {})
        nian_gan = (ctx.get('chart', {}).get('chart', {}) or {}).get('nian', {}).get('gan', '')
        return 1 if ln.get('nian_gan') and ln.get('nian_gan') == nian_gan else 0

    if op == '天干合':
        g1, g2 = (_source_gan(args[0], ctx), _source_gan(args[1], ctx))
        return 1 if g1 and g2 and (const['天干五合'].get(g1) == g2) else 0

_LIU_OP_HANDLERS = {
        "流年长生": _liu_handler_longevity,
        "流年神煞": _liu_handler_shensha,
        "流年合": _liu_handler_target_star,
        "流年值": _liu_handler_target_star,
        "流年冲": _liu_handler_target_star,
        "流年克": _liu_handler_target_star,
        "流年透": _liu_handler_target_star,
        "忌神干": _liu_handler_yongshen,
        "忌神支": _liu_handler_yongshen,
        "财坏印流年": _liu_handler_yongshen,
        "大运窗口流年": _liu_handler_dayun,
        "换运流年": _liu_handler_dayun,
        "引用本命": _liu_handler_ziwei,
        "流曜入宫": _liu_handler_flow_star,
        "流年宫化": _liu_handler_ziwei,
        "流年支受克": _liu_handler_mechanical,
        "干克": _liu_handler_mechanical,
        "干支相等": _liu_handler_mechanical,
        "旬空": _liu_handler_mechanical,
        "支冲": _liu_handler_mechanical,
        "年柱干伏吟": _liu_handler_mechanical,
        "天干合": _liu_handler_mechanical,
        "三刑": _liu_handler_mechanical,
        "半合": _liu_handler_banhe,
}


def _liu_op(op: str, args, gender: str, chart: dict, ctx=None) -> int:
    """执行流年算子；具体语义由领域分组 handler 分派。"""
    ctx = ctx or {}
    ln = ctx.get("liunian", {})
    base = _base_ctx_from_pan(chart) or {}
    target_ops = ("流年透", "流年值", "流年合", "流年冲", "流年克", "大运窗口流年", "换运流年")
    target = args[0] if op in target_ops else ""
    if op in target_ops and not target:
        raise FactorEvaluateError(f"{op} 必须显式传入 target 参数")
    const = load_constants()
    common = dict(
        op=op, args=args, base=base, gender=gender, chart=chart, ctx=ctx,
        current_year=ctx.get("year", 0), const=const,
        ln=ln, nz=ln.get("nian_zhi", ""), nian_gan=ln.get("nian_gan", ""),
        ss_year=ln.get("shi_shen", ""),
        star_keys=_target_stars(target, gender, const), target=target,
    )
    handler = _LIU_OP_HANDLERS.get(op)
    if handler is None:
        return 0
    return handler(**common)
def _current_dayun_gz(ctx: dict) -> str:
    """当前大运干支（机械——查大运步骤公历年段）。"""
    base = ctx.get("base", {})
    year = ctx.get("year", 0)
    for s in base.get("dayun_steps", []):
        if s.get("start_year", 0) <= year <= s.get("end_year", 0):
            return s.get("name", "")
    return ""
def _source_ganzhi(src: str, ctx: dict) -> str:
    """干支来源解析：大运/流年/日柱 → 干支。"""
    return _source_value(src, ctx, "干支")


def _source_gan(src: str, ctx: dict) -> str:
    """干来源：流年干/大运干/日干。"""
    return _source_value(src, ctx, "干")


def _source_zhi(src: str, ctx: dict) -> str:
    """支来源：流年支/大运支/四柱支。"""
    return _source_value(src, ctx, "支")


def _source_value(src: str, ctx: dict, part: str) -> str:
    """按 constants.json「干支来源」解析指定干/支/干支。"""
    spec = load_constants().get("干支来源", {}).get(src)
    if not spec:
        return ""
    if spec.get("部分") != part:
        return ""
    source = spec.get("源", "")
    if source == "流年":
        ln = ctx.get("liunian", {})
        gan, zhi = ln.get("nian_gan", ""), ln.get("nian_zhi", "")
    elif source == "大运":
        gz = _current_dayun_gz(ctx)
        gan, zhi = gz[:1], gz[1:]
    elif source == "四柱":
        pillar = (ctx.get("chart", {}).get("chart", {}) or {}).get(spec.get("柱", ""), {}) or {}
        gan, zhi = pillar.get("gan", ""), pillar.get("zhi", "")
    elif source == "基础":
        value = (ctx.get("base", {}) or {}).get(spec.get("字段", ""), "")
        return str(value) if part == "干" else ""
    else:
        return ""
    if part == "干":
        return gan
    if part == "支":
        return zhi
    return gan + zhi


def _target_stars(target: str, gender: str, const: dict) -> tuple:
    """目标词 → 具体十神：六亲角色 → 十神大类 → 原子十神。"""
    role = const.get("六亲角色", {}).get(target, target)
    if isinstance(role, dict):
        gender_key = const.get("性别别名", {}).get(gender, gender)
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
    classes = const.get("十神大类", {})
    relations = const.get("十神大类日主关系", {})
    for star in star_keys:
        class_name = next((name for name, members in classes.items() if star in members), "")
        relation = relations.get(class_name)
        if not relation:
            continue
        name = relation.get("关系", "")
        if name == "同":
            return day_wx
        if relation.get("方向") == "出":
            return shengke.get(day_wx, {}).get(name, "")
        if relation.get("方向") == "入":
            return next((wx for wx, rel in shengke.items() if rel.get(name) == day_wx), "")
    return ""
