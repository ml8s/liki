"""奇门结构化 answer 契约：校验证据引用和禁止性表述。"""
from __future__ import annotations

import hashlib
import json


SCHEMA_VERSION = "qimen-report-v1"
FORBIDDEN = ["必然", "百分百", "100%", "保证", "一定会"]
ALLOWED_FIELDS = {
    "schema_version",
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
    timing_ids = {
        candidate_id(candidate)
        for candidate in factors.get("ying_qi", [])
    }
    return evidence_ids, timing_ids


def template(snapshot: dict) -> dict:
    if not isinstance(snapshot, dict) or snapshot.get("schema_version") != "qimen-snapshot-v3":
        raise ValueError("snapshot must be qimen-snapshot-v3")
    evidence_ids, timing_ids = available_ids(snapshot)
    method = snapshot.get("method_context", {})
    matter = snapshot.get("matter") or {}
    return {
        "schema_version": SCHEMA_VERSION,
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
        "disclaimer": "结论为传统文化视角下的条件性倾向，不承诺现实结果。",
    }


def validate_report(report: dict, snapshot: dict) -> dict:
    errors: list[dict] = []
    if not isinstance(report, dict):
        raise ValueError("report must be an object")
    if report.get("schema_version") != SCHEMA_VERSION:
        errors.append({"field": "schema_version", "reason": "must be qimen-report-v1"})
    for field in report:
        if field not in ALLOWED_FIELDS:
            errors.append({"field": field, "reason": "unknown field"})
    for field in ("headline", "verdict", "action", "boundary", "disclaimer"):
        if not isinstance(report.get(field), str) or not report.get(field, "").strip():
            errors.append({"field": field, "reason": "required non-empty text"})
    if report.get("confidence") not in {"low", "medium", "high"}:
        errors.append({"field": "confidence", "reason": "must be low/medium/high"})

    evidence_ids, timing_ids = available_ids(snapshot)
    for field, allowed in (
        ("evidence_refs", evidence_ids),
        ("assertion_refs", evidence_ids),
        ("timing_refs", timing_ids),
    ):
        refs = report.get(field, [])
        if not isinstance(refs, list):
            errors.append({"field": field, "reason": "must be array"})
            continue
        for ref in refs:
            if ref not in allowed:
                errors.append({"field": field, "reason": "unknown ref", "ref": ref})

    if not report.get("evidence_refs"):
        errors.append({"field": "evidence_refs", "reason": "at least one snapshot fact is required"})
    method_ref = report.get("method_ref")
    method = snapshot.get("method_context", {})
    if (
        not isinstance(method_ref, dict)
        or method_ref.get("scope") != method.get("scope")
        or method_ref.get("school") != method.get("school")
    ):
        errors.append({"field": "method_ref", "reason": "must match snapshot scope/school"})

    for field in ("headline", "verdict", "action"):
        text = report.get(field, "")
        if not isinstance(text, str):
            continue
        for word in FORBIDDEN:
            if word.lower() in text.lower():
                errors.append({"field": "forbidden_language", "reason": f"contains {word}"})

    return {
        "schema_version": "qimen-report-audit-v1",
        "accepted": not errors,
        "errors": errors,
        "checked": ["schema", "method ref", "evidence refs", "timing refs", "forbidden language"],
        "unchecked": ["格局主次", "吉凶结论", "现实建议", "应期是否必然发生"],
    }
