"""六爻因子投影编排：调用 engine 排盘并投影 case 内部因子。"""

from typing import Any

from app.divination.tools.divination_rpc import engine_data
from app.divination.tools.liuyao_factors import project_factors
from app.guards import require_object


def factors(
    *,
    event: dict,
) -> dict:
    """排盘并返回 engine raw chart 与稳定因子投影。"""
    require_object(event, "event")
    casting = event.get("casting")
    context = event.get("casting_context")
    require_object(casting, "event.casting")
    require_object(context, "event.casting_context")

    params: dict[str, Any] = {
        "event": event,
    }
    result = engine_data("liuyao_chart", params)
    chart = result.get("chart")
    analysis = result.get("analysis")
    if not isinstance(chart, dict):
        raise ValueError("engine returned liuyao.chart without chart view")
    if not isinstance(analysis, dict):
        raise ValueError("engine returned liuyao.chart without analysis")

    focus = context.get("focus", {})
    matter = focus.get("matter")
    explicit_yong_shen = focus.get("yong_shen")
    perspective = focus.get("perspective")
    question_fact = {
        "text": event.get("question", ""),
        "matter": matter,
        "explicit_yong_shen": None if matter else explicit_yong_shen,
        "perspective": perspective,
    }
    factors = project_factors(
        casting=casting,
        chart=chart,
        analysis=analysis,
        question=question_fact,
    )
    factors["liuyao_chart"] = result.get("chart")
    return {
        "question": question_fact,
        "matter": {"matter": matter} if matter else None,
        "factors": factors,
    }
