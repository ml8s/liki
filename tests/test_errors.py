"""错误契约：专用错误可捕获，且保持 agent_cli 的 ValueError 兼容行为。"""
import json
import subprocess
import sys
from pathlib import Path

import pytest

import _helpers  # noqa: F401
import duanyu
from errors import (
    AssertionRuleError, FactorEvaluateError, FactorTableError,
    LikiToolError, PanSchemaError, YearRangeError,
)
from factors import _atomic
from pan_schema import validate_natal_pan

ROOT = Path(__file__).resolve().parents[1]


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
        _atomic("不存在算子[]", "male", {"shishen": {}})
    except ValueError as exc:
        assert isinstance(exc, FactorEvaluateError)
    else:
        raise AssertionError("unknown operator did not raise")


def test_agent_cli_transports_error_as_json():
    process = subprocess.run(
        [sys.executable, str(ROOT / "skills/liki-bazi/tools/agent_cli.py")],
        input=json.dumps(
            {"fn": "query", "args": {"rule": "十神", "pan": {}}}
        ).encode(),
        capture_output=True,
        timeout=10,
    )
    assert process.returncode == 0
    payload = json.loads(process.stdout)
    assert payload["ok"] is False
    assert "PanSchemaError" in payload["error"]
