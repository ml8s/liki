"""counsel 断语引擎：bazi_yongshen / compute_factors / natal_query / period_query。

正交化后 engine 的补集能力（判断，不排盘）：
- bazi_yongshen(chart)       ：engine 八字排盘 → 唯一用神裁决（不计算因子）
- compute_factors(chart)    ：engine 排盘结果 → 因子快照（内部调 engine fullchart）
- natal_query(factors, ...) ：因子快照 → 本命断语
- period_query(factors, ...)：因子快照 → 应期断语（内部调 engine 流年/大限）

输入 chart 为 engine 分领域排盘输出（bazi_chart / ziwei_chart，公共契约）。
"""

import hashlib
import json

from app.natal.tools.chart_schema import validate_side_chart
from app.natal.tools.charts import _bazi_fullchart, _ziwei_daxian, call
from app.natal.tools.duanyu import match_rule
from app.natal.tools.factors import evaluate_factors
from app.natal.tools.yongshen import published_yongshen, resolve_yongshen


def _factors_digest(factors: dict) -> str:
    raw = json.dumps(
        factors, ensure_ascii=False, sort_keys=True,
        separators=(",", ":"), allow_nan=False,
    )
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _chart_digest(chart: dict) -> str:
    raw = json.dumps(
        chart, ensure_ascii=False, sort_keys=True,
        separators=(",", ":"), allow_nan=False,
    )
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _factor_provenance(factors: dict) -> dict:
    provenance = factors.get("_provenance")
    if not isinstance(provenance, dict):
        raise ValueError("factors 缺少 compute_factors 生成的 _provenance")
    return provenance


def verify_factors_digest(factors: dict, factors_digest: str) -> None:
    """校验因子快照未被调用方裁剪、拼接或替换。"""
    if not isinstance(factors_digest, str):
        raise ValueError("factors_digest 必须是 compute_factors 返回的 SHA-256 字符串")
    expected = _factors_digest(factors)
    if factors_digest != expected:
        raise ValueError("factors_digest mismatch；必须原样传递 compute_factors 返回的 factors")


def _is_bazi_chart(chart: dict) -> bool:
    return "ri" in chart  # 八字盘含日柱；紫微盘为宫位结构


def _hour_calibration_facts(chart: dict, full: dict) -> dict:
    """提取考时所需的时柱敏感事实；候选盘即使断语相同也可结构区分。"""
    hour = chart.get("shi") or {}
    atomic = (full.get("atomic_facts") or {}) if isinstance(full, dict) else {}
    shen_sha = ((full.get("shi") or {}).get("shen_sha") or []) if isinstance(full, dict) else []
    return {
        "hour_pillar": {
            "gan": hour.get("gan", ""),
            "zhi": hour.get("zhi", ""),
        },
        "hour_stem_ten_god": atomic.get("hour_stem_ten_god", ""),
        "hour_shen_sha": [
            item.get("name") or item.get("xing") or ""
            for item in shen_sha
            if isinstance(item, dict)
        ],
    }


def bazi_yongshen(chart: dict) -> dict:
    """engine 八字排盘结果 → 唯一用神裁决。

    这是起名等跨域消费者使用的聚焦能力；与 compute_factors 内部调用同一个
    resolve_yongshen，保证全系统只有一份裁决实现。
    """
    if not isinstance(chart, dict) or not _is_bazi_chart(chart):
        raise ValueError("bazi_yongshen 只接受八字盘（chart 含四柱）；紫微盘请走紫微领域")
    full = _bazi_fullchart(chart)
    natal_chart = {"chart": chart, "full": full, "gender": chart.get("gender", "")}
    validate_side_chart(natal_chart, "bazi", "bazi_yongshen")
    return published_yongshen(resolve_yongshen(full))


def compute_factors(chart: dict) -> dict:
    """engine 排盘结果 → { factors, factors_digest }。

    chart 为 bazi_chart（八字）或 ziwei_chart（紫微）输出。内部调 engine
    fullchart 取因子所需详细字段，再求因子快照（与正交化基线一致）。
    来源 context（性别；出生信息如有则附带）写入 factors._provenance.context，
    随快照进入 digest——下游工具直接从 provenance 读取，不再单独传递。
    八字快照把唯一用神裁决写入 factors.selected_yongshen，与独立用神能力同源。
    """
    gender = chart.get("gender", "")
    context: dict = {"性别": gender}
    side = "bazi" if _is_bazi_chart(chart) else "ziwei"
    selected_yongshen: dict | None = None
    if side == "bazi":
        full = _bazi_fullchart(chart)
        selected_yongshen = published_yongshen(resolve_yongshen(full))
        natal_chart = {
            "chart": chart, "full": full, "gender": gender,
            "_selected_yongshen": selected_yongshen,
        }
        validate_side_chart(natal_chart, "bazi", "compute_factors")
        factors = evaluate_factors(gender, natal_chart, shushi="bazi")
        factors["selected_yongshen"] = selected_yongshen
    else:
        zw = call("ziwei.fullchart", {"chart": chart})["data"]
        # ziwei.daxian 契约收 ziwei.chart 输出；fullchart 的增量字段属因子层，不进 daxian。
        daxian = _ziwei_daxian(chart)
        natal_chart = {"chart": chart, "ziwei": zw, "ziwei_daxian": daxian, "gender": gender}
        validate_side_chart(natal_chart, "ziwei", "compute_factors")
        factors = evaluate_factors(gender, natal_chart, shushi="ziwei")
    factors["_provenance"] = {
        "side": side,
        "chart_digest": _chart_digest(chart),
        "context": context,
    }
    if side == "bazi":
        factors["_provenance"].update({
            "yongshen_digest": selected_yongshen["yongshen_digest"],
            "policy_id": selected_yongshen["policy_id"],
        })
    return {
        "factors": factors,
        "factors_digest": _factors_digest(factors),
    }


def natal_query(factors: dict, topics: list[str], factors_digest: str,
                side: str = "bazi") -> dict:
    """因子快照 + topics → 本命断语（用因子快照匹配断语表，不重算 snapshot）。

    与正交化基线一致（context 至少含性别；出生信息影响断语时附带）。
    side 指定因子所属侧（bazi/ziwei），断语只出该侧。
    """
    from app.natal.tools.analytics import _flatten_side_result, _load_routes, _require_topics

    verify_factors_digest(factors, factors_digest)
    provenance = _factor_provenance(factors)
    if provenance.get("side") != side:
        raise ValueError(
            f"factors provenance side mismatch: expected {side!r}, "
            f"got {provenance.get('side')!r}"
        )
    if side not in ("bazi", "ziwei"):
        raise ValueError(f"natal_query side 只支持 bazi/ziwei，收到: {side!r}")
    context = provenance.get("context") or {}
    from app.natal.tools.factor_constants import load_constants

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


def period_query(factors: dict, factors_digest: str, time_scope: dict,
                 topics: list[str], chart: dict, side: str = "bazi") -> dict:
    """已校验因子快照 + 时间层 + topics → 应期断语（大运/大限/流年）。

    chart 为 engine 排盘结果（bazi_chart/ziwei_chart），内部据此调 engine
    流年/大限取应期字段，再匹配断语表（本命因子参与匹配）。
    factors_digest 先验证 factors 未被修改；随后按 chart 重建求值上下文并
    调 engine 获取流年/大限事实。engine 事实必须以 chart 为源，不能从快照伪造。
    side 指定因子所属侧（bazi/ziwei），断语只出该侧。
    输出结构与 analyze_periods 一致（periods 数组）。

    time_scope 的年份参数语义（命理领域模型）：
    - side=bazi（八字流年）：干支年号（公历立春界，年号级 = 公历年号）
    - side=ziwei（紫微流年）：农历年号（春节界，年号级与八字一致，如 2030 → 庚戌）
    年份参数直接作为年号使用，**不做公历/农历日期换算**——换算会引入月份
    信息（用户只按"年"查询），难以对应；年号级八字与紫微一致。
    """
    from app.natal.tools.analytics import _analyze_periods

    verify_factors_digest(factors, factors_digest)
    provenance = _factor_provenance(factors)
    if provenance.get("side") != side:
        raise ValueError(
            f"factors provenance side mismatch: expected {side!r}, "
            f"got {provenance.get('side')!r}"
        )
    if provenance.get("chart_digest") != _chart_digest(chart):
        raise ValueError("chart 与 compute_factors 因子来源不一致")
    if side not in ("bazi", "ziwei"):
        raise ValueError(f"period_query side 只支持 bazi/ziwei，收到: {side!r}")
    gender = chart.get("gender", "")
    if side == "bazi":
        selected_yongshen = factors.get("selected_yongshen")
        if not isinstance(selected_yongshen, dict):
            raise ValueError(
                "factors 缺少 compute_factors 写入的 selected_yongshen；"
                "必须原样传递 compute_factors 返回的 factors，不得重新裁决"
            )
        if provenance.get("yongshen_digest") != selected_yongshen.get("yongshen_digest"):
            raise ValueError("factors _provenance 的用神来源与 selected_yongshen 不一致")
        full = _bazi_fullchart(chart)
        natal_chart = {
            "chart": chart, "full": full, "gender": gender,
            "_selected_yongshen": selected_yongshen,
        }
    else:
        zw = call("ziwei.fullchart", {"chart": chart})["data"]
        # ziwei.daxian 契约收 ziwei.chart 输出；fullchart 的增量字段属因子层，不进 daxian。
        daxian = _ziwei_daxian(chart)
        natal_chart = {"chart": chart, "full": zw, "ziwei": zw, "ziwei_daxian": daxian,
               "gender": gender}
    validate_side_chart(natal_chart, side, "period_query")
    # period_query 固定只出本侧断语：组合盘不含另一侧排盘，另一侧数据由
    # _factor_context_from_natal_chart 从本侧盘误读，会产生假断语，必须过滤。
    result = _analyze_periods(
        natal_chart,
        {"topics": topics, "time_scope": time_scope or {}},
        validate_pan=False,
    )
    for scope in result["periods"]:
        scope["assertions"] = [
            item for item in scope["assertions"] if item.get("side") == side
        ]
        if "counts" in scope:
            scope["counts"]["returned"] = len(scope["assertions"])
    if side == "bazi":
        result["calibration_facts"] = _hour_calibration_facts(chart, full)
    return result
