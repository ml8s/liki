"""一次性六爻问测：创建内部 case 并返回受控 Reading。"""

from app.divination.tools.liuyao_answer_core import build as build_answer_core
from app.divination.tools.liuyao_answer_core import validate_core as validate_answer_core
from app.divination.tools.liuyao_case import create as create_case
from app.guards import require_question

_DOMAIN_EVIDENCE = {
    "general": ["general", "timing", "obstacle"],
    "self": ["self", "timing", "obstacle"],
    "other": ["other_person", "timing"],
    "career": ["career", "obstacle", "timing"],
    "wealth": ["wealth", "obstacle", "timing"],
    "relationship": ["relationship", "timing"],
    "study": ["study", "timing"],
    "home": ["home", "timing"],
    "travel": ["travel", "obstacle", "timing"],
    "family": ["family", "timing"],
    "competition": ["competition", "obstacle", "timing"],
    "children": ["children", "timing"],
    "lost_item": ["lost_item", "timing"],
    "legal": ["legal", "obstacle", "timing"],
}

_FOLLOWUP_POLICY = {
    "same_matter": {
        "clarification": "allowed_from_reading",
        "timing": "allowed_from_reading",
        "obstacle": "allowed_from_reading",
    },
    "new_domain": "requires_new_case",
    "another_matter": "requires_new_case",
    "recast_same_matter": "forbidden",
    "beyond_returned_evidence": "answer_insufficient",
}


def _focus_value(focus: dict) -> str | None:
    if focus.get("kind") == "matter":
        return focus.get("matter")
    return focus.get("yong_shen")


def _build_scope(case: dict) -> dict:
    focus = case["focus"]
    value = _focus_value(focus) or "general"
    included = list(_DOMAIN_EVIDENCE.get(value, ["general", "timing", "obstacle"]))
    known_domains = {
        domain
        for domains in _DOMAIN_EVIDENCE.values()
        for domain in domains
    }
    return {
        "focus_kind": focus.get("kind"),
        "focus_value": value,
        "topic": case.get("topic"),
        "included_domains": included,
        "excluded_domains": sorted(known_domains - set(included)),
        "followup_policy": _FOLLOWUP_POLICY,
    }


def _evidence_index(case: dict) -> dict[str, tuple[str, dict]]:
    evidence = case["evidence"]
    index: dict[str, tuple[str, dict]] = {}
    for layer in ("primary", "secondary", "reference"):
        for item in evidence.get(layer, []):
            if isinstance(item, dict) and item.get("id"):
                index[item["id"]] = (layer, item)
    return index


def _project_evidence(case: dict, answer: dict) -> list[dict]:
    index = _evidence_index(case)
    projected = []
    for ref in answer.get("primary_evidence_refs", []):
        layer, item = index[ref]
        projected.append({"id": ref, "source_kind": _source_kind(item.get("source_ref", "")), "layer": layer, "fact": item.get("fact", item)})
    for ref in answer.get("secondary_evidence_refs", []):
        layer, item = index[ref]
        projected.append({"id": ref, "source_kind": _source_kind(item.get("source_ref", "")), "layer": layer, "fact": item.get("fact", item)})
    return projected


def _source_kind(source_ref: str) -> str:
    if source_ref.startswith("liuyao_chart."):
        return "chart"
    if source_ref.startswith("selected_yongshen_context."):
        return "yongshen_context"
    if source_ref.startswith("chart_facts."):
        return "chart_facts"
    return "chart_facts"


def _reading(case: dict, answer: dict) -> dict:
    reading = {
        "technique": "liuyao",
        "question": case["question"],
        "scope": _build_scope(case),
        "casting": case["casting"],
        "chart_view": case["liuyao_chart"],
        "selected_yongshen_context": case["selected_yongshen_context"],
        "evidence": _project_evidence(case, answer),
        "conflicts": answer["conflicts"],
        "timing_candidates": case["timing_plan"]["candidates"],
        "condition_candidates": [
            {key: value for key, value in item.items() if key != "applies"}
            for item in case["condition_candidates"]
            if item.get("applies") is True
        ],
        "interpretation": {
            "headline": answer["headline"],
            "verdict": answer["verdict"],
            "action": answer["action"],
            "boundary": answer["boundary"],
            "disclaimer": answer["disclaimer"],
        },
        "safety_advisory": answer["safety_advisory"],
    }
    return reading


def ask(**arguments) -> dict:
    """创建一次六爻内部 case，并返回自包含受控 Reading。"""
    require_question(arguments["question"])
    case = create_case(**arguments)
    answer = build_answer_core(case)
    counts = (
        len(answer.get("primary_evidence_refs", [])),
        len(answer.get("secondary_evidence_refs", [])),
        len(answer.get("timing_refs", [])),
    )
    answer.update({
        "headline": "六爻卦例条件性解读",
        "verdict": (
            f"当前卦例提供 {counts[0]} 项主要因子、{counts[1]} 项辅助因子"
            f"和 {counts[2]} 个应期候选；只能表达条件性倾向，不构成结果承诺。"
        ),
        "action": "先核对主要因子对应的现实条件，再决定推进或调整节奏。",
    })
    audit = validate_answer_core(answer, case)
    if not audit.get("accepted"):
        errors = "; ".join(
            str(item.get("reason") or item.get("field") or item)
            for item in audit.get("errors", [])
        )
        raise ValueError(f"liuyao reading failed validation: {errors}")
    answer["safety_advisory"] = case["safety_advisory"]
    return _reading(case, answer)
