"""奇门排盘编排层：只调用 RPC，不做命理判断。"""

from typing import Any

from app.divination.tools.divination_rpc import engine_data
from app.divination.tools.qimen_matters import resolve_matter
from app.guards import require_mutually_exclusive


def city_coords(city: str) -> dict:
    data = engine_data("city.coords", {"city": city})
    candidates = data.get("candidates") or []
    if len(candidates) != 1:
        labels = "、".join(
            f"{item.get('region', '')}{item.get('name', '')}" for item in candidates
        ) or "无候选"
        raise ValueError(
            f"city '{city}' 存在歧义或未找到（{labels}）；请让用户确认城市"
        )
    return candidates[0]


def qimen_chart(
    instant: str,
    timezone: str,
    longitude: float,
    *,
    method: dict[str, str],
    matter: str | None = None,
    yong_shen: list[str] | None = None,
    birth_date: str | None = None,
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
    params: dict[str, Any] = {
        "instant": instant,
        "timezone": timezone,
        "longitude": longitude,
        "method": method,
    }
    if birth_date is not None and birth_date != "":
        params["birth_date"] = birth_date
    if symbols:
        params["yong_shen"] = symbols

    chart = engine_data("qimen.chart", params)
    return {"matter": matter_fact, "chart": chart}
