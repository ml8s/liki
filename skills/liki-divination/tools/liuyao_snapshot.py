"""六爻 snapshot 创建：起卦、排盘、因子投影一次完成。"""
from __future__ import annotations

from liuyao_casting import qigua
from liuyao_conditions import evaluate as evaluate_conditions
from divination_safety import assess, blocked_payload
from divination_snapshot import build_snapshot
from liuyao_paipan import chart as liuyao_chart
from liuyao_timing import rank_timing_candidates as plan_timing
from liuyao_topic_guidance import project_topic_guidance


SCHEMA_VERSION = "liuyao-snapshot-v3"
MATTER_TO_TOPIC = {
    "wealth": "wealth",
    "career": "career",
    "relationship": "marriage",
    "academic": "study",
    "lost_item": "lost_item",
    "travel": "travel",
    "legal_risk": "lawsuit",
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
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload(SCHEMA_VERSION, question, "liuyao")

    casting = qigua(mode=mode, rounds=rounds, yaos=yaos)["casting"]
    result = liuyao_chart(
        casting=casting,
        matter=matter,
        yong_shen=yong_shen,
        perspective=perspective,
        question=question,
    )
    projected = result["snapshot"]
    resolved_topic = topic or MATTER_TO_TOPIC.get(matter or "")

    return build_snapshot(
        method="liuyao",
        schema_version=SCHEMA_VERSION,
        payload={
            "question": projected.get("question", {"text": question}),
            "matter": (result.get("matter") or {}).get("matter"),
            "topic": resolved_topic,
            "solar_time": result.get("solar_time"),
            "casting": result.get("casting"),
            "board": projected.get("board"),
            "focus": projected.get("focus"),
            "evidence": projected.get("evidence"),
            "facts": projected.get("facts"),
            "timing_candidates": projected.get("timing_candidates"),
            "topic_guidance": (
                project_topic_guidance(projected, resolved_topic)
                if resolved_topic else None
            ),
            "timing_plan": plan_timing(projected, topic=resolved_topic),
            "condition_rules": evaluate_conditions(projected, topic=resolved_topic),
            "followup": projected.get("followup"),
            "policy": {
                **(projected.get("policy") or {}),
                "immutable": True,
                "no_recast_without_new_event": True,
            },
        },
    )
