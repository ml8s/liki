"""Counsel decision methods.

All chart-producing input is a resolved subject. Counsel may compute judgment
facts from the resulting canonical charts, but it never accepts a large chart
back from the caller as a substitute for that subject.
"""

from concurrent.futures import ThreadPoolExecutor

from app.integrity.digest import sha256_json
from app.natal.tools.chart_integrity import with_natal_chart_digest
from app.natal.tools.chart_schema import validate_natal_chart
from app.natal.tools.charts import _bazi_chart, _ziwei_chart
from app.natal.tools.duanyu import match_rule
from app.natal.tools.factors import evaluate_factors
from app.natal.tools.yongshen import published_yongshen, resolve_yongshen


def _validate_subject(subject: object, action: str = "subject") -> dict:
    if not isinstance(subject, dict):
        raise ValueError(f"{action} must be an object")
    for key in ("birth_time", "birth_longitude", "gender"):
        if key not in subject:
            raise ValueError(f"{action}.{key} is required")
    return subject


def _build_natal_chart(subject: dict) -> dict:
    subject = _validate_subject(subject)
    with ThreadPoolExecutor(max_workers=2) as executor:
        bazi_future = executor.submit(_bazi_chart, subject)
        ziwei_future = executor.submit(_ziwei_chart, subject)
        bazi = bazi_future.result()
        ziwei = ziwei_future.result()
    natal_chart = with_natal_chart_digest({
        "subject": subject,
        "bazi": bazi,
        "ziwei": ziwei,
        "ziwei_daxian": ziwei.get("da_xian"),
    })
    validate_natal_chart(natal_chart, action="internal factor snapshot")
    return natal_chart


def bazi_yongshen(subject: dict) -> dict:
    """Resolve subject -> Bazi chart -> unique yongshen decision."""
    subject = _validate_subject(subject, "subject")
    bazi = _bazi_chart(subject)
    if "ri" not in bazi:
        raise ValueError("bazi.chart returned an invalid Bazi chart")
    source_digest = sha256_json(bazi)
    return published_yongshen(resolve_yongshen(bazi, source_digest))


def _selected_yongshen(natal_chart: dict, bazi_digest: str | None = None) -> dict:
    digest = bazi_digest if bazi_digest is not None else sha256_json(natal_chart["bazi"])
    return published_yongshen(resolve_yongshen(natal_chart["bazi"], digest))


def _factor_snapshots(subject: dict) -> dict[str, dict]:
    """Internal subject -> canonical charts -> both technique factor snapshots."""
    natal_chart = _build_natal_chart(subject)
    gender = natal_chart["subject"]["gender"]
    bazi_digest = sha256_json(natal_chart["bazi"])
    selected_yongshen = _selected_yongshen(natal_chart, bazi_digest)
    snapshots: dict[str, dict] = {}
    for technique, shushi in (("bazi", "bazi"), ("ziwei", "ziwei")):
        factor_snapshot = {
            "provenance": {
                "technique": technique,
                "subject": subject,
                "context": {"gender": gender},
                "natal_chart_digest": natal_chart["natal_chart_digest"],
            },
            "facts": evaluate_factors(gender, natal_chart, shushi=technique),
        }
        if technique == "bazi":
            factor_snapshot["selected_yongshen"] = {
                "selected_yongshen": selected_yongshen["selected_yongshen"],
                "status": selected_yongshen["status"],
            }
            factor_snapshot["provenance"].update({
                "yongshen_digest": selected_yongshen["yongshen_digest"],
                "policy_id": selected_yongshen["policy_id"],
                "bazi_chart_digest": selected_yongshen["bazi_chart_digest"],
            })
        snapshots[technique] = factor_snapshot
    return snapshots


def _factor_snapshot(subject: dict, technique: str) -> dict:
    """Internal one-side snapshot helper for focused internal callers."""
    return _factor_snapshots(subject)[technique]


def natal_query(subject: dict, life_topics: list[str]) -> dict:
    """Subject + topics -> combined Bazi and Ziwei natal assertions."""
    from app.natal.tools.analytics import (
        _deduplicate,
        _flatten_side_result,
        _load_routes,
        _public_assertion,
        _require_life_topics,
    )
    from app.natal.tools.factor_constants import load_constants

    factor_snapshots = _factor_snapshots(subject)
    context = factor_snapshots["bazi"]["provenance"]["context"] or {}
    side_labels = load_constants()["命理侧"]["标签"]
    selected = _require_life_topics(life_topics)
    routes = _load_routes()
    snapshots = {
        side_labels["bazi"]: factor_snapshots["bazi"]["facts"],
        side_labels["ziwei"]: factor_snapshots["ziwei"]["facts"],
        "context": context,
    }
    grouped: dict[str, list[dict]] = {"bazi": [], "ziwei": []}
    for topic, route in selected:
        for rule in route["natal_rules"]:
            result = match_rule(rule, snapshots)
            result["_rule"] = rule
            part, _count = _flatten_side_result(result, selected, routes)
            for technique in ("bazi", "ziwei"):
                grouped[technique].extend(part[technique])

    side_results = {}
    for technique in ("bazi", "ziwei"):
        side_results[technique] = {
            "assertions": [
                _public_assertion(item)
                for item in _deduplicate(grouped[technique])
            ],
        }
    return {
        "natal_chart_digest": factor_snapshots["bazi"]["provenance"]["natal_chart_digest"],
        **side_results,
    }


def natal_liunian_query(subject: dict, start_year: int,
                        year_count: int, life_topics: list[str]) -> dict:
    """Subject + start/count -> combined Bazi and Ziwei annual assertions."""
    from app.natal.tools.analytics import _analyze_liunian

    subject = _validate_subject(subject, "subject")
    natal_chart = _build_natal_chart(subject)
    selected_yongshen = _selected_yongshen(natal_chart)
    natal_chart["_selected_yongshen"] = {
        "selected": selected_yongshen["selected_yongshen"],
        "status": selected_yongshen["status"],
    }
    return _analyze_liunian(
        natal_chart,
        {
            "life_topics": life_topics,
            "start_year": start_year,
            "year_count": year_count,
        },
        validate_pan=False,
    )
