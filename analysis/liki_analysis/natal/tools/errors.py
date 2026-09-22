"""工具层错误契约。

这些错误都是领域/参数契约错误，同时继承 ValueError，
保证既有调用方 `except ValueError` 与 agent_cli 的错误透传不变。
"""
from __future__ import annotations

__all__ = [
    "LikiToolError", "PanSchemaError", "AssertionRuleError",
    "YearRangeError", "FactorEvaluateError", "FactorTableError",
]


class LikiToolError(Exception):
    """liki 工具层错误的公共基类。"""


class PanSchemaError(LikiToolError, ValueError):
    """pan 不是 full_paipan 返回的完整结构。"""


class AssertionRuleError(LikiToolError, ValueError):
    """断语域、规则别名或断语快照契约错误。"""


class YearRangeError(LikiToolError, ValueError):
    """流年范围、跨度或规则展开错误。"""


class FactorEvaluateError(LikiToolError, ValueError):
    """因子表达式、算子或求值契约错误。"""


class FactorTableError(LikiToolError, ValueError):
    """因子长表结构、分组或引用错误。"""
