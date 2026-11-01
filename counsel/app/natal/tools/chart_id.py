"""本命盘稳定资源 ID。

仅用于展示和日志的稳定短 ID；不作为安全令牌，不参与完整性校验。
"""

from app.integrity.digest import sha256_json

__all__ = ["chart_id"]


def chart_id(natal_chart: dict) -> str:
    """稳定资源 ID；仅用于展示和日志，不作为安全令牌。"""
    return (natal_chart.get("natal_chart_digest") or sha256_json(natal_chart))[:16]
