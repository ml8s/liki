"""奇门专占上下文投影；核心匹配由 engine 提供，本层只补固定呈现字段。"""

from __future__ import annotations

from qimen_projection_common import (
    _hour_polarity,
    _load_palace_person,
    _palace_direction,
    _palace_domain,
    _palace_range,
    _read_path,
)


def _context(chart: dict, key: str) -> list:
    found, value = _read_path(chart, f"specialized.{key}")
    if not found or not isinstance(value, list):
        raise ValueError(f"qimen chart lacks specialized.{key}")
    if any(not isinstance(item, dict) for item in value):
        raise ValueError(f"qimen specialized.{key} contains a non-object")
    return value


def _geng_levels(chart: dict) -> list[dict]:
    result = []
    for row in _context(chart, "geng_ge"):
        gong = row.get("gong")
        result.append({
            "level": row.get("level"),
            "gong": gong,
            "direction": _palace_direction(gong),
            "door": row.get("door"),
            "earth_gan": row.get("earth_gan"),
        })
    return result


def _lost_rows(chart: dict) -> list[dict]:
    return [{
        "gong": row.get("gong"),
        "direction": _palace_direction(row.get("gong")),
        "domain": _palace_domain(chart, row.get("gong")),
        "door": row.get("door"),
        "wang_shuai": row.get("wang_shuai"),
    } for row in _context(chart, "lost_context")]


def _thief_rows(chart: dict) -> list[dict]:
    result = []
    for row in _context(chart, "thief_context"):
        gong = row.get("gong")
        result.append({
            "role": row.get("role"),
            "thief_symbol": row.get("symbol"),
            "gong": gong,
            "direction": _palace_direction(gong),
            "domain": _palace_domain(chart, gong),
            "person_image": _load_palace_person()[gong],
            "wang_shuai_label": row.get("wang_shuai_label"),
            "lucky_pattern": bool(row.get("lucky_pattern_names")),
            "lucky_pattern_names": row.get("lucky_pattern_names", []),
        })
    return result


def _tianwang_rows(chart: dict) -> list[dict]:
    result = []
    for row in _context(chart, "tianwang_context"):
        gong, hour_gong = row.get("gong"), row.get("hour_gong")
        result.append({
            "gong": gong,
            "direction": _palace_direction(gong),
            "range": _palace_range(gong),
            "hour_gong": hour_gong,
            "hour_direction": _palace_direction(hour_gong),
            "relation_to_hour": row.get("relation_to_hour"),
            "hour_polarity": _hour_polarity(chart),
        })
    return result


def project(projection: str, source, chart: dict):
    """Project engine-owned specialized facts with fixed presentation fields."""
    if projection == "geng_ge":
        return _geng_levels(chart)
    if projection == "geng_ge_levels":
        order = {level: rank for rank, level in enumerate(("year", "month", "day", "hour"))}
        return sorted({row["level"] for row in _geng_levels(chart)}, key=order.get)
    if projection == "geng_ge_status":
        return "present" if _geng_levels(chart) else "absent"
    if projection == "lost_context":
        return _lost_rows(chart)
    if projection == "thief_context":
        return _thief_rows(chart)
    if projection == "tianwang_context":
        return _tianwang_rows(chart)
    raise ValueError(f"specialized qimen projection invalid: {projection}")
