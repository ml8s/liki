"""六爻一次解读包：排盘后自动绑定专题、应期、条件还原和报告骨架。"""
from __future__ import annotations

import hashlib
import json


from liuyao_paipan import chart as liuyao_chart
from divination_safety import assess, blocked_payload
from liuyao_conditions import evaluate as evaluate_conditions
from liuyao_report import template as report_template
from liuyao_timing import rank_timing_candidates as plan_timing
from liuyao_topic_guidance import project_topic_guidance


MATTER_TO_TOPIC = {
    "wealth": "wealth",
    "career": "career",
    "relationship": "marriage",
    "academic": "study",
    "lost_item": "lost_item",
    "travel": "travel",
    "legal_risk": "lawsuit",
}


def read(
    *,
    casting: dict | None = None,
    yaos: list[int] | None = None,
    solar_time: str | None = None,
    matter: str | None = None,
    yong_shen: str | None = None,
    perspective: str | None = None,
    question: str | None = None,
    topic: str | None = None,
) -> dict:
    """读取并组装一个不可变六爻 Reading；LLM 只基于该对象解释。"""
    question = question or ""
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload("liuyao-reading-v1", question, "liuyao")
    result = liuyao_chart(
        casting=casting,
        yaos=yaos,
        solar_time=solar_time,
        matter=matter,
        yong_shen=yong_shen,
        perspective=perspective,
        question=question,
    )
    snapshot = result["snapshot"]
    resolved_topic = topic or MATTER_TO_TOPIC.get(matter or "")
    topic_guidance = project_topic_guidance(snapshot, resolved_topic) if resolved_topic else None
    timing_plan = plan_timing(snapshot, topic=resolved_topic)
    condition_rules = evaluate_conditions(snapshot, topic=resolved_topic)
    report = report_template(snapshot)
    core = {
        "question": question or "",
        "matter": (result.get("matter") or {}).get("matter"),
        "topic": resolved_topic,
        "solar_time": result.get("solar_time"),
        "casting": result.get("casting"),
        "chart": result.get("chart"),
        "snapshot": snapshot,
        "topic_guidance": topic_guidance,
        "timing_plan": timing_plan,
        "condition_rules": condition_rules,
        "report_template": report,
    }
    digest_source = json.dumps(core, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return {
        "$schema": "liki:liuyao-reading-v1",
        "schema_version": "liuyao-reading-v1",
        "reading_id": hashlib.sha256(digest_source.encode("utf-8")).hexdigest(),
        "reading_digest": hashlib.sha256(
            json.dumps(core, ensure_ascii=False, sort_keys=True, separators=(",", ":")).encode("utf-8")
        ).hexdigest(),
        **core,
        "next_actions": [
            "LLM 只依据 snapshot、topic_guidance、timing_plan 和 condition_rules 填写 report_template",
            "liuyao_report validate 校验报告",
            "liuyao_audit 校验盘面硬事实",
            "liuyao_session create 固化首次结论",
        ],
    }


# Backward-compatible in-process alias only; not exposed to LLM.
interpret = read
