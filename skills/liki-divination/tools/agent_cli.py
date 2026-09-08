#!/usr/bin/env python3
"""liki-divination 奇门工具的 CLI 适配器。"""
from __future__ import annotations

import json
import os
import sys

from qimen_duanyu import query
from qimen_factors import project_qimen_snapshot
from qimen_interpretations import assert_rule_for_pan
from qimen_paipan import city_coords, qimen_chart, solar_time


def _configure_windows_stdio() -> None:
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
    print(json.dumps(payload, ensure_ascii=True))


def _query(args: dict) -> dict:
    assert_rule_for_pan(args["rule"], args["pan"])
    return query(args["rule"], project_qimen_snapshot(args["pan"]))


_DISPATCH = {
    "city_coords": lambda args: city_coords(args["city"]),
    "solar_time": lambda args: solar_time(args["time"], args["longitude"]),
    "qimen_chart": lambda args: qimen_chart(
        args["solar_time"],
        matter=args.get("matter"),
        yong_shen=args.get("yong_shen"),
        birth_date=args.get("birth_date"),
        scope=args.get("scope"),
        school=args.get("school"),
        dingju_method=args.get("dingju_method"),
        quarter_rule=args.get("quarter_rule"),
        base_dingju_method=args.get("base_dingju_method"),
        dun_source=args.get("dun_source"),
        hour_boundary=args.get("hour_boundary"),
    ),
    "query": _query,
}


def _dispatch(fn: str, args: dict):
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
        request = json.loads(raw)
        args = request.get("args", {})
        if not isinstance(args, dict):
            raise ValueError("args must be an object")
        _emit({"ok": True, "data": _dispatch(request["fn"], args)})
    except KeyError as error:
        _emit({"ok": False, "error": f"missing arg: {error}"})
    except Exception as error:  # noqa: BLE001 — 工具链错误统一转为 JSON error
        _emit({"ok": False, "error": f"{type(error).__name__}: {error}"})
    return 0


if __name__ == "__main__":
    _configure_windows_stdio()
    sys.exit(main())
