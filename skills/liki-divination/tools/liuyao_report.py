"""六爻结构化报告契约：生成骨架、校验证据引用和禁止性表述。"""
from __future__ import annotations

import re


SCHEMA_VERSION = "liuyao-report-v1"
FORBIDDEN = ["必然", "百分百", "100%", "保证", "一定会"]
ALLOWED_FIELDS = {
    "schema_version",
    "headline",
    "verdict",
    "confidence",
    "primary_evidence_refs",
    "secondary_evidence_refs",
    "conflicts",
    "timing_refs",
    "action",
    "boundary",
    "disclaimer",
}


def _evidence_ids(snapshot: dict) -> set[str]:
    evidence = snapshot.get("evidence")
    ids: set[str] = set()
    if isinstance(evidence, dict):
        for layer in evidence.values():
            if not isinstance(layer, list):
                continue
            for item in layer:
                if isinstance(item, dict) and isinstance(item.get("id"), str):
                    ids.add(item["id"])
    return ids


def _timing_ids(snapshot: dict) -> set[str]:
    ids: set[str] = set()
    for item in snapshot.get("timing_candidates", []):
        if isinstance(item, dict) and isinstance(item.get("id"), str):
            ids.add(item["id"])
    return ids


def template(snapshot: dict) -> dict:
    """生成必须由 LLM 填充语义内容的结构化报告骨架。"""
    primary_ids = sorted(
        item.get("id")
        for item in snapshot.get("evidence", {}).get("primary", [])
        if isinstance(item, dict) and item.get("id")
    )
    secondary_ids = sorted(
        item.get("id")
        for item in snapshot.get("evidence", {}).get("secondary", [])
        if isinstance(item, dict) and item.get("id")
    )
    timing_ids = sorted(
        item.get("id")
        for item in snapshot.get("timing_candidates", [])
        if isinstance(item, dict) and item.get("id")
    )
    return {
        "schema_version": SCHEMA_VERSION,
        "headline": "",
        "verdict": "",
        "confidence": "medium",
        "primary_evidence_refs": primary_ids,
        "secondary_evidence_refs": secondary_ids,
        "conflicts": snapshot.get("evidence", {}).get("conflicts", []),
        "timing_refs": timing_ids,
        "action": "",
        "boundary": "传统六爻视角；不构成医疗、法律、财务或其他专业建议。",
        "disclaimer": "结论为传统文化视角下的条件性倾向，不承诺现实结果。",
    }


def validate_report(report: dict, snapshot: dict) -> dict:
    errors: list[dict] = []
    if not isinstance(report, dict):
        raise ValueError("report must be an object")
    if report.get("schema_version") != SCHEMA_VERSION:
        errors.append({"field": "schema_version", "reason": "must be liuyao-report-v1"})
    for field in report:
        if field not in ALLOWED_FIELDS:
            errors.append({"field": field, "reason": "unknown field"})

    for field in ("headline", "verdict", "action", "boundary", "disclaimer"):
        if not isinstance(report.get(field), str) or not report.get(field, "").strip():
            errors.append({"field": field, "reason": "required non-empty text"})
    if report.get("confidence") not in {"low", "medium", "high"}:
        errors.append({"field": "confidence", "reason": "must be low/medium/high"})

    evidence_ids = _evidence_ids(snapshot)
    for field in ("primary_evidence_refs", "secondary_evidence_refs"):
        refs = report.get(field, [])
        if not isinstance(refs, list):
            errors.append({"field": field, "reason": "must be array"})
            continue
        for ref in refs:
            if ref not in evidence_ids:
                errors.append({"field": field, "reason": "unknown evidence id", "ref": ref})

    primary_refs = report.get("primary_evidence_refs", [])
    if isinstance(primary_refs, list) and not primary_refs:
        errors.append({"field": "primary_evidence_refs", "reason": "at least one primary fact is required"})

    timing_ids = _timing_ids(snapshot)
    for ref in report.get("timing_refs", []):
        if ref not in timing_ids:
            errors.append({"field": "timing_refs", "reason": "unknown timing candidate id", "ref": ref})

    texts = [report.get(field, "") for field in ("headline", "verdict", "action")]
    for text in texts:
        if not isinstance(text, str):
            continue
        for word in FORBIDDEN:
            if word.lower() in text.lower():
                errors.append({"field": "forbidden_language", "reason": f"contains {word}"})

    conflicts = snapshot.get("evidence", {}).get("conflicts", [])
    if conflicts and not report.get("conflicts"):
        errors.append({"field": "conflicts", "reason": "snapshot has conflicts that must be covered"})

    return {
        "schema_version": "liuyao-report-audit-v1",
        "accepted": not errors,
        "errors": errors,
        "checked": ["schema", "evidence refs", "timing refs", "conflict coverage", "forbidden language"],
        "unchecked": ["传统取象", "吉凶结论", "现实建议", "应期是否必然发生"],
    }
