"""金函玉镜 answer core；只解释星、门与十二神事实。"""
from __future__ import annotations

from divination_answer import validate_common_core

ALLOWED_FIELDS = {
    "headline",
    "verdict",
    "confidence",
    "question_ref",
    "method_ref",
    "evidence_refs",
    "assertion_refs",
    "timing_refs",
    "conflicts",
    "action",
    "boundary",
    "disclaimer",
}


def _validate_snapshot(snapshot: dict) -> None:
    factors = snapshot.get("factors")
    if not isinstance(factors, dict) or factors.get("kind") != "jinhan":
        raise ValueError("snapshot must contain jinhan factors")
    palaces = factors.get("palaces")
    day_spirits = factors.get("day_spirits")
    if not isinstance(palaces, list) or len(palaces) != 9:
        raise ValueError("jinhan factors require 9 palaces")
    if not isinstance(day_spirits, list) or len(day_spirits) != 12:
        raise ValueError("jinhan factors require 12 day spirits")


def available_ids(snapshot: dict) -> set[str]:
    _validate_snapshot(snapshot)
    return {
        "snapshot:method",
        "snapshot:palaces",
        "snapshot:day_spirits",
    }


def build(snapshot: dict) -> dict:
    method = snapshot.get("method_context") or {}
    palace_count = len((snapshot.get("factors") or {}).get("palaces", []))
    return {
        "headline": "",
        "verdict": "",
        "confidence": "medium",
        "question_ref": snapshot.get("question", ""),
        "method_ref": {
            "scope": method.get("scope"),
            "school": method.get("school"),
            "matter": None,
        },
        "evidence_refs": [
            "snapshot:method",
            "snapshot:palaces",
            "snapshot:day_spirits",
        ],
        "assertion_refs": [],
        "timing_refs": [],
        "conflicts": [],
        "action": "",
        "boundary": "金函玉镜日家视角；不构成医疗、法律、财务、人事或紧急决策建议。",
        "disclaimer": "结论为传统文化视角下的条件性倾向，不构成结果承诺。",
    }


def validate_core(core: dict, snapshot: dict) -> dict:
    errors = validate_common_core(
        core,
        allowed_fields=ALLOWED_FIELDS,
        required_text_fields=("headline", "verdict", "action"),
    )

    allowed_ids = available_ids(snapshot)
    for ref in core.get("evidence_refs", []):
        if ref not in allowed_ids:
            errors.append({"field": "evidence_refs", "reason": "unknown ref", "ref": ref})
    if not core.get("evidence_refs"):
        errors.append({"field": "evidence_refs", "reason": "at least one snapshot fact is required"})

    method_ref = core.get("method_ref")
    method = snapshot.get("method_context") or {}
    if (
        not isinstance(method_ref, dict)
        or method_ref.get("scope") != method.get("scope")
        or method_ref.get("school") != method.get("school")
    ):
        errors.append({"field": "method_ref", "reason": "must match snapshot scope/school"})

    return {
        "schema_version": "qimen-jinhan-answer-core-audit-v1",
        "accepted": not errors,
        "errors": errors,
        "checked": ["fields", "method ref", "evidence refs", "forbidden language"],
        "unchecked": ["星门组合主次", "吉凶结论", "现实建议", "十二神用途"],
    }
