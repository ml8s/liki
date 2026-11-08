"""奇门排盘编排层：只调用 RPC，不做命理判断。"""

from typing import Any

from app.divination.tools.divination_rpc import engine_data
from app.divination.tools.qimen_matters import resolve_matter
from app.guards import require_mutually_exclusive


def qimen_chart(
    standard_time: str,
    longitude: float,
    *,
    method: dict[str, str],
    matter: str | None = None,
    yong_shen: list[str] | None = None,
    subject: dict | None = None,
) -> dict:
    """排奇门盘；普通事象在 Python 表内转为显式用神后调用 engine。"""
    require_mutually_exclusive(matter=matter, yong_shen=yong_shen)
    if yong_shen is not None and len(yong_shen) == 0:
        raise ValueError("yong_shen 不能为空；高级用户未指定符号时应省略该参数")
    if yong_shen is not None:
        if not isinstance(yong_shen, list) or any(
            not isinstance(symbol, str) or not symbol.strip()
            for symbol in yong_shen
        ):
            raise ValueError("yong_shen 必须是字符串数组")
        if len(yong_shen) != len(set(yong_shen)):
            raise ValueError("yong_shen 符号重复")

    matter_fact = None
    symbols = list(yong_shen or [])
    if matter is not None:
        matter_fact = resolve_matter(matter)
        matter_fact = {"matter": matter, **matter_fact}
        symbols = list(matter_fact["symbols"])

    if not isinstance(method, dict):
        raise ValueError("method must be an explicit qimen method object")
    if not standard_time:
        raise ValueError("standard_time is required")
    params: dict[str, Any] = {
        "standard_time": standard_time,
        "longitude": longitude,
        "method": method,
    }
    if subject is not None:
        params["subject"] = subject
    if symbols:
        params["yong_shen"] = symbols

    chart = engine_data("qimen_chart", params)
    return {"matter": matter_fact, "chart": chart}
