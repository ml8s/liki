"""Rule selection and final normalization for a yongshen decision."""

from typing import Any

from .context import (
    MISSING,
    NO_CANDIDATE,
    YongShenEvaluationError,
    is_empty,
    set_path,
)
from .scoring import eval_condition, resolve_value


def select_rule(policy: dict, facts: dict, derived: dict) -> dict:
    for rule in policy.get("rules", []):
        if eval_condition(rule.get("when"), facts, derived):
            return rule
    raise YongShenEvaluationError("policy.rules 未给出兜底规则")


def apply_then(rule: dict, facts: dict, derived: dict) -> dict:
    result: dict = {
        "selected": {
            "yong": NO_CANDIDATE,
            "xi": [],
            "ji": [],
        },
        "status": "unresolved",
        "conflicts": [],
    }
    for key, spec in rule.get("then", {}).items():
        set_path(result, key, resolve_value(spec, facts, derived))
    result["rule_id"] = rule["id"]
    return result


def _dedupe(values: Any) -> list:
    if is_empty(values):
        return []
    source = values if isinstance(values, list) else [values]
    result: list = []
    for value in source:
        if not is_empty(value) and value not in result:
            result.append(value)
    return result


def finalize(result: dict, policy: dict) -> None:
    selected = result["selected"]
    selected["yong"] = NO_CANDIDATE if selected.get("yong") in (None, MISSING, "") else str(selected["yong"])
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

    selected.pop("source", None)
    resolved_field_count = sum(1 for field in ("yong", "xi", "ji") if not is_empty(selected.get(field)))
    if status == "resolved" and resolved_field_count < 3:
        status = "partial"
    if status in ("resolved", "partial") and selected.get("yong") == NO_CANDIDATE:
        status = "needs_review"

    result["status"] = status
    result["conflicts"] = _dedupe_conflicts(conflicts)


def _dedupe_conflicts(conflicts: list[dict]) -> list[dict]:
    deduped: list[dict] = []
    seen: set[str] = set()
    for conflict in conflicts:
        code = conflict.get("code", "")
        if code in seen:
            continue
        seen.add(code)
        deduped.append(conflict)
    return deduped
