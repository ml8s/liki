#!/usr/bin/env python3
"""agent_cli.py — liki 命理工具的 CLI 适配器。

协议：stdin 读一行 JSON {"fn": <工具名>, "args": {<参数>}} → stdout 一行 JSON。
成功：{"ok": true, "data": <结果>}；失败：{"ok": false, "error": "..."}。

设计：
- 白名单分派（显式 if/elif，无 eval/exec/getattr 动态调用）——LLM 不能执行任意代码
- 参数 dict 直接 **args 传给工具函数——参数化调用，零代码注入
- 异常捕获进 error 字段（不 panic、exit 0）——调用方按 ok 字段判断

使用方式：
- 通过 `python3 tools/agent_cli.py` 执行 Python 工具
- stdin 传 JSON：{"fn": "<工具名>", "args": {<参数>}}
- stdout 返回 JSON：{"ok": true, "data": <结果>} 或 {"ok": false, "error": "..."}
"""
from __future__ import annotations

import json
import sys

from paipan import full_paipan, city_coords, bond
from duanyu import query, yearly_range, calibrate


def _load_file_refs(args: dict) -> dict:
    """{"$file": path} 引用展开——大对象（pan/liunian_pan/snapshots）从 UTF-8 文件读，
    免 shell 内联转义（feedback fedd52aa：make_factors 手拼 pan 转义易错）。"""
    out = {}
    for k, v in args.items():
        if isinstance(v, dict) and "$file" in v:
            with open(v["$file"], encoding="utf-8") as f:
                loaded = json.load(f)
                # 自动解包 agent_cli 输出格式 {"ok": true, "data": {...}}
                if isinstance(loaded, dict) and "ok" in loaded and "data" in loaded:
                    loaded = loaded["data"]
                out[k] = loaded
        else:
            out[k] = v
    return out


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
    args = _load_file_refs(args)
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
