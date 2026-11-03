"""金函玉镜 answer core；只解释星、门与十二神事实。"""

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


def _validate_case(case: dict) -> None:
    factors = case.get("factors")
    if not isinstance(factors, dict) or factors.get("kind") != "jinhan":
        raise ValueError("case must contain jinhan factors")
    palaces = factors.get("palaces")
    day_spirits = factors.get("day_spirits")
    if not isinstance(palaces, list) or len(palaces) != 9:
        raise ValueError("jinhan factors require 9 palaces")
    if not isinstance(day_spirits, list) or len(day_spirits) != 12:
        raise ValueError("jinhan factors require 12 day spirits")


def available_ids(case: dict) -> set[str]:
    _validate_case(case)
    return {
        "case:method",
        "case:palaces",
        "case:day_spirits",
    }


def build(case: dict) -> dict:
    return {
        "headline": "",
        "verdict": "",
        "refs": [
            "case:method",
            "case:palaces",
            "case:day_spirits",
        ],
        "timing_refs": [],
        "action": "",
        "boundary": "金函玉镜日家视角；不构成医疗、法律、财务、人事或紧急决策建议。",
        "disclaimer": "结论为传统文化视角下的条件性倾向，不构成结果承诺。",
    }


def validate_core(core: dict, case: dict) -> dict:
    errors = validate_common_core(
        core,
        allowed_fields=ALLOWED_FIELDS,
        required_text_fields=("headline", "verdict", "action"),
    )

    allowed_ids = available_ids(case)
    for ref in core.get("refs", []):
        if ref not in allowed_ids:
            errors.append({"field": "refs", "reason": "unknown ref", "ref": ref})
    if not core.get("refs"):
        errors.append({"field": "refs", "reason": "at least one case fact is required"})
    for field in ("timing_refs",):
        refs = core.get(field, [])
        if not isinstance(refs, list):
            errors.append({"field": field, "reason": "must be array"})
        elif refs:
            errors.append({
                "field": field,
                "reason": "jinhan case has no timing candidates",
            })

    return {
        "accepted": not errors,
        "errors": errors,
        "checked": ["fields", "refs", "forbidden language"],
        "unchecked": ["星门组合主次", "吉凶结论", "现实建议", "十二神用途"],
    }
