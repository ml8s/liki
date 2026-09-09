"""六爻应期候选排序；只整理 engine 事实，不发明日期或结果。"""
from __future__ import annotations


BASE_PRIORITY = {
    "旬空填实": 82,
    "冲空": 74,
    "动爻逢值": 70,
    "动爻逢合": 62,
    "静爻逢冲": 58,
    "月破逢值": 54,
    "月破逢合": 46,
    "出月令": 38,
    "冲飞出伏": 66,
}

HORIZONS = {
    "wealth": "short_to_medium",
    "career": "medium",
    "relationship": "event_based",
    "study": "deadline_based",
    "lost_item": "short",
    "travel": "short",
    "legal": "process_based",
    "health_context": "blocked",
}


def _targets_hit(candidate: dict) -> int:
    text = " ".join(str(item) for item in candidate.get("basis", []))
    return 1 if "yong_shen" in text or candidate.get("position") else 0


def _topic_bonus(mechanism: str, topic: str | None) -> int:
    if not topic:
        return 0
    focus = {
        "wealth": {"moving-value", "void-fill", "yuepo-recovery", "changsheng-support"},
        "career": {"moving-value", "static-clash", "void-fill", "changsheng-support"},
        "study": {"moving-value", "void-fill", "static-clash", "changsheng-support"},
        "lost_item": {"hidden-release", "static-clash", "void-fill"},
    }.get(topic, set())
    normalized = mechanism.lower().replace(" ", "-")
    return 6 if normalized in focus else 0


def rank_timing_candidates(snapshot: dict, *, topic: str | None = None, limit: int = 6) -> dict:
    """按透明规则整理应期候选；score 只表示当前上下文优先级，不是概率。"""

    if topic == "health_context":
        return {
            "schema_version": "liuyao-timing-plan-v1",
            "blocked": True,
            "reason": "健康语境不给应期；先引导专业医疗或紧急帮助。",
            "candidates": [],
        }
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
        base = BASE_PRIORITY.get(mechanism, 30)
        bonus = 8 if _targets_hit(candidate) else 0
        bonus += _topic_bonus(mechanism, topic)
        penalty = 10 if candidate.get("confidence") != "candidate" else 0
        score = base + bonus - penalty
        priority = "high" if score >= 75 else "medium" if score >= 50 else "low"
        ranked.append({
            **candidate,
            "priority": priority,
            "rank_score": score,
            "horizon": HORIZONS.get(topic or "", "event_based"),
            "conclusion_scope": "conditional_timing_candidate_not_date",
            "required_check": "仍须核对旺衰、空破真假和现实时间范围",
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
