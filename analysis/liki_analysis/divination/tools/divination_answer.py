"""问卦 answer 公共校验：字段、文本、置信度与禁语。"""
from __future__ import annotations


FORBIDDEN_PHRASES = ("必然", "百分百", "100%", "保证", "一定会")
CONFIDENCE_LEVELS = {"low", "medium", "high"}


def unknown_fields(core: dict, allowed_fields: set[str]) -> list[str]:
    return [field for field in core if field not in allowed_fields]


def has_forbidden_phrase(text: str) -> str | None:
    lowered = text.lower()
    for phrase in FORBIDDEN_PHRASES:
        if phrase.lower() in lowered:
            return phrase
    return None


def validate_common_core(
    core: dict,
    *,
    allowed_fields: set[str],
    required_text_fields: tuple[str, ...],
) -> list[dict]:
    errors: list[dict] = []
    if not isinstance(core, dict):
        raise ValueError("answer core must be an object")

    for field in unknown_fields(core, allowed_fields):
        errors.append({"field": field, "reason": "unknown field"})

    for field in (*required_text_fields, "boundary", "disclaimer"):
        if not isinstance(core.get(field), str) or not core.get(field, "").strip():
            errors.append({"field": field, "reason": "required non-empty text"})
    if core.get("confidence") not in CONFIDENCE_LEVELS:
        errors.append({"field": "confidence", "reason": "must be low/medium/high"})

    for field in required_text_fields:
        phrase = has_forbidden_phrase(core.get(field, ""))
        if phrase:
            errors.append({"field": "forbidden_language", "reason": f"contains {phrase}"})

    return errors
