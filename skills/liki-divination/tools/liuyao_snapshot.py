"""六爻 snapshot 创建：起卦、排盘、因子投影一次完成。"""
from __future__ import annotations

from divination_contracts import validate_document
from divination_safety import assess, blocked_payload
from divination_snapshot import build_snapshot
from liuyao_casting import qigua
from liuyao_conditions import evaluate as evaluate_conditions
from liuyao_matters import load_matter_table, resolve_matter
from liuyao_paipan import factors as build_liuyao_factors
from liuyao_timing import rank_timing_candidates as plan_timing
from liuyao_topic_guidance import project_topic_guidance


SCHEMA_VERSION = "liuyao-snapshot-v4"
ALLOWED_MATTERS = set(load_matter_table())
ALLOWED_TOPICS = {
    "wealth", "career", "marriage", "study", "lost_item", "travel",
    "lawsuit", "health_context",
}
MATTER_TO_TOPIC = {
    "wealth": "wealth",
    "career": "career",
    "relationship": "marriage",
    "study": "study",
    "lost_item": "lost_item",
    "travel": "travel",
    "legal": "lawsuit",
}


def create(
    *,
    question: str,
    mode: str = "auto",
    rounds: list[list[str]] | None = None,
    yaos: list[int] | None = None,
    matter: str | None = None,
    yong_shen: str | None = None,
    perspective: str | None = None,
    topic: str | None = None,
) -> dict:
    """创建 immutable liuyao snapshot；LLM 后续只依赖该对象。"""
    if not isinstance(question, str):
        raise ValueError("question must be non-empty text")
    question = question.strip()
    if not question:
        raise ValueError("question must be non-empty text")
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload(question=question, method="liuyao", safety=safety)
    if len(question) < 2:
        raise ValueError("question must be at least 2 characters")
    if mode not in {"auto", "coins", "yaos"}:
        raise ValueError("mode must be auto, coins, or yaos")
    if (matter is None) == (yong_shen is None):
        raise ValueError("provide exactly one of matter or yong_shen")
    if matter is not None and matter not in ALLOWED_MATTERS:
        raise ValueError(f"unknown matter: {matter}")
    if perspective is not None and matter != "relationship":
        raise ValueError("perspective is only valid with matter=relationship")
    if topic is not None and topic not in ALLOWED_TOPICS:
        raise ValueError(f"unknown topic: {topic}")
    if matter is not None:
        resolve_matter(matter, perspective=perspective)

    casting = qigua(mode=mode, rounds=rounds, yaos=yaos)
    result = build_liuyao_factors(
        casting=casting,
        matter=matter,
        yong_shen=yong_shen,
        perspective=perspective,
        question=question,
    )
    projected = result["factors"]
    resolved_topic = topic or MATTER_TO_TOPIC.get(matter or "")
    snapshot = build_snapshot(
        method="liuyao",
        schema_version=SCHEMA_VERSION,
        payload={
            "question": projected["question"],
            "matter": (result.get("matter") or {}).get("matter"),
            "topic": resolved_topic,
            "casting": projected["casting"],
            "board": projected["board"],
            "focus": projected["focus"],
            "evidence": projected["evidence"],
            "facts": projected["facts"],
            "timing_candidates": projected["timing_candidates"],
            "topic_guidance": (
                project_topic_guidance(projected, resolved_topic)
                if resolved_topic else None
            ),
            "timing_plan": plan_timing(projected, topic=resolved_topic),
            "condition_rules": evaluate_conditions(projected, topic=resolved_topic),
            "policy": {
                **projected["policy"],
                "immutable": True,
                "no_recast_without_new_event": True,
            },
        },
    )
    validate_document("liuyao_snapshot", snapshot)
    return snapshot
