"""本命 natal_chart 的 canonical digest 与完整性校验。"""


import hashlib
import json
from typing import Any

DIGEST_FIELD = "natal_chart_digest"


def _canonical_json(value: Any) -> str:
    return json.dumps(
        value,
        ensure_ascii=False,
        sort_keys=True,
        separators=(",", ":"),
        allow_nan=False,
    )


def build_natal_chart_digest(natal_chart: dict) -> str:
    """计算 canonical natal_chart digest；digest 字段本身不参与计算。"""
    if not isinstance(natal_chart, dict):
        raise ValueError("natal_chart must be an object")
    core = {key: value for key, value in natal_chart.items() if key != DIGEST_FIELD}
    raw = _canonical_json(core).encode("utf-8")
    return hashlib.sha256(raw).hexdigest()


def with_natal_chart_digest(natal_chart: dict) -> dict:
    """返回带 digest 的 natal_chart；不修改调用方对象。"""
    result = dict(natal_chart)
    result[DIGEST_FIELD] = build_natal_chart_digest(result)
    return result


def validate_natal_chart_digest(natal_chart: dict, action: str = "natal_chart") -> None:
    """校验 natal_chart 是否完整且未被调用方修改。"""
    digest = natal_chart.get(DIGEST_FIELD)
    if not isinstance(digest, str) or len(digest) != 64:
        raise ValueError(f"{action} natal_chart.{DIGEST_FIELD} 必须是 build_natal_chart 返回的 64 位 digest")
    if digest != build_natal_chart_digest(natal_chart):
        raise ValueError(f"{action} natal_chart digest mismatch；禁止修改或手工拼装 natal_chart")
