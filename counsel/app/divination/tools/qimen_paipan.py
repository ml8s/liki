"""奇门排盘编排层：只调用 RPC，不做命理判断。"""
from __future__ import annotations

from typing import Any

from divination_rpc import engine_data
from qimen_matters import resolve_matter


def city_coords(city: str) -> dict:
    return engine_data("city.coords", {"city": city})


def solar_time(time: str, longitude: float) -> dict:
    return engine_data("tianwen.time", {"time": time, "longitude": longitude})


def qimen_chart(
    solar_time: str,
    *,
    matter: str | None = None,
    yong_shen: list[str] | None = None,
    birth_date: str | None = None,
    scope: str | None = None,
    school: str | None = None,
    dingju_method: str | None = None,
    quarter_rule: str | None = None,
    base_dingju_method: str | None = None,
    dun_source: str | None = None,
    hour_boundary: str | None = None,
) -> dict:
    """排奇门盘；普通事象在 Python 表内转为显式用神后调用 engine。"""
    if matter is not None and yong_shen is not None:
        raise ValueError("matter 与 yong_shen 互斥；普通问事传 matter，高级用户传 yong_shen")
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

    params: dict[str, Any] = {"solar_time": solar_time}
    optional = {
        "scope": scope,
        "school": school,
        "dingju_method": dingju_method,
        "quarter_rule": quarter_rule,
        "base_dingju_method": base_dingju_method,
        "dun_source": dun_source,
        "hour_boundary": hour_boundary,
        "birth_date": birth_date,
    }
    for key, value in optional.items():
        if value is not None and value != "":
            params[key] = value
    if symbols:
        params["yong_shen"] = symbols

    chart = engine_data("qimen.chart", params)
    return {"matter": matter_fact, "chart": chart}
