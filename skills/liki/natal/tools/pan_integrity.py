"""本命 pan 的 canonical digest 与完整性校验。"""

from __future__ import annotations

import hashlib
import json
from typing import Any

DIGEST_FIELD = "pan_digest"


def _canonical_json(value: Any) -> str:
    return json.dumps(
        value,
        ensure_ascii=False,
        sort_keys=True,
        separators=(",", ":"),
        allow_nan=False,
    )


def build_natal_digest(pan: dict) -> str:
    """计算 canonical pan digest；digest 字段本身不参与计算。"""
    if not isinstance(pan, dict):
        raise ValueError("pan must be an object")
    core = {key: value for key, value in pan.items() if key != DIGEST_FIELD}
    raw = _canonical_json(core).encode("utf-8")
    return hashlib.sha256(raw).hexdigest()


def with_natal_digest(pan: dict) -> dict:
    """返回带 digest 的 pan；不修改调用方对象。"""
    result = dict(pan)
    result[DIGEST_FIELD] = build_natal_digest(result)
    return result


def validate_natal_digest(pan: dict, action: str = "pan") -> None:
    """校验 pan 是否完整且未被调用方修改。"""
    digest = pan.get(DIGEST_FIELD)
    if not isinstance(digest, str) or len(digest) != 64:
        raise ValueError(f"{action} pan.{DIGEST_FIELD} 必须是 full_paipan 返回的 64 位 digest")
    if digest != build_natal_digest(pan):
        raise ValueError(f"{action} pan digest mismatch；禁止修改或手工拼装 pan")
