"""六爻 case 组装：排盘、因子投影一次完成；本模块不创建新卦。"""

from app.divination.tools.divination_case import build_case
from app.divination.tools.divination_contracts import validate_document
from app.divination.tools.liuyao_board import factors as build_liuyao_factors
from app.divination.tools.liuyao_conditions import evaluate as evaluate_conditions
from app.divination.tools.liuyao_timing import project_timing_plan


def _allow_advisory() -> dict:
    # ask 不读取 question；现实风险识别由持有用户问题的调用方完成。
    return {
        "status": "allow",
        "category": None,
        "support_first": False,
        "message": "",
        "guidance": "",
    }


def create(*, event: dict) -> dict:
    """组装 immutable liuyao case；本方法不创建新卦，也不读取 question/focus。"""
    if not isinstance(event, dict):
        raise ValueError("event must be an object")
    casting = event.get("casting")
    if not isinstance(casting, dict):
        raise ValueError("event.casting must be an object")
    if casting.get("mode") == "auto":
        raise ValueError("liuyao case cannot create a new casting event")
    result = build_liuyao_factors(event=event)
    projected = result["factors"]
    case = build_case(
        technique="liuyao",
        payload={
            "case_version": "liuyao-case-v3",
            "event": event,
            "chart": projected["liuyao_chart"],
            "evidence": projected["evidence"],
            "timing_plan": project_timing_plan(projected),
            "condition_candidates": evaluate_conditions(projected)["results"],
            "safety_advisory": _allow_advisory(),
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
