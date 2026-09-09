"""奇门结构化报告契约：生成骨架、校验引用和禁止性表述。"""
from __future__ import annotations

import hashlib
import re


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


def _digest(value) -> str:
    import json
    raw = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()[:16]


def candidate_id(candidate: dict, index: int) -> str:
    raw = f"{candidate.get('type')}|{candidate.get('branch')}|{candidate.get('gong')}"
    return f"timing-{_digest(raw)}"


def available_ids(read_result: dict) -> tuple[set[str], set[str]]:
    """返回 evidence ids 和 timing ids。"""
    snapshot = read_result.get("snapshot", {})
    evidence_ids = {f"snapshot:{key}" for key in snapshot}
    assertion_ids = {f"assertion:{item['id']}" for item in (read_result.get("special") or {}).get("assertions", []) if item.get("id")}
    evidence_ids |= assertion_ids
    special = read_result.get("special") or {}
    for assertion in special.get("assertions", []):
        if assertion.get("id"):
            evidence_ids.add(f"assertion:{assertion['id']}")
    timing_ids = {
        candidate_id(candidate, index)
        for index, candidate in enumerate(snapshot.get("ying_qi", []), 1)
    }
    return evidence_ids, timing_ids


def template(read_result: dict) -> dict:
    if not isinstance(read_result, dict) or read_result.get("schema_version") != "qimen-read-v1":
        raise ValueError("read_result must be qimen-read-v1")
    evidence_ids, timing_ids = available_ids(read_result)
    matter = read_result.get("matter") or {}
    method = read_result.get("snapshot", {}).get("method", {})
    return {
        "schema_version": SCHEMA_VERSION,
        "headline": "",
        "verdict": "",
        "confidence": "medium",
        "question_ref": read_result.get("question", ""),
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


def validate_report(report: dict, read_result: dict) -> dict:
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

    evidence_ids, timing_ids = available_ids(read_result)
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
    method = read_result.get("snapshot", {}).get("method", {})
    if not isinstance(method_ref, dict) or method_ref.get("scope") != method.get("scope") or method_ref.get("school") != method.get("school"):
        errors.append({"field": "method_ref", "reason": "must match read result scope/school"})

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
