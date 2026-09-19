"""流年因子算子：目标星、宫位、干支、神煞、大运与紫微四化判定。"""
from __future__ import annotations

from errors import FactorEvaluateError
from factor_constants import load_constants
from factor_tokens import FACTOR_WILDCARD
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
    if op == '流年长生':
        cs = (chart.get('full', {}) or {}).get('chang_sheng', []) or []
        nz = ln.get('nian_zhi', '')
        for it in cs:
            if it.get('index') == nz:
                return it.get('name', '') if args and args[0] == FACTOR_WILDCARD else 1 if it.get('name') == args[0] else 0
        return '' if args and args[0] == FACTOR_WILDCARD else 0

def _liu_handler_shensha(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    if op == '流年神煞':
        ss = ln.get('shensha', []) or []
        return 1 if any(((s.get('name') or '') == args[0] for s in ss)) else 0

def _liu_handler_target_star(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
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
        if op == '流年值':
            return 1 if nz == palace_zhi else 0
        atomic = _atomic_facts(ctx)
        relations = atomic.get('year_branch_relations', {})
        if op == '流年冲':
            return 1 if relations.get(palace_zhi) == 'liu_chong' else 0
        if op == '流年合':
            if relations.get(palace_zhi) == 'liu_he':
                return 1
            return 1 if any(
                item.get('kind') in {'san_he', 'san_hui'}
                and item.get('includes_year')
                and palace_zhi in item.get('branches', [])
                for item in atomic.get('combinations', [])
            ) else 0

    if op == '流年克':
        role = const.get('六亲角色', {}).get(target, target)
        if isinstance(role, dict):
            gender_key = const.get('性别别名', {}).get(gender, gender)
            role = role.get(gender_key, '')
        atomic_key = const.get('流年克目标', {}).get(role, 'day_master')
        return 1 if _atomic_facts(ctx).get('controls_targets', {}).get(atomic_key) else 0

def _liu_handler_jishen(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    if op == '忌神干':
        return 1 if _atomic_facts(ctx).get('unfavorable_gan') else 0

    if op == '忌神支':
        return 1 if _atomic_facts(ctx).get('unfavorable_branch') else 0

    if op == '财坏印流年':
        return 1 if _atomic_facts(ctx).get('wealth_breaks_seal') else 0

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
    if op == '流年宫化':
        # [宫位, 星曜, 四化]——与「宫含」同构的三维 key：星曜是显式维度，
        # 因子值 0/1 自身即构成完整命题（如「廉贞化忌落流年官禄宫」），
        # 不需要任何返回值之外的旁路来补全事实。
        if len(args) != 3:
            raise FactorEvaluateError(
                f"流年宫化需 3 参 [宫位,星曜,四化]，实得 {len(args)}: {args}"
            )
        gong, star, hua = args
        zw = ctx.get('zw_liunian', {}) or {}
        si_hua_gong = zw.get('si_hua_gong', {}) or {}
        si_hua = zw.get('si_hua', {}) or {}
        for hit_star, gname in si_hua_gong.items():
            if gong not in (FACTOR_WILDCARD, gname):
                continue
            if star not in (FACTOR_WILDCARD, hit_star):
                continue
            if si_hua.get(hit_star) == hua:
                return 1
        return 0

    if op == '引用本命':
        return (ctx.get('snapshot', {}) or {}).get(args[0], 0)


def _liu_handler_flow_star(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                           current_year: int, const: dict, ln: dict, nz: str,
                           nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> int:
    """机械检查流曜是否入指定流年宫位；星曜含义由断语表表达。"""
    if len(args) != 2:
        raise FactorEvaluateError(f"流曜入宫需 2 参 [星曜,宫位]，实得 {len(args)}: {args}")
    star, palace = str(args[0]), str(args[1])
    for item in (ctx.get('zw_liunian', {}).get('gong_wei', []) or []):
        if palace != item.get('name'):
            continue
        if star in (item.get('xing_yao', []) or []):
            return 1
    return 0


def _liu_handler_banhe(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                       current_year: int, const: dict, ln: dict, nz: str,
                       nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> int:
    """读取 engine 三合半合原子事实；旺支规则不在 Python 复算。"""
    if len(args) < 2:
        raise FactorEvaluateError(f"半合需至少 2 参 [来源/地支,...]，实得 {len(args)}: {args}")
    branches = set()
    for value in args:
        if value in const.get('地支', []):
            branches.add(value)
        else:
            resolved = _source_zhi(value, ctx)
            if resolved:
                branches.add(resolved)
    return 1 if any(
        item.get('kind') == 'half_he' and branches <= set(item.get('branches', []))
        for item in _atomic_facts(ctx).get('combinations', [])
    ) else 0

def _liu_handler_mechanical(op: str, args: list, base: dict, gender: str, chart: dict, ctx: dict,
                         current_year: int, const: dict, ln: dict, nz: str,
                         nian_gan: str, ss_year: str, star_keys: tuple, target: str) -> "int | str":
    if op == '干支相等':
        pair = tuple(args)
        if pair == ("大运", "流年"):
            return 1 if _atomic_facts(ctx).get('dayun_equals_year_pillar') else 0
        if pair == ("流年", "日柱"):
            return 1 if _atomic_facts(ctx).get('year_equals_day_pillar') else 0
        raise FactorEvaluateError(f"干支相等不支持的来源组合: {args}")

    if op == '干克':
        pair = tuple(args)
        if pair == ("流年干", "日干"):
            return 1 if _atomic_facts(ctx).get('year_gan_controls_day_gan') else 0
        if pair == ("日干", "流年干"):
            return 1 if _atomic_facts(ctx).get('day_gan_controls_year_gan') else 0
        raise FactorEvaluateError(f"干克不支持的来源组合: {args}")

    if op == '支冲':
        day_zhi_source = next(
            key for key, spec in const['干支来源'].items()
            if spec == {'源': '四柱', '柱': 'ri', '部分': '支'}
        )
        if tuple(args) == ("大运支", "流年支"):
            return 1 if _atomic_facts(ctx).get('dayun_zhi_clashes_year') else 0
        if tuple(args) != ("流年支", day_zhi_source):
            raise FactorEvaluateError(f"支冲不支持的来源组合: {args}")
        day_zhi = _source_zhi(day_zhi_source, ctx)
        relation = _atomic_facts(ctx).get('year_branch_relations', {}).get(day_zhi)
        return 1 if relation == 'liu_chong' else 0

    if op == '旬空':
        if tuple(args) != ("日柱", "流年支"):
            raise FactorEvaluateError(f"旬空不支持的来源组合: {args}")
        year_zhi = _source_zhi('流年支', ctx)
        return 1 if year_zhi in _atomic_facts(ctx).get('day_void_branches', []) else 0

    if op == '三刑':
        # 三刑[来源, 刑组]：来源决定参与判定的地支，刑组决定是哪一组刑。
        # 两者都进 key——答案（刑组）不在参数里时的证据旁路已被删除。
        if len(args) != 2:
            raise FactorEvaluateError(
                f"三刑需 2 参 [来源,刑组]，实得 {len(args)}: {args}"
            )
        source, group = args[0], args[1]
        if source != "流年支":
            raise FactorEvaluateError(f"三刑来源只支持 流年支: {source}")
        members = _xing_members(group)
        if members is None:
            raise FactorEvaluateError(f"三刑刑组无效: {group}")
        source_zhi = _source_zhi(source, ctx)
        for item in _atomic_facts(ctx).get('combinations', []):
            if item.get('kind') != 'xing':
                continue
            if not item.get("includes_year"):
                continue
            branches = item.get('branches') or []
            if source_zhi in branches and set(branches) == members:
                return 1
        return 0

    if op == '流年支受克':
        if not args:
            return 0
        wx = str(args[0])
        snap = ctx.get('snapshot', {})
        natal_factor = wx + const['五行旺因子后缀']
        return 1 if snap.get(natal_factor) and wx in _atomic_facts(ctx).get('year_branch_controlled_by', []) else 0

    if op == '年柱干伏吟':
        return 1 if _atomic_facts(ctx).get('year_gan_equals_natal_year_gan') else 0

    if op == '天干合':
        pair = tuple(args)
        if pair == ('大运干', '流年干'):
            return 1 if _atomic_facts(ctx).get('dayun_gan_combines') else 0
        if pair != ('流年干', '日干'):
            raise FactorEvaluateError(f"天干合不支持的来源组合: {args}")
        return 1 if _atomic_facts(ctx).get('gan_combines') else 0

_LIU_OP_HANDLERS = {
        "流年长生": _liu_handler_longevity,
        "流年神煞": _liu_handler_shensha,
        "流年合": _liu_handler_target_star,
        "流年值": _liu_handler_target_star,
        "流年冲": _liu_handler_target_star,
        "流年克": _liu_handler_target_star,
        "流年透": _liu_handler_target_star,
        "忌神干": _liu_handler_jishen,
        "忌神支": _liu_handler_jishen,
        "财坏印流年": _liu_handler_jishen,
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
def _atomic_facts(ctx: dict) -> dict:
    facts = ctx.get('liunian', {}).get('atomic_facts')
    return facts if isinstance(facts, dict) else {}


def _source_zhi(src: str, ctx: dict) -> str:
    """支来源：流年支/大运支/四柱支。"""
    return _source_value(src, ctx, "支")


def _xing_members(group: str) -> frozenset | None:
    """校验刑组名并返回其成员集合；刑组名不自洽时返回 None。

    刑组闭集取自 constants「三刑」（地支 → 同组其余地支），因此不写死任何
    地支字面量。成员以集合比较，与引擎组合事实的支序、与本命侧组名写法
    （丑戌未 / 丑未戌 同组）均无关。
    """
    table = load_constants().get("三刑", {})
    if not isinstance(group, str):
        return None
    members = set(group)
    if not members or len(members) > 3:
        return None
    if len(members) == 1 and len(group) != 2:
        return None
    if len(members) > 1 and len(group) != len(members):
        return None
    for branch in members:
        partners = table.get(branch)
        if partners is None or set(partners) - members:
            return None
    return frozenset(members)


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
