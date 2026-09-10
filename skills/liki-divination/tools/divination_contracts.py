"""问卦契约集中加载与运行时校验。"""
from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path

from jsonschema import Draft202012Validator


PATH = Path(__file__).parent

CONTRACT_FILES = {
    "divination_blocked": "divination_blocked_contract.json",
    "liuyao_snapshot": "liuyao_snapshot_contract.json",
    "qimen_snapshot": "qimen_snapshot_contract.json",
    "huangli_days": "huangli_days_contract.json",
    "liuyao_answer": "liuyao_answer_contract.json",
    "qimen_answer": "qimen_answer_contract.json",
}


@lru_cache
def load_contract(name: str) -> dict:
    if name not in CONTRACT_FILES:
        raise ValueError(f"unknown contract: {name}")
    return json.loads((PATH / CONTRACT_FILES[name]).read_text(encoding="utf-8"))


@lru_cache
def _validator(name: str):
    return Draft202012Validator(load_contract(name))


def validate_document(name: str, document: dict) -> None:
    """校验失败抛 ValueError；成功返回 None。"""
    validator = _validator(name)
    errors = sorted(validator.iter_errors(document), key=lambda item: list(item.absolute_path))
    if not errors:
        return
    details = []
    for error in errors:
        path = ".".join(str(part) for part in error.absolute_path) or "$"
        details.append(f"{path}: {error.message}")
    raise ValueError(f"{name} contract failed: " + "; ".join(details))
