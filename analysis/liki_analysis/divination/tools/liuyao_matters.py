"""六爻事项路由表加载；表内只做类别 → 默认用神，不做排盘或断卦。"""
from __future__ import annotations

import csv
from functools import lru_cache
from pathlib import Path

MATTERS_PATH = Path(__file__).with_name("data") / "liuyao_matters.csv"
_MATTER_TABLE: dict[str, dict] | None = None


@lru_cache
def load_matter_table() -> dict[str, dict]:
    """加载封闭事项表，并校验用神与视角规则。"""
    global _MATTER_TABLE
    if _MATTER_TABLE is not None:
        return _MATTER_TABLE
    result: dict[str, dict] = {}
    with MATTERS_PATH.open(encoding="utf-8-sig", newline="") as stream:
        for row in csv.DictReader(stream):
            matter = (row.get("matter") or "").strip()
            name = (row.get("name") or "").strip()
            notes = (row.get("notes") or "").strip()
            default = (row.get("yong_shen") or "").strip()
            requires = (row.get("requires_perspective") or "").strip().lower() == "true"
            male = (row.get("male_yong_shen") or "").strip()
            female = (row.get("female_yong_shen") or "").strip()
            if not matter or not name or not notes:
                raise ValueError(f"liuyao matter row lacks required fields: {row}")
            if requires and (not male or not female):
                raise ValueError(f"liuyao matter {matter} lacks gender-specific yong_shen")
            if not requires and (male or female):
                raise ValueError(f"liuyao matter {matter} cannot declare gender-specific yong_shen")
            if matter in result:
                raise ValueError(f"duplicate liuyao matter: {matter}")
            result[matter] = {
                "name": name,
                "yong_shen": default,
                "requires_perspective": requires,
                "male_yong_shen": male,
                "female_yong_shen": female,
                "notes": notes,
            }
    if not result:
        raise ValueError("liuyao matter table is empty")
    _MATTER_TABLE = result
    return result


def resolve_matter(matter: str, *, perspective: str | None = None) -> dict:
    """按表返回事项路由事实；relationship 必须显式确认视角。"""
    row = load_matter_table().get(matter)
    if row is None:
        raise ValueError(
            f"unknown matter: {matter}; valid: {', '.join(sorted(load_matter_table()))}"
        )
    fact = row.copy()
    if fact.pop("requires_perspective"):
        if perspective not in {"male", "female"}:
            raise ValueError(
                "relationship 需要 perspective=male/female；"
                "不按性别口径时请显式指定 yong_shen"
            )
        fact["yong_shen"] = row["male_yong_shen" if perspective == "male" else "female_yong_shen"]
    fact.pop("male_yong_shen", None)
    fact.pop("female_yong_shen", None)
    return fact
