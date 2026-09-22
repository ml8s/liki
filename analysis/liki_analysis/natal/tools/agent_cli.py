"""Liki natal analysis tools 的唯一 CLI 入口。

stdin 接收一行 JSON：{"fn": "<工具名>", "args": {...}}。
stdout 固定输出 {"ok":true,"data":...} 或 {"ok":false,"error":{code,message}}。
"""
from __future__ import annotations

import json
import os
import sys

from analytics import (
    LocationError,
    analyze_natal,
    analyze_periods,
    calibrate_birth_time,
    compare_birth_charts,
    create_birth_chart,
)
from chart_token import ChartRefError
from errors import (
    AssertionRuleError,
    FactorEvaluateError,
    FactorTableError,
    LikiToolError,
    PanSchemaError,
    YearRangeError,
)
from paipan import RPCError, ensure_engine_compatible


def _configure_windows_stdio() -> None:
    """Windows CLI 的 JSON 流按 UTF-8 处理；诊断流不可编码时降级为转义。"""
    if os.name != "nt":
        return
    for stream in (sys.stdin, sys.stdout):
        try:
            stream.reconfigure(encoding="utf-8")
        except (AttributeError, OSError):
            pass
    try:
        sys.stderr.reconfigure(encoding="utf-8", errors="backslashreplace")
    except (AttributeError, OSError):
        pass


def _emit(payload: dict) -> None:
    """输出 JSON；ASCII 转义避免 Windows 控制台代码页破坏中文。"""
    print(json.dumps(payload, ensure_ascii=True))


# 公共工具白名单：通过模块属性包装分派；没有 eval/exec/getattr 动态调用。
_DISPATCH = {
    "create_birth_chart": lambda args: create_birth_chart(args),
    "analyze_natal": lambda args: analyze_natal(args),
    "analyze_periods": lambda args: analyze_periods(args),
    "compare_birth_charts": lambda args: compare_birth_charts(args),
    "calibrate_birth_time": lambda args: calibrate_birth_time(args),
}
_REQUIRED_ARGS = {
    "create_birth_chart": ("gender", "source"),
    "analyze_natal": ("chart_ref", "topics"),
    "analyze_periods": ("chart_ref", "time_scope", "topics"),
    "compare_birth_charts": ("chart_ref_a", "chart_ref_b"),
    "calibrate_birth_time": ("candidates", "events"),
}


def _dispatch(fn: str, args: dict):
    """白名单分派；schema 先做形状约束，服务层负责领域约束。"""
    handler = _DISPATCH.get(fn)
    if handler is None:
        raise ValueError(f"unknown tool: {fn}")
    return handler(args)


def _error_payload(exc: Exception) -> dict:
    """领域异常 → 稳定错误码；未知异常不泄漏调用栈。"""
    code_by_type = {
        ChartRefError: "CHART_DIGEST_MISMATCH",
        PanSchemaError: "CHART_INVALID",
        AssertionRuleError: "INVALID_TOPIC",
        YearRangeError: "PERIOD_OUT_OF_RANGE",
        FactorEvaluateError: "FACTOR_EVALUATION_FAILED",
        FactorTableError: "ASSERTION_TABLE_INVALID",
        LocationError: "LOCATION_NOT_RESOLVED",
        RPCError: "ENGINE_UNAVAILABLE",
    }
    if isinstance(exc, LikiToolError):
        for exc_type, code in code_by_type.items():
            if isinstance(exc, exc_type):
                return {"code": code, "message": str(exc)}
        return {"code": "INVALID_INPUT", "message": str(exc)}
    if isinstance(exc, RPCError):
        return {"code": "ENGINE_UNAVAILABLE", "message": str(exc)}
    if isinstance(exc, (KeyError, TypeError, ValueError)):
        return {"code": "INVALID_INPUT", "message": str(exc)}
    return {"code": "TOOL_EXECUTION_FAILED", "message": str(exc)}


def main() -> int:
    raw = sys.stdin.read().strip()
    if not raw:
        _emit({"ok": False, "error": {
            "code": "INVALID_INPUT", "message": "empty stdin",
        }})
        return 0
    try:
        req = json.loads(raw)
        fn = req["fn"]
        args = req.get("args", {})
        if not isinstance(args, dict):
            raise ValueError("args must be an object")
        if fn not in _DISPATCH:
            raise ValueError(f"unknown tool: {fn}")
        missing = [key for key in _REQUIRED_ARGS[fn] if key not in args]
        if missing:
            raise ValueError(f"missing arg: {', '.join(missing)}")
        ensure_engine_compatible()
        _emit({"ok": True, "data": _dispatch(fn, args)})
    except Exception as exc:  # noqa: BLE001 —— CLI 边界统一转稳定 JSON 错误契约
        _emit({"ok": False, "error": _error_payload(exc)})
    return 0


if __name__ == "__main__":
    _configure_windows_stdio()
    sys.exit(main())
