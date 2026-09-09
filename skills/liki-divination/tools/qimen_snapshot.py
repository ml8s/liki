"""奇门 snapshot 创建：时间、地点、排盘与因子投影一次完成。"""
from __future__ import annotations

from divination_safety import assess, blocked_payload
from divination_snapshot import build_snapshot
from qimen_duanyu import query
from qimen_factors import project_qimen_snapshot
from qimen_interpretations import assert_rule_for_pan
from qimen_paipan import city_coords, engine_data, qimen_chart, solar_time


SCHEMA_VERSION = "qimen-snapshot-v3"


def _server_time() -> str:
    response = engine_data("time.now", {})
    cst = response.get("cst")
    if not isinstance(cst, str) or not cst:
        raise ValueError("time.now returned no cst")
    return cst


def _resolve_location(city: str | None, longitude: float | None) -> tuple[str | None, float]:
    if longitude is not None:
        if isinstance(longitude, bool) or not isinstance(longitude, (int, float)):
            raise ValueError("longitude must be a number")
        if not -180 <= float(longitude) <= 180:
            raise ValueError("longitude must be between -180 and 180")
        return None, float(longitude)
    if city:
        coords = city_coords(city)
        value = coords.get("longitude")
        if isinstance(value, bool) or not isinstance(value, (int, float)):
            raise ValueError("city.coords returned no longitude")
        return city, float(value)
    raise ValueError("qimen requires city or longitude")


def create(
    *,
    question: str,
    city: str | None = None,
    longitude: float | None = None,
    time: str | None = None,
    matter: str | None = None,
    yong_shen: list[str] | None = None,
    rule: str | None = None,
    scope: str | None = None,
    school: str | None = None,
    dingju_method: str | None = None,
    quarter_rule: str | None = None,
    base_dingju_method: str | None = None,
    dun_source: str | None = None,
    hour_boundary: str | None = None,
    birth_date: str | None = None,
) -> dict:
    """创建 immutable qimen snapshot；LLM 后续只依赖该对象。"""
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload(SCHEMA_VERSION, question, "qimen")
    if not city and longitude is None:
        raise ValueError("qimen requires city or longitude")
    if matter is not None and yong_shen is not None:
        raise ValueError("matter 与 yong_shen 互斥")

    if not time:
        time = _server_time()
    resolved_city, resolved_longitude = _resolve_location(city, longitude)
    local = solar_time(time, resolved_longitude)
    solar = local.get("solar")
    if not isinstance(solar, str) or not solar:
        raise ValueError("tianwen.time returned no solar")

    pan = qimen_chart(
        solar,
        matter=matter,
        yong_shen=yong_shen,
        birth_date=birth_date,
        scope=scope,
        school=school,
        dingju_method=dingju_method,
        quarter_rule=quarter_rule,
        base_dingju_method=base_dingju_method,
        dun_source=dun_source,
        hour_boundary=hour_boundary,
    )
    factors = project_qimen_snapshot(pan)
    method_context = pan.get("chart", {}).get("method")
    special = None
    if rule is not None:
        assert_rule_for_pan(rule, pan)
        special = query(rule, factors)

    return build_snapshot(
        method="qimen",
        schema_version=SCHEMA_VERSION,
        payload={
            "question": question.strip(),
            "input": {
                "city": resolved_city,
                "longitude": resolved_longitude,
                "local_time": time,
                "solar_time": solar,
            },
            "matter": pan.get("matter"),
            "method_context": method_context,
            "factors": factors,
            "special": special,
            "policy": {
                "immutable": True,
                "no_rechart_without_new_event": True,
            },
        },
    )
