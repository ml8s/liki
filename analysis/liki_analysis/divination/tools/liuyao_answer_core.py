"""六爻 answer core：从 snapshot 投影证据并校验引用与禁语。"""
from __future__ import annotations

from divination_answer import validate_common_core


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


def _evidence_ids(snapshot: dict, layer: str) -> set[str]:
    items = snapshot.get("evidence", {}).get(layer, [])
    return {
        item["id"]
        for item in items
        if isinstance(item, dict) and isinstance(item.get("id"), str)
    }


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
    errors = validate_common_core(
        core,
        allowed_fields=ALLOWED_FIELDS,
        required_text_fields=("headline", "verdict", "action"),
    )

    for field, layer in (
        ("primary_evidence_refs", "primary"),
        ("secondary_evidence_refs", "secondary"),
    ):
        evidence_ids = _evidence_ids(snapshot, layer)
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
