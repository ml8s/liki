"""工具层错误契约。

这些错误都是领域/参数契约错误，同时继承 ValueError，
保证既有调用方 `except ValueError` 的错误透传语义不变。
"""

__all__ = [
    "LikiToolError", "ChartSchemaError", "AssertionRuleError",
    "YearRangeError", "FactorEvaluateError", "FactorTableError",
]


class LikiToolError(ValueError):
    """liki 工具层错误的公共基类。"""


class ChartSchemaError(LikiToolError):
    """natal_chart 不是 build_natal_chart 返回的完整结构。"""


class AssertionRuleError(LikiToolError):
    """断语域、规则别名或断语快照契约错误。"""


class YearRangeError(LikiToolError):
    """流年范围、跨度或规则展开错误。"""


class FactorEvaluateError(LikiToolError):
    """因子表达式、算子或求值契约错误。"""


class FactorTableError(LikiToolError):
    """因子长表结构、分组或引用错误。"""
