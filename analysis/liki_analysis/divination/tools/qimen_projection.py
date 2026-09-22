"""标准奇门稳定因子投影；专占上下文由 qimen_specialized.py 提供。"""
from __future__ import annotations

from qimen_projection_common import (  # noqa: F401
    load_factors_contract,
    _load_wuxing_relations,
    _palace_direction,
    _palace_domain,
    _palace_name,
    _palace_range,
    _palace_wuxing,
    _project_fields,
    _project_source,
    _read_path,
    _spirit_items,
    _type_matches,
    _validate_value,
    _hour_polarity,
)
from qimen_specialized import project as project_specialized_context


def _project_root_field(chart: dict, name: str, spec: dict):
    root = spec["path"].split(".", 1)[0]
    if not spec.get("required", True) and root not in chart:
        if spec["kind"] in {"object_array", "value_array"}:
            return []
        if spec["kind"] in {"scalar", "boolean"} and spec.get("nullable"):
            return None
        if spec["kind"] == "object":
            return {}
        raise ValueError(f"qimen chart lacks optional source field: {spec['path']}")
    if spec.get("projection"):
        found, source = _read_path(chart, spec["path"])
        if not found:
            raise ValueError(f"qimen chart lacks source field: {spec['path']}")
        return _project_stable_factors(source, chart, spec)
    found, source = _read_path(chart, spec["path"])
    if not found:
        if spec.get("required", True):
            raise ValueError(f"qimen chart lacks source field: {spec['path']}")
        if spec.get("nullable"):
            return None
        if spec["kind"] in {"object_array", "value_array"}:
            return []
        raise ValueError(f"qimen chart lacks optional source field: {spec['path']}")
    return _project_source(source, spec)

def _project_stable_factors(source, chart: dict, spec: dict):
    projection = spec["projection"]
    if projection not in {"gong_domain", "hour_polarity"} and (
            not isinstance(source, list)
            or any(not isinstance(item, dict) for item in source)):
        raise ValueError("chart source palace array invalid")
    runtime_fields = {
        field: {
            "path": field,
            "type": field_spec["type"],
            "required": bool(field_spec.get("required", True)),
            "nullable": bool(field_spec.get("nullable", False)),
        }
        for field, field_spec in spec.get("fields", {}).items()
    } if projection not in {"gong_domain", "hour_polarity"} else {}

    if projection in {"geng_ge", "geng_ge_levels", "geng_ge_status", "lost_context", "thief_context", "tianwang_context"}:
        return project_specialized_context(projection, source, chart)

    if projection == "stars":
        result = []
        for palace in source:
            gong = _palace_name(palace)
            for symbol in palace.get("tian_pan") or []:
                if not isinstance(symbol, dict):
                    raise ValueError("chart source heaven symbol invalid")
                result.append(_project_fields({
                    "star": symbol.get("xing"),
                    "gong": gong,
                    "heaven_gan": symbol.get("gan"),
                }, runtime_fields))
        return result

    if projection == "doors":
        return [
            _project_fields({
                "door": palace.get("men"), "gong": _palace_name(palace),
            }, runtime_fields)
            for palace in source if palace.get("men") is not None
        ]

    if projection == "spirits":
        return [
            _project_fields(item, runtime_fields)
            for item in _spirit_items(source, chart)
        ]

    if projection in {"spirit_star_relations", "spirit_spirit_relations"}:
        palace_wuxing = {}
        for item in chart.get("palace_wang_shuai", []):
            if isinstance(item, dict) and _type_matches(item.get("gong"), "string") \
                    and _type_matches(item.get("wuxing"), "string"):
                palace_wuxing[item["gong"]] = item["wuxing"]
        relations = _load_wuxing_relations()
        spirits = _spirit_items(source, chart)

        def relation(left_gong: str, right_gong: str) -> str:
            left = palace_wuxing.get(left_gong)
            right = palace_wuxing.get(right_gong)
            if not left or not right:
                raise ValueError(f"chart source lacks palace wuxing: {left_gong}/{right_gong}")
            key = (left, right)
            if key not in relations:
                raise ValueError(f"qimen wuxing relation missing: {left}/{right}")
            return relations[key]

        if projection == "spirit_star_relations":
            result = []
            stars = []
            for palace in source:
                gong = _palace_name(palace)
                for symbol in palace.get("tian_pan") or []:
                    if not isinstance(symbol, dict):
                        raise ValueError("chart source heaven symbol invalid")
                    star = symbol.get("xing")
                    if _type_matches(star, "string"):
                        stars.append((star, gong))
            for left in spirits:
                for star, gong in stars:
                    result.append({
                        "spirit": left["archetype"], "spirit_name": left["name"],
                        "spirit_gong": left["gong"], "star": star, "star_gong": gong,
                        "relation": relation(left["gong"], gong),
                        "same_gong": left["gong"] == gong,
                    })
            return result

        result = []
        for left_index, left in enumerate(spirits):
            for right_index, right in enumerate(spirits):
                if left_index == right_index:
                    continue
                result.append({
                    "spirit": left["archetype"], "spirit_name": left["name"],
                    "spirit_gong": left["gong"], "other": right["archetype"],
                    "other_name": right["name"], "other_gong": right["gong"],
                    "relation": relation(left["gong"], right["gong"]),
                    "same_gong": left["gong"] == right["gong"],
                })
        return result

    if projection == "palace_domains":
        result = []
        for palace in source:
            gong = _palace_name(palace)
            if _palace_domain(chart, gong) == "center":
                continue
            result.append({"gong": gong, "domain": _palace_domain(chart, gong)})
        return result

    if projection == "palace_directions":
        result = []
        for palace in source:
            gong = _palace_name(palace)
            direction = _palace_direction(gong)
            if direction == "center":
                continue
            result.append({"gong": gong, "direction": direction})
        return result

    if projection == "gong_domain":
        if not _type_matches(source, "string"):
            raise ValueError("chart source gong invalid")
        return _palace_domain(chart, source)

    if projection == "symbol_palaces":
        if not isinstance(source, list):
            raise ValueError("chart source symbol array invalid")
        source_palaces = chart.get("pan", {}).get("gong_wei", [])
        star_states = {
            row.get("xing"): row.get("traditional_label")
            for row in chart.get("xing_gong_wu_xing", [])
            if isinstance(row, dict)
        }
        result = []
        for item in source:
            if not isinstance(item, dict):
                raise ValueError("chart source symbol invalid")
            symbol = item.get("symbol")
            palace = item.get("palace")
            if not _type_matches(symbol, "string") or not _type_matches(palace, "string"):
                raise ValueError("chart source symbol palace invalid")
            self_palace = next((
                row for row in source_palaces
                if _palace_name(row) == palace
            ), {})
            stars = [
                symbol_row.get("xing")
                for symbol_row in self_palace.get("tian_pan") or []
                if isinstance(symbol_row, dict) and symbol_row.get("xing")
            ] or [None]
            for star in stars:
                result.append({
                    "symbol": symbol,
                    "palace": palace,
                    "domain": _palace_domain(chart, palace),
                    "direction": _palace_direction(palace),
                    "spirit": self_palace.get("shen"),
                    "door": self_palace.get("men"),
                    "wang_shuai": next((
                        row.get("wang_shuai") for row in chart.get("palace_wang_shuai", [])
                        if isinstance(row, dict) and row.get("gong") == palace
                    ), None),
                    "star": star,
                    "star_traditional_label": star_states.get(star),
                })
        return result

    if projection == "gan_interactions":
        if not isinstance(source, list):
            raise ValueError("chart source gan interaction array invalid")
        result = []
        for item in source:
            if not isinstance(item, dict):
                raise ValueError("chart source gan interaction invalid")
            gong = item.get("gong")
            if not _type_matches(gong, "string"):
                raise ValueError("chart source gan interaction palace invalid")
            result.append({
                "name": item.get("name"), "gong": gong,
                "range": _palace_range(gong),
                "direction": _palace_direction(gong),
                "heaven_gan": item.get("tian_pan_gan"),
                "earth_gan": item.get("di_pan_gan"),
                "auspicious": item.get("auspicious"),
            })
        return result

    if projection == "hour_polarity":
        if not _type_matches(source, "string"):
            raise ValueError("chart source hour stem invalid")
        return _hour_polarity(chart)

    if projection == "spirit_door_relations":
        doors = []
        for palace in source:
            door = palace.get("men")
            if door is None:
                continue
            if not _type_matches(door, "string"):
                raise ValueError("chart source door invalid")
            doors.append((door, _palace_name(palace)))
        result = []
        for left in _spirit_items(source, chart):
            for door, door_gong in doors:
                key = (_palace_wuxing(chart, left["gong"]), _palace_wuxing(chart, door_gong))
                relation = _load_wuxing_relations()[key]
                result.append({
                    "spirit": left["archetype"], "spirit_gong": left["gong"],
                    "door": door, "door_gong": door_gong,
                    "relation": relation, "same_gong": left["gong"] == door_gong,
                })
        return result

    raise ValueError(f"factors projection invalid: {projection}")

def project(chart: dict) -> dict:
    """把 qimen_chart 输出的 chart 投影为标准奇门稳定因子。"""
    if not isinstance(chart, dict):
        raise ValueError("chart must be an object")
    contract = load_factors_contract()
    return {
        name: _project_root_field(chart, name, spec)
        for name, spec in contract["fields"].items()
    }

def validate(factors: dict) -> None:
    """校验标准奇门解释层输入必须是完整 factors。"""
    contract = load_factors_contract()
    if not isinstance(factors, dict):
        raise ValueError("standard qimen factors must be an object")
    expected = set(contract["fields"])
    if set(factors) != expected:
        missing = sorted(expected - set(factors))
        extra = sorted(set(factors) - expected)
        raise ValueError(f"standard qimen factors invalid: missing={missing}, extra={extra}")
    for name, spec in contract["fields"].items():
        _validate_value(factors[name], spec)
