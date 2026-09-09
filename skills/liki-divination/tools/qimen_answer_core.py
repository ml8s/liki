"""奇门 answer core：从 snapshot 投影证据并校验引用与禁语。"""
from __future__ import annotations

import hashlib
import json


FORBIDDEN = ["必然", "百分百", "100%", "保证", "一定会"]
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


def _digest(value: dict) -> str:
    raw = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()[:16]


def candidate_id(candidate: dict) -> str:
    raw = {
        "type": candidate.get("type"),
        "branch": candidate.get("branch"),
        "gong": candidate.get("gong"),
    }
    return f"timing-{_digest(raw)}"


def available_ids(snapshot: dict) -> tuple[set[str], set[str]]:
    factors = snapshot.get("factors", {})
    evidence_ids = {f"snapshot:{key}" for key in factors}
    assertion_ids = {
        f"assertion:{item['id']}"
        for item in (snapshot.get("special") or {}).get("assertions", [])
        if item.get("id")
    }
    evidence_ids |= assertion_ids
    timing_ids = {candidate_id(candidate) for candidate in factors.get("ying_qi", [])}
    return evidence_ids, timing_ids


def build(snapshot: dict) -> dict:
    if not isinstance(snapshot, dict) or snapshot.get("schema_version") != "qimen-snapshot-v3":
        raise ValueError("snapshot must be qimen-snapshot-v3")
    evidence_ids, timing_ids = available_ids(snapshot)
    method = snapshot.get("method_context", {})
    matter = snapshot.get("matter") or {}
    return {
        "headline": "",
        "verdict": "",
        "confidence": "medium",
        "question_ref": snapshot.get("question", ""),
        "method_ref": {
            "scope": method.get("scope"),
            "school": method.get("school"),
            "matter": matter.get("matter"),
        },
        "evidence_refs": sorted(ref for ref in evidence_ids if ref.startswith("snapshot:")),
        "assertion_refs": sorted(ref for ref in evidence_ids if ref.startswith("assertion:")),
        "timing_refs": sorted(timing_ids),
        "conflicts": [],
        "action": "",
        "boundary": "传统奇门视角；不构成医疗、法律、财务、人事或紧急决策建议。",
        "disclaimer": "结论为传统文化视角下的条件性倾向，不构成结果承诺。",
    }


def validate_core(core: dict, snapshot: dict) -> dict:
    errors: list[dict] = []
    if not isinstance(core, dict):
        raise ValueError("answer core must be an object")
    for field in core:
        if field not in ALLOWED_FIELDS:
            errors.append({"field": field, "reason": "unknown field"})

    for field in ("headline", "verdict", "action", "boundary", "disclaimer"):
        if not isinstance(core.get(field), str) or not core.get(field, "").strip():
            errors.append({"field": field, "reason": "required non-empty text"})
    if core.get("confidence") not in {"low", "medium", "high"}:
        errors.append({"field": "confidence", "reason": "must be low/medium/high"})

    evidence_ids, timing_ids = available_ids(snapshot)
    for field, allowed in (
        ("evidence_refs", evidence_ids),
        ("assertion_refs", evidence_ids),
        ("timing_refs", timing_ids),
    ):
        refs = core.get(field, [])
        if not isinstance(refs, list):
            errors.append({"field": field, "reason": "must be array"})
            continue
        for ref in refs:
            if ref not in allowed:
                errors.append({"field": field, "reason": "unknown ref", "ref": ref})

    if not core.get("evidence_refs"):
        errors.append({"field": "evidence_refs", "reason": "at least one snapshot fact is required"})
    method_ref = core.get("method_ref")
    method = snapshot.get("method_context", {})
    if (
        not isinstance(method_ref, dict)
        or method_ref.get("scope") != method.get("scope")
        or method_ref.get("school") != method.get("school")
    ):
        errors.append({"field": "method_ref", "reason": "must match snapshot scope/school"})

    for field in ("headline", "verdict", "action"):
        text = core.get(field, "")
        if not isinstance(text, str):
            continue
        for word in FORBIDDEN:
            if word.lower() in text.lower():
                errors.append({"field": "forbidden_language", "reason": f"contains {word}"})

    return {
        "schema_version": "qimen-answer-core-audit-v1",
        "accepted": not errors,
        "errors": errors,
        "checked": ["fields", "method ref", "evidence refs", "timing refs", "forbidden language"],
        "unchecked": ["格局主次", "吉凶结论", "现实建议", "应期是否必然发生"],
    }
