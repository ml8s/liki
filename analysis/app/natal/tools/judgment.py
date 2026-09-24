"""judgment 判断层：compute_factors / natal_query / period_query。

正交化后 engine-pro 的补集能力（判断，不排盘）：
- compute_factors(chart)    ：engine 排盘结果 → 因子快照（内部调 engine fullchart）
- natal_query(factors, ...) ：因子快照 → 本命断语
- period_query(factors, ...)：因子快照 → 应期断语（内部调 engine 流年/大限）

输入 chart 为 engine 分领域排盘输出（bazi_chart / ziwei_chart，公共契约）。
"""
from __future__ import annotations

import hashlib
import json

from duanyu import match_rule, query, yearly_range
from factors import evaluate_factors
from paipan import _bazi_fullchart, _ziwei_daxian, call


def _factors_digest(factors: dict) -> str:
    raw = json.dumps(factors, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _is_bazi_chart(chart: dict) -> bool:
    return "ri" in chart  # 八字盘含日柱；紫微盘为宫位结构


def compute_factors(chart: dict) -> dict:
    """engine 排盘结果 → { factors, factors_digest, context }。

    chart 为 bazi_chart（八字）或 ziwei_chart（紫微）输出。内部调 engine
    fullchart 取因子所需详细字段，再求因子快照（与正交化基线一致）。
    context 携带断语所需的最小领域事实（性别；出生信息如有则附带）。
    """
    gender = chart.get("gender", "")
    context: dict = {"性别": gender}
    if chart.get("solar"):
        context["公历出生"] = chart["solar"]
    if chart.get("lunar"):
        context["农历出生"] = chart["lunar"]
    if _is_bazi_chart(chart):
        full = _bazi_fullchart(chart)
        pan = {"chart": chart, "full": full, "gender": gender}
        factors = evaluate_factors(gender, pan, shushi="bazi")
    else:
        zw = call("ziwei.fullchart", {"chart": chart})["data"]
        daxian = _ziwei_daxian(zw)
        pan = {"chart": chart, "ziwei": zw, "ziwei_daxian": daxian, "gender": gender}
        factors = evaluate_factors(gender, pan, shushi="ziwei")
    return {
        "factors": factors,
        "factors_digest": _factors_digest(factors),
        "context": context,
    }


def natal_query(factors: dict, topics: list[str], context: dict | None = None,
                side: str = "bazi") -> dict:
    """因子快照 + topics → 本命断语（用因子快照匹配断语表，不重算 snapshot）。

    与正交化基线一致（context 至少含性别；出生信息影响断语时附带）。
    side 指定因子所属侧（bazi/ziwei），断语只出该侧。
    """
    from analytics import _flatten_side_result, _load_routes, _require_topics

    if side not in ("bazi", "ziwei"):
        raise ValueError(f"natal_query side 只支持 bazi/ziwei，收到: {side!r}")
    from factor_constants import load_constants

    side_labels = load_constants()["命理侧"]["标签"]
    bz_label = side_labels["bazi"]
    zw_label = side_labels["ziwei"]
    selected = _require_topics(topics)
    routes = _load_routes()
    snapshots = {bz_label: factors if side == "bazi" else {},
                 zw_label: factors if side == "ziwei" else {}, "context": context or {}}
    all_assertions: list[dict] = []
    seen: set[tuple[str, str]] = set()
    for topic, route in selected:
        for rule in route["natal_rules"]:
            result = match_rule(rule, snapshots)
            result["_rule"] = rule
            part, _count = _flatten_side_result(result, selected, routes, "natal")
            for item in part:
                key = (item["assertion_id"], item["side"])
                if key not in seen:
                    seen.add(key)
                    all_assertions.append(item)
    return {"assertions": all_assertions}


def period_query(factors: dict, time_scope: dict, topics: list[str], chart: dict,
                 side: str = "bazi") -> dict:
    """因子快照 + 时间层 + topics → 应期断语（大运/大限/流年）。

    chart 为 engine 排盘结果（bazi_chart/ziwei_chart），内部据此调 engine
    流年/大限取应期字段，再匹配断语表（本命因子参与匹配）。
    side 指定因子所属侧（bazi/ziwei），断语只出该侧。
    输出结构与 analyze_periods 一致（periods 数组）。

    time_scope 的年份参数语义（命理领域模型）：
    - side=bazi（八字流年）：干支年号（公历立春界，年号级 = 公历年号）
    - side=ziwei（紫微流年）：农历年号（春节界，年号级与八字一致，如 2030 → 庚戌）
    年份参数直接作为年号使用，**不做公历/农历日期换算**——换算会引入月份
    信息（用户只按"年"查询），难以对应；年号级八字与紫微一致。
    """
    from analytics import _analyze_periods

    if side not in ("bazi", "ziwei"):
        raise ValueError(f"period_query side 只支持 bazi/ziwei，收到: {side!r}")
    gender = chart.get("gender", "")
    if side == "bazi":
        full = _bazi_fullchart(chart)
        pan = {"chart": chart, "full": full, "gender": gender}
    else:
        zw = call("ziwei.fullchart", {"chart": chart})["data"]
        daxian = _ziwei_daxian(zw)
        pan = {"chart": chart, "full": zw, "ziwei": zw, "ziwei_daxian": daxian, "gender": gender}
    # period_query 固定只出本侧断语：组合盘不含另一侧排盘，另一侧数据由
    # _factor_context_from_pan 从本侧盘误读，会产生假断语，必须过滤。
    result = _analyze_periods(
        pan,
        {"topics": topics, "time_scope": time_scope},
        validate_pan=False,
    )
    for scope in result["periods"]:
        scope["assertions"] = [
            item for item in scope["assertions"] if item.get("side") == side
        ]
        if "counts" in scope:
            scope["counts"]["returned"] = len(scope["assertions"])
    return result