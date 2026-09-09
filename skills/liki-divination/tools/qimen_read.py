"""奇门读盘编排：城市 / 时间 / 排盘 / snapshot / 专占解释一次完成。"""
from __future__ import annotations

from typing import Any

from divination_safety import assess, blocked_payload
from qimen_duanyu import query
from qimen_factors import project_qimen_snapshot
from qimen_interpretations import assert_rule_for_pan
from qimen_paipan import city_coords, engine_data, qimen_chart, solar_time
from qimen_report import template as report_template


def _server_time() -> str:
    response = engine_data("time.now", {})
    cst = response.get("cst")
    if not isinstance(cst, str) or not cst:
        raise ValueError("time.now returned no cst")
    return cst


def _resolve_location(
    city: str | None,
    longitude: float | None,
) -> tuple[str | None, float]:
    if longitude is not None:
        if not isinstance(longitude, (int, float)) or isinstance(longitude, bool):
            raise ValueError("longitude must be a number")
        if not -180 <= float(longitude) <= 180:
            raise ValueError("longitude must be between -180 and 180")
        return None, float(longitude)
    if city:
        coords = city_coords(city)
        value = coords.get("longitude")
        if not isinstance(value, (int, float)) or isinstance(value, bool):
            raise ValueError("city.coords returned no longitude")
        return city, float(value)
    raise ValueError("奇门需要地点口径：请提供 city 或 longitude")


def read(
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
    """一次调用完成奇门时间解析、排盘、快照投影和可选专占查询。"""
    if not isinstance(question, str) or not question.strip():
        raise ValueError("question must be non-empty text")
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload("qimen-read-v1", question, "qimen")
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
    snapshot = project_qimen_snapshot(pan)

    special = None
    if rule is not None:
        assert_rule_for_pan(rule, pan)
        special = {
            "rule": rule,
            "assertions": query(rule, snapshot),
        }

    return {
        "schema_version": "qimen-read-v1",
        "question": question,
        "input": {
            "city": resolved_city,
            "longitude": resolved_longitude,
            "local_time": time,
            "solar_time": solar,
        },
        "matter": pan.get("matter"),
        "chart": pan.get("chart"),
        "snapshot": snapshot,
        "special": special,
        "report_template": report_template({
            "schema_version": "qimen-read-v1",
            "question": question,
            "snapshot": snapshot,
            "special": special,
        }),
    }
