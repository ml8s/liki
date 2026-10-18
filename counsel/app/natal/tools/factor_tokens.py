"""因子 DSL 的语法 token；与命理闭集 constants.json 分离。"""
from __future__ import annotations

from typing import Final

FACTOR_WILDCARD: Final = "任意"
FACTOR_UNKNOWN: Final = "unknown"

__all__ = ["FACTOR_UNKNOWN", "FACTOR_WILDCARD"]
