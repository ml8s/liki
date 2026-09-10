"""六爻因子投影编排：调用 engine 排盘并投影 snapshot 内部因子。"""
from __future__ import annotations

from typing import Any

from liuyao_factors import project_factors
from liuyao_matters import resolve_matter
from divination_rpc import engine_data, server_time


def _resolve_yong_shen(
    matter: str | None,
    yong_shen: str | None,
    perspective: str | None,
) -> tuple[dict | None, str]:
    if matter is not None and yong_shen is not None:
        raise ValueError("matter 与 yong_shen 互斥；普通问事传 matter，高级用户传 yong_shen")
    if matter is None and yong_shen is None:
        raise ValueError("matter 或 yong_shen 必须提供一个")
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
    solar_time: str | None = None,
    matter: str | None = None,
    yong_shen: str | None = None,
    perspective: str | None = None,
    question: str | None = None,
) -> dict:
    """排盘并返回 engine raw chart 与稳定因子投影；不直接暴露 matter 给 engine。"""
    matter_fact, resolved_yong_shen = _resolve_yong_shen(
        matter, yong_shen, perspective
    )
    if not isinstance(casting, dict):
        raise ValueError("casting must be an object")
    if solar_time is None:
        solar_time = server_time()

    params: dict[str, Any] = {
        "solar_time": solar_time,
        "yong_shen": resolved_yong_shen,
    }
    params["casting"] = casting
    raw = engine_data("liuyao.chart", params)
    question_fact = {
        "text": question or "",
        "matter": matter,
        "explicit_yong_shen": None if matter else resolved_yong_shen,
        "perspective": perspective,
        "solar_time": solar_time,
    }
    factors = project_factors(casting, raw, question_fact)
    return {
        "question": question_fact,
        "matter": matter_fact,
        "factors": factors,
    }
