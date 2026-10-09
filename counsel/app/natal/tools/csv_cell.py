"""CSV 单元格解析：expected 值既可能是整数，也可能是原样字符串。"""
from __future__ import annotations


def typed_expected(value: str | None):
    value = (value or "").strip()
    try:
        return int(value)
    except ValueError:
        return value
