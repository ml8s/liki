"""问卦 / 择日统一安全边界；规则与指导语来自数据表。"""

from __future__ import annotations

import json
import re
from functools import lru_cache
from pathlib import Path

from divination_contracts import validate_document

RULES_PATH = Path(__file__).with_name("divination_safety_rules.json")


@lru_cache
def _load_rules() -> dict:
    return json.loads(RULES_PATH.read_text(encoding="utf-8"))


@lru_cache
def _compiled_rules() -> list[tuple[str, re.Pattern[str], str]]:
    compiled: list[tuple[str, re.Pattern[str], str]] = []
    for rule in _load_rules()["rules"]:
        patterns = "|".join(f"(?:{pattern})" for pattern in rule["patterns"])
        compiled.append((rule["category"], re.compile(patterns, re.IGNORECASE), rule["guidance"]))
    return compiled


_HIGH_SUPPORT_CATEGORIES = {
    "self_harm",
    "emergency_medical",
    "domestic_violence",
    "missing_person_safety",
}


def assess(question: str) -> dict:
    """识别现实风险主题，返回 advisory；调用方不得因此阻断排盘。"""
    if not isinstance(question, str):
        raise ValueError("question must be a string")
    rules = _load_rules()
    message = rules["message"]
    for category, pattern, guidance in _compiled_rules():
        if pattern.search(question):
            return {
                "status": "advisory",
                "blocking": False,
                "category": category,
                "severity": "high" if category in _HIGH_SUPPORT_CATEGORIES else "normal",
                "support_first": category in _HIGH_SUPPORT_CATEGORIES,
                "message": message,
                "guidance": guidance,
            }
    return {
        "status": "allow",
        "blocking": False,
        "category": None,
        "severity": "none",
        "support_first": False,
        "message": "",
        "guidance": "",
    }
