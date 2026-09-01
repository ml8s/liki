#!/usr/bin/env python3
"""agent_cli.py — liki 命理工具的 CLI 适配器。

协议：stdin 读一行 JSON {"fn": <工具名>, "args": {<参数>}} → stdout 一行 JSON。
成功：{"ok": true, "data": <结果>}；失败：{"ok": false, "error": "..."}。

设计：
- 白名单分派（显式 if/elif，无 eval/exec/getattr 动态调用）——LLM 不能执行任意代码
- 参数 dict 直接 **args 传给工具函数——参数化调用，零代码注入
- 异常捕获进 error 字段（不 panic、exit 0）——调用方按 ok 字段判断

使用方式：
- Windows 优先通过 `tools\agent_cli.cmd`；POSIX 通过 `python3 tools/agent_cli.py` 执行
- stdin 传 JSON：{"fn": "<工具名>", "args": {<参数>}}
- stdout 返回 JSON：{"ok": true, "data": <结果>} 或 {"ok": false, "error": "..."}
"""
from __future__ import annotations

import json
import os
import sys

from paipan import full_paipan, city_coords, bond
from duanyu import query, yearly_range
from calibrate import calibrate


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


# 白名单：工具名 → 参数提取器（无 eval/exec/getattr 动态调用）
_DISPATCH = {
    "city_coords":  lambda a: city_coords(a["city"]),
    "full_paipan": lambda a: full_paipan(a["gregorian"], a["gender"],
                                         longitude=a.get("longitude"),
                                         correct=a.get("correct", True)),
    "query":        lambda a: query(a["rule"], a["pan"]),
    "yearly_range": lambda a: yearly_range(a["pan"], a["start"], a["end"],
                                           rules=a.get("rules"),
                                           detail=a.get("detail", False)),
    "calibrate":    lambda a: calibrate(a["candidates"], a["events"], detail=a.get("detail", False)),
    "bond":         lambda a: bond(a["pan_a"], a["pan_b"]),
}


def _dispatch(fn: str, args: dict):
    """白名单分派（dict 映射）。args 为参数字典（由 schema 约束，此处直接传函数）。"""
    handler = _DISPATCH.get(fn)
    if handler is None:
        raise ValueError(f"unknown tool: {fn}")
    return handler(args)


def main() -> int:
    raw = sys.stdin.read().strip()
    if not raw:
        _emit({"ok": False, "error": "empty stdin"})
        return 0
    try:
        req = json.loads(raw)
        fn = req["fn"]
        args = req.get("args", {})
        if not isinstance(args, dict):
            raise ValueError("args must be an object")
        data = _dispatch(fn, args)
        _emit({"ok": True, "data": data})
    except KeyError as e:
        _emit({"ok": False, "error": f"missing arg: {e}"})
    except Exception as e:  # noqa: BLE001 —— 工具链异常（网络/参数/真值表）统一转错误
        _emit({"ok": False, "error": f"{type(e).__name__}: {e}"})
    return 0


if __name__ == "__main__":
    _configure_windows_stdio()
    sys.exit(main())
