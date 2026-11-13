"""Shared facts, policy identity, and path helpers for yongshen resolution."""

import json
from pathlib import Path
from typing import Any

MISSING = object()
POLICY_PATH = Path(__file__).resolve().parent / "policy.json"
_POLICY_CACHE: dict | None = None
NO_CANDIDATE = "无候选"
NO_SECONDARY = "无辅调候"


class YongShenEvaluationError(ValueError):
    """policy 表达式无法求值。"""


def load_policy() -> dict:
    """Load and validate the table-driven policy once per process."""
    global _POLICY_CACHE
    if _POLICY_CACHE is None:
        policy = json.loads(POLICY_PATH.read_text(encoding="utf-8"))
        from .schema import validate_policy

        validate_policy(policy)
        _POLICY_CACHE = policy
    return _POLICY_CACHE


def get_path(path: str, facts: dict, derived: dict, root: Any = None) -> Any:
    """读取 dotted path。root 非空时用于 select 的 item 作用域。"""
    if not isinstance(path, str) or not path:
        return MISSING
    if root is not None:
        if path.startswith("global."):
            return get_path(path[len("global.") :], facts, derived)
        if path.startswith("item."):
            return walk(root, path[len("item.") :])
        return walk(root, path)
    scope = {**facts, **derived}
    return walk(scope, path)


def walk(value: Any, path: str) -> Any:
    current = value
    for part in path.split("."):
        if current is MISSING:
            return MISSING
        if isinstance(current, dict):
            current = current.get(part, MISSING)
        elif isinstance(current, list) and part.isdigit():
            index = int(part)
            current = current[index] if 0 <= index < len(current) else MISSING
        else:
            return MISSING
    return current


def is_empty(value: Any) -> bool:
    return (
        value is MISSING
        or value is None
        or value == ""
        or value == []
        or value == {}
        or value in (NO_CANDIDATE, NO_SECONDARY)
    )


def length(value: Any) -> int:
    if value is MISSING or value is None:
        return 0
    try:
        return len(value)
    except TypeError:
        return 0


def materialize(value: Any) -> Any:
    """把求值哨兵转成可序列化空值，避免内部对象进入输出。"""
    if value is MISSING:
        return ""
    if isinstance(value, dict):
        return {key: materialize(item) for key, item in value.items()}
    if isinstance(value, list):
        return [materialize(item) for item in value]
    return value


def candidate(value: Any) -> Any:
    """Treat a missing/legacy-empty candidate as an explicit closure value."""
    value = materialize(value)
    return NO_CANDIDATE if value in ("", None) else value


def set_path(target: dict, key: str, value: Any) -> None:
    parts = key.split(".")
    current = target
    for part in parts[:-1]:
        child = current.get(part)
        if not isinstance(child, dict):
            child = {}
            current[part] = child
        current = child
    resolved = materialize(value)
    if key == "selected.yong" and resolved == "":
        resolved = None
    current[parts[-1]] = resolved


def enrich_facts(chart: dict, policy: dict) -> dict:
    """把 engine 事实合并成解释器可读的纯数据视图（不做命理判断）。"""
    fu_yi = chart.get("fu_yi") or {}
    tiao_hou = chart.get("tiao_hou") or {}
    ge_ju = chart.get("ge_ju") or {}
    from app.natal.tools.bazi_projection import build_atomic_facts

    atomic = build_atomic_facts(chart)
    counts = fu_yi.get("wuxing_count") or {}
    relations = policy["tables"].get("wuxing_relations") or {}
    day_master_element = atomic.get("day_master_element", "")
    if not day_master_element:
        day_master_element = (chart.get("ri") or {}).get("gan_wuxing", "")

    elements: list[dict] = []
    for item in chart.get("element_states") or []:
        if not isinstance(item, dict):
            continue
        wuxing = item.get("wuxing", "")
        relation = relations.get(wuxing) or {}
        elements.append(
            {
                **item,
                "count": counts.get(wuxing, 0),
                "is_day_master": wuxing == day_master_element,
                **relation,
            }
        )

    stems = [(chart.get(pillar) or {}).get("gan", "") for pillar in ("nian", "yue", "ri", "shi")]
    stems = [stem for stem in stems if stem]
    month_zhi = (chart.get("yue") or {}).get("zhi", "")
    day_master_stem = atomic.get("day_master_stem", "") or chart.get("ri", {}).get("gan", "")
    fu_yi = {
        **fu_yi,
        "season": (fu_yi.get("wang_shuai") or {}).get(day_master_element, ""),
    }

    return {
        "fu_yi": fu_yi,
        "tiao_hou": tiao_hou,
        "ge_ju": ge_ju,
        "atomic_facts": atomic,
        "element_states": elements,
        "ten_god_states": [item for item in (chart.get("ten_god_states") or []) if isinstance(item, dict)],
        "wuxing_count": counts,
        "month_zhi": month_zhi,
        "day_master_stem": day_master_stem,
        "day_master_element": day_master_element,
        "stems": stems,
    }
