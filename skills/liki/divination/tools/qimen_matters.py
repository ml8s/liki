"""奇门事象路由表；本层只做 matter → 显式用神查表。"""
from __future__ import annotations

import csv
from pathlib import Path

from qimen_errors import TableError


MATTERS_PATH = Path(__file__).with_name("data") / "qimen_matters.csv"
_MATTER_TABLE = None


def load_matter_table() -> dict[str, dict]:
    """加载 `{matter: name/symbols/basis}` 的封闭事象路由表。"""
    global _MATTER_TABLE
    if _MATTER_TABLE is not None:
        return _MATTER_TABLE
    result: dict[str, dict] = {}
    with MATTERS_PATH.open(encoding="utf-8-sig", newline="") as stream:
        for row in csv.DictReader(stream):
            matter = (row.get("matter") or "").strip()
            name = (row.get("name") or "").strip()
            basis = (row.get("basis") or "").strip()
            symbols = [
                item.strip()
                for item in (row.get("symbols") or "").split("、")
                if item.strip()
            ]
            if not matter or not name or not basis or not symbols:
                raise TableError(f"奇门事象表存在空字段: {row}")
            if len(symbols) != len(set(symbols)):
                raise TableError(f"奇门事象表存在重复符号: {matter}")
            if matter in result:
                raise TableError(f"奇门事象重复: {matter}")
            result[matter] = {"name": name, "symbols": symbols, "basis": basis}
    if not result:
        raise TableError("奇门事象表为空")
    _MATTER_TABLE = result
    return result


def resolve_matter(matter: str) -> dict:
    table = load_matter_table()
    if matter not in table:
        raise ValueError(f"unknown matter: {matter}")
    fact = table[matter]
    return {"matter": matter, **fact}
