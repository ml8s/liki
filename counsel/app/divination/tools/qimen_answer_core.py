"""奇门 answer core：从 case 投影证据并校验引用与禁语。"""

from app.divination.tools.divination_answer import validate_common_core

ALLOWED_FIELDS = {
    "headline",
    "verdict",
    "refs",
    "timing_refs",
    "action",
    "boundary",
    "disclaimer",
}


def available_ids(case: dict) -> tuple[set[str], set[str]]:
    evidence_ids = {
        item["id"]
        for item in case.get("evidence", [])
        if isinstance(item, dict) and item.get("id")
    }
    assertion_ids = {
        f"assertion:{item['id']}"
        for item in (case.get("special") or {}).get("assertions", [])
        if item.get("id")
    }
    assertion_ids |= {
        f"assertion:{item['id']}"
        for item in (case.get("judgments") or {}).get("results", [])
        if item.get("applies") and item.get("id")
    }
    evidence_ids |= assertion_ids
    timing_ids = {
        item["id"]
        for item in case.get("timing_candidates", [])
        if isinstance(item, dict) and item.get("id")
    }
    return evidence_ids, timing_ids


def build(case: dict) -> dict:
    if not isinstance(case, dict) or case.get("technique") != "qimen":
        raise ValueError("case must use technique=qimen")
    evidence_ids, timing_ids = available_ids(case)
    return {
        "headline": "",
        "verdict": "",
        "refs": sorted(evidence_ids),
        "timing_refs": sorted(timing_ids),
        "action": "",
        "boundary": "传统奇门视角；不构成医疗、法律、财务、人事或紧急决策建议。",
        "disclaimer": "结论为传统文化视角下的条件性倾向，不构成结果承诺。",
    }


def validate_core(core: dict, case: dict) -> dict:
    errors = validate_common_core(
        core,
        allowed_fields=ALLOWED_FIELDS,
        required_text_fields=("headline", "verdict", "action"),
    )

    all_ids, timing_ids = available_ids(case)
    for field, allowed in (
        ("refs", all_ids),
        ("timing_refs", timing_ids),
    ):
        refs = core.get(field, [])
        if not isinstance(refs, list):
            errors.append({"field": field, "reason": "must be array"})
            continue
        for ref in refs:
            if ref not in allowed:
                errors.append({"field": field, "reason": "unknown ref", "ref": ref})

    if not core.get("refs"):
        errors.append({"field": "refs", "reason": "at least one case or assertion ref is required"})
    return {
        "accepted": not errors,
        "errors": errors,
        "checked": ["fields", "refs", "timing refs", "forbidden language"],
        "unchecked": ["格局主次", "吉凶结论", "现实建议", "应期是否必然发生"],
    }
