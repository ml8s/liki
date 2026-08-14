"""排盘工具（数据准备——非命理逻辑，零命理判定）。

liki 命理 skill 的排盘层：
- full_paipan：本命盘（八字 + 紫微一次排全）+ 内嵌 build_factors 归并（返回盘含 fac）
- liunian：流年盘（八字 + 紫微单年合并——应期按候选年逐调）

命理逻辑不在本层（见 duanyu.py 的因子生成 + 断语查询）。
本层只做「读引擎字段 + 编排 RPC + 归并」，零命理判断。
"""
from __future__ import annotations
import json
import os
import urllib.request
from typing import Optional

RPC_URL = os.environ.get("LIKI_RPC_URL", "https://liki.hk/jsonrpc")
TIMEOUT = 30


class RPCError(Exception):
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
        except Exception as e:  # noqa: BLE001
            last_err = e
    raise RPCError(f"{method} 失败: {last_err}")


# ── 内部 RPC 封装（full_paipan / liunian 编排用，agent 不直接碰）──

def _solar_time(gregorian: str, longitude: float) -> dict:
    r = call("tianwen.time", {"time": gregorian, "longitude": longitude})
    return r["data"]


def _bazi_chart(solar: str, gender: str, longitude: Optional[float] = None) -> dict:
    params = {"solar_time": solar, "gender": gender}
    if longitude is not None:
        params["longitude"] = longitude
    return call("bazi.chart", params)["data"]


def _bazi_fullchart(chart: dict) -> dict:
    return call("bazi.fullchart", {"chart": chart})["data"]


def _bazi_yongshen(full: dict) -> dict:
    # 2.6.14 起用神三派归完整命盘（bazi.fullchart 承载，chart 纯排盘不含）
    return full.get("yong_shen", {})


def _ziwei_chart(lunar: dict, gender: str) -> dict:
    return call("ziwei.chart", {"lunar": lunar, "gender": gender})["data"]


def _bazi_liunian(chart: dict, year: int) -> dict:
    return call("bazi.liunian", {"chart": chart, "year": year})["data"]


def _ziwei_liunian(ziwei: dict, liu_nian: int) -> dict:
    return call("ziwei.liunian", {"chart": ziwei, "liu_nian": liu_nian})["data"]


# ── agent 工具（排盘 2 个）──

def full_paipan(gregorian: str, gender: str, longitude: Optional[float] = None, correct: bool = True) -> dict:
    """本命盘（八字 + 紫微一次排全）+ 内嵌 fac（build_factors 归并结果）。

    correct=True：真太阳时校正（路 A，用户给具体时刻）；
    correct=False：直接排盘不校正（路 B，题干已定时辰——再校正会二次偏移，日柱/时柱全错）。

    返回盘结构：{solar, lunar, chart, full, yongshen, ziwei, gender, fac}
    fac = build_factors(盘)——排盘数据归并（十神/五行/用神/紫微宫位），供因子生成直接读。
    """
    from aggregate import build_factors

    if longitude is None:
        longitude = 120.0
    if correct:
        t = _solar_time(gregorian, longitude)
        solar = t["solar"]
        lunar = t["lunar"]
        chart = _bazi_chart(solar, gender)   # 校正后时间直接排盘，不传 longitude（防二次校正）
    else:
        # 路 B：题干已定时辰，直接排盘不校正；紫微农历用经度 120（无真太阳时偏移）查
        solar = gregorian
        chart = _bazi_chart(gregorian, gender)
        t = _solar_time(gregorian, 120.0)
        lunar = t["lunar"]
    full = _bazi_fullchart(chart)
    ys = _bazi_yongshen(full)
    zw = _ziwei_chart(lunar, gender)
    pan = {
        "solar": solar,
        "lunar": lunar,
        "chart": chart,      # 含 birth_year / da_yun
        "full": full,        # 十神/藏干/神煞/合会冲刑/三元
        "yongshen": ys,      # 身强弱/旺衰/三派用神
        "ziwei": zw,         # 十二宫/四化/格局
        "gender": gender,
    }
    pan["fac"] = build_factors(pan)   # 内嵌归并——排盘产出即含 fac
    return pan


def liunian(pan: dict, year: int) -> dict:
    """流年盘（八字 + 紫微单年合并）。

    pan：full_paipan 返回的本命盘（只取其中 chart / ziwei 两个字段，agent 传整个盘即可）。
    year：要排的流年年份（应期按候选年逐个调——本命静态一次、流年动态按需）。

    返回：{bazi: 八字流年, ziwei: 紫微流年}。
    """
    return {
        "bazi": _bazi_liunian(pan["chart"], year),
        "ziwei": _ziwei_liunian(pan["ziwei"], year),
    }
