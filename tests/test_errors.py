"""错误契约：专用错误可捕获，且保持 LikiToolError 的 ValueError 兼容行为。"""
import pytest

import _helpers  # noqa: F401
import duanyu
from errors import (
    AssertionRuleError, FactorEvaluateError, FactorTableError,
    LikiToolError, PanSchemaError, YearRangeError,
)
from factors import _atomic
from pan_schema import validate_natal_pan


def test_error_hierarchy_is_valueerror_compatible():
    for error_type in (
        PanSchemaError, AssertionRuleError, YearRangeError,
        FactorEvaluateError, FactorTableError,
    ):
        assert issubclass(error_type, LikiToolError)
        assert issubclass(error_type, ValueError)


def test_pan_schema_error():
    try:
        validate_natal_pan({}, action="test")
    except ValueError as exc:
        assert isinstance(exc, PanSchemaError)
    else:
        raise AssertionError("missing pan did not raise")


def test_year_range_error():
    try:
        duanyu.yearly_range({}, 2027, 2026, rules=["yingqi"])
    except ValueError as exc:
        assert isinstance(exc, YearRangeError)
    else:
        raise AssertionError("reversed range did not raise")


def test_factor_table_error():
    try:
        raise FactorTableError("bad table")
    except ValueError as exc:
        assert isinstance(exc, FactorTableError)


def test_factor_evaluate_error():
    try:
        _atomic("不存在算子[]", "male", {"ten_god_states": {}})
    except ValueError as exc:
        assert isinstance(exc, FactorEvaluateError)
    else:
        raise AssertionError("unknown operator did not raise")
