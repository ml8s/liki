"""排盘编排层：调用 engine、归并本命盘；不做命理判定。

build_natal_chart 用 resolve 输出的 subject 一次排八字与紫微。所有下游
复用同一个 subject，不再把大 chart 传回 engine。本层输出的领域对象是：
subject（考时后的出生主体）、bazi（八字命盘）、ziwei（紫微命盘）、
ziwei_daxian（大限）。
"""

from typing import Optional

from app.engine_mcp_client import MCPError, call
from app.natal.tools.chart_integrity import with_natal_chart_digest
from app.natal.tools.chart_schema import validate_natal_chart

RPCError = MCPError
MAX_LIUNIAN_BATCH = 50


def _resolve_civil(
    civil_time: str, timezone: str, longitude: float, gender: str,
) -> dict:
    """把报告的民用钟表时间归一化为候选 subject。"""
    civil = civil_time[:19]
    date, clock = civil.split("T", 1)
    return call("tianwen.resolve", {
        "calendar": "solar", "date": date, "time": clock,
        "timezone": timezone, "longitude": longitude, "gender": gender,
    })["data"]


def _bazi_chart(subject: dict) -> dict:
    return call("bazi.chart", {"subject": subject})["data"]


def _ziwei_chart(subject: dict) -> dict:
    return call("ziwei.chart", {"subject": subject})["data"]


def _ziwei_daxian(subject: dict) -> list:
    return _series_items(
        "ziwei.daxian",
        call("ziwei.daxian", {"subject": subject})["data"],
    )


_SERIES_ROOTS = {
    "bazi.liunian": "liu_nian", "bazi.liuyue": "liu_yue",
    "bazi.liuri": "liu_ri", "bazi.liushi": "liu_shi",
    "bazi.xiaoyun": "xiao_yun", "ziwei.daxian": "da_xian",
    "ziwei.liunian": "liu_nian", "ziwei.liuyue": "liu_yue",
    "ziwei.liuri": "liu_ri", "ziwei.liushi": "liu_shi",
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
    return call("time.now", {"timezone": timezone})["data"]


def current_years(timezone: str) -> tuple[int, int]:
    """返回八字立春年号和紫微农历年号。"""
    payload = time_now(timezone)
    standard = payload.get("local_standard_time")
    if not standard:
        raise ValueError("time.now 返回缺少 local_standard_time")
    subject = {
        "birth_time": standard,
        "birth_longitude": _CURRENT_YEAR_LONGITUDE,
        "gender": _CURRENT_YEAR_GENDER,
    }

    bazi_item = _series_items(
        "bazi.liunian",
        call("bazi.liunian", {
            "subject": subject,
            "liunian_start_time": standard,
        })["data"],
    )[0]
    bazi_year = ((bazi_item.get("nian_jie") or {}).get("lichun_year")
                 or bazi_item.get("year"))
    if not bazi_year:
        raise ValueError("bazi.liunian 返回缺少 nian_jie.lichun_year")

    ziwei_item = _series_items(
        "ziwei.liunian",
        call("ziwei.liunian", {
            "subject": subject,
            "liunian_start_time": standard,
        })["data"],
    )[0]
    lunar_year = (ziwei_item.get("nian_jie") or {}).get("lunar_year")
    if not lunar_year:
        raise ValueError("ziwei.liunian 返回缺少 nian_jie.lunar_year")
    return int(bazi_year), int(lunar_year)


def _bazi_liunian(subject: dict, year: int) -> dict:
    """按公历年号排一条八字流年。"""
    return _series_items(
        "bazi.liunian",
        call("bazi.liunian", {
            "subject": subject,
            "liunian_start_time": f"{year:04d}-06-15T12:00:00+08:00",
        })["data"],
    )[0]


def _ziwei_liunian(subject: dict, lunar_year: int) -> dict:
    """按农历年号排一条紫微流年。"""
    return _series_items(
        "ziwei.liunian",
        call("ziwei.liunian", {
            "subject": subject,
            "liunian_start_time": f"{lunar_year:04d}-06-15T12:00:00+08:00",
        })["data"],
    )[0]


def _bazi_liunian_range(subject: dict, start_year: int, count: int) -> list[dict]:
    if not 0 < count <= MAX_LIUNIAN_BATCH:
        raise ValueError(f"count 必须在 1-{MAX_LIUNIAN_BATCH} 之间：{count}")
    result = call("bazi.liunian", {
        "subject": subject,
        "liunian_start_time": f"{start_year:04d}-06-15T12:00:00+08:00",
        "liunian_count": count,
    })
    return _liunian_range_items("bazi.liunian", result, count)


def _ziwei_liunian_range(subject: dict, start_lunar_year: int, count: int) -> list[dict]:
    if not 0 < count <= MAX_LIUNIAN_BATCH:
        raise ValueError(f"count 必须在 1-{MAX_LIUNIAN_BATCH} 之间：{count}")
    result = call("ziwei.liunian", {
        "subject": subject,
        "liunian_start_time": f"{start_lunar_year:04d}-06-15T12:00:00+08:00",
        "liunian_count": count,
    })
    return _liunian_range_items("ziwei.liunian", result, count)


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
    return items


def build_natal_chart(
    civil_time: str,
    timezone: str,
    gender: str,
    longitude: Optional[float] = None,
) -> dict:
    """用 input 候选 subject 排八字、紫微与大限。

    多候选时保留 time_candidates；选定 subject 仍是本盘唯一计算输入。
    """
    if longitude is None:
        raise ValueError("longitude 必填；请先通过 city_coords 查询城市经度。")
    if not civil_time or not timezone:
        raise ValueError("civil_time 和 timezone 必填（IANA）")

    resolved = _resolve_civil(civil_time, timezone, longitude, gender)
    candidates = resolved.get("candidates") or []
    if not candidates:
        raise ValueError("tianwen.resolve returned no candidates")
    candidate = next(
        (item for item in candidates
         if any(reason.get("domain") == "input" for reason in item.get("reasons", []))),
        candidates[0],
    )
    subject = candidate["subject"]

    bazi = _bazi_chart(subject)
    ziwei = _ziwei_chart(subject)
    daxian = _ziwei_daxian(subject)

    result = {
        "subject": subject,
        "bazi": bazi,
        "ziwei": ziwei,
        "ziwei_daxian": daxian,
    }
    if len(candidates) > 1:
        domains = sorted({
            reason.get("domain", "")
            for item in candidates
            for reason in item.get("reasons", [])
            if reason.get("domain") not in (None, "", "input")
        })
        suffix = f"（{ '/'.join(domains) }）" if domains else ""
        result["time_candidates"] = candidates
        result["calibration_hint"] = (
            f"出生时间落在历法边界上{suffix}：存在多个候选（见 time_candidates），"
            "本盘以 input 候选为基准；请按原因领域用事件校准。"
        )
    else:
        detail = (candidate.get("reasons") or [{}])[0].get("detail", "")
        if "距" in detail or "恰在" in detail:
            result["calibration_hint"] = f"输入时间接近历法边界：{detail}；时辰敏感结论须确认。"
    result = with_natal_chart_digest(result)
    validate_natal_chart(result, action="build_natal_chart result")
    return result


def liunian(natal_chart: dict, year: int, *, validate: bool = True) -> dict:
    subject = natal_chart["subject"] if not validate else None
    if validate:
        validate_natal_chart(natal_chart, action="liunian")
        subject = natal_chart["subject"]
    return {
        "bazi": _bazi_liunian(subject, year),
        "ziwei": _ziwei_liunian(subject, year),
    }


def city_coords(city: str) -> dict:
    return call("city.coords", {"city": city}, retries=3)["data"]


def bond(pan_a: dict, pan_b: dict) -> dict:
    """两张本命盘的八字与紫微合盘。"""
    validate_natal_chart(pan_a, action="bond pan_a")
    validate_natal_chart(pan_b, action="bond pan_b")
    return {
        "bazi": call("bazi.bond", {
            "a": {"subject": pan_a["subject"]},
            "b": {"subject": pan_b["subject"]},
        })["data"],
        "ziwei": call("ziwei.bond", {
            "a": {"subject": pan_a["subject"]},
            "b": {"subject": pan_b["subject"]},
        })["data"],
    }
