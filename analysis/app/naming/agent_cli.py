"""Liki naming (起名) tools CLI 入口。

stdin 接收一行 JSON：{"fn": "<工具名>", "args": {...}}。
stdout 固定输出 {"ok":true,"data":...} 或 {"ok":false,"error":{code,message}}。
纯函数实现（app.naming.qiming），与 engine Go qiming 逐字段对齐。
"""
from __future__ import annotations

import json
import os
import sys

from qiming import (
    compose_names,
    evaluate_names,
    lookup_char,
    match_surnames,
    pick_chars,
)

_DISPATCH = {
    "qiming_pick": pick_chars,
    "qiming_surname": match_surnames,
    "qiming_char": lookup_char,
    "qiming_compose": compose_names,
    "qiming_check": evaluate_names,
}

_REQUIRED_ARGS = {
    "qiming_pick": ("wuxing1",),
    "qiming_surname": ("source_surname",),
    "qiming_char": ("char",),
    "qiming_compose": ("first",),
    "qiming_check": ("given_names",),
}


def _error_payload(exc: Exception) -> dict:
    if isinstance(exc, (ValueError, TypeError, KeyError)):
        return {"code": "VALIDATION", "message": str(exc)}
    return {"code": "INTERNAL", "message": str(exc)}


def _emit(payload: dict) -> None:
    sys.stdout.write(json.dumps(payload, ensure_ascii=False) + "\n")


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
        _emit({"ok": True, "data": _dispatch(fn, args)})
    except Exception as exc:  # noqa: BLE001 —— CLI 边界统一转稳定 JSON 错误契约
        _emit({"ok": False, "error": _error_payload(exc)})
    return 0


def _dispatch(fn: str, args: dict):
    if fn == "qiming_check":
        # 工具契约参数（对齐 Go）→ 函数签名
        args = dict(args)
        if "yongshen" in args:
            args["yong_shen"] = args.pop("yongshen")
        if "xishen" in args:
            args["xi_shen"] = args.pop("xishen")
        if "jishen" in args:
            args["ji_shen"] = args.pop("jishen")
    return _DISPATCH[fn](**args)


if __name__ == "__main__":
    sys.exit(main())