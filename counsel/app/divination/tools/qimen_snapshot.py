"""奇门 snapshot 创建：时间、地点、排盘与因子投影一次完成。"""

from app.divination.tools.divination_contracts import validate_document
from app.divination.tools.divination_rpc import server_time
from app.divination.tools.divination_safety import assess
from app.divination.tools.divination_snapshot import build_snapshot
from app.divination.tools.qimen_charts import city_coords, qimen_chart
from app.divination.tools.qimen_duanyu import query
from app.divination.tools.qimen_interpretations import assert_rule_compatibility, load_rule_table
from app.divination.tools.qimen_jinhan_factors import project as project_jinhan_factors
from app.divination.tools.qimen_projection import project as project_standard_factors
from app.divination.tools.qimen_projection import validate as validate_standard_factors
from app.guards import require_mutually_exclusive, require_question

#
# 专占规则是否接受 matter / yong_shen 由 qimen_rules.csv 的 focus_policy 声明。




def _resolve_location(city: str | None, longitude: float | None) -> tuple[str | None, float]:
    has_city = city is not None
    has_longitude = longitude is not None
    if has_city == has_longitude:
        raise ValueError("provide exactly one of city or longitude")
    if has_city and not city.strip():
        raise ValueError("city must be non-empty")
    if longitude is not None:
        if isinstance(longitude, bool) or not isinstance(longitude, (int, float)):
            raise ValueError("longitude must be a number")
        if not -180 <= float(longitude) <= 180:
            raise ValueError("longitude must be between -180 and 180")
        return None, float(longitude)
    coords = city_coords(city or "")
    value = coords.get("longitude")
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ValueError("city.coords returned no longitude")
    return city.strip(), float(value)


def create(
    *,
    question: str,
    city: str | None = None,
    longitude: float | None = None,
    instant: str | None = None,
    timezone: str = "",
    matter: str | None = None,
    yong_shen: list[str] | None = None,
    rule: str | None = None,
    method: dict[str, str],
    birth_instant: str | None = None,
    birth_timezone: str | None = None,
) -> dict:
    """创建 immutable qimen snapshot；LLM 后续只依赖该对象。"""
    question = require_question(question)
    safety = assess(question)
    if not isinstance(method, dict):
        raise ValueError("method must be an explicit qimen method object")
    effective_scope = method.get("scope", "")
    effective_school = method.get("school", "")
    if effective_school == "jinhan_yujing":
        forbidden = (matter, yong_shen, rule, birth_instant, birth_timezone)
        if any(value is not None for value in forbidden):
            raise ValueError("school jinhan_yujing does not support focus or advanced method fields")

    is_jinhan = effective_school == "jinhan_yujing"
    require_mutually_exclusive(matter=matter, yong_shen=yong_shen)
    has_focus = (matter is not None) or (yong_shen is not None)
    quarter_allows_no_focus = effective_scope == "quarter"
    if is_jinhan:
        pass
    elif not rule and not has_focus and not quarter_allows_no_focus:
        raise ValueError("ordinary qimen requires exactly one of matter or yong_shen; specialized rule may omit both")

    if yong_shen is not None and not yong_shen:
        raise ValueError("yong_shen cannot be empty")
    if rule is not None:
        assert_rule_compatibility(rule, effective_scope, effective_school)
        focus_policy = load_rule_table()[rule]["focus_policy"]
        if focus_policy == "none" and (matter is not None or yong_shen is not None):
            raise ValueError(f"qimen rule {rule} does not accept matter or yong_shen")
        if focus_policy.startswith("required_matter:"):
            required_matter = focus_policy.split(":", 1)[1]
            if matter != required_matter:
                raise ValueError(f"qimen rule {rule} requires matter={required_matter}")
    resolved_city, resolved_longitude = _resolve_location(city, longitude)
    if not timezone:
        raise ValueError("timezone is required (IANA)")
    if not instant:
        instant = server_time(timezone)["instant"]

    chart_result = qimen_chart(
        instant,
        timezone,
        resolved_longitude,
        matter=matter,
        yong_shen=yong_shen,
        birth_instant=birth_instant,
        birth_timezone=birth_timezone,
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

    snapshot = build_snapshot(
        method="qimen",
        payload={
            "question": question,
            "input": {
                "city": resolved_city,
                "longitude": resolved_longitude,
                "instant": instant, "timezone": timezone,
            },
            "matter": chart_result.get("matter"),
            "method_context": method_context,
            "factors": factors,
            "special": special,
            "judgments": judgments,
            "snapshot_kind": "jinhan" if effective_school == "jinhan_yujing" else "standard",
            "safety_advisory": safety,
            "policy": {
                "immutable": True,
                "no_rechart_without_new_event": True,
            },
        },
    )
    validate_document("qimen_snapshot", snapshot)
    return snapshot
