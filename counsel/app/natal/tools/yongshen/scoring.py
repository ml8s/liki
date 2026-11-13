"""Expression scoring for the table-driven yongshen policy."""

import json
from typing import Any

from .context import (
    MISSING,
    YongShenEvaluationError,
    get_path,
    is_empty,
    length,
    load_policy,
    set_path,
)


def resolve_value(spec: Any, facts: dict, derived: dict, root: Any = None) -> Any:
    if not isinstance(spec, dict):
        return spec
    if "$path" in spec:
        return get_path(spec["$path"], facts, derived, root)
    if "$table" in spec:
        policy = load_policy()
        return policy["tables"].get(spec["$table"], spec["$table"])
    if "$if" in spec:
        branch = spec["$if"]
        if eval_condition(branch.get("when"), facts, derived, root):
            return resolve_value(branch.get("then"), facts, derived, root)
        return resolve_value(branch.get("else"), facts, derived, root)
    if "$lookup" in spec:
        return _lookup(spec["$lookup"], facts, derived, root)
    if "$as_list" in spec:
        value = resolve_value(spec["$as_list"], facts, derived, root)
        if is_empty(value):
            return []
        return value if isinstance(value, list) else [value]
    if "$list" in spec:
        return [
            value
            for value in (resolve_value(item, facts, derived, root) for item in spec["$list"])
            if not is_empty(value)
        ]
    if "$union" in spec:
        return _union(spec["$union"], facts, derived, root)
    if "$eq_paths" in spec:
        left, right = spec["$eq_paths"]
        return resolve_value(left, facts, derived, root) == resolve_value(right, facts, derived, root)
    if "$select" in spec:
        return select(spec["$select"], facts, derived, root)
    if "$select_count" in spec:
        return select(spec["$select_count"], facts, derived, root)["count"]
    if "$select_exists" in spec:
        return select(spec["$select_exists"], facts, derived, root)["count"] > 0
    raise YongShenEvaluationError(f"未知 value 表达式: {sorted(spec)}")


def _lookup(spec: dict, facts: dict, derived: dict, root: Any) -> Any:
    table = load_policy()["tables"].get(spec.get("table"), {})
    key = resolve_value(spec.get("key"), facts, derived, root)
    if key is MISSING or key is None:
        return MISSING
    row = table.get(str(key)) if isinstance(table, dict) else None
    if row is None:
        return MISSING
    field = spec.get("field")
    if field is None:
        return row
    if not isinstance(row, dict):
        return MISSING
    return row.get(field, MISSING)


def _union(specs: list, facts: dict, derived: dict, root: Any) -> list:
    result: list[Any] = []
    for item in specs:
        value = resolve_value(item, facts, derived, root)
        values = value if isinstance(value, list) else ([] if is_empty(value) else [value])
        for value_item in values:
            if not is_empty(value_item) and value_item not in result:
                result.append(value_item)
    return result


def select(spec: dict, facts: dict, derived: dict, root: Any = None) -> dict:
    source = get_path(spec.get("from", ""), facts, derived, root)
    if not isinstance(source, list):
        source = []
    where = spec.get("where")
    items = [item for item in source if where is None or eval_condition(where, facts, derived, item)]
    order_by = spec.get("order_by")
    if order_by:
        direction = spec.get("direction", "asc")
        items = sorted(
            items,
            key=lambda item: (
                get_path(order_by, facts, derived, item) is MISSING,
                get_path(order_by, facts, derived, item),
            ),
            reverse=direction == "desc",
        )
    field = spec.get("field")
    if field:
        values = [get_path(field, facts, derived, item) for item in items]
        values = [value for value in values if value is not MISSING]
    else:
        values = items
    return {
        "count": len(values),
        "values": values,
        "value": values[0] if len(values) == 1 else None,
        "unique": len(values) == 1,
    }


def eval_condition(condition: Any, facts: dict, derived: dict, root: Any = None) -> bool:
    if condition is None:
        return True
    if not isinstance(condition, dict):
        raise YongShenEvaluationError(f"condition 必须是 object: {condition!r}")
    if "const" in condition:
        return bool(condition["const"])
    if "$eq_paths" in condition:
        return resolve_value({"$eq_paths": condition["$eq_paths"]}, facts, derived, root)
    if "all" in condition:
        return all(eval_condition(item, facts, derived, root) for item in condition["all"])
    if "any" in condition:
        return any(eval_condition(item, facts, derived, root) for item in condition["any"])
    if "not" in condition:
        return not eval_condition(condition["not"], facts, derived, root)

    actual = get_path(condition.get("path", ""), facts, derived, root)
    op = condition.get("op")
    expected = resolve_value(condition.get("value"), facts, derived, root)
    return _compare(actual, op, expected)


def _compare(actual: Any, op: str | None, expected: Any) -> bool:
    if op == "exists":
        return not is_empty(actual)
    if op == "empty":
        return is_empty(actual)
    if actual is MISSING:
        return False
    if op == "eq":
        return actual == expected
    if op == "ne":
        return actual != expected
    if op == "in":
        return isinstance(expected, (list, tuple, set)) and actual in expected
    if op == "not_in":
        return isinstance(expected, (list, tuple, set)) and actual not in expected
    if op == "contains":
        return isinstance(actual, (list, tuple, set, dict, str)) and expected in actual
    if op == "not_contains":
        if isinstance(actual, (list, tuple, set, dict, str)):
            return expected not in actual
        return True
    if op == "starts_with":
        return isinstance(actual, str) and actual.startswith(str(expected))
    if op == "ends_with":
        return isinstance(actual, str) and actual.endswith(str(expected))
    if op == "gte":
        return actual >= expected
    if op == "lte":
        return actual <= expected
    if op == "gt":
        return actual > expected
    if op == "lt":
        return actual < expected
    if op == "length_eq":
        return length(actual) == expected
    if op == "length_gte":
        return length(actual) >= expected
    if op == "length_gt":
        return length(actual) > expected
    raise YongShenEvaluationError(f"未知 condition op: {op}")


def apply_derived(policy: dict, facts: dict) -> dict:
    derived = {**policy.get("defaults", {})}
    rules = policy.get("pre_factors", [])
    for _ in range(len(rules) + 2):
        before = json.dumps(derived, ensure_ascii=False, sort_keys=True, default=str)
        for rule in rules:
            if eval_condition(rule.get("when"), facts, derived):
                for key, spec in rule.get("set", {}).items():
                    set_path(derived, key, resolve_value(spec, facts, derived))
        after = json.dumps(derived, ensure_ascii=False, sort_keys=True, default=str)
        if after == before:
            break
    return derived
