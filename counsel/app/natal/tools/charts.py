"""排盘工具（数据准备——非命理逻辑，零命理判定）。

liki 命理 skill 的排盘层：
- build_natal_chart：本命盘（八字 + 紫微一次排全）
- liunian：流年盘（八字 + 紫微单年合并——应期按候选年逐调）
- city_coords：城市名→经纬度（交互式查询）
- bond：合盘（八字合盘 + 紫微合盘）

命理逻辑不在本层（见 factors.py / duanyu.py）。
本层只做「读引擎字段 + 编排 MCP 工具调用 + 归并」，零命理判断。
"""

from pathlib import Path
from typing import Optional

# Engine MCP 客户端（counsel 统一走 MCP）
from app.engine_mcp_client import (  # noqa: E402
    MCPError,
    _version_key,
    call,
    engine_version,
)
from app.engine_mcp_client import (  # noqa: E402
    ensure_engine_compatible as _ensure_engine_compatible,
)
from app.natal.tools.chart_integrity import with_natal_chart_digest
from app.natal.tools.chart_schema import validate_natal_chart

VERSION_PATH = Path(__file__).resolve().parents[2] / "VERSION.txt"


RPCError = MCPError


def skill_version() -> str:
    """Read the distributed skill version; fail closed when missing or invalid."""
    try:
        version = VERSION_PATH.read_text(encoding="utf-8").strip()
    except OSError as error:
        raise RPCError(f"skill VERSION.txt is unavailable: {error}") from error
    if not version:
        raise RPCError("skill VERSION.txt is empty")
    _version_key(version)
    return version


def required_engine_version() -> str:
    """Require the engine to meet the installed skill's CalVer, not a stale floor."""
    return skill_version()


def ensure_engine_compatible() -> None:
    """Reject old engines before malformed half pans reach the factor layer."""
    _ensure_engine_compatible(
        version=engine_version(),
        required=required_engine_version(),
    )


# ── 内部 RPC 封装（build_natal_chart / liunian 编排用，agent 不直接碰）──

def _resolve_utc(instant: str, timezone: str, longitude: float) -> dict:
    """Canonical time facts: {instant, timezone} -> candidates {instant, timezone, local_time, offset, reasons}."""
    r = call("tianwen.resolve", {"instant": instant, "timezone": timezone, "longitude": longitude})
    return r["data"]


def _resolve_local(date: str, clock: str, timezone: str, longitude: float) -> dict:
    """Wall-clock input (used where only the local calendar date is known)."""
    r = call("tianwen.resolve", {"calendar": "solar", "date": date, "time": clock, "timezone": timezone, "longitude": longitude})
    return r["data"]


def _bazi_chart(instant: str, timezone: str, longitude: float, gender: str) -> dict:
    # Year/month come from solar terms (instant); day/hour from LAST (longitude).
    params = {"instant": instant, "timezone": timezone, "longitude": longitude, "gender": gender}
    return call("bazi.chart", params)["data"]


def _bazi_fullchart(chart: dict) -> dict:
    return call("bazi.fullchart", {"chart": chart})["data"]


def _ziwei_chart(instant: str, timezone: str, longitude: float, gender: str) -> dict:
    params = {"instant": instant, "timezone": timezone, "longitude": longitude, "gender": gender}
    return call("ziwei.chart", params)["data"]


def _ziwei_fullchart(
    instant: str, timezone: str, longitude: float, gender: str,
) -> tuple[dict, dict]:
    """返回 (chart, fullchart)：chart 是 ziwei.chart 契约对象，flow 工具复用。"""
    chart = _ziwei_chart(instant, timezone, longitude, gender)
    full = call("ziwei.fullchart", {"chart": chart})["data"]
    return chart, full


def _ziwei_daxian(ziwei: dict) -> list:
    return call("ziwei.daxian", {"chart": ziwei})["data"]


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
    instant = payload.get("instant")
    if not instant:
        raise ValueError("time.now 返回缺少 instant")
    current = _bazi_chart(instant, timezone, _CURRENT_YEAR_LONGITUDE, _CURRENT_YEAR_GENDER)
    gregorian_year = int(current.get("birth_year") or 0)
    if not gregorian_year:
        raise ValueError("bazi.chart 返回缺少 birth_year")
    mid_candidate = _resolve_local(
        f"{gregorian_year}-07-01", "12:00", timezone, _CURRENT_YEAR_LONGITUDE,
    )["candidates"][0]
    mid = _bazi_chart(mid_candidate["instant"], timezone, _CURRENT_YEAR_LONGITUDE, _CURRENT_YEAR_GENDER)
    bazi_year = gregorian_year if _nian_gz(current) == _nian_gz(mid) else gregorian_year - 1
    ziwei = _ziwei_chart(instant, timezone, _CURRENT_YEAR_LONGITUDE, _CURRENT_YEAR_GENDER)
    lunar_year = ziwei.get("birth_year")
    if not lunar_year:
        raise ValueError("ziwei.chart 返回缺少 birth_year")
    return bazi_year, int(lunar_year)


def _bazi_liunian(chart: dict, year: int) -> dict:
    """八字流年盘。

    year = 公历年号。engine 内部用年中（6 月）确定年干支，避开立春边界
    （公历 year 的干支在立春后即当年号干支）。与紫微流年（_ziwei_liunian
    按干支年号）在年粒度查询下同号一致；两者年界口径统一为「查询年份号」。
    """
    return call("bazi.liunian", {"chart": chart, "year": year})["data"]


def _ziwei_liunian(ziwei: dict, lunar_year: int) -> dict:
    """紫微流年盘。

    lunar_year = 干支年号（农历年号数字，如 2030 → 庚戌）。engine 内部用
    yearGanZhi(lunar_year) 直接算年干支（年号即干支），不做公历/农历换算。
    流年查询以年为粒度：公历 year 与农历 year 同号（年号 2030 两边都指庚戌），
    故调用方直接传查询年份即可；年初（春节前）的日粒度偏差属年界约定，
    见 yearly_range 的流年年界说明。
    """
    return call("ziwei.liunian", {"chart": ziwei, "lunar_year": lunar_year})["data"]


# ── agent 工具（排盘 2 个）──

def build_natal_chart(instant: str, timezone: str, gender: str, longitude: Optional[float] = None) -> dict:
    """本命盘（八字 + 紫微一次排全）。

    instant/timezone：规范时间事实（UTC + IANA）；longitude：出生地经度（必填）。
    时间归一化走 tianwen.resolve（返回候选与原因）；本盘以 input 候选为基准，
    多候选时把候选与原因写入 time_candidates / calibration_hint（考时由上层按候选编排）。

    返回盘结构：{instant, timezone, gregorian_time, chart, full, ziwei, ziwei_daxian, gender,
    pan_digest[, time_candidates][, calibration_hint]}；gregorian_time 是 local_time 展示值。
    返回结构是 factors 层的唯一输入；领域快照由 factors 层按 natal_chart 生成。
    """
    if longitude is None:
        raise ValueError(
            "longitude 必填——真太阳时校正需要出生地经度。请先通过 city_coords 查询城市经度。"
        )
    if not instant or not timezone:
        raise ValueError("instant 和 timezone 必填（IANA）")
    resolved = _resolve_utc(instant, timezone, longitude)
    candidates = resolved.get("candidates") or []
    if not candidates:
        raise ValueError("tianwen.resolve returned no candidates")
    candidate = next(
        (item for item in candidates
         if any(reason.get("domain") == "input" for reason in item.get("reasons", []))),
        candidates[0],
    )
    instant = candidate["instant"]

    chart = _bazi_chart(instant, timezone, longitude, gender)
    full = _bazi_fullchart(chart)
    # 2.6.14 起用神三派归完整命盘（bazi.fullchart 承载，chart 纯排盘不含）
    zw_chart, zw = _ziwei_fullchart(instant, timezone, longitude, gender)
    # ziwei.daxian 契约收 ziwei.chart 输出；fullchart 增量字段属因子层。
    daxian = _ziwei_daxian(zw_chart)
    result = {
        "instant": instant,
        "timezone": timezone,
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


def liunian(natal_chart: dict, year: int) -> dict:
    """流年盘（八字 + 紫微单年合并）。

    natal_chart：build_natal_chart 返回的本命盘（只取其中 chart / ziwei 两个字段，agent 传整个盘即可）。
    year：要排的流年年份（应期按候选年逐个调——本命静态一次、流年动态按需）。

    返回：{bazi: 八字流年, ziwei: 紫微流年}。
    """
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
