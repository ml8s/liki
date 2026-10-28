"""六爻 snapshot 创建：起卦、排盘、因子投影一次完成。"""

from app.divination.tools.divination_contracts import validate_document
from app.divination.tools.divination_safety import assess
from app.divination.tools.divination_snapshot import build_snapshot
from app.divination.tools.liuyao_board import factors as build_liuyao_factors
from app.divination.tools.liuyao_casting import qigua
from app.divination.tools.liuyao_conditions import evaluate as evaluate_conditions
from app.divination.tools.liuyao_matters import load_matter_table, resolve_matter
from app.divination.tools.liuyao_timing import rank_timing_candidates as plan_timing
from app.divination.tools.liuyao_topic_guidance import load_topic_table, project_topic_guidance
from app.guards import require_exactly_one, require_question

ALLOWED_MATTERS = set(load_matter_table())
ALLOWED_TOPICS = set(load_topic_table()["topics"])
MATTER_TO_TOPIC = {
    matter: matter for matter in ALLOWED_MATTERS if matter in ALLOWED_TOPICS
}


def create(
    *,
    question: str,
    longitude: float,
    casting: dict | None = None,
    standard_time: str = "",
    matter: str | None = None,
    yong_shen: str | None = None,
    perspective: str | None = None,
    topic: str | None = None,
) -> dict:
    """创建 immutable liuyao snapshot；LLM 后续只依赖该对象。

    casting 是 MCP 契约的判别联合字典：{mode: "auto"} /
    {mode: "coins", rounds: [...]} / {mode: "yaos", yaos: [...]}。
    """
    question = require_question(question)
    safety = assess(question)
    if casting is None:
        casting = {"mode": "auto"}
    if not isinstance(casting, dict) or "mode" not in casting:
        raise ValueError("casting must be an object with a mode discriminator")
    mode = casting["mode"]
    rounds = casting.get("rounds")
    yaos = casting.get("yaos")
    if mode not in {"auto", "coins", "yaos"}:
        raise ValueError("mode must be auto, coins, or yaos")
    # coins→rounds / yaos→yaos 由 casting 判别联合在 schema 边界保证，不重复校验。
    require_exactly_one(matter=matter, yong_shen=yong_shen)
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
        standard_time=standard_time,
        longitude=longitude,
    )
    projected = result["factors"]
    resolved_topic = topic or MATTER_TO_TOPIC.get(matter or "")
    snapshot = build_snapshot(
        method="liuyao",
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
            "safety_advisory": safety,
            "policy": {
                **projected["policy"],
                "immutable": True,
                "no_recast_without_new_event": True,
            },
        },
    )
    validate_document("liuyao_snapshot", snapshot)
    return snapshot
