"""奇门解释表加载器；条件与适用范围均由表驱动。"""
from __future__ import annotations

import csv
from collections import defaultdict
from pathlib import Path

from qimen_errors import TableError
from qimen_factors import load_snapshot_contract


TOOLS_DIR = Path(__file__).resolve().parent
RULES_PATH = TOOLS_DIR / "assertions" / "qimen_rules.csv"
INTERPRETATIONS_PATH = TOOLS_DIR / "assertions" / "qimen_assertions.csv"
INTERPRETATION_CONDITIONS_PATH = TOOLS_DIR / "assertions" / "qimen_conditions.csv"

_RULE_TABLE = None
_INTERPRETATION_INDEX = None


def _split_values(value: str) -> list[str]:
    return [item.strip() for item in value.split("|") if item.strip()]


def load_rule_table() -> dict[str, dict]:
    """加载专占规则的 scope / school 适用范围。"""
    global _RULE_TABLE
    if _RULE_TABLE is not None:
        return _RULE_TABLE
    result: dict[str, dict] = {}
    with RULES_PATH.open(encoding="utf-8-sig", newline="") as stream:
        for row in csv.DictReader(stream):
            rule = (row.get("rule") or "").strip()
            name = (row.get("name") or "").strip()
            scopes = _split_values(row.get("scopes") or "")
            schools = _split_values(row.get("schools") or "")
            basis = (row.get("basis") or "").strip()
            if not rule or not name or not scopes or not schools or not basis:
                raise TableError(f"奇门解释规则表存在空字段: {row}")
            if rule in result:
                raise TableError(f"奇门解释规则重复: {rule}")
            if len(scopes) != len(set(scopes)):
                raise TableError(f"奇门解释规则 scopes 重复: {rule}")
            if len(schools) != len(set(schools)):
                raise TableError(f"奇门解释规则 schools 重复: {rule}")
            result[rule] = {
                "name": name,
                "scopes": scopes,
                "schools": schools,
                "basis": basis,
            }
    if not result:
        raise TableError("奇门解释规则表为空")
    _RULE_TABLE = result
    return result


def assert_rule_compatibility(rule: str, scope: str, school: str) -> None:
    metadata = load_rule_table().get(rule)
    if metadata is None:
        raise ValueError(f"unknown qimen interpretation rule: {rule}")
    if scope not in metadata["scopes"] or school not in metadata["schools"]:
        raise ValueError(
            f"qimen {rule} requires scope in {metadata['scopes']} and "
            f"school in {metadata['schools']}; got scope={scope}, school={school}"
        )


def assert_rule_for_pan(rule: str, pan: dict) -> None:
    if not isinstance(pan, dict) or not isinstance(pan.get("chart"), dict):
        raise ValueError("pan 必须是 qimen_chart 返回的结构，且包含 chart 字段")
    method = pan["chart"].get("method")
    if not isinstance(method, dict):
        raise ValueError("qimen chart 缺少可用字段: method")
    scope = method.get("scope")
    school = method.get("school")
    if not isinstance(scope, str) or not isinstance(school, str):
        raise ValueError("qimen chart 缺少可用字段: method.scope / method.school")
    assert_rule_compatibility(rule, scope, school)


def load_interpretation_index() -> dict[str, list[dict]]:
    """按 rule 聚合解释行；条件组内 AND、组间 OR。"""
    global _INTERPRETATION_INDEX
    if _INTERPRETATION_INDEX is not None:
        return _INTERPRETATION_INDEX
    rules = load_rule_table()
    rows: dict[str, list[dict]] = defaultdict(list)
    row_by_id: dict[str, dict] = {}
    with INTERPRETATIONS_PATH.open(encoding="utf-8-sig", newline="") as stream:
        for source in csv.DictReader(stream):
            assertion_id = (source.get("assertion_id") or "").strip()
            rule = (source.get("rule") or "").strip()
            name = (source.get("name") or "").strip()
            conclusion = (source.get("conclusion") or "").strip()
            basis = (source.get("basis") or "").strip()
            if not assertion_id or not rule or not name or not conclusion or not basis:
                raise TableError(f"奇门解释表存在空字段: {source}")
            if assertion_id in row_by_id:
                raise TableError(f"奇门解释重复: {assertion_id}")
            if rule not in rules:
                raise TableError(f"奇门解释规则未登记: {rule}")
            item = {
                "id": assertion_id,
                "rule": rule,
                "name": name,
                "conclusion": conclusion,
                "basis": basis,
                "conditions": [],
            }
            rows[rule].append(item)
            row_by_id[assertion_id] = item

    groups: dict[tuple[str, int], list[dict]] = defaultdict(list)
    snapshot_fields = load_snapshot_contract()["fields"]
    with INTERPRETATION_CONDITIONS_PATH.open(encoding="utf-8-sig", newline="") as stream:
        for source in csv.DictReader(stream):
            assertion_id = (source.get("assertion_id") or "").strip()
            group_id = (source.get("condition_group_id") or "").strip()
            condition = {
                "field": (source.get("field") or "").strip(),
                "item_key": (source.get("item_key") or "").strip(),
                "operator": (source.get("operator") or "").strip(),
                "expected": (source.get("expected") or "").strip(),
            }
            if assertion_id not in row_by_id or not group_id:
                raise TableError(f"奇门解释条件归属无效: {source}")
            if not group_id.isdigit() or int(group_id) <= 0:
                raise TableError(f"奇门解释 condition_group_id 必须为正整数: {group_id}")
            field = snapshot_fields.get(condition["field"])
            if field is None:
                raise TableError(f"奇门解释字段不在快照契约中: {condition['field']}")
            kind = field["kind"]
            if kind == "object_array":
                if condition["operator"] not in {"array_some_equals", "array_some_in"}:
                    raise TableError(f"奇门解释数组条件无效: {assertion_id} {condition}")
                item_fields = field["fields"]
                if condition["item_key"] not in item_fields:
                    raise TableError(f"奇门解释 item_key 不在快照契约中: {condition['item_key']}")
            elif kind == "object":
                if condition["operator"] not in {"equals", "is_null"}:
                    raise TableError(f"奇门解释 object 条件无效: {assertion_id} {condition}")
                if condition["item_key"] not in field["fields"]:
                    raise TableError(f"奇门解释 item_key 不在快照契约中: {condition['item_key']}")
            else:
                if condition["item_key"]:
                    raise TableError(f"奇门解释标量条件不能有 item_key: {assertion_id}")
                if kind == "value_array" and condition["operator"] in {"array_some_equals", "array_some_in"}:
                    pass
                elif condition["operator"] not in {"equals", "is_null"}:
                    raise TableError(f"奇门解释条件无效: {assertion_id} {condition}")
            if condition["operator"] != "is_null" and not condition["expected"]:
                raise TableError(f"奇门解释条件缺 expected: {assertion_id} {condition}")
            group_conditions = groups[(assertion_id, int(group_id))]
            duplicate_key = (condition["field"], condition["item_key"])
            if any((item["field"], item["item_key"]) == duplicate_key for item in group_conditions):
                raise TableError(
                    f"奇门解释条件组重复因子: {assertion_id}/{group_id}/{condition['field']}"
                )
            group_conditions.append(condition)

    for item in row_by_id.values():
        grouped = sorted(
            (
                (key, conditions)
                for key, conditions in groups.items()
                if key[0] == item["id"]
            ),
            key=lambda pair: pair[0][1],
        )
        item["conditions"] = [conditions for _key, conditions in grouped]
        if not item["conditions"]:
            raise TableError(f"奇门解释没有条件，会恒命中: {item['id']}")

    if set(rows) != set(rules):
        raise TableError("奇门解释规则与断言行不一致")
    _INTERPRETATION_INDEX = dict(rows)
    return _INTERPRETATION_INDEX
