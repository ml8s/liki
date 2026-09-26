"""排盘工具（数据准备——非命理逻辑，零命理判定）。

liki 命理 skill 的排盘层：
- full_paipan：本命盘（八字 + 紫微一次排全）
- liunian：流年盘（八字 + 紫微单年合并——应期按候选年逐调）
- city_coords：城市名→经纬度（交互式查询）
- bond：合盘（八字合盘 + 紫微合盘）

命理逻辑不在本层（见 factors.py / duanyu.py）。
本层只做「读引擎字段 + 编排 RPC + 归并」，零命理判断。
"""
from __future__ import annotations

import sys
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

from factor_constants import load_constants
from pan_integrity import with_natal_digest
from pan_schema import validate_natal_pan

# MCP 引擎客户端（dev 走 MCP；RPC 仅引擎保留给线上兼容）
_COUNSEL_APP = Path(__file__).resolve().parents[2]
if str(_COUNSEL_APP) not in sys.path:
    sys.path.insert(0, str(_COUNSEL_APP))
from engine_client import (  # noqa: E402
    MCPError,
    _version_key,
    call,
    engine_version,
)
from engine_client import (  # noqa: E402
    ensure_engine_compatible as _ensure_engine_compatible,
)

SHICHEN_BOUNDARY_THRESHOLD_MINUTES = 8
SHICHEN_BOUNDARY_START_HOURS = (23, 1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21)
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


# ── 内部 RPC 封装（full_paipan / liunian 编排用，agent 不直接碰）──

def _solar_time(gregorian: str, longitude: float) -> dict:
    r = call("tianwen.time", {"time": gregorian, "longitude": longitude})
    return r["data"]


def _bazi_chart(solar: str, gender: str) -> dict:
    params = {"solar_time": solar, "gender": gender}
    return call("bazi.chart", params)["data"]


def _bazi_fullchart(chart: dict) -> dict:
    return call("bazi.fullchart", {"chart": chart})["data"]


def _ziwei_chart(lunar: dict, gender: str) -> dict:
    return call("ziwei.chart", {"lunar": lunar, "gender": gender})["data"]


def _ziwei_fullchart(lunar: dict, gender: str) -> dict:
    chart = _ziwei_chart(lunar, gender)
    return call("ziwei.fullchart", {"chart": chart})["data"]


def _ziwei_daxian(ziwei: dict) -> list:
    return call("ziwei.daxian", {"chart": ziwei})["data"]


def _bazi_liunian(chart: dict, year: int) -> dict:
    """八字流年盘。

    year = 公历年号。engine 内部用年中（6 月）确定年干支，避开立春边界
    （公历 year 的干支在立春后即当年号干支）。与紫微流年（_ziwei_liunian
    按干支年号）在年粒度查询下同号一致；两者年界口径统一为「查询年份号」。
    """
    return call("bazi.liunian", {"chart": chart, "year": year})["data"]


def _shichen_window(index: int, boundaries: list, zhi: list) -> dict:
    """交界索引 → 时辰窗口（纯机械映射，不涉及换日口径，不做吉凶判断）。

    名称与起始小时均由 boundaries 与 constants.json 的「地支」推导，不写死
    命理成员字面量（见 tests/test_domain_semantic_contracts.py）。
    """
    start = boundaries[index]
    return {
        "name": zhi[index] + "时",
        "branch": zhi[index],
        "span": f"{start:02d}:00-{(start + 2) % 24:02d}:00",
    }


def _shichen_boundary_hint(solar: str) -> dict | None:
    """返回距时辰交界的确定性提示；不做吉凶判断。

    同时给出当前时辰与跨过最近交界后的时辰，调用方无需自行二次推断
    「往哪边偏会翻」。boundary_offset_minutes 有符号：负数=早于交界，
    正数=晚于交界，与 direction 一一对应。

    晚子时 / 早子时的换日口径仍由 engine 与 skill 文档决定；这里只提示
    输入分钟接近传统两小时交界，建议用户用事件校准。
    """
    try:
        moment = datetime.fromisoformat(solar.replace("Z", "+00:00"))
    except (TypeError, ValueError):
        return None

    boundaries = SHICHEN_BOUNDARY_START_HOURS
    candidates = []
    for day_offset in (-1, 0, 1):
        for index, hour in enumerate(boundaries):
            candidate = moment.replace(
                hour=hour, minute=0, second=0, microsecond=0
            ) + timedelta(days=day_offset)
            candidates.append((index, candidate))

    index, nearest = min(
        candidates, key=lambda item: abs((moment - item[1]).total_seconds())
    )
    delta = moment - nearest
    minutes = int(abs(delta.total_seconds()) // 60)
    if minutes > SHICHEN_BOUNDARY_THRESHOLD_MINUTES:
        return None
    before_boundary = delta.total_seconds() < 0
    zhi = load_constants()["地支"]
    # 交界 index 是「跨过去之后」那个时辰的起点：未跨=前一窗，已跨=本窗。
    current_index = (index - 1) % len(boundaries) if before_boundary else index
    alternate_index = index if before_boundary else (index - 1) % len(boundaries)
    return {
        "near_boundary": True,
        "minutes_to_boundary": minutes,
        "boundary_offset_minutes": -minutes if before_boundary else minutes,
        "boundary_solar": nearest.isoformat(),
        "threshold_minutes": SHICHEN_BOUNDARY_THRESHOLD_MINUTES,
        "current_shichen": _shichen_window(current_index, boundaries, zhi),
        "alternate_shichen": _shichen_window(alternate_index, boundaries, zhi),
        "direction": "later" if before_boundary else "earlier",
        "message": "出生时间接近时辰交界；建议提供 3-5 件已发生大事校准时辰。",
    }


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

def full_paipan(gregorian: str, gender: str, longitude: Optional[float] = None, correct: bool = True) -> dict:
    """本命盘（八字 + 紫微一次排全）。

    correct=True：真太阳时校正（路 A，用户给具体时刻）；
    correct=False：直接排盘不校正（路 B，用户已定时辰——再校正会二次偏移，日柱/时柱全错）。

    返回盘结构：{solar, lunar, chart, full, ziwei, ziwei_daxian, gender,
    pan_digest[, calibration_hint]}；用神结论位于 full.fu_yi，调候位于 full.tiao_hou，格局位于 full.ge_ju。
    返回结构是 factors 层的唯一输入；领域快照由 factors 层按 pan 生成。
    """
    if correct and longitude is None:
        raise ValueError(
            "correct=true 时 longitude 必填——真太阳时校正需要出生地经度。"
            "请先通过 city_coords 查询城市经度，或改用 correct=false（按给定时辰直接排）。"
        )
    if correct:
        t = _solar_time(gregorian, longitude)
        solar = t["solar"]
        lunar = t["lunar"]
        chart = _bazi_chart(solar, gender)   # 校正后时间直接排盘，不传 longitude（防二次校正）
    else:
        # 路 B：用户已定时辰，直接排盘不做真太阳时校正；农历换算使用
        # 东八区中央经线 120°，仅作历法转换，不引入出生地偏移。
        solar = gregorian
        chart = _bazi_chart(gregorian, gender)
        t = _solar_time(gregorian, 120.0)
        lunar = t["lunar"]
    full = _bazi_fullchart(chart)
    # 2.6.14 起用神三派归完整命盘（bazi.fullchart 承载，chart 纯排盘不含）
    zw = _ziwei_fullchart(lunar, gender)
    daxian = _ziwei_daxian(zw)
    result = {
        "solar": solar,
        "lunar": lunar,
        "chart": chart,      # 含 birth_year / da_yun
        "full": full,        # 十神/藏干/神煞/合会冲刑/三元
        "ziwei": zw,         # 十二宫/四化/杂曜/格局
        "ziwei_daxian": daxian,  # 十年大限（公历年段与宫位）
        "gender": gender,
    }
    if hint := _shichen_boundary_hint(solar):
        result["calibration_hint"] = hint
    result = with_natal_digest(result)
    validate_natal_pan(result, action="full_paipan result")
    return result


def liunian(pan: dict, year: int) -> dict:
    """流年盘（八字 + 紫微单年合并）。

    pan：full_paipan 返回的本命盘（只取其中 chart / ziwei 两个字段，agent 传整个盘即可）。
    year：要排的流年年份（应期按候选年逐个调——本命静态一次、流年动态按需）。

    返回：{bazi: 八字流年, ziwei: 紫微流年}。
    """
    validate_natal_pan(pan, action="liunian")
    return {
        "bazi": _bazi_liunian(pan["chart"], year),
        "ziwei": _ziwei_liunian(pan["ziwei"], year),
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

    pan_a / pan_b 为 full_paipan 返回的完整盘。
    返回: {"bazi": {...}, "ziwei": {...}}
    """
    validate_natal_pan(pan_a, action="bond pan_a")
    validate_natal_pan(pan_b, action="bond pan_b")
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
