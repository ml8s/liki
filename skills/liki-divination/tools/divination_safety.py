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


def assess(question: str) -> dict:
    """返回 allow/redirect 状态；调用方必须在排盘前检查。"""
    if not isinstance(question, str):
        raise ValueError("question must be a string")
    message = _load_rules()["message"]
    for category, pattern, guidance in _compiled_rules():
        if pattern.search(question):
            return {
                "status": "redirect",
                "category": category,
                "message": message,
                "guidance": guidance,
            }
    return {"status": "allow", "category": None, "message": "", "guidance": ""}


def blocked_payload(*, question: str, method: str, safety: dict) -> dict:
    """返回统一安全拦截对象；其中不包含任何盘面或应期因子。"""
    if safety.get("status") != "redirect":
        raise ValueError("blocked payload requires redirect safety")
    payload = {
        "$schema": "liki:divination-blocked-v1",
        "schema_version": "divination-blocked-v1",
        "blocked": True,
        "method": method,
        "question": question,
        "safety": safety,
        "policy": {
            "no_casting": True,
            "no_chart": True,
            "no_timing": True,
            "no_guilty_or_fatal_prediction": True,
        },
    }
    validate_document("divination_blocked", payload)
    return payload
