"""六爻因子投影编排：调用 engine 排盘并投影 snapshot 内部因子。"""

from typing import Any

from app.divination.tools.divination_rpc import engine_data, server_time
from app.divination.tools.liuyao_factors import project_factors
from app.divination.tools.liuyao_matters import resolve_matter
from app.guards import require_exactly_one, require_mutually_exclusive, require_object


def _resolve_yong_shen(
    matter: str | None,
    yong_shen: str | None,
    perspective: str | None,
) -> tuple[dict | None, str]:
    require_mutually_exclusive(matter=matter, yong_shen=yong_shen)
    require_exactly_one(matter=matter, yong_shen=yong_shen)
    if matter is not None:
        fact = resolve_matter(matter, perspective=perspective)
        return {"matter": matter, **fact}, fact["yong_shen"]
    if perspective is not None:
        raise ValueError("perspective 仅在 matter=relationship 时使用")
    allowed = {"父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"}
    if yong_shen not in allowed:
        raise ValueError(f"yong_shen must be one of: {', '.join(sorted(allowed))}")
    return None, yong_shen





def factors(
    *,
    casting: dict,
    longitude: float,
    instant: str | None = None,
    timezone: str = "",
    matter: str | None = None,
    yong_shen: str | None = None,
    perspective: str | None = None,
    question: str | None = None,
) -> dict:
    """排盘并返回 engine raw chart 与稳定因子投影；不直接暴露 matter 给 engine。"""
    matter_fact, resolved_yong_shen = _resolve_yong_shen(
        matter, yong_shen, perspective
    )
    require_object(casting, "casting")
    if not timezone:
        raise ValueError("timezone is required (IANA)")
    if instant is None:
        instant = server_time(timezone)["instant"]
    if longitude < -180 or longitude > 180:
        raise ValueError("longitude must be within [-180,180]")
    resolved_longitude = longitude

    params: dict[str, Any] = {
        "instant": instant,
        "timezone": timezone,
        "longitude": resolved_longitude,
        "yong_shen": resolved_yong_shen,
    }
    params["casting"] = casting
    raw = engine_data("liuyao.chart", params)
    question_fact = {
        "text": question or "",
        "matter": matter,
        "explicit_yong_shen": None if matter else resolved_yong_shen,
        "perspective": perspective,
    }
    factors = project_factors(casting, raw, question_fact)
    return {
        "question": question_fact,
        "matter": matter_fact,
        "factors": factors,
    }
