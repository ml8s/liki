"""排盘工具（数据准备——非命理逻辑，零命理判定）。

liki 命理 skill 的排盘层：
- build_natal_chart：本命盘（八字 + 紫微一次排全）
- liunian：流年盘（八字 + 紫微单年合并——应期按候选年逐调）
- city_coords：城市名→经纬度（交互式查询）
- bond：合盘（八字合盘 + 紫微合盘）

命理逻辑不在本层（见 factors.py / duanyu.py）。
本层只做「读引擎字段 + 编排 MCP 工具调用 + 归并」，零命理判断。
"""

from typing import Optional

from app.engine_mcp_client import (  # noqa: E402
    MCPError,
    _version_key,
    call,
)
from app.natal.tools.chart_integrity import with_natal_chart_digest
from app.natal.tools.chart_schema import validate_natal_chart

# Engine MCP 客户端（counsel 统一走 MCP）
from app.version import runtime_version

RPCError = MCPError
MAX_LIUNIAN_BATCH = 50


def skill_version() -> str:
    """Read the distributed skill version; fail closed when missing or invalid."""
    try:
        version = runtime_version()
    except OSError as error:
        raise RPCError(f"skill VERSION.txt is unavailable: {error}") from error
    if not version:
        raise RPCError("skill VERSION.txt is empty")
    _version_key(version)
    return version


# ── 内部 RPC 封装（build_natal_chart / liunian 编排用，agent 不直接碰）──

def _resolve_civil(civil_time: str, timezone: str, longitude: float) -> dict:
    """Resolve a reported civil wall clock to standard-time candidates."""
    civil = civil_time[:19]
    date, clock = civil.split("T", 1)
    r = call("tianwen.resolve", {
        "calendar": "solar", "date": date, "time": clock,
        "timezone": timezone, "longitude": longitude,
    })
    return r["data"]


def _resolve_local(date: str, clock: str, timezone: str, longitude: float) -> dict:
    """Wall-clock input (used where only the local calendar date is known)."""
    r = call("tianwen.resolve", {"calendar": "solar", "date": date, "time": clock, "timezone": timezone, "longitude": longitude})
    return r["data"]


def _bazi_chart(standard_time: str, longitude: float, gender: str) -> dict:
    params = {"standard_time": standard_time, "longitude": longitude, "gender": gender}
    return call("bazi.chart", params)["data"]


def _bazi_fullchart(chart: dict) -> dict:
    return call("bazi.fullchart", {"chart": chart})["data"]


def _ziwei_chart(standard_time: str, longitude: float, gender: str) -> dict:
    params = {"standard_time": standard_time, "longitude": longitude, "gender": gender}
    return call("ziwei.chart", params)["data"]


def _ziwei_fullchart(
    standard_time: str, longitude: float, gender: str,
) -> tuple[dict, dict]:
    """返回 (chart, fullchart)：chart 是 ziwei.chart 契约对象，flow 工具复用。"""
    chart = _ziwei_chart(standard_time, longitude, gender)
    full = call("ziwei.fullchart", {"chart": chart})["data"]
    return chart, full


def _ziwei_daxian(ziwei: dict) -> list:
    return _series_items("ziwei.daxian", call("ziwei.daxian", {"chart": ziwei})["data"])


def _series_items(tool: str, data: object) -> list:
    """engine series 工具恒以对象根 {items:[...]} 返回（MCP 出参契约）。"""
    if not isinstance(data, dict) or not isinstance(data.get("items"), list):
        raise RPCError(f"{tool}: engine series response must be {{items:[...]}}")
    return data["items"]


# 当前年份的时点事实只用于年界判定；经度取标准时 120°，性别只需满足 chart 输入契约。
_CURRENT_YEAR_LONGITUDE = 120.0
_CURRENT_YEAR_GENDER = "male"


def time_now(timezone: str) -> dict:
    """引擎当前时间（不回落本地时钟）；时区必填，调用方负责缓存。"""
    if not timezone:
        raise ValueError("timezone is required")
    return call("time.now", {"timezone": timezone})["data"]


def _nian_gz(chart: dict) -> str:
    nian = chart.get("nian") or {}
    return f"{nian.get('gan', '')}{nian.get('zhi', '')}"


def current_years(timezone: str) -> tuple[int, int]:
    """引擎时点事实 → (八字年号[立春界], 农历年号[春节界])；不做缓存。"""
    payload = time_now(timezone)
    standard = payload.get("standard_time")
    if not standard:
        raise ValueError("time.now 返回缺少 standard_time")
    current = _bazi_chart(standard, _CURRENT_YEAR_LONGITUDE, _CURRENT_YEAR_GENDER)
    gregorian_year = int(current.get("birth_year") or 0)
    if not gregorian_year:
        raise ValueError("bazi.chart 返回缺少 birth_year")
    mid_candidate = _resolve_local(
        f"{gregorian_year}-07-01", "12:00", timezone, _CURRENT_YEAR_LONGITUDE,
    )["candidates"][0]
    mid = _bazi_chart(mid_candidate["standard_time"], _CURRENT_YEAR_LONGITUDE, _CURRENT_YEAR_GENDER)
    bazi_year = gregorian_year if _nian_gz(current) == _nian_gz(mid) else gregorian_year - 1
    ziwei = _ziwei_chart(standard, _CURRENT_YEAR_LONGITUDE, _CURRENT_YEAR_GENDER)
    lunar_year = ziwei.get("birth_year")
    if not lunar_year:
        raise ValueError("ziwei.chart 返回缺少 birth_year")
    return bazi_year, int(lunar_year)


def _bazi_liunian(chart: dict, year: int) -> dict:
    """八字流年盘。

    year = 公历年号。engine 以日期选择立春年；这里用该年 6 月 15 日作
    年粒度代表日，落在立春之后，保持「查询年份号」的既有口径。
    """
    return _series_items(
        "bazi.liunian",
        call("bazi.liunian", {"chart": chart, "standard_time": f"{year:04d}-06-15T12:00:00+08:00"})["data"],
    )[0]


def _ziwei_liunian(ziwei: dict, lunar_year: int) -> dict:
    """紫微流年盘。

    lunar_year = 干支年号（农历年号数字，如 2030 → 庚戌）。engine 以日期
    选择农历年；这里用该年 6 月 15 日作年粒度代表日（春节之后），保持
    「查询年份号即农历年号」的既有口径。
    """
    return _series_items(
        "ziwei.liunian",
        call("ziwei.liunian", {"chart": ziwei, "standard_time": f"{lunar_year:04d}-06-15T12:00:00+08:00"})["data"],
    )[0]


def _bazi_liunian_range(chart: dict, start_year: int, count: int) -> list[dict]:
    """批量八字流年（engine count 参数，1 次 RPC）。

    返回按年序排列的流年盘列表，长度 = count。
    """
    if not 0 < count <= MAX_LIUNIAN_BATCH:
        raise ValueError(f"count 必须在 1-{MAX_LIUNIAN_BATCH} 之间：{count}")
    result = call("bazi.liunian", {"chart": chart, "standard_time": f"{start_year:04d}-06-15T12:00:00+08:00", "count": count})
    return _liunian_range_items("bazi.liunian", result, count)


def _ziwei_liunian_range(ziwei: dict, start_lunar_year: int, count: int) -> list[dict]:
    """批量紫微流年（engine count 参数，1 次 RPC）。

    返回按农历年序排列的流年盘列表，长度 = count。
    """
    if not 0 < count <= MAX_LIUNIAN_BATCH:
        raise ValueError(f"count 必须在 1-{MAX_LIUNIAN_BATCH} 之间：{count}")
    result = call("ziwei.liunian", {"chart": ziwei, "standard_time": f"{start_lunar_year:04d}-06-15T12:00:00+08:00", "count": count})
    return _liunian_range_items("ziwei.liunian", result, count)


def _liunian_range_items(tool: str, result: object, count: int) -> list[dict]:
    """Normalize an engine series response; malformed payloads fail closed."""
    if not isinstance(result, dict):
        raise RPCError(f"{tool}: engine response must be an object")
    data = result.get("data", result)
    if not isinstance(data, dict):
        raise RPCError(f"{tool}: engine response data must be an object")
    items = data.get("items")
    if items is None:
        if count != 1:
            raise RPCError(f"{tool}: engine response missing items for count={count}")
        items = [data]
    if not isinstance(items, list) or len(items) != count:
        actual = "non-array" if not isinstance(items, list) else str(len(items))
        raise RPCError(f"{tool}: engine response expected {count} items, got {actual}")
    if any(not isinstance(item, dict) for item in items):
        raise RPCError(f"{tool}: engine response items must be objects")
    return items


# ── agent 工具（排盘 2 个）──

def build_natal_chart(civil_time: str, timezone: str, gender: str, longitude: Optional[float] = None) -> dict:
    """本命盘（八字 + 紫微一次排全）。

    civil_time/timezone：报告的当地民用钟表时间 + IANA 时区；longitude：出生地经度（必填）。
    时间归一化走 tianwen.resolve（返回候选与原因）；本盘以 input 候选为基准，
    多候选时把候选与原因写入 time_candidates / calibration_hint（考时由上层按候选编排）。

    返回盘结构：{standard_time, chart, full, ziwei, ziwei_daxian, gender,
    pan_digest[, time_candidates][, calibration_hint]}；时间字段只有 standard_time。
    返回结构是 factors 层的唯一输入；领域快照由 factors 层按 natal_chart 生成。
    """
    if longitude is None:
        raise ValueError(
            "longitude 必填——真太阳时校正需要出生地经度。请先通过 city_coords 查询城市经度。"
        )
    if not civil_time or not timezone:
        raise ValueError("civil_time 和 timezone 必填（IANA）")
    resolved = _resolve_civil(civil_time, timezone, longitude)
    candidates = resolved.get("candidates") or []
    if not candidates:
        raise ValueError("tianwen.resolve returned no candidates")
    candidate = next(
        (item for item in candidates
         if any(reason.get("domain") == "input" for reason in item.get("reasons", []))),
        candidates[0],
    )
    standard_time = candidate["standard_time"]

    chart = _bazi_chart(standard_time, longitude, gender)
    full = _bazi_fullchart(chart)
    # 用神候选与格局证据归完整命盘（bazi.fullchart 承载，chart 纯排盘不含）
    zw_chart, zw = _ziwei_fullchart(standard_time, longitude, gender)
    # ziwei.daxian 契约收 ziwei.chart 输出；fullchart 增量字段属因子层。
    daxian = _ziwei_daxian(zw_chart)
    result = {
        "standard_time": candidate.get("standard_time"),
        "chart": chart,      # 含 birth_year / da_yun
        "full": full,        # 十神/藏干/神煞/合会冲刑/三元
        "ziwei_chart": zw_chart,  # ziwei.chart 契约对象；flow 工具输入
        "ziwei": zw,         # 十二宫/四化/杂曜/格局
        "ziwei_daxian": daxian,  # 十年大限（公历年段与宫位）
        "gender": gender,
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
    """流年盘（八字 + 紫微单年合并）。

    natal_chart：build_natal_chart 返回的本命盘（只取其中 chart / ziwei 两个字段，agent 传整个盘即可）。
    year：要排的流年年份（应期按候选年逐个调——本命静态一次、流年动态按需）。

    返回：{bazi: 八字流年, ziwei: 紫微流年}。
    """
    if validate:
        validate_natal_chart(natal_chart, action="liunian")
    return {
        "bazi": _bazi_liunian(natal_chart["chart"], year),
        "ziwei": _ziwei_liunian(natal_chart["ziwei_chart"], year),
    }


# ── agent 工具（交互式查询 + 合盘）──

def city_coords(city: str) -> dict:
    """城市名→经纬度（交互式查询——找不到时抛 RPCError，LLM 负责问用户替代城市）。

    返回: {"name": "桦川县", "longitude": 130.3, "latitude": ..., "country": "..."}
    """
    r = call("city.coords", {"city": city}, retries=3)
    return r["data"]


def bond(pan_a: dict, pan_b: dict) -> dict:
    """合盘：两张本命盘 → 八字合盘 + 紫微合盘。

    pan_a / pan_b 为 build_natal_chart 返回的完整盘。
    返回: {"bazi": {...}, "ziwei": {...}}
    """
    validate_natal_chart(pan_a, action="bond pan_a")
    validate_natal_chart(pan_b, action="bond pan_b")
    bazi_r = call("bazi.bond", {
        "a": {"chart": pan_a["chart"]},
        "b": {"chart": pan_b["chart"]},
    })
    ziwei_r = call("ziwei.bond", {
        "a": pan_a["ziwei"],
        "b": pan_b["ziwei"],
    })
    return {
        "bazi": bazi_r["data"],
        "ziwei": ziwei_r["data"],
    }
