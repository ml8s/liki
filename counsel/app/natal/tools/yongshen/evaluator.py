"""表驱动用神裁决解释器。

解释器只认识通用表达式：path / eq / in / all / any / select / lookup。
从格、调候、格局、中和、喜忌冲突等命理规则全部在 policy.json。
"""

import hashlib
import json
from pathlib import Path
from typing import Any

from .schema import validate_policy

POLICY_PATH = Path(__file__).resolve().parent / "policy.json"
_MISSING = object()
_POLICY_CACHE: dict | None = None
_DIRECT_FACTOR_KEYS = (
    "month_zhi", "day_master_stem", "day_master_element", "stems",
    "fu_yi", "tiao_hou", "ge_ju", "atomic_facts",
    "element_states", "ten_god_states", "wuxing_count",
)


class YongShenEvaluationError(ValueError):
    """policy 表达式无法求值。"""


def load_policy() -> dict:
    global _POLICY_CACHE
    if _POLICY_CACHE is None:
        policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))
        validate_policy(policy)
        _POLICY_CACHE = policy
    return _POLICY_CACHE


def _get_path(path: str, facts: dict, derived: dict, root: Any = None) -> Any:
    """读取 dotted path。root 非空时用于 select 的 item 作用域。"""
    if not isinstance(path, str) or not path:
        return _MISSING
    if root is not None:
        if path.startswith("global."):
            return _get_path(path[len("global."):], facts, derived)
        if path.startswith("item."):
            return _walk(root, path[len("item."):])
        return _walk(root, path)
    scope = {**facts, **derived}
    return _walk(scope, path)


def _walk(value: Any, path: str) -> Any:
    current = value
    for part in path.split("."):
        if current is _MISSING:
            return _MISSING
        if isinstance(current, dict):
            current = current.get(part, _MISSING)
        elif isinstance(current, list) and part.isdigit():
            index = int(part)
            current = current[index] if 0 <= index < len(current) else _MISSING
        else:
            return _MISSING
    return current


def _is_empty(value: Any) -> bool:
    return value is _MISSING or value is None or value == "" or value == [] or value == {} or value in (NO_CANDIDATE, NO_SECONDARY)


def _len(value: Any) -> int:
    if value is _MISSING or value is None:
        return 0
    try:
        return len(value)
    except TypeError:
        return 0


def _resolve_value(spec: Any, facts: dict, derived: dict, root: Any = None) -> Any:
    if not isinstance(spec, dict):
        return spec
    if "$path" in spec:
        return _get_path(spec["$path"], facts, derived, root)
    if "$table" in spec:
        policy = load_policy()
        return policy["tables"].get(spec["$table"], spec["$table"])
    if "$if" in spec:
        branch = spec["$if"]
        condition = branch.get("when")
        if _eval_condition(condition, facts, derived, root):
            return _resolve_value(branch.get("then"), facts, derived, root)
        return _resolve_value(branch.get("else"), facts, derived, root)
    if "$lookup" in spec:
        lookup = spec["$lookup"]
        table = load_policy()["tables"].get(lookup.get("table"), {})
        key = _resolve_value(lookup.get("key"), facts, derived, root)
        if key is _MISSING or key is None:
            return _MISSING
        row = table.get(str(key)) if isinstance(table, dict) else None
        if row is None:
            return _MISSING
        field = lookup.get("field")
        if field is None:
            return row
        if not isinstance(row, dict):
            return _MISSING
        return row.get(field, _MISSING)
    if "$as_list" in spec:
        value = _resolve_value(spec["$as_list"], facts, derived, root)
        if _is_empty(value):
            return []
        return value if isinstance(value, list) else [value]
    if "$list" in spec:
        return [
            value for value in (
                _resolve_value(item, facts, derived, root)
                for item in spec["$list"]
            )
            if not _is_empty(value)
        ]
    if "$union" in spec:
        result: list[Any] = []
        for item in spec["$union"]:
            value = _resolve_value(item, facts, derived, root)
            values = value if isinstance(value, list) else ([] if _is_empty(value) else [value])
            for value_item in values:
                if not _is_empty(value_item) and value_item not in result:
                    result.append(value_item)
        return result
    if "$eq_paths" in spec:
        left, right = spec["$eq_paths"]
        return _resolve_value(left, facts, derived, root) == _resolve_value(right, facts, derived, root)
    if "$select" in spec:
        return _select(spec["$select"], facts, derived, root)
    if "$select_count" in spec:
        return _select(spec["$select_count"], facts, derived, root)["count"]
    if "$select_exists" in spec:
        return _select(spec["$select_exists"], facts, derived, root)["count"] > 0
    raise YongShenEvaluationError(f"未知 value 表达式: {sorted(spec)}")


def _select(spec: dict, facts: dict, derived: dict, root: Any = None) -> dict:
    source = _get_path(spec.get("from", ""), facts, derived, root)
    if not isinstance(source, list):
        source = []
    where = spec.get("where")
    items = [
        item for item in source
        if where is None or _eval_condition(where, facts, derived, item)
    ]
    order_by = spec.get("order_by")
    if order_by:
        direction = spec.get("direction", "asc")
        reverse = direction == "desc"
        items = sorted(
            items,
            key=lambda item: (
                _get_path(order_by, facts, derived, item) is _MISSING,
                _get_path(order_by, facts, derived, item),
            ),
            reverse=reverse,
        )
    field = spec.get("field")
    if field:
        values = [
            _get_path(field, facts, derived, item)
            for item in items
        ]
        values = [value for value in values if value is not _MISSING]
    else:
        values = items
    return {
        "count": len(values),
        "values": values,
        "value": values[0] if len(values) == 1 else None,
        "unique": len(values) == 1,
    }


def _eval_condition(condition: Any, facts: dict, derived: dict, root: Any = None) -> bool:
    if condition is None:
        return True
    if not isinstance(condition, dict):
        raise YongShenEvaluationError(f"condition 必须是 object: {condition!r}")
    if "const" in condition:
        return bool(condition["const"])
    if "$eq_paths" in condition:
        return _resolve_value({"$eq_paths": condition["$eq_paths"]}, facts, derived, root)
    if "all" in condition:
        return all(_eval_condition(item, facts, derived, root) for item in condition["all"])
    if "any" in condition:
        return any(_eval_condition(item, facts, derived, root) for item in condition["any"])
    if "not" in condition:
        return not _eval_condition(condition["not"], facts, derived, root)

    actual = _get_path(condition.get("path", ""), facts, derived, root)
    op = condition.get("op")
    expected = _resolve_value(condition.get("value"), facts, derived, root)
    if op == "exists":
        return not _is_empty(actual)
    if op == "empty":
        return _is_empty(actual)
    if actual is _MISSING:
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
        if isinstance(actual, (list, tuple, set, dict, str)):
            return expected in actual
        return False
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
        return _len(actual) == expected
    if op == "length_gte":
        return _len(actual) >= expected
    if op == "length_gt":
        return _len(actual) > expected
    raise YongShenEvaluationError(f"未知 condition op: {op}")


def _materialize(value: Any) -> Any:
    """把求值哨兵转成可序列化空值，避免内部对象进入输出。"""
    if value is _MISSING:
        return ""
    if isinstance(value, dict):
        return {key: _materialize(item) for key, item in value.items()}
    if isinstance(value, list):
        return [_materialize(item) for item in value]
    return value


NO_CANDIDATE = "无候选"
NO_SECONDARY = "无辅调候"


def _candidate(value: Any) -> Any:
    """Treat a missing/legacy-empty candidate as an explicit closure value."""
    value = _materialize(value)
    return NO_CANDIDATE if value in ("", None) else value


def _set_path(target: dict, key: str, value: Any) -> None:
    parts = key.split(".")
    current = target
    for part in parts[:-1]:
        child = current.get(part)
        if not isinstance(child, dict):
            child = {}
            current[part] = child
        current = child
    materialized = _materialize(value)
    if key == "selected.yong" and materialized == "":
        materialized = None
    current[parts[-1]] = materialized


def _enrich_facts(fullchart: dict, policy: dict) -> dict:
    """把 engine 事实合并成解释器可读的纯数据视图（不做命理判断）。"""
    fu_yi = fullchart.get("fu_yi") or {}
    tiao_hou = fullchart.get("tiao_hou") or {}
    ge_ju = fullchart.get("ge_ju") or {}
    atomic = fullchart.get("atomic_facts") or {}
    counts = fu_yi.get("wuxing_count") or {}
    relations = policy["tables"].get("wuxing_relations") or {}
    day_master_element = atomic.get("day_master_element", "")
    if not day_master_element:
        day_master_element = (
            ((fullchart.get("ri") or {}).get("gan_wuxing") or "")
        )

    elements: list[dict] = []
    for item in fullchart.get("element_states") or []:
        if not isinstance(item, dict):
            continue
        wuxing = item.get("wuxing", "")
        relation = relations.get(wuxing) or {}
        elements.append({
            **item,
            "count": counts.get(wuxing, 0),
            "is_day_master": wuxing == day_master_element,
            **relation,
        })

    stems = [
        (fullchart.get(pillar) or {}).get("gan", "")
        for pillar in ("nian", "yue", "ri", "shi")
    ]
    stems = [stem for stem in stems if stem]
    month_zhi = (fullchart.get("yue") or {}).get("zhi", "")
    day_master_stem = atomic.get("day_master_stem", "") or (fullchart.get("ri") or {}).get("gan", "")

    return {
        "fu_yi": fu_yi,
        "tiao_hou": tiao_hou,
        "ge_ju": ge_ju,
        "atomic_facts": atomic,
        "element_states": elements,
        "ten_god_states": [
            item for item in (fullchart.get("ten_god_states") or [])
            if isinstance(item, dict)
        ],
        "wuxing_count": counts,
        "month_zhi": month_zhi,
        "day_master_stem": day_master_stem,
        "day_master_element": day_master_element,
        "stems": stems,
    }


def _apply_derived(policy: dict, facts: dict) -> dict:
    derived = {**policy.get("defaults", {})}
    rules = policy.get("pre_factors", [])
    for _ in range(len(rules) + 2):
        before = json.dumps(derived, ensure_ascii=False, sort_keys=True, default=str)
        for rule in rules:
            if _eval_condition(rule.get("when"), facts, derived):
                for key, spec in rule.get("set", {}).items():
                    _set_path(derived, key, _resolve_value(spec, facts, derived))
        after = json.dumps(derived, ensure_ascii=False, sort_keys=True, default=str)
        if after == before:
            break
    return derived


def _select_rule(policy: dict, facts: dict, derived: dict) -> dict:
    for rule in policy.get("rules", []):
        if _eval_condition(rule.get("when"), facts, derived):
            return rule
    raise YongShenEvaluationError("policy.rules 未给出兜底规则")


def _apply_then(rule: dict, facts: dict, derived: dict) -> dict:
    result: dict = {
        "selected": {
            "yong": NO_CANDIDATE,
            "xi": [],
            "ji": [],
            "source": "",
            "confidence": "low",
            "known_fields": [],
        },
        "source": "",
        "confidence": "low",
        "known_fields": [],
        "status": "unresolved",
        "conflicts": [],
    }
    for key, spec in rule.get("then", {}).items():
        _set_path(result, key, _resolve_value(spec, facts, derived))
    result["rule_id"] = rule["id"]
    return result


def _dedupe(values: Any) -> list:
    if _is_empty(values):
        return []
    source = values if isinstance(values, list) else [values]
    result: list = []
    for value in source:
        if not _is_empty(value) and value not in result:
            result.append(value)
    return result


def _finalize(result: dict, policy: dict) -> None:
    selected = result["selected"]
    selected["yong"] = (
        NO_CANDIDATE
        if selected.get("yong") in (None, _MISSING, "")
        else str(selected["yong"])
    )
    selected["xi"] = _dedupe(selected.get("xi"))
    selected["ji"] = _dedupe(selected.get("ji"))
    conflicts = list(result.get("conflicts") or [])
    status = result.get("status", "unresolved")

    disjoint = (policy.get("finalize") or {}).get("disjoint") or {}
    yong = selected.get("yong")
    if yong != NO_CANDIDATE and yong in selected["xi"]:
        selected["xi"] = [item for item in selected["xi"] if item != yong]
    if yong != NO_CANDIDATE and yong in selected["ji"]:
        selected["ji"] = [item for item in selected["ji"] if item != yong]
        spec = disjoint.get("yong_vs_xi_ji", {}).get("on_overlap", {})
        status = spec.get("status", status)
        if conflict := spec.get("conflict"):
            conflicts.append(dict(conflict))

    overlap = [item for item in selected["xi"] if item in selected["ji"]]
    if overlap:
        spec = disjoint.get("xi_vs_ji", {}).get("on_overlap", {})
        remove_from = set(spec.get("remove_from") or [])
        if "xi" in remove_from:
            selected["xi"] = [item for item in selected["xi"] if item not in overlap]
        if "ji" in remove_from:
            selected["ji"] = [item for item in selected["ji"] if item not in overlap]
        status = spec.get("status", status)
        if conflict := spec.get("conflict"):
            conflicts.append(dict(conflict))

    known_fields = [
        field for field in ("yong", "xi", "ji")
        if not _is_empty(selected.get(field))
    ]
    if status == "resolved" and len(known_fields) < 3:
        status = "partial"
    if status in ("resolved", "partial") and selected.get("yong") == NO_CANDIDATE:
        status = "needs_review"

    selected["source"] = result.pop("source", "")
    selected["confidence"] = result.pop("confidence", "low")
    selected["known_fields"] = known_fields
    result.pop("known_fields", None)
    result["status"] = status
    # 去重 conflict code，保留首次出现。
    deduped: list[dict] = []
    seen: set[str] = set()
    for conflict in conflicts:
        code = conflict.get("code", "")
        if code in seen:
            continue
        seen.add(code)
        deduped.append(conflict)
    result["conflicts"] = deduped


def _alternatives(facts: dict, derived: dict, policy: dict) -> dict:
    fu_yi = facts.get("fu_yi") or {}
    tiao_hou = facts.get("tiao_hou") or {}
    ge_ju = facts.get("ge_ju") or {}
    basis = fu_yi.get("basis") or {}
    themes = policy["tables"].get("ge_ju_themes") or {}
    raw_ten_god = ge_ju.get("pattern_god_ten_god") or ""
    ten_god = raw_ten_god or "杂格"
    return {
        "fu_yi": {
            "qiangruo": fu_yi.get("qiangruo", ""),
            "yong": _candidate(fu_yi.get("yong")),
            "xi": _candidate(fu_yi.get("xi")),
            "ji": _candidate(fu_yi.get("ji")),
            "pattern": fu_yi.get("pattern", ""),
            "root_type": basis.get("root_type", ""),
            "season": basis.get("season", ""),
            "yin_bi_count": basis.get("yin_bi_count", 0),
        },
        "tiao_hou": {
            "primary_wuxing": tiao_hou.get("primary_wuxing", ""),
            "secondary_wuxing": _candidate(tiao_hou.get("secondary_wuxing")),
            "secondary_condition": tiao_hou.get("secondary_condition", ""),
            "condition_ok": derived.get("tiaohou_secondary_condition_ok", False),
            "secondary_usable": derived.get("tiaohou_secondary_usable", False),
            "primary_present": derived.get("tiaohou_primary_present", False),
            "primary_visible": derived.get("tiaohou_primary_visible", False),
            "secondary_present": derived.get("tiaohou_secondary_present", False),
            "secondary_visible": derived.get("tiaohou_secondary_visible", False),
            "yong": _candidate(derived.get("tiaohou_yong")),
            "yong_source": derived.get("tiaohou_yong_source", "none"),
            "yong_present": derived.get("tiaohou_yong_present", False),
            "yong_visible": derived.get("tiaohou_yong_visible", False),
            "control_wuxing": _candidate(derived.get("tiaohou_control_wuxing")),
        },
        "ge_ju": {
            "ge_ju": ge_ju.get("ge_ju", ""),
            "pattern_god_ten_god": raw_ten_god or "无格神",
            "pattern_god_source": ge_ju.get("pattern_god_source", ""),
            "yong_fa": ge_ju.get("yong_fa", ""),
            "clear": derived.get("ge_ju_clear", False),
            "theme": themes.get(ten_god, themes.get("杂格", "")),
        },
        "cong_ge": {
            "pattern": fu_yi.get("pattern", ""),
            "verified": derived.get("congge_verified", False),
            "fake": derived.get("congge_fake", False),
            "guan_sha_effective_count": derived.get("guan_sha_effective_count", 0),
            "cai_effective_count": derived.get("cai_effective_count", 0),
            "shi_shang_effective_count": derived.get("shi_shang_effective_count", 0),
            "yin_rooted_count": derived.get("yin_rooted_count", 0),
            "bi_jie_rooted_count": derived.get("bi_jie_rooted_count", 0),
        },
        "neutral": {
            "overstrong": derived.get("neutral_overstrong", {}),
            "weak": derived.get("neutral_weak", {}),
            "yong": _candidate(derived.get("neutral_yong")),
            "status": derived.get("neutral_status", "unresolved"),
        },
    }


def _evidence(facts: dict, derived: dict, policy: dict) -> list[dict]:
    return [
        {
            "source": "facts",
            "month_zhi": facts.get("month_zhi", ""),
            "day_master_stem": facts.get("day_master_stem", ""),
            "day_master_element": facts.get("day_master_element", ""),
            "season_is_extreme": derived.get("season_is_extreme", False),
        },
        {
            "source": "fu_yi",
            "qiangruo": (facts.get("fu_yi") or {}).get("qiangruo", ""),
            "pattern": (facts.get("fu_yi") or {}).get("pattern", ""),
            "root_type": ((facts.get("fu_yi") or {}).get("basis") or {}).get("root_type", ""),
            "season": ((facts.get("fu_yi") or {}).get("basis") or {}).get("season", ""),
            "yin_bi_count": ((facts.get("fu_yi") or {}).get("basis") or {}).get("yin_bi_count", 0),
            "yong": _candidate((facts.get("fu_yi") or {}).get("yong")),
            "xi": _candidate((facts.get("fu_yi") or {}).get("xi")),
            "ji": _candidate((facts.get("fu_yi") or {}).get("ji")),
        },
        {
            "source": "tiao_hou",
            "primary_wuxing": (facts.get("tiao_hou") or {}).get("primary_wuxing", ""),
            "secondary_wuxing": _candidate((facts.get("tiao_hou") or {}).get("secondary_wuxing")),
            "secondary_condition": (facts.get("tiao_hou") or {}).get("secondary_condition", ""),
            "condition_ok": derived.get("tiaohou_secondary_condition_ok", False),
            "secondary_usable": derived.get("tiaohou_secondary_usable", False),
            "primary_present": derived.get("tiaohou_primary_present", False),
            "primary_visible": derived.get("tiaohou_primary_visible", False),
            "secondary_present": derived.get("tiaohou_secondary_present", False),
            "secondary_visible": derived.get("tiaohou_secondary_visible", False),
            "yong": _candidate(derived.get("tiaohou_yong")),
            "yong_source": derived.get("tiaohou_yong_source", "none"),
            "yong_present": derived.get("tiaohou_yong_present", False),
            "yong_visible": derived.get("tiaohou_yong_visible", False),
            "control_wuxing": _candidate(derived.get("tiaohou_control_wuxing")),
        },
        {
            "source": "ge_ju",
            "pattern": (facts.get("ge_ju") or {}).get("ge_ju", ""),
            "pattern_god_ten_god": (facts.get("ge_ju") or {}).get("pattern_god_ten_god") or "无格神",
            "pattern_god_source": (facts.get("ge_ju") or {}).get("pattern_god_source", ""),
            "clear": derived.get("ge_ju_clear", False),
        },
        {
            "source": "cong_ge",
            "pattern": (facts.get("fu_yi") or {}).get("pattern", ""),
            "verified": derived.get("congge_verified", False),
            "fake": derived.get("congge_fake", False),
            "guan_sha_effective_count": derived.get("guan_sha_effective_count", 0),
            "cai_effective_count": derived.get("cai_effective_count", 0),
            "shi_shang_effective_count": derived.get("shi_shang_effective_count", 0),
            "yin_rooted_count": derived.get("yin_rooted_count", 0),
            "bi_jie_rooted_count": derived.get("bi_jie_rooted_count", 0),
        },
    ]


def _digest(value: Any) -> str:
    raw = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"), default=str)
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _pre_factor_snapshot(facts: dict, derived: dict) -> dict:
    """按阶段暴露取用神所需因子：engine 直读 + policy 二次因子。"""
    return {
        "phase": "pre_yongshen",
        "direct": {key: facts[key] for key in _DIRECT_FACTOR_KEYS if key in facts},
        "secondary": dict(derived),
    }


PUBLISHED_FIELDS = (
    "selected",
    "status",
    "conflicts",
    "rule_id",
    "ge_ju_context",
    "evidence",
    "policy_id",
    "chart_digest",
)


def published_yongshen(result: dict) -> dict:
    """Project the internal resolver result to the MCP decision object.

    The published digest binds the decision, its evidence, and the source chart.
    Internal pre-factor snapshots and alternatives remain outside this wire object.
    """
    published = {field: result[field] for field in PUBLISHED_FIELDS}
    if "yongshen_digest" in result:
        published["yongshen_digest"] = result["yongshen_digest"]
    return published


def resolve_yongshen(fullchart: dict) -> dict:
    """从 bazi.fullchart 事实裁决唯一 selected_yongshen。"""
    policy = load_policy()
    facts = _enrich_facts(fullchart or {}, policy)
    derived = _apply_derived(policy, facts)
    rule = _select_rule(policy, facts, derived)
    result = _apply_then(rule, facts, derived)
    _finalize(result, policy)
    result["factors"] = _pre_factor_snapshot(facts, derived)
    result["alternatives"] = _alternatives(facts, derived, policy)
    result["ge_ju_context"] = result["alternatives"]["ge_ju"]
    result["evidence"] = _evidence(facts, derived, policy)
    result["policy_id"] = policy["policy_id"]
    result["chart_digest"] = _digest(fullchart or {})
    result["yongshen_digest"] = _digest(published_yongshen(result))
    return result


__all__ = [
    "YongShenEvaluationError",
    "load_policy",
    "published_yongshen",
    "resolve_yongshen",
]
