"""Counsel decision methods.

All chart-producing input is a resolved subject. Counsel may compute judgment
facts from the resulting canonical charts, but it never accepts a large chart
back from the caller as a substitute for that subject.
"""

from app.integrity.digest import sha256_json
from app.natal.tools.chart_integrity import with_natal_chart_digest
from app.natal.tools.chart_schema import validate_natal_chart
from app.natal.tools.charts import _bazi_chart, _ziwei_chart, _ziwei_daxian
from app.natal.tools.duanyu import match_rule
from app.natal.tools.factors import evaluate_factors
from app.natal.tools.yongshen import published_yongshen, resolve_yongshen


def _factors_digest(factors: dict) -> str:
    return sha256_json(factors)


def _validate_subject(subject: object, action: str = "subject") -> dict:
    if not isinstance(subject, dict):
        raise ValueError(f"{action} must be an object")
    for key in ("birth_time", "birth_longitude", "gender"):
        if key not in subject:
            raise ValueError(f"{action}.{key} is required")
    return subject


def _factor_provenance(factors: dict) -> dict:
    provenance = factors.get("_provenance")
    if not isinstance(provenance, dict):
        raise ValueError("factors 缺少 natal_compute_factors 生成的 _provenance")
    return provenance


def verify_factors_digest(factors: dict, factors_digest: str) -> None:
    if not isinstance(factors_digest, str):
        raise ValueError("factors_digest 必须是 SHA-256 字符串")
    if factors_digest != _factors_digest(factors):
        raise ValueError("factors_digest mismatch；必须原样传递 natal_compute_factors 返回的 factors")


def _hour_calibration_facts(bazi: dict) -> dict:
    from app.natal.tools.bazi_projection import build_atomic_facts
    hour = bazi.get("shi") or {}
    atomic = build_atomic_facts(bazi)
    shen_sha = (bazi.get("shi") or {}).get("shen_sha") or []
    return {
        "hour_pillar": {"gan": hour.get("gan", ""), "zhi": hour.get("zhi", "")},
        "hour_stem_ten_god": atomic.get("hour_stem_ten_god", ""),
        "hour_shen_sha": [
            item.get("name") or item.get("xing") or ""
            for item in shen_sha if isinstance(item, dict)
        ],
    }


def _build_natal_chart(subject: dict) -> dict:
    subject = _validate_subject(subject)
    bazi = _bazi_chart(subject)
    ziwei = _ziwei_chart(subject)
    natal_chart = with_natal_chart_digest({
        "subject": subject,
        "bazi": bazi,
        "ziwei": ziwei,
        "ziwei_daxian": _ziwei_daxian(subject),
    })
    validate_natal_chart(natal_chart, action="natal_compute_factors")
    return natal_chart


def bazi_yongshen(subject: dict) -> dict:
    """Resolve subject -> Bazi chart -> unique yongshen decision."""
    subject = _validate_subject(subject, "subject")
    bazi = _bazi_chart(subject)
    if "ri" not in bazi:
        raise ValueError("bazi.chart returned an invalid Bazi chart")
    source_digest = sha256_json(bazi)
    return published_yongshen(resolve_yongshen(bazi, source_digest))


def natal_compute_factors(subject: dict, side: str) -> dict:
    """Resolve subject -> canonical charts -> one-side factor snapshot."""
    natal_chart = _build_natal_chart(subject)
    gender = natal_chart["subject"]["gender"]
    context = {"性别": gender}

    selected_yongshen = None
    if side == "bazi":
        source_digest = sha256_json(natal_chart["bazi"])
        selected_yongshen = published_yongshen(resolve_yongshen(
            natal_chart["bazi"], source_digest,
        ))
        factors = evaluate_factors(gender, natal_chart, shushi="bazi")
        factors["selected_yongshen"] = selected_yongshen
    elif side == "ziwei":
        factors = evaluate_factors(gender, natal_chart, shushi="ziwei")
    else:
        raise ValueError(f"side 只支持 bazi/ziwei，收到: {side!r}")

    factors["_provenance"] = {
        "side": side,
        "subject": subject,
        "context": context,
    }
    if side == "bazi":
        factors["_provenance"].update({
            "yongshen_digest": selected_yongshen["yongshen_digest"],
            "policy_id": selected_yongshen["policy_id"],
        })
    return {"factors": factors, "factors_digest": _factors_digest(factors)}


def natal_query(factors: dict, topics: list[str], factors_digest: str,
                side: str = "bazi") -> dict:
    """Factor snapshot + topics -> natal assertions."""
    from app.natal.tools.analytics import _flatten_side_result, _load_routes, _require_topics

    verify_factors_digest(factors, factors_digest)
    provenance = _factor_provenance(factors)
    if provenance.get("side") != side:
        raise ValueError(
            f"factors provenance side mismatch: expected {side!r}, "
            f"got {provenance.get('side')!r}"
        )
    if side not in ("bazi", "ziwei"):
        raise ValueError(f"natal_query side 只支持 bazi/ziwei，收到: {side!r}")
    context = provenance.get("context") or {}
    from app.natal.tools.factor_constants import load_constants

    side_labels = load_constants()["命理侧"]["标签"]
    selected = _require_topics(topics)
    routes = _load_routes()
    snapshots = {
        side_labels["bazi"]: factors if side == "bazi" else {},
        side_labels["ziwei"]: factors if side == "ziwei" else {},
        "context": context or {},
    }
    all_assertions: list[dict] = []
    seen: set[tuple[str, str]] = set()
    for topic, route in selected:
        for rule in route["natal_rules"]:
            result = match_rule(rule, snapshots)
            result["_rule"] = rule
            part, _count = _flatten_side_result(result, selected, routes, "natal")
            for item in part:
                key = (item["assertion_id"], item["side"])
                if key not in seen:
                    seen.add(key)
                    all_assertions.append(item)
    return {"assertions": all_assertions}


def natal_period_query(factors: dict, factors_digest: str, time_scope: dict,
                       topics: list[str], subject: dict,
                       side: str = "bazi") -> dict:
    """Factor snapshot + time scope + subject -> period assertions."""
    from app.natal.tools.analytics import _analyze_periods

    verify_factors_digest(factors, factors_digest)
    provenance = _factor_provenance(factors)
    if provenance.get("side") != side:
        raise ValueError(
            f"factors provenance side mismatch: expected {side!r}, "
            f"got {provenance.get('side')!r}"
        )
    _validate_subject(subject, "subject")
    if provenance.get("subject") != subject:
        raise ValueError("subject 与 natal_compute_factors 因子来源不一致")
    if side not in ("bazi", "ziwei"):
        raise ValueError(f"natal_period_query side 只支持 bazi/ziwei，收到: {side!r}")

    if side == "bazi":
        selected_yongshen = factors.get("selected_yongshen")
        if not isinstance(selected_yongshen, dict):
            raise ValueError("factors 缺少 selected_yongshen；必须原样传递 factors")
        if provenance.get("yongshen_digest") != selected_yongshen.get("yongshen_digest"):
            raise ValueError("factors _provenance 的用神来源与 selected_yongshen 不一致")
    natal_chart = _build_natal_chart(subject)
    if side == "bazi":
        natal_chart["_selected_yongshen"] = selected_yongshen

    result = _analyze_periods(
        natal_chart,
        {"topics": topics, "time_scope": time_scope or {}},
        validate_pan=False,
    )
    for scope in result["periods"]:
        scope["assertions"] = [
            item for item in scope["assertions"] if item.get("side") == side
        ]
        if "counts" in scope:
            scope["counts"]["returned"] = len(scope["assertions"])
    if side == "bazi":
        result["calibration_facts"] = _hour_calibration_facts(natal_chart["bazi"])
    return result
