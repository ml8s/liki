"""六爻问测：解读已有卦象事件并返回受控 Reading。"""

from app.divination.tools.liuyao_answer_core import build as build_answer_core
from app.divination.tools.liuyao_answer_core import validate_core as validate_answer_core
from app.divination.tools.liuyao_case import create as create_case


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
    if source_ref.startswith("analysis."):
        return "analysis"
    return "analysis"


def _reading(case: dict, answer: dict) -> dict:
    reading = {
        "technique": "liuyao",
        "evidence": _project_evidence(case, answer),
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
    event = arguments.get("event")
    if not isinstance(event, dict):
        raise ValueError("liuyao ask requires event")
    if set(arguments) != {"event"}:
        raise ValueError("liuyao ask accepts only event")
    case = create_case(event=event)
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
