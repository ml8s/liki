"""Compatibility façade for the split yongshen resolver."""

from .context import MISSING as _MISSING
from .context import YongShenEvaluationError, load_policy
from .context import enrich_facts as _enrich_facts
from .evidence import evidence as _evidence
from .evidence import ge_ju_context as _ge_ju_context
from .evidence import published_yongshen
from .scoring import apply_derived as _apply_derived
from .selection import apply_then as _apply_then
from .selection import finalize as _finalize
from .selection import select_rule as _select_rule

__all__ = [
    "_MISSING",
    "YongShenEvaluationError",
    "load_policy",
    "published_yongshen",
    "resolve_yongshen",
]


def resolve_yongshen(bazi_chart: dict) -> dict:
    """从 canonical bazi.chart 事实裁决唯一 selected_yongshen。"""
    policy = load_policy()
    facts = _enrich_facts(bazi_chart or {}, policy)
    derived = _apply_derived(policy, facts)
    rule = _select_rule(policy, facts, derived)
    result = _apply_then(rule, facts, derived)
    _finalize(result, policy)
    result["ge_ju_context"] = _ge_ju_context(facts, derived, policy)
    result["evidence"] = _evidence(facts, derived, policy)
    result["policy_id"] = policy["policy_id"]
    return result
