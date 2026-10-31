"""六爻应期候选排序；策略来自 JSON 表，只整理 engine 事实。"""

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


def rank_timing_candidates(snapshot: dict, *, topic: str | None = None, limit: int = 6) -> dict:
    """按表驱动 policy 整理应期候选；数组顺序表达上下文排序。"""
    rules = _rules()
    candidates = snapshot.get("timing_candidates", [])
    if not isinstance(candidates, list):
        raise ValueError("snapshot.timing_candidates must be an array")
    if limit < 1 or limit > 20:
        raise ValueError("limit must be 1-20")

    ranked = []
    for candidate in candidates:
        if not isinstance(candidate, dict) or not candidate.get("id"):
            continue
        mechanism = candidate.get("mechanism", "")
        order_score = rules["mechanism_order"].get(mechanism, rules["default_order"])
        order_score += rules["target_bonus"] if _targets_hit(candidate) else 0
        order_score += _topic_bonus(mechanism, str(candidate.get("id", "")), topic)
        ranked.append({
            **candidate,
            "horizon": rules["topic_horizon"].get(topic or "", rules["default_horizon"]),
            "conclusion_scope": "conditional_timing_candidate_not_date",
            "required_check": rules["required_check"],
            "_order_score": order_score,
        })

    ranked.sort(key=lambda item: (-item["_order_score"], str(item["id"])))
    ranked = ranked[:limit]
    for item in ranked:
        item.pop("_order_score", None)

    return {
        "blocked": False,
        "topic": topic,
        "candidates": ranked,
        "boundary": rules["boundary"],
        "policy": {
            "may_output": ["time_window", "condition", "watchpoint"],
            "forbidden": ["exact_destiny_date", "guarantee", "probability_percent"],
            "advice_scope": "traditional_reference_not_professional_advice",
        },
    }
