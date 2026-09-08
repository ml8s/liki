"""奇门排盘编排层：只调用 RPC，不做命理判断。"""
from __future__ import annotations

import json
import os
import urllib.request
from typing import Any
from urllib.error import HTTPError, URLError

from qimen_errors import RPCError
from qimen_matters import resolve_matter


TIMEOUT = 30
RETRYABLE_HTTP_CODES = {408, 429}


def call(method: str, params: dict, retries: int = 1) -> dict:
    endpoint = os.environ.get("LIKI_RPC_URL", "https://liki.hk/jsonrpc")
    body = json.dumps(
        {"jsonrpc": "2.0", "method": method, "params": params, "id": 1},
        ensure_ascii=False,
    ).encode("utf-8")
    last_error = None
    for _ in range(retries + 1):
        try:
            request = urllib.request.Request(
                endpoint,
                data=body,
                headers={"Content-Type": "application/json; charset=utf-8"},
            )
            with urllib.request.urlopen(request, timeout=TIMEOUT) as response:
                try:
                    result = json.loads(response.read().decode("utf-8"))
                except (UnicodeDecodeError, json.JSONDecodeError) as error:
                    raise RPCError(
                        f"{method}: malformed JSON-RPC response: {error}"
                    ) from error
            if not isinstance(result, dict):
                raise RPCError(f"{method}: malformed JSON-RPC response")
            if "error" in result:
                raise RPCError(f"{method}: {result['error']}")
            if "result" not in result:
                raise RPCError(f"{method}: JSON-RPC response missing result")
            return result["result"]
        except HTTPError as error:
            try:
                detail = error.read().decode("utf-8", errors="replace")[:500]
            except Exception:  # noqa: BLE001 — 错误体不可读时不阻断原始错误
                detail = ""
            if error.code in RETRYABLE_HTTP_CODES or error.code >= 500:
                last_error = error
                continue
            raise RPCError(f"{method}: HTTP {error.code}: {detail or error.reason}") from error
        except (URLError, ConnectionError, TimeoutError, OSError) as error:
            last_error = error
    raise RPCError(f"{method} 失败: {last_error}")


def engine_data(method: str, params: dict) -> dict:
    response = call(method, params)
    if not isinstance(response, dict) or not isinstance(response.get("data"), dict):
        raise RPCError(f"{method}: engine response missing data object")
    return response["data"]


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
