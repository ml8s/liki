"""六爻应期候选排序；策略来自 JSON 表，只整理 engine 事实。"""
from __future__ import annotations

import json
from functools import lru_cache
from pathlib import Path

RULES_PATH = Path(__file__).with_name("liuyao_timing_rules.json")
TOPIC_PATH = Path(__file__).with_name("liuyao_topic_methods.json")


@lru_cache
def _rules() -> dict:
    return json.loads(RULES_PATH.read_text(encoding="utf-8"))


@lru_cache
def _topic_methods() -> dict:
    return json.loads(TOPIC_PATH.read_text(encoding="utf-8"))


def _targets_hit(candidate: dict) -> int:
    indicators = _rules()["target_indicators"]
    text = " ".join(str(item) for item in candidate.get("basis", []))
    return 1 if (
        any(token in text for token in indicators["basis_tokens"])
        or any(candidate.get(field) for field in indicators["fields"])
    ) else 0


def _topic_focus(topic: str | None) -> set[str]:
    if not topic:
        return set()
    return set(_topic_methods().get("topics", {}).get(topic, {}).get("timing_focus", []))


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
    """按表驱动策略整理应期候选；score 是上下文优先级，不是概率。"""
    rules = _rules()
    blocked = rules["blocked_topics"].get(topic or "")
    if blocked:
        return {
            "schema_version": "liuyao-timing-plan-v1",
            "blocked": True,
            "reason": blocked["reason"],
            "candidates": [],
        }
    candidates = snapshot.get("timing_candidates", [])
    if not isinstance(candidates, list):
        raise ValueError("snapshot.timing_candidates must be an array")
    if limit < 1 or limit > 20:
        raise ValueError("limit must be 1-20")

    thresholds = rules["priority_thresholds"]
    ranked = []
    for candidate in candidates:
        if not isinstance(candidate, dict) or not candidate.get("id"):
            continue
        mechanism = candidate.get("mechanism", "")
        score = rules["mechanism_priority"].get(mechanism, rules["default_priority"])
        score += rules["target_bonus"] if _targets_hit(candidate) else 0
        score += _topic_bonus(mechanism, str(candidate.get("id", "")), topic)
        score -= rules["non_candidate_penalty"] if candidate.get("confidence") != "candidate" else 0
        priority = (
            "high" if score >= thresholds["high"]
            else "medium" if score >= thresholds["medium"]
            else "low"
        )
        ranked.append({
            **candidate,
            "priority": priority,
            "rank_score": score,
            "horizon": rules["topic_horizon"].get(topic or "", rules["default_horizon"]),
            "conclusion_scope": "conditional_timing_candidate_not_date",
            "required_check": rules["required_check"],
        })
    ranked.sort(key=lambda item: (-item["rank_score"], str(item["id"])))
    ranked = ranked[:limit]
    for index, item in enumerate(ranked, 1):
        item["rank"] = index
    return {
        "schema_version": "liuyao-timing-plan-v1",
        "blocked": False,
        "topic": topic,
        "candidates": ranked,
        "policy": {
            "may_output": ["time_window", "condition", "watchpoint"],
            "forbidden": ["exact_destiny_date", "guarantee", "probability_percent"],
        },
    }
