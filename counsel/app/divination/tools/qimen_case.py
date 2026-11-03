"""奇门 case 创建：时间、地点、排盘与因子投影一次完成。"""

from app.divination.tools.divination_case import build_case
from app.divination.tools.divination_contracts import validate_document
from app.divination.tools.divination_safety import assess
from app.divination.tools.qimen_charts import city_coords, qimen_chart
from app.divination.tools.qimen_duanyu import query
from app.divination.tools.qimen_interpretations import assert_rule_compatibility, load_rule_table
from app.divination.tools.qimen_jinhan_factors import project as project_jinhan_factors
from app.divination.tools.qimen_projection import project as project_standard_factors
from app.divination.tools.qimen_projection import validate as validate_standard_factors
from app.divination.tools.qimen_projection_common import public_palace_keys
from app.guards import require_question
from app.integrity.digest import short_object_id

#
# 专占规则是否接受 matter / yong_shen 由 qimen_rules.csv 的 focus_policy 声明。




def _resolve_location(location: dict) -> tuple[str | None, float]:
    kind = location.get("kind")
    if kind == "city":
        city = location.get("city")
        if not isinstance(city, str) or not city.strip():
            raise ValueError("location.city must be non-empty")
        coords = city_coords(city)
        value = coords.get("longitude")
        if isinstance(value, bool) or not isinstance(value, (int, float)):
            raise ValueError("city.coords returned no longitude")
        return city.strip(), float(value)
    if kind == "longitude":
        longitude = location.get("longitude")
        if isinstance(longitude, bool) or not isinstance(longitude, (int, float)):
            raise ValueError("location.longitude must be a number")
        if not -180 <= float(longitude) <= 180:
            raise ValueError("longitude must be between -180 and 180")
        return None, float(longitude)
    raise ValueError("location must use kind=city or kind=longitude")


def _resolve_focus(focus: dict | None) -> tuple[str | None, list[str] | None]:
    """Extract the already contract-validated focus facts."""
    if focus is None:
        return None, None
    if focus["kind"] == "matter":
        return focus["matter"], None
    return None, focus["yong_shen"]



def create(
    *,
    question: str,
    location: dict,
    qi_ju_time: str = "",
    focus: dict | None = None,
    rule: str | None = None,
    method: dict[str, str],
    subject: dict | None = None,
) -> dict:
    """创建 immutable qimen case；LLM 后续只依赖该对象。"""
    question = require_question(question)
    safety = assess(question)
    effective_scope = method["scope"]
    effective_school = method["school"]

    matter, yong_shen = _resolve_focus(focus)
    if rule is not None:
        assert_rule_compatibility(rule, effective_scope, effective_school)
        focus_policy = load_rule_table()[rule]["focus_policy"]
        if focus_policy == "none" and (matter is not None or yong_shen is not None):
            raise ValueError(f"qimen rule {rule} does not accept matter or yong_shen")
        if focus_policy.startswith("required_matter:"):
            required_matter = focus_policy.split(":", 1)[1]
            if matter != required_matter:
                raise ValueError(f"qimen rule {rule} requires matter={required_matter}")
    resolved_city, resolved_longitude = _resolve_location(location)
    if not qi_ju_time:
        raise ValueError("qi_ju_time is required")

    chart_result = qimen_chart(
        qi_ju_time,
        resolved_longitude,
        matter=matter,
        yong_shen=yong_shen,
        subject=subject,
        method=method,
    )
    chart = chart_result["chart"]
    method_context = chart.get("method")
    if not isinstance(method_context, dict):
        raise ValueError("qimen chart lacks method context")

    if effective_school == "jinhan_yujing":
        factors = project_jinhan_factors(chart)
        special = None
        from app.divination.tools.qimen_judgment import empty as empty_judgments
        judgments = empty_judgments()
    else:
        factors = project_standard_factors(chart)
        validate_standard_factors(factors)
        special = query(rule, factors) if rule is not None else None
        from app.divination.tools.qimen_judgment import evaluate as evaluate_judgments
        judgments = evaluate_judgments(factors)
    if (
        method_context.get("scope") != factors.get("scope")
        or method_context.get("school") != factors.get("school")
    ):
        raise ValueError("qimen method context does not match factors")

    resolved_focus = None
    if chart_result.get("matter"):
        resolved_focus = {"kind": "matter", **chart_result["matter"]}
    elif focus is not None:
        resolved_focus = focus

    timing_candidates = [
        {
            "id": short_object_id("timing", candidate),
            "ying_qi_kind": candidate.get("ying_qi_kind") or None,
            "palace": candidate.get("palace") or candidate.get("gong"),
            "branch": candidate.get("branch"),
            "related_to": candidate.get("related_to"),
        }
        for candidate in factors.get("ying_qi", [])
    ]
    case_factors = {
        key: value
        for key, value in factors.items()
        if key not in {"scope", "school", "ying_qi"}
    }
    published_factors = public_palace_keys(case_factors)
    evidence_specs = (
        ("context-pillars", "factors.pillars", "context"),
        ("structure-geju", "factors.ge_ju", "structure"),
        ("relation-ri-shi", "factors.ri_shi_relation", "relation"),
        ("relation-gan", "factors.gan_interactions", "relation"),
        ("relation-men", "factors.men_interactions", "relation"),
        ("relation-xing", "factors.xing_interactions", "relation"),
        ("structure-stars", "factors.stars", "structure"),
        ("structure-doors", "factors.doors", "structure"),
        ("structure-spirits", "factors.spirits", "structure"),
        ("relation-star-spirit", "factors.spirit_star_relations", "relation"),
        ("relation-spirit-spirit", "factors.spirit_spirit_relations", "relation"),
        ("relation-spirit-door", "factors.spirit_door_relations", "relation"),
        ("structure-kong-wang", "factors.kong_wang", "structure"),
        ("structure-ma-xing", "factors.ma_xing", "structure"),
        ("structure-palace-strength", "factors.palace_wang_shuai", "structure"),
        ("focus-yongshen", "factors.yong_shen", "focus"),
        ("focus-year-gan", "factors.nian_gan_palace", "focus"),
        ("focus-year-gan-facts", "factors.nian_gan_palace_facts", "focus"),
        ("focus-yongshen-gong", "factors.yong_shen_palace", "focus"),
    )
    evidence = [
        {
            "id": f"case:{evidence_id}",
            "fact_ref": fact_ref,
            "layer": layer,
        }
        for evidence_id, fact_ref, layer in evidence_specs
        if published_factors.get(fact_ref.removeprefix("factors.")) not in (None, [], {}, "")
    ]
    case = build_case(
        technique="qimen",
        payload={
            "question": question,
            "qi_ju": {
                "city": resolved_city,
                "longitude": resolved_longitude,
                "qi_ju_time": qi_ju_time,
            },
            "focus": resolved_focus,
            "method": method,
            "evidence": evidence,
            "timing_candidates": timing_candidates,
            "factors": case_factors,
            "special": special,
            "judgments": {"results": judgments["results"]},
            "safety_advisory": safety,
            "policy": {
                "immutable": True,
                "no_rechart_without_new_event": True,
                "candidate_not_absolute_verdict": True,
                "forbid_downscaling_obvious_big_failure": True,
            },
        },
    )
    validate_document("qimen_case", case)
    return case
