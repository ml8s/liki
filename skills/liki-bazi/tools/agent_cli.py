#!/usr/bin/env python3
"""agent_cli.py — liki 命理工具的 CLI 适配器。

协议：stdin 读一行 JSON {"fn": <工具名>, "args": {<参数>}} → stdout 一行 JSON。
成功：{"ok": true, "data": <结果>}；失败：{"ok": false, "error": "..."}。

设计：
- 白名单分派（显式 if/elif，无 eval/exec/getattr 动态调用）——LLM 不能执行任意代码
- 参数 dict 直接 **args 传给工具函数——参数化调用，零代码注入
- 异常捕获进 error 字段（不 panic、exit 0）——Go 侧按 ok 字段判断

使用方式：
- 通过 `python3 tools/agent_cli.py` 执行 Python 工具
- stdin 传 JSON：{"fn": "<工具名>", "args": {<参数>}}
- stdout 返回 JSON：{"ok": true, "data": <结果>} 或 {"ok": false, "error": "..."}
"""
from __future__ import annotations

import json
import sys

from paipan import full_paipan, liunian
from duanyu import make_factors, make_liunian_factors, query


# 白名单：工具名 → 参数提取器（无 eval/exec/getattr 动态调用）
_DISPATCH = {
    "full_paipan": lambda a: full_paipan(a["time"], a["gender"],
                                         longitude=a.get("longitude"),
                                         correct=a.get("correct", True)),
    "liunian": lambda a: liunian(a["pan"], a["year"]),
    "make_factors": lambda a: make_factors(a["pan"]),
    "make_liunian_factors": lambda a: make_liunian_factors(a["pan"], a["liunian_pan"],
                                                           target=a.get("target", "配偶星"),
                                                           year=a.get("year", 0)),
    "query": lambda a: query(a["rule"], a["snapshots"]),
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
        print(json.dumps({"ok": False, "error": "empty stdin"}, ensure_ascii=False))
        return 0
    try:
        req = json.loads(raw)
        fn = req["fn"]
        args = req.get("args", {})
        if not isinstance(args, dict):
            raise ValueError("args must be an object")
        data = _dispatch(fn, args)
        print(json.dumps({"ok": True, "data": data}, ensure_ascii=False))
    except KeyError as e:
        print(json.dumps({"ok": False, "error": f"missing arg: {e}"}, ensure_ascii=False))
    except Exception as e:  # noqa: BLE001 —— 工具链异常（网络/参数/真值表）统一转错误
        print(json.dumps({"ok": False, "error": f"{type(e).__name__}: {e}"}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    sys.exit(main())
