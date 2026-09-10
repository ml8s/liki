"""奇门 snapshot 创建：时间、地点、排盘与因子投影一次完成。"""
from __future__ import annotations

from divination_contracts import validate_document
from divination_rpc import server_time
from divination_safety import assess, blocked_payload
from divination_snapshot import build_snapshot
from qimen_duanyu import query
from qimen_projection import project as project_standard_factors
from qimen_projection import validate as validate_standard_factors
from qimen_jinhan_factors import project as project_jinhan_factors
from qimen_interpretations import assert_rule_compatibility
from qimen_paipan import city_coords, qimen_chart, solar_time


SCHEMA_VERSION = "qimen-snapshot-v3"

# 专占规则与 focus 的领域边界：
# - 失物 / 捕盗 / 捕亡 / 贼人画像只读日干、时干与盘面固有因子；
# - 走失人口通过 matter=missing_person 取六合用神。
NO_FOCUS_RULES = frozenset({
    "lost_property",
    "thief_capture",
    "capture_escape",
    "thief_profile",
})




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
    time: str | None = None,
    matter: str | None = None,
    yong_shen: list[str] | None = None,
    rule: str | None = None,
    scope: str | None = None,
    school: str | None = None,
    dingju_method: str | None = None,
    quarter_rule: str | None = None,
    base_dingju_method: str | None = None,
    dun_source: str | None = None,
    hour_boundary: str | None = None,
    birth_date: str | None = None,
) -> dict:
    """创建 immutable qimen snapshot；LLM 后续只依赖该对象。"""
    if not isinstance(question, str):
        raise ValueError("question must be non-empty text")
    question = question.strip()
    if not question:
        raise ValueError("question must be non-empty text")
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload(question=question, method="qimen", safety=safety)
    if len(question) < 2:
        raise ValueError("question must be at least 2 characters")
    effective_scope = scope or "hour"
    effective_school = school or "zhuanpan"
    if effective_school == "jinhan_yujing":
        if effective_scope != "day":
            raise ValueError("school jinhan_yujing requires scope=day")
        forbidden = (matter, yong_shen, rule, birth_date, dingju_method, quarter_rule, base_dingju_method, dun_source, hour_boundary)
        if any(value is not None for value in forbidden):
            raise ValueError("school jinhan_yujing does not support focus or advanced method fields")

    is_jinhan = effective_school == "jinhan_yujing"
    if (matter is not None) and (yong_shen is not None):
        raise ValueError("matter 与 yong_shen 互斥")
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
        if rule in NO_FOCUS_RULES and (matter is not None or yong_shen is not None):
            raise ValueError(f"qimen rule {rule} does not accept matter or yong_shen")
        if rule == "missing_person" and matter != "missing_person":
            raise ValueError("qimen rule missing_person requires matter=missing_person")
    resolved_city, resolved_longitude = _resolve_location(city, longitude)
    if not time:
        time = server_time()
    local = solar_time(time, resolved_longitude)
    solar = local.get("solar")
    if not isinstance(solar, str) or not solar:
        raise ValueError("tianwen.time returned no solar")

    chart_result = qimen_chart(
        solar,
        matter=matter,
        yong_shen=yong_shen,
        birth_date=birth_date,
        scope=effective_scope,
        school=effective_school,
        dingju_method=dingju_method,
        quarter_rule=quarter_rule,
        base_dingju_method=base_dingju_method,
        dun_source=dun_source,
        hour_boundary=hour_boundary,
    )
    chart = chart_result["chart"]
    method_context = chart.get("method")
    if not isinstance(method_context, dict):
        raise ValueError("qimen chart lacks method context")

    if effective_school == "jinhan_yujing":
        factors = project_jinhan_factors(chart)
        special = None
    else:
        factors = project_standard_factors(chart)
        validate_standard_factors(factors)
        special = query(rule, factors) if rule is not None else None
    if (
        method_context.get("scope") != factors.get("scope")
        or method_context.get("school") != factors.get("school")
    ):
        raise ValueError("qimen method context does not match factors")

    snapshot = build_snapshot(
        method="qimen",
        schema_version=SCHEMA_VERSION,
        payload={
            "question": question,
            "input": {
                "city": resolved_city,
                "longitude": resolved_longitude,
                "local_time": time,
                "solar_time": solar,
            },
            "matter": chart_result.get("matter"),
            "method_context": method_context,
            "factors": factors,
            "special": special,
            "snapshot_kind": "jinhan" if effective_school == "jinhan_yujing" else "standard",
            "policy": {
                "immutable": True,
                "no_rechart_without_new_event": True,
            },
        },
    )
    validate_document("qimen_snapshot", snapshot)
    return snapshot
