"""Canonical JSON 与 SHA-256 digest 的唯一实现。

规则：UTF-8、对象键排序、无多余空白、数组有序、拒绝 NaN/Infinity。
digest 字段由 artifact helper 负责剔除，本模块只提供 canonical 编码和哈希。
"""

import hashlib
import json
from typing import Any

Digest = str


def canonical_json(value: Any) -> str:
    """返回稳定 JSON 字符串。"""
    return json.dumps(
        value,
        ensure_ascii=False,
        sort_keys=True,
        separators=(",", ":"),
        allow_nan=False,
    )


def sha256_json(value: Any) -> Digest:
    """计算 canonical JSON 的 SHA-256，返回小写 hex。"""
    return hashlib.sha256(canonical_json(value).encode("utf-8")).hexdigest()


def sha256_excluding(value: dict, *fields: str) -> Digest:
    """计算剔除元数据字段后的 canonical SHA-256。"""
    core = {key: item for key, item in value.items() if key not in fields}
    return sha256_json(core)


def require_sha256(value: Any, field: str) -> Digest:
    """校验并返回 64 位小写十六进制 digest。"""
    if not isinstance(value, str) or len(value) != 64:
        raise ValueError(f"{field} 必须是 64 位 SHA-256 digest")
    try:
        bytes.fromhex(value)
    except ValueError as exc:
        raise ValueError(f"{field} 必须是十六进制 SHA-256 digest") from exc
    if value != value.lower():
        raise ValueError(f"{field} 必须是小写 SHA-256 digest")
    return value


def short_object_id(prefix: str, value: Any) -> str:
    """生成稳定短 ID；用于引用，不用于安全判断。"""
    return f"{prefix}-{sha256_json(value)[:16]}"


__all__ = [
    "canonical_json",
    "short_object_id",
    "require_sha256",
    "sha256_excluding",
    "sha256_json",
]
