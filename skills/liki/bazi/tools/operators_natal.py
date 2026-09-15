"""本命因子算子与 pan 基础上下文聚合。"""
from __future__ import annotations

from typing import Optional

from errors import FactorEvaluateError
from factor_constants import load_constants
from factor_context import FactorContext

# 本命算子名清单：_atomic 显式分派；新增算子必须同步登记与测试。
_OP_NAMES = frozenset({
    "现", "透", "藏", "得令", "有根", "旺", "弱", "缺", "克", "直读", "含", "宫含", "关系",
    "大运十神", "大运十神有根", "数量至少", "五行数量至少", "官杀取清", "为用", "为忌",
    "月支长生", "夫妻宫状态", "日支类型", "财库现", "财星入墓", "克者旺",
    "格神透", "月令本气", "时柱十神", "年柱十神", "禄根", "年柱官杀", "柱刑",
    "大限宫位",
})


def _ten_god_states_from_pan(chart: dict) -> dict:
    """直读 engine 十神状态，按十神名建立机械索引。"""
    full = chart.get("full", {}) or {}
    return {
        item.get("shi_shen", ""): item
        for item in full.get("ten_god_states", []) or []
        if isinstance(item, dict) and item.get("shi_shen")
    }


def _ten_god_state(base, ten: str) -> Optional[dict]:
    """取某十神的 engine 状态。"""
    return (base.get("ten_god_states") or {}).get(ten)


def _element_states_from_pan(chart: dict) -> dict:
    """直读 engine 五行状态，按五行名建立机械索引。"""
    full = chart.get("full", {}) or {}
    return {
        item.get("wuxing", ""): item
        for item in full.get("element_states", []) or []
        if isinstance(item, dict) and item.get("wuxing")
    }


def _pillar_key(branch: str, const: dict | None = None) -> str:
    """按四柱序号表解析柱字段名。"""
    const = const or load_constants()
    index = const["四柱序号"].get(branch)
    return const["四柱"][index] if index is not None else ""


def _class_wuxing(base, ten_class: str) -> str:
    """十神大类的五行（取该类第一个出现的十神之五行）。"""
    classes = load_constants()["十神大类"]
    for ten in classes.get(ten_class, []):
        st = _ten_god_state(base, ten)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""
def _ten_to_wx(base, tens: list) -> str:
    """十神列表的五行（聚合）。"""
    for t in tens:
        st = _ten_god_state(base, t)
        if st and st.get("wuxing"):
            return st["wuxing"]
    return ""

def _element_state(base, wuxing: str) -> dict:
    return (base.get("element_states") or {}).get(wuxing) or {}
def _resolve_tens(tens, gender):
    """解析十神参数：六亲角色与十神大类 → 具体十神列表；配偶星/子女星按性别。"""
    const = load_constants()
    classes = const["十神大类"]
    roles = const["六亲角色"]
    gender_key = load_constants()["性别别名"].get(gender, gender)
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
        return 1 if any((_ten_god_state(base, t) or {}).get("count", 0) >= 1 for t in tens) else 0
    if op == "透":
        tens = _resolve_tens(args, gender)
        return 1 if any((_ten_god_state(base, t) or {}).get("transparent") for t in tens) else 0
    if op == "藏":
        tens = _resolve_tens(args, gender)
        return 1 if any((_ten_god_state(base, t) or {}).get("hidden") for t in tens) else 0
    if op == "得令":
        tens = _resolve_tens(args, gender)
        return 1 if any((_ten_god_state(base, t) or {}).get("timely") for t in tens) else 0
    if op == "有根":
        tens = _resolve_tens(args, gender)
        return 1 if any((_ten_god_state(base, t) or {}).get("rooted") for t in tens) else 0
    if op in ("旺", "弱"):
        tens = args
        # 五行名读取季节旺弱；十神/大类/六亲读取 engine 组合旺弱。
        if tens and tens[0] in const["五行"]:
            wx = tens[0]
            expected = "strong" if op == "旺" else "weak"
            return 1 if _element_state(base, wx).get("season_strength") == expected else 0
        resolved = _resolve_tens(tens, gender)
        expected = "strong" if op == "旺" else "weak"
        return 1 if any(
            (_ten_god_state(base, t) or {}).get("strength") == expected for t in resolved
        ) else 0
    if op == "缺":
        count = base.get("wuxing", {}).get("count", {}) or {}
        return 1 if args and count.get(args[0], 0) == 0 else 0
    if op == "克":
        # 克(A,B)：读取 engine 五行生克方向。
        a, b = args[0], args[1]
        wx_a = _resolve_wx(base, gender, a)
        wx_b = _resolve_wx(base, gender, b)
        if not wx_a or not wx_b:
            return 0
        return 1 if _element_state(base, wx_a).get("controls") == wx_b else 0
    if op == "直读":
        path, expect = args[0], args[1]
        if path == "gender":
            # gender 是求值上下文；mock/单元测试可只传 gender 参数而无完整 pan。
            val = gender
            if expect == "任意":
                return val if val else 0
            return 1 if str(val) == str(expect) else 0
        if path == "ri_gan_wx":
            val = base.get("day_master_element", "")
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
            for zhu in const["四柱"]:
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
        return _dayun_op(base, chart, args, current_year, gender)
    if op == "大运十神有根":
        return _dayun_root_op(base, chart, args, current_year, gender)
    if op == "数量至少":
        # 数量至少(N, 十神...)：十神出现总数 ≥ N（事实计数——印杂等"多"的定量）
        n = int(args[0])
        tens = _resolve_tens(args[1:], gender)
        total = sum((_ten_god_state(base, t) or {}).get("count", 0) for t in tens)
        return 1 if total >= n else 0
    if op == "关系":
        # 关系[field, 组名]——读取引擎完整合会冲刑组。
        field, group = args[0], args[1]
        relation_groups = (chart.get("full", {}) or {}).get("relation_groups", [])
        return 1 if any(
            item.get("field") == field and item.get("group") == group
            for item in relation_groups
            if isinstance(item, dict)
        ) else 0
    if op == "柱刑":
        # 读取 engine 柱位刑伤事实；日支由夫妻宫状态专管。
        pillar = _pillar_key(args[0] if args else "")
        return 1 if _atomic_facts(chart).get("pillar_punishments", {}).get(pillar) else 0
    if op == "五行数量至少":
        # 五行数量至少(N, 五行)：四柱五行计数 ≥ N（数量事实，阈值在因子表）
        n = int(args[0])
        count = base.get("wuxing", {}).get("count", {}).get(args[1], 0)
        return 1 if count >= n else 0
    if op == "官杀取清":
        # Engine atomic fact: 官杀混杂且七杀柱被合/冲后取清。
        return 1 if _atomic_facts(chart).get("officer_killing_cleaned") else 0
    if op in ("为用", "为忌"):
        # 为用(十神类)：该十神五行 ∈ {用, 喜}；为忌：== 忌（引擎五神体系 yong/xi/ji）
        fy = base.get("yongshen", {}).get("fu_yi", {}) or {}
        favorable_fields = const["用忌映射"][op]
        tens = _resolve_tens(args, gender)
        wx = _ten_to_wx(base, tens)
        if not wx:
            return 0
        favorable = {fy.get(field, "") for field in favorable_fields}
        return 1 if wx in favorable else 0
    if op == "月支长生":
        value = _atomic_facts(chart).get("month_longevity", "")
        return value if args and args[0] == "任意" else 1 if value == args[0] else 0
    if op == "夫妻宫状态":
        # Engine atomic fact: 日支按冲、合、刑、害优先级归并夫妻宫状态。
        value = _atomic_facts(chart).get("spouse_palace_state", "")
        return value if args and args[0] == "任意" else 1 if value == args[0] else 0
    if op == "日支类型":
        value = _atomic_facts(chart).get("day_branch_type", "")
        return value if args and args[0] == "任意" else 1 if value == args[0] else 0
    if op == "财库现":
        return 1 if _atomic_facts(chart).get("wealth_tomb_present") else 0
    if op == "财星入墓":
        return 1 if _atomic_facts(chart).get("wealth_star_in_tomb") else 0
    if op == "克者旺":
        # A 的五行被克，且克者五行得令而旺。
        resolved = _resolve_tens(args, gender)
        wx = _ten_to_wx(base, resolved)
        if not wx:
            return 0
        return 1 if _element_state(base, wx).get("controller_strength") == "strong" else 0
    if op == "格神透":
        return 1 if _atomic_facts(chart).get("pattern_god_transparent") else 0
    if op == "月令本气":
        value = _atomic_facts(chart).get("month_main_ten_god", "")
        if args and args[0] == "任意":
            return value
        return 1 if value == args[0] else 0
    if op == "时柱十神":
        value = _atomic_facts(chart).get("hour_stem_ten_god", "")
        if args and args[0] == "任意":
            return value
        return 1 if value == args[0] else 0
    if op == "年柱十神":
        value = _atomic_facts(chart).get("year_stem_ten_god", "")
        if args and args[0] == "任意":
            return value
        return 1 if value == args[0] else 0
    if op == "禄根":
        # engine lu_roots 只记录已经落在四柱的十干禄。
        tens = _resolve_tens(args, gender)
        lu_branches = _target_lu_branches(chart, tens)
        return 1 if lu_branches else 0
    if op == "年柱官杀":
        return 1 if _atomic_facts(chart).get("year_officer_killing") else 0
    if op == "大限宫位":
        return _daxian_op(chart, current_year, args)
    # 流年算子（流年透/值/合/冲/克/忌神/财坏印/大运窗口/换运/岁运并临/干合等）由 _liu_op 处理
    raise FactorEvaluateError(f"未知算子: {op}")

def _resolve_wx(base, gender, arg):
    """解析任意参数为五行：具体五行名原样；十神/大类 → 五行。"""
    const = load_constants()
    if arg in const["五行"]:
        return arg
    if arg in const.get("十神大类", {}):
        return _class_wuxing(base, arg)
    if _ten_god_state(base, arg):
        return (_ten_god_state(base, arg) or {}).get("wuxing")
    # 配偶星等
    resolved = _resolve_tens([arg], gender)
    return _ten_to_wx(base, resolved)
def _path_get(base, chart, path: str):
    """按路径取值：优先 factors（基础因子），其次 chart 原始数据。"""
    obj = base if path.startswith(("ten_god_states", "element_states", "wuxing", "yongshen", "ri_gan")) else chart
    cur = obj
    for part in path.split("."):
        if isinstance(cur, dict):
            cur = cur.get(part)
        else:
            return None
    return cur
def _zw_gong_op(base, chart, args):
    """紫微宫位查询：exact match engine palace_atomic_facts，Python 不复算星曜条件。"""
    const = load_constants()
    gong_name = args[0]
    star = args[1]
    cond = args[2] if len(args) > 2 else "任意"
    palaces = set(const["紫微宫位"]) | {"任意"}
    if gong_name not in palaces:
        raise FactorEvaluateError(
            f"宫含宫位无效: {gong_name}; 有效: {sorted(palaces)}"
        )
    facts = ((chart.get("ziwei") or {}).get("palace_facts") or [])

    def hit(kind: str, target: str, star_name: str = "") -> bool:
        return any(
            (gong_name == "任意" or fact.get("palace") == gong_name)
            and fact.get("kind") == kind
            and fact.get("target") == target
            and (not star_name or fact.get("star") == star_name)
            for fact in facts
            if isinstance(fact, dict)
        )

    if cond in {"禄", "权", "科", "忌"}:
        target_star = "" if star == "任意" else star
        return 1 if hit("si_hua", cond, target_star) else 0
    if star in const.get("紫微星曜特殊值", {}):
        return 1 if hit("special", star) else 0
    if cond in const.get("紫微宫位特殊条件", {}):
        return 1 if hit("special", cond, star) else 0
    if cond in {"庙旺", "落陷"}:
        if star == "任意":
            return 1 if hit("brightness", cond) else 0
        if star == "紫微主星":
            main_stars = {
                fact.get("star")
                for fact in facts
                if isinstance(fact, dict)
                and (gong_name == "任意" or fact.get("palace") == gong_name)
                and fact.get("kind") == "brightness"
                and fact.get("target") == "紫微主星"
            }
            return 1 if any(
                (gong_name == "任意" or fact.get("palace") == gong_name)
                and fact.get("kind") == "brightness"
                and fact.get("target") == cond
                and fact.get("star") in main_stars
                for fact in facts
                if isinstance(fact, dict)
            ) else 0
        return 1 if hit("brightness", cond, star) else 0
    return 1 if hit("star", star) else 0
def _ten_class(name: str) -> str:
    """具体十神 → 十神大类；无映射时原样返回。"""
    for class_name, members in load_constants()["十神大类"].items():
        if name in members:
            return class_name
    return name


def _selected_dayun_step(base, current_year: int = 0):
    steps = base.get("dayun_steps", [])
    if current_year:
        return next((
            step for step in steps
            if step.get("start_year", 0) <= current_year <= step.get("end_year", 0)
        ), None)
    idx = base.get("dayun_current_index", -1)
    return steps[idx] if 0 <= idx < len(steps) else None


def _dayun_op(base, chart, args, current_year: int = 0, gender: str = ""):
    """大运十神查询：大运十神(当前, 大类/任意)。任意模式返回十神大类标量。"""
    when, star_class = args[0], args[1]
    if when != "当前":
        return "" if star_class == "任意" else 0

    selected = _selected_dayun_step(base, current_year)

    if selected is None:
        return "" if star_class == "任意" else 0
    shi_shen = selected.get("shi_shen", "") or ""
    suffix = load_constants()["大运十神后缀"]
    if suffix and shi_shen.endswith(suffix):
        shi_shen = shi_shen[:-len(suffix)]
    if star_class == "任意":
        return _ten_class(shi_shen)
    resolved = _resolve_tens([star_class], gender)
    return 1 if shi_shen in resolved else 0


def _dayun_root_op(base, chart, args, current_year: int = 0, gender: str = ""):
    """当前大运干透十神是否通根；读取 engine root 事实。"""
    if args[0] != "当前":
        return 0
    selected = _selected_dayun_step(base, current_year)
    if not selected:
        return 0

    const = load_constants()
    shi_shen = selected.get("shi_shen", "") or ""
    suffix = const["大运十神后缀"]
    if suffix and shi_shen.endswith(suffix):
        shi_shen = shi_shen[:-len(suffix)]
    resolved = _resolve_tens([args[1]], gender)
    if shi_shen not in resolved:
        return 0
    return 1 if selected.get("rooted") else 0


def _daxian_op(chart: dict, current_year: int, args) -> "int | str":
    """当前公历年所在的紫微大限宫位；args=[当前, 任意/宫名]。"""
    when, palace = args[0], args[1]
    if when != "当前":
        return "" if palace == "任意" else 0
    year = current_year
    if not year:
        return "" if palace == "任意" else 0
    steps = chart.get("ziwei_daxian") or []
    selected = next((
        step for step in steps
        if year and step.get("start_year", 0) <= year <= step.get("end_year", 0)
    ), None)
    value = (selected or {}).get("gong", "")
    if palace == "任意":
        return value
    return 1 if value == palace else 0


def _atomic_facts(chart: dict) -> dict:
    full = chart.get("full", {}) or {}
    facts = full.get("atomic_facts")
    return facts if isinstance(facts, dict) else {}
def _base_ctx_from_pan(chart: dict) -> dict:
    """从 pan 直接构建算子求值所需的基础上下文（从 pan 直读，无中间层）。

    只读 pan、只做机械索引且不修改调用方 pan；十神/五行状态、日干/日支/大运/用神从 pan 取。
    输入为基础求值上下文（含 ten_god_states 等聚合键）时原样返回。
    """
    if isinstance(chart, FactorContext):
        return chart.base
    if chart and "ten_god_states" in chart:
        return chart
    chart = chart or {}
    full = chart.get("full", {}) or {}
    bazi_chart = chart.get("chart", {}) or {}
    chart_da_yun = bazi_chart.get("da_yun") or chart.get("da_yun") or {}
    full_da_yun = full.get("da_yun") or {}
    da_yun = chart_da_yun or full_da_yun
    full_steps = full_da_yun.get("steps", []) or []
    steps = chart.get("dayun_steps") or da_yun.get("steps", [])
    if chart_da_yun.get("steps") and full_steps:
        steps = [
            {**step, "rooted": step.get("rooted", full_steps[index].get("rooted", False))}
            if index < len(full_steps) else step
            for index, step in enumerate(chart_da_yun["steps"])
        ]
    const = load_constants()
    fu_yi = (full.get("yong_shen") or {}).get("fu_yi", {}) or {}
    atomic = _atomic_facts(chart)
    ctx = {
        "ten_god_states": _ten_god_states_from_pan(chart),
        "element_states": _element_states_from_pan(chart),
        "wuxing": {
            "count": fu_yi.get("wuxing_count", {}) or {},
        },
        "yongshen": full.get("yong_shen") or {},
        "day_master_element": atomic.get("day_master_element", ""),
        "ri_gan": atomic.get("day_master_stem", "") or chart.get("ri_gan", ""),
        "palace_ri": {"zhi": atomic.get("day_branch", "") or chart.get("palace_ri", {}).get("zhi", "")},
        "dayun_steps": [
            {"name": s.get("name", ""), "start_year": s.get("start_year", 0),
             "end_year": s.get("end_year", 0), "shi_shen": s.get("shi_shen", ""),
             "rooted": s.get("rooted", False)}
            for s in steps
        ],
        "dayun_current_index": chart_da_yun.get(
            "current_step_index", da_yun.get("current_step_index", -1)
        ),
    }
    return ctx


def _op(op: str, args, gender: str, chart: dict,
        current_year: int = 0) -> "int | str":
    """执行本命算子；base context 由 pan/FactorContext 内部构建。"""
    return _eval_natal_op(
        op, args, _base_ctx_from_pan(chart) or {}, gender, chart, current_year
    )


def _target_lu_branches(chart: dict, tens: list[str]) -> set[str]:
    """读取 engine 已计算的十干禄原子事实。"""
    target = set(tens)
    if not target:
        return set()
    full = chart.get("full", {}) or {}
    return {
        item.get("zhi", "")
        for item in full.get("lu_roots", []) or []
        if isinstance(item, dict) and item.get("shi_shen") in target
    }
