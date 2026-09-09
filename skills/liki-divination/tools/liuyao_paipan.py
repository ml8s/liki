"""六爻排盘编排：LLM 工具层的唯一入口；RPC 由本模块内部调用。"""
from __future__ import annotations

from typing import Any

from liuyao_snapshot import project_snapshot
from liuyao_matters import resolve_matter
from liuyao_rpc import engine_data


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


def _server_time() -> str:
    response = engine_data("time.now", {})
    cst = response.get("cst")
    if not isinstance(cst, str) or not cst:
        raise ValueError("time.now returned no cst")
    return cst


def _casting_yaos(casting: dict | None, yaos: list[int] | None) -> tuple[dict | None, list[int]]:
    if casting is not None and yaos is not None:
        raise ValueError("casting 与 yaos 互斥")
    if casting is not None:
        if not isinstance(casting, dict):
            raise ValueError("casting must be an object")
        values = casting.get("yaos")
        if not isinstance(values, list) or len(values) != 6:
            raise ValueError("casting.yaos must contain exactly 6 values")
        return casting, values
    if yaos is None:
        raise ValueError("casting 或 yaos 必须提供一个")
    if not isinstance(yaos, list) or len(yaos) != 6:
        raise ValueError("yaos must contain exactly 6 values")
    return None, yaos


def chart(
    *,
    casting: dict | None = None,
    yaos: list[int] | None = None,
    solar_time: str | None = None,
    matter: str | None = None,
    yong_shen: str | None = None,
    perspective: str | None = None,
    question: str | None = None,
) -> dict:
    """排盘并返回 raw chart + 稳定 snapshot；不直接暴露 matter 给 engine。"""
    matter_fact, resolved_yong_shen = _resolve_yong_shen(
        matter, yong_shen, perspective
    )
    receipt, values = _casting_yaos(casting, yaos)
    if solar_time is None:
        solar_time = _server_time()

    params: dict[str, Any] = {
        "solar_time": solar_time,
        "yong_shen": resolved_yong_shen,
    }
    if receipt is not None:
        params["casting"] = receipt
    else:
        params["yaos"] = values
    raw = engine_data("liuyao.chart", params)
    question_fact = {
        "text": question or "",
        "domain": matter or "advanced",
        "perspective": perspective,
        "solar_time": solar_time,
    }
    if receipt is None:
        receipt = raw.get("casting")
    if not isinstance(receipt, dict):
        receipt = {"mode": "unknown", "yaos": values}

    snapshot = project_snapshot(receipt, raw, question_fact)
    return {
        "question": question_fact,
        "matter": matter_fact,
        "solar_time": solar_time,
        "casting": receipt,
        "chart": raw,
        "snapshot": snapshot,
    }
