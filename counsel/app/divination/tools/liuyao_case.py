"""六爻 case 组装：排盘、因子投影一次完成；本模块不创建新卦。"""

from app.divination.tools.divination_case import build_case
from app.divination.tools.divination_contracts import validate_document
from app.divination.tools.divination_safety import assess
from app.divination.tools.liuyao_board import factors as build_liuyao_factors
from app.divination.tools.liuyao_conditions import evaluate as evaluate_conditions
from app.divination.tools.liuyao_timing import project_timing_plan
from app.divination.tools.liuyao_topic_guidance import load_topic_table, project_topic_guidance
from app.guards import require_object, require_question

ALLOWED_TOPICS = set(load_topic_table()["topics"])


def create(
    *,
    event: dict,
    topic: str | None = None,
) -> dict:
    """组装 immutable liuyao case；本方法不创建新卦。"""
    require_object(event, "event")
    question = require_question(event.get("question"))
    casting = event.get("casting")
    if not isinstance(casting, dict):
        raise ValueError("event.casting must be an object")
    safety = assess(question)
    if casting.get("mode") == "auto":
        raise ValueError("liuyao case cannot create a new casting event")
    resolved_topic = topic if topic in ALLOWED_TOPICS else None
    result = build_liuyao_factors(
        event=event,
    )
    projected = result["factors"]
    case = build_case(
        technique="liuyao",
        payload={
            "case_version": "liuyao-case-v3",
            "event": event,
            "topic": resolved_topic,
            "chart": projected["liuyao_chart"],
            "evidence": projected["evidence"],
            "analysis": projected["analysis"],
            "topic_guidance": (
                project_topic_guidance(projected, resolved_topic)
                if resolved_topic else None
            ),
            "timing_plan": project_timing_plan(projected, topic=resolved_topic),
            "condition_candidates": (
                evaluate_conditions(projected, topic=resolved_topic)["results"]
            ),
            "safety_advisory": safety,
            "policy": {
                **projected["policy"],
                "immutable": True,
                "no_recast_without_new_event": True,
                "timing_may_output": ["time_window", "condition", "watchpoint"],
                "timing_forbidden": [
                    "exact_destiny_date", "guarantee", "probability_percent",
                ],
                "timing_advice_scope": "traditional_reference_not_professional_advice",
                "must_restore_hidden_conditions": True,
                "forbid_absolute_old_phrases": True,
                "reference_may_not_override_primary": True,
            },
        },
    )
    validate_document("liuyao_case", case)
    return case
