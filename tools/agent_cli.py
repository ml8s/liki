#!/usr/bin/env python3
"""agent_cli.py — liki 规则引擎工具的 CLI 适配器（web agent 子进程入口）。

协议：stdin 读一行 JSON {"fn": <工具名>, "args": {<参数>}} → stdout 一行 JSON。
成功：{"ok": true, "data": <结果>}；失败：{"ok": false, "error": "..."}。

设计：
- 白名单分派（显式 if/elif，无 eval/exec/getattr 动态调用）——LLM 不能执行任意代码
- 参数 dict 直接 **args 传给工具函数——参数化调用，零代码注入
- 异常捕获进 error 字段（不 panic、exit 0）——Go 侧按 ok 字段判断

与本地 agent 的关系：本地 agent 有 shell，直接 import 工具链使用；
本文件只服务无 shell 的 web agent（Go exec python3 agent_cli.py）。
"""
from __future__ import annotations

import json
import sys

from paipan import full_paipan, liunian
from duanyu import make_factors, make_liunian_factors, query


def _dispatch(fn: str, args: dict):
    """白名单分派。args 为参数字典（由 schema 约束，此处直接传函数）。"""
    if fn == "full_paipan":
        return full_paipan(args["time"], args["gender"],
                           longitude=args.get("longitude"),
                           correct=args.get("correct", True))
    if fn == "liunian":
        return liunian(args["pan"], args["year"])
    if fn == "make_factors":
        return make_factors(args["pan"])
    if fn == "make_liunian_factors":
        return make_liunian_factors(args["pan"], args["liunian_pan"],
                                    target=args.get("target", "配偶星"),
                                    year=args.get("year", 0))
    if fn == "query":
        return query(args["rule"], args["snapshots"])
    raise ValueError(f"unknown tool: {fn}")


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
