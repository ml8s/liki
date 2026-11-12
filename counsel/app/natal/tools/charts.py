"""Counsel 内部使用的 engine 排盘 RPC helper。"""

from app.engine_mcp_client import MCPError, call

RPCError = MCPError
MAX_LIUNIAN_BATCH = 50

def _bazi_chart(subject: dict) -> dict:
    return call("bazi_chart", {"subject": subject})["data"]


def _ziwei_chart(subject: dict) -> dict:
    return call("ziwei_chart", {"subject": subject})["data"]


_SERIES_ROOTS = {
    "bazi_liupan": "liu_nian",
    "ziwei_liupan": "liu_nian",
}


def _series_items(tool: str, data: object) -> list:
    root = _SERIES_ROOTS.get(tool)
    if root and isinstance(data, dict) and isinstance(data.get(root), list):
        return data[root]
    raise RPCError(f"{tool}: engine series response must contain {root}")


_CURRENT_YEAR_LONGITUDE = 120.0
_CURRENT_YEAR_GENDER = "male"


def time_now(timezone: str) -> dict:
    """引擎当前时间；timezone 必填。"""
    if not timezone:
        raise ValueError("timezone is required")
    return call("time_now", {"timezone": timezone})["data"]


def current_years(timezone: str) -> tuple[int, int]:
    """返回八字立春年号和紫微农历年号。两次流盘并行调用。"""
    payload = time_now(timezone)
    standard = payload.get("local_standard_time")
    if not standard:
        raise ValueError("time.now 返回缺少 local_standard_time")
    subject = {
        "standard_time": standard,
        "longitude": _CURRENT_YEAR_LONGITUDE,
        "gender": _CURRENT_YEAR_GENDER,
    }
    params = {"scope": "liu_nian", "subject": subject, "standard_time": standard, "count": 1}

    from concurrent.futures import ThreadPoolExecutor
    with ThreadPoolExecutor(max_workers=2) as executor:
        bazi_future = executor.submit(call, "bazi_liupan", params)
        ziwei_future = executor.submit(call, "ziwei_liupan", params)
        bazi_items = _series_items("bazi_liupan", bazi_future.result()["data"])
        ziwei_items = _series_items("ziwei_liupan", ziwei_future.result()["data"])

    bazi_item = bazi_items[0]
    ziwei_item = ziwei_items[0]
    bazi_year = ((bazi_item.get("nian_jie") or {}).get("lichun_year")
                 or bazi_item.get("year"))
    if not bazi_year:
        raise ValueError("bazi_liupan 返回缺少 nian_jie.lichun_year")
    lunar_year = (ziwei_item.get("nian_jie") or {}).get("lunar_year")
    if not lunar_year:
        raise ValueError("ziwei_liupan 返回缺少 nian_jie.lunar_year")
    return int(bazi_year), int(lunar_year)


def _bazi_liunian_range(subject: dict, start_year: int, count: int) -> list[dict]:
    if not 0 < count <= MAX_LIUNIAN_BATCH:
        raise ValueError(f"count 必须在 1-{MAX_LIUNIAN_BATCH} 之间：{count}")
    result = call("bazi_liupan", {
        "scope": "liu_nian",
        "subject": subject,
        "standard_time": f"{start_year:04d}-06-15T12:00:00+08:00",
        "count": count,
    })
    return _liunian_range_items("bazi_liupan", result, count)


def _ziwei_liunian_range(subject: dict, start_lunar_year: int, count: int) -> list[dict]:
    if not 0 < count <= MAX_LIUNIAN_BATCH:
        raise ValueError(f"count 必须在 1-{MAX_LIUNIAN_BATCH} 之间：{count}")
    result = call("ziwei_liupan", {
        "scope": "liu_nian",
        "subject": subject,
        "standard_time": f"{start_lunar_year:04d}-06-15T12:00:00+08:00",
        "count": count,
    })
    return _liunian_range_items("ziwei_liupan", result, count)


def _liunian_range_items(tool: str, result: object, count: int) -> list[dict]:
    if not isinstance(result, dict):
        raise RPCError(f"{tool}: engine response must be an object")
    data = result.get("data", result)
    if not isinstance(data, dict):
        raise RPCError(f"{tool}: engine response data must be an object")
    root = _SERIES_ROOTS.get(tool, "liu_nian")
    items = data.get(root)
    if items is None and count == 1:
        items = [data]
    if not isinstance(items, list) or len(items) != count:
        actual = "non-array" if not isinstance(items, list) else str(len(items))
        raise RPCError(f"{tool}: engine response expected {count} items, got {actual}")
    if any(not isinstance(item, dict) for item in items):
        raise RPCError(f"{tool}: engine response items must be objects")
    required = {"gan", "zhi"} if tool == "bazi_liupan" else {"zhi"}
    for index, item in enumerate(items):
        if any(not item.get(field) for field in required):
            missing = sorted(field for field in required if not item.get(field))
            raise RPCError(f"{tool}: engine item {index} missing {missing}")
    return items
