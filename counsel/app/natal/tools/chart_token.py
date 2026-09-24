"""不可变本命盘资源的无状态引用编解码器。

CLI 每次调用都是独立进程，因此 chart_ref 采用自包含压缩令牌；令牌始终携带
pan_digest，并在解码时重新校验完整盘结构与摘要。LLM 只复制 token，不组装盘。
"""
from __future__ import annotations

import base64
import binascii
import hashlib
import json
import zlib

from errors import LikiToolError
from pan_integrity import with_natal_digest
from pan_schema import validate_natal_pan

__all__ = ["encode_chart_ref", "decode_chart_ref", "chart_id"]

_TOKEN_PREFIX = "liki-chart-v1"


class ChartRefError(LikiToolError, ValueError):
    """chart_ref 缺失、被裁剪、被修改或摘要不匹配。"""


def _canonical(pan: dict) -> bytes:
    return json.dumps(pan, ensure_ascii=False, sort_keys=True,
                      separators=(",", ":")).encode("utf-8")


def _digest(payload: bytes) -> str:
    return hashlib.sha256(payload).hexdigest()


def _decode_segment(token: str) -> tuple[bytes, str]:
    if not isinstance(token, str) or not token.startswith(_TOKEN_PREFIX + "."):
        raise ChartRefError("chart_ref.token 版本无效")
    parts = token.split(".")
    if len(parts) != 3:
        raise ChartRefError("chart_ref.token 结构无效")
    _version, body, digest = parts
    try:
        padding = "=" * (-len(body) % 4)
        return zlib.decompress(base64.urlsafe_b64decode(body + padding)), digest
    except (ValueError, zlib.error, binascii.Error) as e:
        raise ChartRefError(f"chart_ref.token 解码失败: {e}") from e


def encode_chart_ref(pan: dict) -> dict:
    """完整盘 → {version, token, digest}。token 可原样传给分析工具。"""
    validate_natal_pan(pan, action="chart token encode")
    payload = _canonical(pan)
    digest = _digest(payload)
    body = base64.urlsafe_b64encode(zlib.compress(payload, 9)).decode("ascii").rstrip("=")
    return {
        "version": 1,
        "token": f"{_TOKEN_PREFIX}.{body}.{digest}",
        "digest": pan.get("pan_digest") or digest,
    }


def decode_chart_ref(chart_ref: dict) -> dict:
    """校验 chart_ref 并还原完整本命盘；摘要不匹配时 fail closed。"""
    if not isinstance(chart_ref, dict):
        raise ChartRefError("chart_ref 必须是对象")
    token = chart_ref.get("token")
    expected_digest = chart_ref.get("digest")
    if not token or not expected_digest:
        raise ChartRefError("chart_ref 缺少 token 或 digest")
    payload, token_digest = _decode_segment(token)
    if _digest(payload) != token_digest:
        raise ChartRefError("chart_ref.token 摘要不匹配")
    try:
        pan = json.loads(payload.decode("utf-8"))
    except (UnicodeDecodeError, json.JSONDecodeError) as e:
        raise ChartRefError(f"chart_ref.token 内容无效: {e}") from e
    pan = with_natal_digest(pan)
    validate_natal_pan(pan, action="chart token decode")
    if pan.get("pan_digest") != expected_digest:
        raise ChartRefError("chart_ref.digest 与完整盘不一致")
    return pan


def chart_id(pan: dict) -> str:
    """稳定资源 ID；仅用于展示和日志，不作为安全令牌。"""
    return (pan.get("pan_digest") or _digest(_canonical(pan)))[:16]
