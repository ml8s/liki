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
import json
import os
import urllib.request
from urllib.error import URLError
from typing import Optional

from errors import LikiToolError
from pan_schema import validate_natal_pan

RPC_URL = os.environ.get("LIKI_RPC_URL", "https://liki.hk/jsonrpc")
TIMEOUT = 30


class RPCError(LikiToolError):
    pass


def call(method: str, params: dict, retries: int = 1) -> dict:
    """调 JSON-RPC。失败重试 retries 次。"""
    body = json.dumps({"jsonrpc": "2.0", "method": method, "params": params, "id": 1}).encode()
    last_err = None
    for _ in range(retries + 1):
        try:
            req = urllib.request.Request(RPC_URL, data=body, headers={"Content-Type": "application/json"})
            with urllib.request.urlopen(req, timeout=TIMEOUT) as resp:
                data = json.loads(resp.read().decode())
            if "error" in data:
                raise RPCError(f"{method}: {data['error']}")
            return data["result"]
        except (URLError, ConnectionError, TimeoutError, OSError) as e:
            last_err = e
    raise RPCError(f"{method} 失败: {last_err}")


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


def _bazi_liunian(chart: dict, year: int) -> dict:
    return call("bazi.liunian", {"chart": chart, "year": year})["data"]


def _ziwei_liunian(ziwei: dict, lunar_year: int) -> dict:
    return call("ziwei.liunian", {"chart": ziwei, "lunar_year": lunar_year})["data"]


# ── agent 工具（排盘 2 个）──

def full_paipan(gregorian: str, gender: str, longitude: Optional[float] = None, correct: bool = True) -> dict:
    """本命盘（八字 + 紫微一次排全）。

    correct=True：真太阳时校正（路 A，用户给具体时刻）；
    correct=False：直接排盘不校正（路 B，用户已定时辰——再校正会二次偏移，日柱/时柱全错）。

    返回盘结构：{solar, lunar, chart, full, yongshen, ziwei, gender}
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
        # 路 B：用户已定时辰，直接排盘不校正；紫微农历用经度 120（东八区中央经线，无真太阳时偏移——
        # 与路 A 默认 116.4 北京经度区分：路 A 校正用出生地经度，路 B 不校正用标准时经度）查
        solar = gregorian
        chart = _bazi_chart(gregorian, gender)
        t = _solar_time(gregorian, 120.0)
        lunar = t["lunar"]
    full = _bazi_fullchart(chart)
    # 2.6.14 起用神三派归完整命盘（bazi.fullchart 承载，chart 纯排盘不含）
    ys = full.get("yong_shen", {})
    zw = _ziwei_chart(lunar, gender)
    result = {
        "solar": solar,
        "lunar": lunar,
        "chart": chart,      # 含 birth_year / da_yun
        "full": full,        # 十神/藏干/神煞/合会冲刑/三元
        "yongshen": ys,      # 身强弱/旺衰/三派用神
        "ziwei": zw,         # 十二宫/四化/格局
        "gender": gender,
    }
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
    r = call("city.coords", {"city": city})
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
