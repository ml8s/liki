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
    if not isinstance(chart, dict):
        raise ValueError("engine returned liuyao.chart without chart view")
    analysis = chart.get("analysis")
    if not isinstance(analysis, dict):
        raise ValueError("engine returned liuyao.chart without chart.analysis")

    factors = project_factors(
        casting=casting,
        chart=chart,
        analysis=analysis,
    )
    factors["liuyao_chart"] = chart
    return {"factors": factors}
