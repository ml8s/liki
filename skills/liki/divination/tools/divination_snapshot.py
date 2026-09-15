"""问卦 snapshot 公共契约：canonical digest 与类型校验。"""
from __future__ import annotations

from typing import Any

from divination_hashing import sha256_json


def canonical_digest(snapshot: dict) -> str:
    """计算 snapshot 的稳定 digest；digest 字段本身不参与计算。"""
    core = {key: value for key, value in snapshot.items() if key != "snapshot_digest"}
    return sha256_json(core)


def build_snapshot(
    *,
    method: str,
    schema_version: str,
    payload: dict[str, Any],
) -> dict:
    """给 method-specific snapshot 加公共 envelope 和 digest。"""
    if not isinstance(method, str) or not method:
        raise ValueError("snapshot method must be non-empty")
    if not isinstance(schema_version, str) or not schema_version:
        raise ValueError("snapshot schema_version must be non-empty")
    if not isinstance(payload, dict):
        raise ValueError("snapshot payload must be an object")
    reserved = {"method", "schema_version", "snapshot_digest"}
    conflicts = reserved.intersection(payload)
    if conflicts:
        raise ValueError(f"snapshot payload cannot override envelope fields: {sorted(conflicts)}")
    snapshot = dict(payload)
    snapshot["schema_version"] = schema_version
    snapshot["method"] = method
    snapshot["snapshot_digest"] = canonical_digest(snapshot)
    return snapshot


def validate_snapshot(
    snapshot: dict,
    *,
    method: str,
    schema_version: str,
) -> None:
    """校验 snapshot 只能属于当前领域，且未被修改。"""
    if not isinstance(snapshot, dict):
        raise ValueError("snapshot must be an object")
    if snapshot.get("method") != method:
        raise ValueError(f"snapshot method must be {method}")
    if snapshot.get("schema_version") != schema_version:
        raise ValueError(f"snapshot schema_version must be {schema_version}")
    digest = snapshot.get("snapshot_digest")
    if not isinstance(digest, str) or not digest:
        raise ValueError("snapshot lacks snapshot_digest")
    if digest != canonical_digest(snapshot):
        raise ValueError("snapshot digest mismatch")
