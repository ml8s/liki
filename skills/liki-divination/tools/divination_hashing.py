"""问卦领域通用 canonical JSON 与摘要工具。"""

from __future__ import annotations

import hashlib
import json
from typing import Any


def canonical_json(value: Any) -> str:
    """返回稳定 JSON 序列化；用于 digest 和可复现 ID。"""
    return json.dumps(
        value,
        ensure_ascii=False,
        sort_keys=True,
        separators=(",", ":"),
        allow_nan=False,
    )


def sha256_json(value: Any) -> str:
    return hashlib.sha256(canonical_json(value).encode("utf-8")).hexdigest()


def short_object_id(prefix: str, value: Any) -> str:
    """生成稳定短 ID；用于引用，不用于安全判断。"""
    return f"{prefix}-{sha256_json(value)[:16]}"


def message_digest(message: str) -> str:
    return sha256_json({"message": message})
