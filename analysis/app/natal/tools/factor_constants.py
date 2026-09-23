"""因子常量加载：基础闭集、关系表、大类角色与紫微词表。"""
from __future__ import annotations

import json
import os

__all__ = ["load_constants"]

_CONST = None


def load_constants() -> dict:
    """懒加载 constants.json；进程内复用同一不可变约定数据。"""
    global _CONST
    if _CONST is None:
        path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "constants.json")
        with open(path, encoding="utf-8") as fh:
            _CONST = json.load(fh)
    return _CONST
