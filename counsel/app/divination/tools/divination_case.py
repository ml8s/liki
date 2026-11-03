"""问卦 case 公共 envelope。"""

from typing import Any


def build_case(
    *,
    technique: str,
    payload: dict[str, Any],
) -> dict:
    """给 technique-specific case 加公共 envelope。"""
    if not isinstance(technique, str) or not technique:
        raise ValueError("case technique must be non-empty")
    if not isinstance(payload, dict):
        raise ValueError("case payload must be an object")
    reserved = {"technique"}
    conflicts = reserved.intersection(payload)
    if conflicts:
        raise ValueError(f"case payload cannot override envelope fields: {sorted(conflicts)}")
    case = dict(payload)
    case["technique"] = technique
    return case
