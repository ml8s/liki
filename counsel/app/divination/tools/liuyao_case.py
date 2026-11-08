"""六爻 case 创建：起卦、排盘、因子投影一次完成。"""

from app.divination.tools.divination_case import build_case
from app.divination.tools.divination_contracts import validate_document
from app.divination.tools.divination_safety import assess
from app.divination.tools.liuyao_board import factors as build_liuyao_factors
from app.divination.tools.liuyao_casting import qigua
from app.divination.tools.liuyao_conditions import evaluate as evaluate_conditions
from app.divination.tools.liuyao_matters import load_matter_table, resolve_matter
from app.divination.tools.liuyao_timing import project_timing_plan
from app.divination.tools.liuyao_topic_guidance import load_topic_table, project_topic_guidance
from app.guards import require_question

MATTER_TABLE = load_matter_table()
ALLOWED_TOPICS = set(load_topic_table()["topics"])
MATTER_TO_TOPIC = {
    matter: matter for matter in MATTER_TABLE if matter in ALLOWED_TOPICS
}


def create(
    *,
    question: str,
    longitude: float,
    casting: dict | None = None,
    standard_time: str = "",
    focus: dict | None = None,
    topic: str | None = None,
) -> dict:
    """创建 immutable liuyao case；LLM 后续只依赖该对象。

    casting 是 MCP 契约的判别联合字典：{mode: "auto"} /
    {mode: "coins", rounds: [...]} / {mode: "yaos", yaos: [...]}。
    """
    question = require_question(question)
    safety = assess(question)
    if casting is None:
        casting = {"mode": "auto"}
    mode = casting["mode"]
    rounds = casting.get("rounds")
    yaos = casting.get("yaos")
    matter = focus.get("matter")
    yong_shen = focus.get("yong_shen")
    perspective = focus.get("perspective")
    if matter is not None:
        resolve_matter(matter, perspective=perspective)

    casting = qigua(mode=mode, rounds=rounds, yaos=yaos)
    result = build_liuyao_factors(
        casting=casting,
        matter=matter,
        yong_shen=yong_shen,
        perspective=perspective,
        question=question,
        standard_time=standard_time,
        longitude=longitude,
    )
    projected = result["factors"]
    resolved_topic = topic or MATTER_TO_TOPIC.get(matter or "")
    case = build_case(
        technique="liuyao",
        payload={
            "question": question,
            "focus": focus,
            "topic": resolved_topic,
            "casting": projected["casting"],
            "liuyao_chart": projected["liuyao_chart"],
            "selected_yongshen_context": projected["selected_yongshen_context"],
            "evidence": projected["evidence"],
            "chart_facts": projected["chart_facts"],
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
