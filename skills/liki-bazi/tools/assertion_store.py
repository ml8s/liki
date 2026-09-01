"""断语长表索引：一次性加载元数据与条件，并按命理域聚合。"""
from __future__ import annotations

import csv
import os

from errors import AssertionRuleError

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ASSERTIONS_PATH = os.path.join(TOOLS_DIR, "assertions", "assertions.csv")
CONDITIONS_PATH = os.path.join(TOOLS_DIR, "assertions", "assertion_conditions.csv")

_INDEX = None


def _typed_expected(value: str) -> "int | str":
    value = (value or "").strip()
    try:
        return int(value)
    except ValueError:
        return value


def _load_index() -> dict:
    """构建 `{(side, rule): [evaluator row]}` 索引；进程内只读一次长表。"""
    global _INDEX
    if _INDEX is not None:
        return _INDEX
    if not os.path.exists(ASSERTIONS_PATH) or not os.path.exists(CONDITIONS_PATH):
        _INDEX = {}
        return _INDEX

    rows_by_key: dict[tuple[str, str], list[dict]] = {}
    row_by_id: dict[str, dict] = {}
    with open(ASSERTIONS_PATH, encoding="utf-8") as fh:
        for row in csv.DictReader(fh):
            side = (row.get("side") or "").strip()
            rule = (row.get("rule") or "").strip()
            assertion_id = (row.get("assertion_id") or "").strip()
            item = {
                "id": assertion_id,
                "事件": row.get("事件", ""),
                "约束": {},
                "结论": row.get("结论", ""),
                "依据": row.get("依据", ""),
                "经典原文": row.get("经典原文", ""),
            }
            rows_by_key.setdefault((side, rule), []).append(item)
            row_by_id[assertion_id] = item

    with open(CONDITIONS_PATH, encoding="utf-8") as fh:
        for row in csv.DictReader(fh):
            item = row_by_id.get((row.get("assertion_id") or "").strip())
            factor = (row.get("factor") or "").strip()
            if item and factor:
                item["约束"][factor] = _typed_expected(row.get("expected", ""))

    _INDEX = rows_by_key
    return _INDEX


def load_rule_table(name: str, required: bool = True) -> list[dict]:
    """加载 `{side}_{rule}` 对应的断语行。"""
    stem = name[:-4] if name.endswith(".csv") else name
    if "_" not in stem:
        raise AssertionRuleError(f"断语表名无效: {name}")
    side, rule = stem.split("_", 1)
    if side not in ("bazi", "ziwei"):
        raise AssertionRuleError(f"断语表 side 无效: {side}")
    rows = _load_index().get((side, rule), [])
    if not rows and required:
        raise FileNotFoundError(f"断语表不存在: {side}_{rule}.csv")
    return rows
