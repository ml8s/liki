"""奇门投影共享表加载与机械取值 helper。"""
from __future__ import annotations

import csv
import json
from functools import lru_cache
from pathlib import Path

CONTRACT_PATH = Path(__file__).with_name("qimen_projection_contract.json")
SPIRIT_ARCHETYPES_PATH = Path(__file__).with_name("data") / "qimen_spirit_archetypes.csv"
WUXING_RELATIONS_PATH = Path(__file__).with_name("data") / "qimen_wuxing_relations.csv"
PALACE_DOMAINS_PATH = Path(__file__).with_name("data") / "qimen_palace_domains.csv"
PALACE_DIRECTIONS_PATH = Path(__file__).with_name("data") / "qimen_palace_directions.csv"
PALACE_RANGES_PATH = Path(__file__).with_name("data") / "qimen_palace_ranges.csv"
PALACE_PERSON_PATH = Path(__file__).with_name("data") / "qimen_palace_person.csv"
HOUR_POLARITIES_PATH = Path(__file__).with_name("data") / "qimen_hour_polarities.csv"
_CONTRACT = None
_SPIRIT_ARCHETYPES = None
_WUXING_RELATIONS = None
_PALACE_DOMAINS = None
_PALACE_DIRECTIONS = None
_PALACE_RANGES = None
_PALACE_PERSON = None
_HOUR_POLARITIES = None


def load_factors_contract() -> dict:
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
    if expected == "integer":
        return isinstance(value, int) and not isinstance(value, bool)
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
                raise ValueError(f"chart source lacks field: {spec['path']}")
            value = None
        if not _type_matches(value, spec["type"], spec.get("nullable", False)):
            raise ValueError(f"chart source field type invalid: {spec['path']}")
        result[name] = value
    return result

def _validate_value(value, spec: dict) -> None:
    kind = spec["kind"]
    nullable = spec.get("nullable", False)
    if kind == "scalar":
        if not _type_matches(value, spec["type"], nullable=nullable):
            raise ValueError("factor scalar value invalid")
        return
    if kind == "boolean":
        if not isinstance(value, bool):
            raise ValueError("factor boolean value invalid")
        return
    if kind == "object":
        if not isinstance(value, dict) or set(value) != set(spec["fields"]):
            raise ValueError("factor object fields invalid")
        for name, field_spec in spec["fields"].items():
            if "kind" in field_spec:
                _validate_value(value[name], field_spec)
            elif not _type_matches(value[name], field_spec["type"], field_spec.get("nullable", False)):
                raise ValueError("factor object field value invalid")
        return
    if kind == "object_array":
        if not isinstance(value, list):
            raise ValueError("factor object array invalid")
        for item in value:
            if not isinstance(item, dict) or set(item) != set(spec["fields"]):
                raise ValueError("factor object array item fields invalid")
            for name, field_spec in spec["fields"].items():
                if "kind" in field_spec:
                    _validate_value(item[name], field_spec)
                elif not _type_matches(item[name], field_spec["type"], field_spec.get("nullable", False)):
                    raise ValueError("factor object array field value invalid")
        return
    if kind == "value_array":
        if not isinstance(value, list) or any(
            not _type_matches(item, spec["type"]) for item in value
        ):
            raise ValueError("factor value array invalid")
        return
    raise ValueError("factor kind invalid")

def _project_source(value, spec: dict):
    kind = spec["kind"]
    if kind == "scalar":
        if spec.get("nullable") and value is None:
            return None
        if not _type_matches(value, spec["type"]):
            raise ValueError("chart source scalar invalid")
        return value
    if kind == "boolean":
        if not isinstance(value, bool):
            raise ValueError("chart source boolean invalid")
        return value
    if kind == "object":
        if not isinstance(value, dict):
            raise ValueError("chart source object invalid")
        return _project_fields(value, spec["fields"])
    if kind == "value_array":
        if not isinstance(value, list) or any(
            not _type_matches(item, spec["type"]) for item in value
        ):
            raise ValueError("chart source value array invalid")
        return list(value)
    if kind == "object_array":
        if not isinstance(value, list):
            raise ValueError("chart source object array invalid")
        result = []
        for item in value:
            if not isinstance(item, dict):
                raise ValueError("chart source object array item invalid")
            result.append(_project_fields(item, spec["fields"]))
        return result
    raise ValueError("chart source kind invalid")

def _palace_name(palace: dict) -> str:
    value = palace.get("gong")
    if not isinstance(value, dict):
        raise ValueError("chart source palace identity invalid")
    name = value.get("name")
    if not _type_matches(name, "string"):
        raise ValueError("chart source palace name invalid")
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
            raise ValueError("chart source spirit invalid")
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
    raise ValueError(f"chart source lacks palace wuxing: {gong}")

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
