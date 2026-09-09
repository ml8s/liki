"""六爻 answer core：从 snapshot 投影证据并校验引用与禁语。"""
from __future__ import annotations


FORBIDDEN = ["必然", "百分百", "100%", "保证", "一定会"]
ALLOWED_FIELDS = {
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
    ids: set[str] = set()
    evidence = snapshot.get("evidence")
    if isinstance(evidence, dict):
        for layer in evidence.values():
            if not isinstance(layer, list):
                continue
            for item in layer:
                if isinstance(item, dict) and isinstance(item.get("id"), str):
                    ids.add(item["id"])
    return ids


def _timing_ids(snapshot: dict) -> set[str]:
    return {
        item["id"]
        for item in snapshot.get("timing_candidates", [])
        if isinstance(item, dict) and isinstance(item.get("id"), str)
    }


def build(snapshot: dict) -> dict:
    """生成 answer 的确定性证据骨架；语义判断由 LLM 在边界内完成。"""
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
    timing_ids = sorted(_timing_ids(snapshot))
    return {
        "headline": "",
        "verdict": "",
        "confidence": "medium",
        "primary_evidence_refs": primary_ids,
        "secondary_evidence_refs": secondary_ids,
        "conflicts": snapshot.get("evidence", {}).get("conflicts", []),
        "timing_refs": timing_ids,
        "action": "",
        "boundary": "传统六爻视角；不构成医疗、法律、财务或其他专业建议。",
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

    evidence_ids = _evidence_ids(snapshot)
    for field in ("primary_evidence_refs", "secondary_evidence_refs"):
        refs = core.get(field, [])
        if not isinstance(refs, list):
            errors.append({"field": field, "reason": "must be array"})
            continue
        for ref in refs:
            if ref not in evidence_ids:
                errors.append({"field": field, "reason": "unknown evidence id", "ref": ref})

    primary_refs = core.get("primary_evidence_refs", [])
    if isinstance(primary_refs, list) and not primary_refs:
        errors.append({"field": "primary_evidence_refs", "reason": "at least one primary fact is required"})

    timing_ids = _timing_ids(snapshot)
    for ref in core.get("timing_refs", []):
        if ref not in timing_ids:
            errors.append({"field": "timing_refs", "reason": "unknown timing candidate id", "ref": ref})

    texts = [core.get(field, "") for field in ("headline", "verdict", "action")]
    for text in texts:
        if not isinstance(text, str):
            continue
        for word in FORBIDDEN:
            if word.lower() in text.lower():
                errors.append({"field": "forbidden_language", "reason": f"contains {word}"})

    conflicts = snapshot.get("evidence", {}).get("conflicts", [])
    if conflicts and not core.get("conflicts"):
        errors.append({"field": "conflicts", "reason": "snapshot has conflicts that must be covered"})

    return {
        "schema_version": "liuyao-answer-core-audit-v1",
        "accepted": not errors,
        "errors": errors,
        "checked": ["fields", "evidence refs", "timing refs", "conflict coverage", "forbidden language"],
        "unchecked": ["传统取象", "吉凶结论", "现实建议", "应期是否必然发生"],
    }
