"""六爻应期计划投影；策略来自 JSON 表，只整理 engine 事实。"""

import json
from functools import lru_cache
from pathlib import Path

from app.divination.tools.liuyao_topic_guidance import load_topic_table

RULES_PATH = Path(__file__).with_name("liuyao_timing_rules.json")


@lru_cache
def _rules() -> dict:
    return json.loads(RULES_PATH.read_text(encoding="utf-8"))


def _targets_hit(candidate: dict) -> int:
    indicators = _rules()["target_indicators"]
    text = " ".join(str(item) for item in candidate.get("basis", []))
    return 1 if (
        any(token in text for token in indicators["basis_tokens"])
        or any(
            candidate.get(field)
            for field in indicators.get("fields", ())
        )
    ) else 0


def _topic_focus(topic: str | None) -> set[str]:
    if not topic:
        return set()
    return set(load_topic_table().get("topics", {}).get(topic, {}).get("timing_focus", []))


def _topic_bonus(mechanism: str, candidate_id: str, topic: str | None) -> int:
    focus = _topic_focus(topic)
    if not focus:
        return 0
    normalized = mechanism.lower().replace(" ", "-")
    prefixes = _rules()["focus_id_prefixes"]
    identifier = candidate_id.lower()
    return 6 if (
        normalized in focus
        or any(
            identifier.startswith(prefix)
            for token in focus
            for prefix in prefixes.get(token, ())
        )
    ) else 0


def project_timing_plan(factors: dict, *, topic: str | None = None, limit: int = 6) -> dict:
    """按表驱动 policy 投影应期计划；数组顺序表达上下文排序。"""
    rules = _rules()
    source_candidates = factors.get("timing_candidates", [])
    if not isinstance(source_candidates, list):
        raise ValueError("factors.timing_candidates must be an array")
    if limit < 1 or limit > 20:
        raise ValueError("limit must be 1-20")

    candidates = []
    for candidate in source_candidates:
        if not isinstance(candidate, dict) or not candidate.get("id"):
            continue
        mechanism = candidate.get("mechanism", "")
        order_score = rules["mechanism_order"].get(mechanism, rules["default_order"])
        order_score += rules["target_bonus"] if _targets_hit(candidate) else 0
        order_score += _topic_bonus(mechanism, str(candidate.get("id", "")), topic)
        candidates.append({
            "id": candidate["id"],
            "mechanism": candidate.get("mechanism", ""),
            "position": candidate.get("position"),
            "trigger_branch": candidate.get("trigger_branch"),
            "window": candidate.get("window"),
            "basis": candidate.get("basis", []),
            "horizon": rules["topic_horizon"].get(topic or "", rules["default_horizon"]),
            "conclusion_scope": "conditional_timing_candidate_not_date",
            "required_check": rules["required_check"],
            "_order_score": order_score,
        })

    candidates.sort(key=lambda item: (-item["_order_score"], str(item["id"])))
    candidates = candidates[:limit]
    for item in candidates:
        item.pop("_order_score", None)

    return {
        "candidates": candidates,
        "boundary": rules["boundary"],
    }
