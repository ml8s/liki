"""奇门因子快照层：engine chart 的只读、表驱动结构化投影。"""
from __future__ import annotations

import csv
import json
from pathlib import Path


CONTRACT_PATH = Path(__file__).with_name("qimen_snapshot_contract.json")
SPIRIT_ARCHETYPES_PATH = Path(__file__).with_name("data") / "qimen_spirit_archetypes.csv"
WUXING_RELATIONS_PATH = Path(__file__).with_name("data") / "qimen_wuxing_relations.csv"
PALACE_DOMAINS_PATH = Path(__file__).with_name("data") / "qimen_palace_domains.csv"
PALACE_DIRECTIONS_PATH = Path(__file__).with_name("data") / "qimen_palace_directions.csv"
PALACE_RANGES_PATH = Path(__file__).with_name("data") / "qimen_palace_ranges.csv"
PALACE_PERSON_PATH = Path(__file__).with_name("data") / "qimen_palace_person.csv"
HOUR_POLARITIES_PATH = Path(__file__).with_name("data") / "qimen_hour_polarities.csv"
JIA_DUN_PATH = Path(__file__).with_name("data") / "qimen_jia_dun.csv"
_CONTRACT = None
_SPIRIT_ARCHETYPES = None
_WUXING_RELATIONS = None
_PALACE_DOMAINS = None
_PALACE_DIRECTIONS = None
_PALACE_RANGES = None
_PALACE_PERSON = None
_HOUR_POLARITIES = None
_JIA_DUN = None


def load_snapshot_contract() -> dict:
    global _CONTRACT
    if _CONTRACT is None:
        with CONTRACT_PATH.open(encoding="utf-8") as stream:
            _CONTRACT = json.load(stream)
    return _CONTRACT


def _load_lookup(
    path: Path,
    key_fields: tuple[str, ...],
    value_field: str,
    allowed_values: set[str] | None,
    expected_count: int,
    key_transform=lambda *values: values[0] if len(values) == 1 else values,
) -> dict:
    result = {}
    with path.open(encoding="utf-8-sig", newline="") as stream:
        for row in csv.DictReader(stream):
            key_values = tuple((row.get(field) or "").strip() for field in key_fields)
            value = (row.get(value_field) or "").strip()
            if any(not item for item in key_values) or not value:
                raise ValueError(f"qimen {path.name} has an empty field: {row}")
            if allowed_values is not None and value not in allowed_values:
                raise ValueError(f"qimen {path.name} has an invalid value: {row}")
            key = key_transform(*key_values)
            if key in result:
                raise ValueError(f"qimen {path.name} has a duplicate key: {row}")
            result[key] = value
    if len(result) != expected_count:
        raise ValueError(f"qimen {path.name} must contain {expected_count} rows")
    return result


def _load_spirit_archetypes() -> dict[tuple[str, str], str]:
    global _SPIRIT_ARCHETYPES
    if _SPIRIT_ARCHETYPES is None:
        _SPIRIT_ARCHETYPES = _load_lookup(
            SPIRIT_ARCHETYPES_PATH, ("spirit_mode", "name"), "archetype", None, 2
        )
    return _SPIRIT_ARCHETYPES


def _load_wuxing_relations() -> dict[tuple[str, str], str]:
    global _WUXING_RELATIONS
    if _WUXING_RELATIONS is None:
        _WUXING_RELATIONS = _load_lookup(
            WUXING_RELATIONS_PATH,
            ("source", "target"),
            "relation",
            {"same", "generates", "controls", "generated_by", "controlled_by"},
            25,
        )
    return _WUXING_RELATIONS


def _load_palace_domains() -> dict[tuple[bool, str], str]:
    global _PALACE_DOMAINS
    if _PALACE_DOMAINS is None:
        _PALACE_DOMAINS = _load_lookup(
            PALACE_DOMAINS_PATH,
            ("yin_dun", "gong"),
            "domain",
            {"inner", "outer", "center"},
            18,
            lambda yin_dun, gong: (yin_dun.lower() == "true", gong),
        )
    return _PALACE_DOMAINS


def _load_palace_ranges() -> dict[str, str]:
    global _PALACE_RANGES
    if _PALACE_RANGES is None:
        _PALACE_RANGES = _load_lookup(
            PALACE_RANGES_PATH, ("gong",), "range", {"low", "high"}, 9
        )
    return _PALACE_RANGES


def _load_palace_directions() -> dict[str, str]:
    global _PALACE_DIRECTIONS
    if _PALACE_DIRECTIONS is None:
        _PALACE_DIRECTIONS = _load_lookup(
            PALACE_DIRECTIONS_PATH, ("gong",), "direction", None, 9
        )
    return _PALACE_DIRECTIONS


def _load_palace_person() -> dict[str, str]:
    global _PALACE_PERSON
    if _PALACE_PERSON is None:
        _PALACE_PERSON = _load_lookup(
            PALACE_PERSON_PATH, ("gong",), "person_image", None, 8
        )
    return _PALACE_PERSON


def _load_jia_dun() -> dict[str, str]:
    global _JIA_DUN
    if _JIA_DUN is None:
        _JIA_DUN = _load_lookup(
            JIA_DUN_PATH, ("branch",), "liuyi", {"戊","己","庚","辛","壬","癸"}, 6
        )
    return _JIA_DUN


def _load_hour_polarities() -> dict[str, str]:
    global _HOUR_POLARITIES
    if _HOUR_POLARITIES is None:
        _HOUR_POLARITIES = _load_lookup(
            HOUR_POLARITIES_PATH, ("gan",), "polarity", {"yin", "yang"}, 10
        )
    return _HOUR_POLARITIES


def _read_path(value: dict, path: str) -> tuple[bool, object]:
    current = value
    for token in path.split("."):
        if not isinstance(current, dict) or token not in current:
            return False, None
        current = current[token]
    return True, current


def _type_matches(value, expected: str, nullable: bool = False) -> bool:
    if value is None:
        return nullable
    if expected == "string":
        return isinstance(value, str) and value != ""
    if expected == "boolean":
        return isinstance(value, bool)
    if expected == "array":
        return isinstance(value, list)
    if expected == "object":
        return isinstance(value, dict)
    return False


def _project_fields(source: dict, fields: dict) -> dict:
    result = {}
    for name, spec in fields.items():
        found, value = _read_path(source, spec["path"])
        if not found:
            if spec.get("required", True):
                raise ValueError(f"snapshot source lacks field: {spec['path']}")
            value = None
        if not _type_matches(value, spec["type"], spec.get("nullable", False)):
            raise ValueError(f"snapshot source field type invalid: {spec['path']}")
        result[name] = value
    return result


def _validate_value(value, spec: dict) -> None:
    kind = spec["kind"]
    nullable = spec.get("nullable", False)
    if kind == "scalar":
        if not _type_matches(value, spec["type"], nullable=nullable):
            raise ValueError("snapshot scalar value invalid")
        return
    if kind == "boolean":
        if not isinstance(value, bool):
            raise ValueError("snapshot boolean value invalid")
        return
    if kind == "object":
        if not isinstance(value, dict) or set(value) != set(spec["fields"]):
            raise ValueError("snapshot object fields invalid")
        for name, field_spec in spec["fields"].items():
            if "kind" in field_spec:
                _validate_value(value[name], field_spec)
            elif not _type_matches(value[name], field_spec["type"], field_spec.get("nullable", False)):
                raise ValueError("snapshot object field value invalid")
        return
    if kind == "object_array":
        if not isinstance(value, list):
            raise ValueError("snapshot object array invalid")
        for item in value:
            if not isinstance(item, dict) or set(item) != set(spec["fields"]):
                raise ValueError("snapshot object array item fields invalid")
            for name, field_spec in spec["fields"].items():
                if "kind" in field_spec:
                    _validate_value(item[name], field_spec)
                elif not _type_matches(item[name], field_spec["type"], field_spec.get("nullable", False)):
                    raise ValueError("snapshot object array field value invalid")
        return
    if kind == "value_array":
        if not isinstance(value, list) or any(
            not _type_matches(item, spec["type"]) for item in value
        ):
            raise ValueError("snapshot value array invalid")
        return
    raise ValueError("snapshot kind invalid")


def _project_source(value, spec: dict):
    kind = spec["kind"]
    if kind == "scalar":
        if spec.get("nullable") and value is None:
            return None
        if not _type_matches(value, spec["type"]):
            raise ValueError("snapshot source scalar invalid")
        return value
    if kind == "boolean":
        if not isinstance(value, bool):
            raise ValueError("snapshot source boolean invalid")
        return value
    if kind == "object":
        if not isinstance(value, dict):
            raise ValueError("snapshot source object invalid")
        return _project_fields(value, spec["fields"])
    if kind == "value_array":
        if not isinstance(value, list) or any(
            not _type_matches(item, spec["type"]) for item in value
        ):
            raise ValueError("snapshot source value array invalid")
        return list(value)
    if kind == "object_array":
        if not isinstance(value, list):
            raise ValueError("snapshot source object array invalid")
        result = []
        for item in value:
            if not isinstance(item, dict):
                raise ValueError("snapshot source object array item invalid")
            result.append(_project_fields(item, spec["fields"]))
        return result
    raise ValueError("snapshot source kind invalid")


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
        raise ValueError(f"qimen chart lacks optional source field: {spec['path']}")
    return _project_source(source, spec)


def _palace_name(palace: dict) -> str:
    value = palace.get("gong")
    if not isinstance(value, dict):
        raise ValueError("snapshot source palace identity invalid")
    name = value.get("name")
    if not _type_matches(name, "string"):
        raise ValueError("snapshot source palace name invalid")
    return name


def _spirit_archetype(spirit_mode: str, name: str) -> str:
    return _load_spirit_archetypes().get((spirit_mode, name), name)


def _spirit_items(palaces: list[dict], chart: dict) -> list[dict]:
    method = chart.get("method", {})
    spirit_mode = method.get("spirit_mode") if isinstance(method, dict) else None
    if not _type_matches(spirit_mode, "string"):
        raise ValueError("qimen chart lacks spirit_mode")
    result = []
    for palace in palaces:
        name = palace.get("shen")
        if name is None:
            continue
        if not _type_matches(name, "string"):
            raise ValueError("snapshot source spirit invalid")
        gong = _palace_name(palace)
        result.append({
            "name": name,
            "archetype": _spirit_archetype(spirit_mode, name),
            "gong": gong,
        })
    return result


def _yin_dun(chart: dict) -> bool:
    pan = chart.get("pan")
    value = pan.get("yin_dun") if isinstance(pan, dict) else None
    if not isinstance(value, bool):
        raise ValueError("qimen chart lacks yin_dun")
    return value


def _palace_domain(chart: dict, gong: str) -> str:
    domain = _load_palace_domains().get((_yin_dun(chart), gong))
    if domain is None:
        raise ValueError(f"qimen palace domain missing: {gong}")
    return domain


def _palace_range(gong: str) -> str:
    palace_range = _load_palace_ranges().get(gong)
    if palace_range is None:
        raise ValueError(f"qimen palace range missing: {gong}")
    return palace_range


def _palace_direction(gong: str) -> str | None:
    direction = _load_palace_directions().get(gong)
    if direction is None:
        raise ValueError(f"qimen palace direction missing: {gong}")
    return direction


def _palace_wuxing(chart: dict, gong: str) -> str:
    for item in chart.get("palace_wang_shuai", []):
        if isinstance(item, dict) and item.get("gong") == gong:
            wuxing = item.get("wuxing")
            if _type_matches(wuxing, "string"):
                return wuxing
    raise ValueError(f"snapshot source lacks palace wuxing: {gong}")


def _wuxing_relation(chart: dict, left_gong: str, right_gong: str) -> str:
    key = (_palace_wuxing(chart, left_gong), _palace_wuxing(chart, right_gong))
    relations = _load_wuxing_relations()
    if key not in relations:
        raise ValueError(f"qimen wuxing relation missing: {left_gong}/{right_gong}")
    return relations[key]


def _hour_polarity(chart: dict) -> str:
    pan = chart.get("pan")
    stem = pan.get("shi_gan") if isinstance(pan, dict) else None
    if not _type_matches(stem, "string"):
        raise ValueError("qimen chart lacks shi_gan")
    polarity = _load_hour_polarities().get(stem)
    if polarity is None:
        raise ValueError(f"qimen hour polarity missing: {stem}")
    return polarity


def _palace(chart: dict, gong: str) -> dict:
    pan = chart.get("pan", {})
    for row in pan.get("gong_wei", []) if isinstance(pan, dict) else []:
        if isinstance(row, dict) and _palace_name(row) == gong:
            return row
    raise ValueError(f"qimen palace missing: {gong}")


def _resolved_geng_target(gan: str, zhi: str) -> str:
    if gan != "甲":
        return gan
    return _load_jia_dun().get(zhi, "")


def _geng_levels(chart: dict) -> list[dict]:
    pillars = chart.get("pan", {})
    targets = []
    for level, gan_key, zhi_key in (
        ("year", "nian_gan", "nian_zhi"),
        ("month", "yue_gan", "yue_zhi"),
        ("day", "ri_gan", "ri_zhi"),
        ("hour", "shi_gan", "shi_zhi"),
    ):
        gan = pillars.get(gan_key) if isinstance(pillars, dict) else None
        zhi = pillars.get(zhi_key) if isinstance(pillars, dict) else None
        if _type_matches(gan, "string") and _type_matches(zhi, "string"):
            targets.append((level, _resolved_geng_target(gan, zhi)))
    result = []
    for row in chart.get("gan_interaction", []):
        if not isinstance(row, dict) or row.get("tian_pan_gan") != "庚":
            continue
        gong = row.get("gong")
        earth_gan = row.get("di_pan_gan")
        if not _type_matches(gong, "string") or not _type_matches(earth_gan, "string"):
            continue
        for level, target in targets:
            if earth_gan != target:
                continue
            door = _palace(chart, gong).get("men")
            result.append({
                "level": level,
                "gong": gong,
                "direction": _palace_direction(gong),
                "door": door,
                "earth_gan": earth_gan,
            })
    return result


def _patterns_at(chart: dict, gong: str) -> list[dict]:
    return [
        row for row in chart.get("patterns", [])
        if isinstance(row, dict) and gong in (row.get("gong_wei") or [])
    ]


def _project_stable_factors(source, chart: dict, spec: dict):
    projection = spec["projection"]
    if projection not in {"gong_domain", "hour_polarity"} and (
            not isinstance(source, list)
            or any(not isinstance(item, dict) for item in source)):
        raise ValueError("snapshot source palace array invalid")
    runtime_fields = {
        field: {
            "path": field,
            "type": field_spec["type"],
            "required": bool(field_spec.get("required", True)),
            "nullable": bool(field_spec.get("nullable", False)),
        }
        for field, field_spec in spec.get("fields", {}).items()
    } if projection not in {"gong_domain", "hour_polarity"} else {}

    if projection == "stars":
        result = []
        for palace in source:
            gong = _palace_name(palace)
            for symbol in palace.get("tian_pan") or []:
                if not isinstance(symbol, dict):
                    raise ValueError("snapshot source heaven symbol invalid")
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
                raise ValueError(f"snapshot source lacks palace wuxing: {left_gong}/{right_gong}")
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
                        raise ValueError("snapshot source heaven symbol invalid")
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
            raise ValueError("snapshot source gong invalid")
        return _palace_domain(chart, source)

    if projection == "symbol_palaces":
        if not isinstance(source, list):
            raise ValueError("snapshot source symbol array invalid")
        source_palaces = chart.get("pan", {}).get("gong_wei", [])
        star_states = {
            row.get("xing"): row.get("traditional_label")
            for row in chart.get("xing_gong_wu_xing", [])
            if isinstance(row, dict)
        }
        result = []
        for item in source:
            if not isinstance(item, dict):
                raise ValueError("snapshot source symbol invalid")
            symbol = item.get("symbol")
            palace = item.get("palace")
            if not _type_matches(symbol, "string") or not _type_matches(palace, "string"):
                raise ValueError("snapshot source symbol palace invalid")
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
            raise ValueError("snapshot source gan interaction array invalid")
        result = []
        for item in source:
            if not isinstance(item, dict):
                raise ValueError("snapshot source gan interaction invalid")
            gong = item.get("gong")
            if not _type_matches(gong, "string"):
                raise ValueError("snapshot source gan interaction palace invalid")
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
            raise ValueError("snapshot source hour stem invalid")
        return _hour_polarity(chart)

    if projection == "geng_ge":
        return _geng_levels(chart)

    if projection == "geng_ge_levels":
        return sorted({row["level"] for row in _geng_levels(chart)})

    if projection == "geng_ge_status":
        return "present" if _geng_levels(chart) else "absent"

    if projection == "lost_context":
        shi_gong = chart.get("shi_gan_gong")
        if not _type_matches(shi_gong, "string"):
            raise ValueError("qimen chart lacks shi_gan_gong")
        return [{
            "gong": shi_gong,
            "direction": _palace_direction(shi_gong),
            "domain": _palace_domain(chart, shi_gong),
            "door": next((
                row.get("men") for row in chart.get("pan", {}).get("gong_wei", [])
                if isinstance(row, dict) and row.get("gong", {}).get("name") == shi_gong
            ), None),
            "wang_shuai": next((
                row.get("wang_shuai") for row in source
                if isinstance(row, dict) and row.get("gong") == shi_gong
            ), None),
        }]

    if projection == "thief_context":
        result = []
        star_rows = {
            row.get("gong"): row
            for row in chart.get("xing_gong_wu_xing", [])
            if isinstance(row, dict) and row.get("xing") == "天蓬"
        }
        for palace_row in source:
            gong = _palace_name(palace_row) if isinstance(palace_row, dict) else None
            if not _type_matches(gong, "string"):
                continue
            if gong == "中":
                continue
            patterns = _patterns_at(chart, gong)
            lucky_patterns = [
                item.get("name") for item in patterns if item.get("auspicious") is True
            ]
            palace_wang = next((
                row.get("wang_shuai") for row in chart.get("palace_wang_shuai", [])
                if isinstance(row, dict) and row.get("gong") == gong
            ), "未知")
            star_row = star_rows.get(gong)
            if star_row:
                result.append({
                    "role": "大贼", "thief_symbol": "天蓬", "gong": gong,
                    "direction": _palace_direction(gong),
                    "domain": _palace_domain(chart, gong),
                    "person_image": _load_palace_person()[gong],
                    "wang_shuai_label": star_row.get("traditional_label") or "未知",
                    "lucky_pattern": bool(lucky_patterns),
                    "lucky_pattern_names": lucky_patterns,
                })
            spirit = next((item for item in _spirit_items(source, chart) if item["gong"] == gong and item["archetype"] == "玄武"), None)
            if spirit:
                result.append({
                    "role": "小贼", "thief_symbol": "玄武", "gong": gong,
                    "direction": _palace_direction(gong),
                    "domain": _palace_domain(chart, gong),
                    "person_image": _load_palace_person()[gong],
                    "wang_shuai_label": palace_wang or "未知",
                    "lucky_pattern": bool(lucky_patterns),
                    "lucky_pattern_names": lucky_patterns,
                })
        return result

    if projection == "tianwang_context":
        hour_gong = chart.get("shi_gan_gong")
        if not _type_matches(hour_gong, "string"):
            raise ValueError("qimen chart lacks shi_gan_gong")
        result = []
        for row in source:
            if not isinstance(row, dict) or row.get("tian_pan_gan") != "癸":
                continue
            gong = row.get("gong")
            if not _type_matches(gong, "string"):
                raise ValueError("snapshot source gan interaction palace invalid")
            result.append({
                "gong": gong,
                "direction": _palace_direction(gong),
                "range": _palace_range(gong),
                "hour_gong": hour_gong,
                "hour_direction": _palace_direction(hour_gong),
                "relation_to_hour": _wuxing_relation(chart, gong, hour_gong),
                "hour_polarity": _hour_polarity(chart),
            })
        return result

    if projection == "spirit_door_relations":
        doors = []
        for palace in source:
            door = palace.get("men")
            if door is None:
                continue
            if not _type_matches(door, "string"):
                raise ValueError("snapshot source door invalid")
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

    raise ValueError(f"snapshot stable projection invalid: {projection}")


def project_qimen_snapshot(pan: dict) -> dict:
    """把 qimen_chart 返回的 chart 投影为解释层稳定结构化快照。"""
    if not isinstance(pan, dict) or not isinstance(pan.get("chart"), dict):
        raise ValueError("pan 必须是 qimen_chart 返回的结构，且包含 chart 字段")
    chart = pan["chart"]
    contract = load_snapshot_contract()
    return {
        name: _project_root_field(chart, name, spec)
        for name, spec in contract["fields"].items()
    }


def validate_qimen_snapshot(snapshot: dict) -> None:
    """校验解释层输入必须是完整 snapshot v3。"""
    contract = load_snapshot_contract()
    if not isinstance(snapshot, dict):
        raise ValueError("qimen snapshot 必须是 object")
    expected = set(contract["fields"])
    if set(snapshot) != expected:
        missing = sorted(expected - set(snapshot))
        extra = sorted(set(snapshot) - expected)
        raise ValueError(f"qimen snapshot 字段无效: missing={missing}, extra={extra}")
    for name, spec in contract["fields"].items():
        _validate_value(snapshot[name], spec)
