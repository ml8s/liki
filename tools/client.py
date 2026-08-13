from typing import Optional
"""liki 排盘 RPC 客户端（纯标准库，零依赖）。

调 liki.hk 引擎，返回结构化命盘数据（Layer 0 基础因子）。
"""
import json
import urllib.request

RPC_URL = "https://liki.hk/jsonrpc"
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


def solar_time(gregorian: str, longitude: float) -> dict:
    """真太阳时校正。gregorian: RFC3339；longitude: 出生地经度。"""
    r = call("tianwen.time", {"time": gregorian, "longitude": longitude})
    return r["data"]


def bazi_chart(solar: str, gender: str, longitude: Optional[float] = None) -> dict:
    """排八字最小命盘（含 birth_year 与大运）。"""
    params = {"solar_time": solar, "gender": gender}
    if longitude is not None:
        params["longitude"] = longitude
    return call("bazi.chart", params)["data"]


def bazi_fullchart(chart: dict) -> dict:
    """八字全量（十神/藏干/神煞/合会冲刑/三元）。"""
    return call("bazi.fullchart", {"chart": chart})["data"]


def bazi_yongshen(chart: dict) -> dict:
    """三派用神（扶抑/调候/格局，各含 yong/xi/ji + 身强弱 + 旺衰）。"""
    return call("bazi.yongshen", {"chart": chart})["data"]


def ziwei_chart(lunar: dict, gender: str) -> dict:
    """紫微命盘（十二宫/四化/格局）。"""
    return call("ziwei.chart", {"lunar": lunar, "gender": gender})["data"]


def bazi_liunian(chart: dict, year: int) -> dict:
    """流年（natal/dayun_interactions/fuyin_fanyin/shensha）。"""
    return call("bazi.liunian", {"chart": chart, "year": year})["data"]


def full_panchang(gregorian: str, gender: str, longitude: Optional[float] = None, correct: bool = True) -> dict:
    """一次排全：真太阳时 → 八字 → 全量 → 用神 → 紫微。返回 Layer 0 基础因子。

    correct=True：真太阳时校正（路 A，用户给具体时刻）；
    correct=False：直接排盘不校正（路 B，题干已定时辰——再校正会二次偏移，日柱/时柱全错）。
    """
    if longitude is None:
        longitude = 120.0
    if correct:
        t = solar_time(gregorian, longitude)
        solar = t["solar"]
        lunar = t["lunar"]
        chart = bazi_chart(solar, gender)   # 校正后时间直接排盘，不传 longitude（防二次校正）
    else:
        # 路 B：题干已定时辰，直接排盘不校正；紫微农历用经度 120（无真太阳时偏移）查
        solar = gregorian
        chart = bazi_chart(gregorian, gender)
        t = solar_time(gregorian, 120.0)
        lunar = t["lunar"]
    full = bazi_fullchart(chart)
    ys = bazi_yongshen(chart)
    zw = ziwei_chart(lunar, gender)
    return {
        "solar": solar,
        "lunar": lunar,
        "chart": chart,      # 含 birth_year / da_yun
        "full": full,        # 十神/藏干/神煞/合会冲刑/三元
        "yongshen": ys,      # 身强弱/旺衰/三派用神
        "ziwei": zw,         # 十二宫/四化/格局
        "gender": gender,
    }


def ziwei_daxian(ziwei_chart: dict) -> dict:
    """紫微大限（十年大限各宫，起岁=五行局数）。chart 为 ziwei.chart 完整对象。"""
    return call("ziwei.daxian", {"chart": ziwei_chart})


def ziwei_liunian(ziwei_chart: dict, year: int) -> dict:
    """紫微流年（流年四化落宫 + 流年十二宫星曜）。"""
    return call("ziwei.liunian", {"chart": ziwei_chart, "liu_nian": year})["data"]
