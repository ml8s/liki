"""奇门解释层：稳定快照 × 结构化条件表 → 结构化候选；不做最终裁决。"""
from __future__ import annotations

import json

from qimen_factors import load_snapshot_contract, validate_qimen_snapshot
from qimen_interpretations import assert_rule_compatibility, load_interpretation_index


def query(rule: str, snapshot: dict) -> dict:
    """返回某个奇门占法的命中断语候选；未命中返回空列表。"""
    validate_qimen_snapshot(snapshot)
    snapshot_contract = load_snapshot_contract()
    index = load_interpretation_index()
    if rule not in index:
        raise ValueError(f"unknown qimen interpretation rule: {rule}")
    assert_rule_compatibility(rule, snapshot["scope"], snapshot["school"])

    matched = []
    for row in index[rule]:
        evidence = match_groups(row["conditions"], snapshot, snapshot_contract)
        if evidence is None:
            continue
        matched.append({
            "id": row["id"],
            "name": row["name"],
            "conclusion": row["conclusion"],
            "basis": row["basis"],
            "evidence": evidence,
        })
    matched.sort(key=lambda item: item["id"])
    return {"rule": rule, "assertions": matched}


def match_groups(
    groups: list[list[dict]], snapshot: dict, snapshot_contract: dict
) -> list[dict] | None:
    object_array_fields = {
        name for name, spec in snapshot_contract["fields"].items()
        if spec["kind"] == "object_array"
    }
    for conditions in groups:
        if evidence := match_condition_group(
            conditions, snapshot, snapshot_contract, object_array_fields
        ):
            return evidence
    return None


def match_condition_group(
    conditions: list[dict],
    snapshot: dict,
    snapshot_contract: dict,
    object_array_fields: set[str],
) -> list[dict] | None:
    array_conditions: dict[str, list[dict]] = {}
    for condition in conditions:
        if condition["field"] in object_array_fields:
            array_conditions.setdefault(condition["field"], []).append(condition)

    matched_items: dict[str, dict] = {}
    for field, field_conditions in array_conditions.items():
        matched_item = None
        for item in snapshot[field]:
            if all(
                condition_matches(
                    item.get(condition["item_key"]),
                    "equals" if condition["operator"] == "array_some_equals" else condition["operator"],
                    condition["expected"],
                )
                for condition in field_conditions
            ):
                matched_item = item
                break
        if matched_item is None:
            return None
        matched_items[field] = matched_item

    evidence = []
    for condition in conditions:
        if condition["field"] in matched_items:
            actual = matched_items[condition["field"]].get(condition["item_key"])
            operator = (
                "equals"
                if condition["operator"] == "array_some_equals"
                else condition["operator"]
            )
        else:
            actual = condition_value(snapshot, condition, snapshot_contract)
            operator = condition["operator"]
        if not condition_matches(actual, operator, condition["expected"]):
            return None
        evidence.append({
            "field": condition["field"],
            "item_key": condition.get("item_key"),
            "operator": condition["operator"],
            "expected": condition["expected"] or None,
            "actual": normalize_json(actual),
        })
        if condition["field"] in matched_items:
            evidence[-1]["matched"] = normalize_json(
                matched_items[condition["field"]]
            )
    return evidence


def condition_value(snapshot: dict, condition: dict, field_spec: dict):
    actual = snapshot[condition["field"]]
    field_spec = field_spec["fields"][condition["field"]]
    kind = field_spec["kind"]
    if kind == "object":
        return actual.get(condition["item_key"])
    if kind == "object_array":
        return [item.get(condition["item_key"]) for item in actual]
    return actual


def condition_matches(actual, operator: str, expected: str) -> bool:
    if operator == "is_null":
        return actual is None
    if operator == "equals":
        if isinstance(actual, bool):
            return expected.lower() == str(actual).lower()
        return actual == expected
    if operator == "array_some_equals":
        if any(isinstance(item, bool) for item in actual):
            return any(
                isinstance(item, bool) and expected.lower() == str(item).lower()
                for item in actual
            )
        return expected in actual
    if operator == "array_some_in":
        allowed = expected.split("|")
        if isinstance(actual, list):
            return any(item in allowed for item in actual)
        return actual in allowed
    return False


def normalize_json(value):
    """为证据输出保留 JSON 可序列化的标量与容器。"""
    return json.loads(json.dumps(value, ensure_ascii=False, sort_keys=True))
