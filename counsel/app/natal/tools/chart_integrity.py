"""本命 natal_chart 的 canonical digest 与完整性校验。"""

from app.integrity.digest import require_sha256, sha256_excluding

DIGEST_FIELD = "natal_chart_digest"


def build_natal_chart_digest(natal_chart: dict) -> str:
    """计算 canonical natal_chart digest；digest 字段本身不参与计算。"""
    if not isinstance(natal_chart, dict):
        raise ValueError("natal_chart must be an object")
    return sha256_excluding(natal_chart, DIGEST_FIELD)


def with_natal_chart_digest(natal_chart: dict) -> dict:
    """返回带 digest 的 natal_chart；不修改调用方对象。"""
    result = dict(natal_chart)
    result[DIGEST_FIELD] = build_natal_chart_digest(result)
    return result


def validate_natal_chart_digest(natal_chart: dict, action: str = "natal_chart") -> None:
    """校验 natal_chart 是否完整且未被调用方修改。"""
    try:
        digest = require_sha256(natal_chart.get(DIGEST_FIELD), f"{action} natal_chart.{DIGEST_FIELD}")
    except ValueError as exc:
        raise ValueError(str(exc)) from exc
    if digest != build_natal_chart_digest(natal_chart):
        raise ValueError(f"{action} natal_chart digest mismatch；禁止修改或手工拼装 natal_chart")
