"""本命盘稳定资源 ID。

仅用于展示和日志的稳定短 ID；不作为安全令牌，不参与完整性校验。
"""
from __future__ import annotations

import hashlib
import json

__all__ = ["chart_id"]


def _canonical(pan: dict) -> bytes:
    return json.dumps(pan, ensure_ascii=False, sort_keys=True,
                      separators=(",", ":")).encode("utf-8")


def _digest(payload: bytes) -> str:
    return hashlib.sha256(payload).hexdigest()


def chart_id(pan: dict) -> str:
    """稳定资源 ID；仅用于展示和日志，不作为安全令牌。"""
    return (pan.get("pan_digest") or _digest(_canonical(pan)))[:16]
