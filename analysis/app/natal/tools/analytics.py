"""面向用户问题的本命分析服务。

TopicRouter 只做受控配置投影：topic → 命理规则 + 断语领域。这里不解释命理，
不修改断语真值表；输出统一使用英文契约字段，中文只保留在展示内容中。
"""
from __future__ import annotations

import json
import os
from datetime import datetime
from functools import lru_cache
from typing import Any

from chart_token import chart_id, decode_chart_ref, encode_chart_ref
from duanyu import (
    CURRENT_LIMIT_RULES,
    query,
    resolve_current_year,
    yearly_range,
)
from errors import LikiToolError
from factor_constants import load_constants
from paipan import RPCError, city_coords, full_paipan

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROUTE_PATH = os.path.join(TOOLS_DIR, "topic_routes.json")

def _side_pairs() -> list[tuple[str, str]]:
    config = load_constants()["命理侧"]
    return [(label, code) for code, label in config["标签"].items()]


class TopicRouteError(LikiToolError, ValueError):
    """topic 或分析参数不在受控契约内。"""


class BirthInputError(LikiToolError, ValueError):
    """出生输入或时间精度契约错误。"""


class LocationError(LikiToolError, ValueError):
    """出生地解析失败。"""


@lru_cache(maxsize=1)
def _load_routes() -> dict:
    with open(ROUTE_PATH, encoding="utf-8") as fh:
        return json.load(fh)


def _require_topics(topics: list[str]) -> list[tuple[str, dict]]:
    routes = _load_routes()
    configured = routes["topics"]
    if not isinstance(topics, list) or not topics:
        raise TopicRouteError("topics 必须是非空数组")
    if len(topics) != len(set(topics)):
        raise TopicRouteError("topics 不能重复")
    if any(not isinstance(item, str) for item in topics):
        raise TopicRouteError("topics 每项必须是非空字符串")
    unknown = [topic for topic in topics if topic not in configured]
    if unknown:
        raise TopicRouteError(
            f"topics 含无效值: {unknown}。有效 topics: {sorted(configured)}"
        )
    return [(topic, configured[topic]) for topic in topics]


def _location_coords(location: dict | None) -> dict | None:
    if location is None:
        return None
    if not isinstance(location, dict) or isinstance(location, bool):
        raise BirthInputError("source.location 必须是对象")
    allowed = {"city", "longitude", "latitude"}
    unknown = set(location) - allowed
    if unknown:
        raise BirthInputError(f"source.location 含无效字段: {sorted(unknown)}")
    if not location:
        raise BirthInputError("source.location 不能为空")
    city = location.get("city")
    longitude = location.get("longitude")
    if city is not None and (not isinstance(city, str) or not city.strip()):
        raise BirthInputError("source.location.city 必须是非空字符串")
    if longitude is not None and (
        not isinstance(longitude, (int, float)) or
        isinstance(longitude, bool) or not -180 <= float(longitude) <= 180
    ):
        raise BirthInputError("source.location.longitude 必须是 [-180,180] 数值")
    if city:
        try:
            coords = city_coords(city.strip())
        except RPCError as e:
            raise LocationError(f"出生地解析失败: {e}") from e
        return {**coords, "input": {"city": city.strip()}}
    if longitude is not None:
        return {
            "name": "provided-longitude",
            "longitude": float(longitude),
            "latitude": location.get("latitude"),
        }
    raise BirthInputError("source.location 需要 city 或 longitude")


def _location_optional(source: dict) -> dict | None:
    location = source.get("location")
    return location if isinstance(location, dict) and location else None


def _validate_timestamp(value: Any) -> None:
    if not isinstance(value, str) or not value:
        raise BirthInputError("source.timestamp 必须是 RFC3339 字符串")
    try:
        moment = datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError as e:
        raise BirthInputError(f"source.timestamp 无效: {e}") from e
    if moment.tzinfo is None or moment.utcoffset() is None:
        raise BirthInputError("source.timestamp 必须包含时区偏移")


def _resolve_birth_source(source: dict) -> tuple[str, bool, float | None, dict | None]:
    if not isinstance(source, dict):
        raise BirthInputError("source 必须是对象")
    allowed = {
        "type", "timestamp", "precision", "location",
        "solar_time_correction",
    }
    unknown = set(source) - allowed
    if unknown:
        raise BirthInputError(f"source 含无效字段: {sorted(unknown)}")
    source_type = source.get("type")
    if source_type not in ("timestamp", "hour"):
        raise BirthInputError("source.type 只支持 timestamp 或 hour")
    _validate_timestamp(source.get("timestamp"))
    precision = source.get("precision", "minute" if source_type == "timestamp" else "hour")
    if precision not in ("minute", "hour"):
        raise BirthInputError("source.precision 只支持 minute 或 hour")
    correction = source.get("solar_time_correction", "auto")
    if correction not in ("auto", "off"):
        raise BirthInputError("source.solar_time_correction 只支持 auto 或 off")
    if source_type == "hour" and precision == "minute":
        raise BirthInputError('source.type=hour 需要 precision="hour"')
    if source_type == "hour" or precision == "hour":
        if correction == "auto":
            raise BirthInputError(
                "hour 输入表示用户已定时辰；solar_time_correction 必须为 off"
            )
        return source["timestamp"], False, None, _location_optional(source)
    if correction == "off":
        return source["timestamp"], False, None, _location_optional(source)
    coords = _location_coords(source.get("location"))
    if coords is None:
        raise BirthInputError(
            "clock time + solar_time_correction=auto 需要 source.location"
        )
    if coords.get("longitude") is None:
        raise BirthInputError("出生地解析结果缺少 longitude")
    return source["timestamp"], True, float(coords["longitude"]), coords


def create_birth_chart(args: dict) -> dict:
    """出生输入 → 可复用不可变 BirthChart 资源。"""
    gender = args.get("gender")
    if gender not in ("male", "female"):
        raise BirthInputError("gender 只支持 male 或 female")
    timestamp, correct, longitude, location = _resolve_birth_source(args["source"])
    pan = full_paipan(
        timestamp, gender, longitude=longitude, correct=correct
    )
    ref = encode_chart_ref(pan)
    chart = pan["chart"]
    pillars = {key: f"{chart[key]['gan']}{chart[key]['zhi']}"
               for key in ("nian", "yue", "ri", "shi")}
    zw = pan["ziwei"]
    resource = {
        "chart": {
            "id": chart_id(pan),
            "digest": ref["digest"],
            "completeness": "full",
            "gender": gender,
            "birth": {
                "input_type": args["source"]["type"],
                "precision": args["source"].get(
                    "precision", "minute" if args["source"]["type"] == "timestamp" else "hour"
                ),
                "solar_time_correction": "auto" if correct else "off",
                "location": location,
                "solar": pan["solar"],
                "lunar": pan["lunar"],
            },
            "bazi": {"pillars": pillars},
            "ziwei": {
                "ming_gong": zw.get("ming_gong"),
                "shen_gong": zw.get("shen_gong"),
            },
        },
        "chart_ref": ref,
    }
    if hint := pan.get("calibration_hint"):
        resource["chart"]["calibration_hint"] = hint
    return resource


def _topic_for_row(row: dict, selected: list[tuple[str, dict]]) -> str | None:
    domain = row.get("领域")
    for topic_id, route in selected:
        if domain in route["domains"]:
            return topic_id
    return None


def _normalize_row(
    row: dict, *, side: str, topic: str, method: str, time_scope: str,
    year: int | None = None, source_rule: str | None = None,
) -> dict:
    result = {
        "assertion_id": row.get("id"),
        "side": side,
        "topic": topic,
        "method": method,
        "time_scope": time_scope,
        "event_type": row.get("事件类型", ""),
        "event": row.get("事件", ""),
        "conclusion": row.get("结论", ""),
        "source": {
            "kind": "assertion_table",
            "domain": row.get("领域", ""),
            "rule": source_rule or row.get("rule", ""),
            "basis": row.get("依据", ""),
            "classic_basis": row.get("经典依据", ""),
        },
    }
    if year is not None:
        result["year"] = year
    if trace := row.get("trace"):
        evidence = []
        for group in trace:
            for factor, value in group.get("factors", {}).items():
                evidence.append({
                    "condition_group": group.get("condition_group"),
                    "factor": factor,
                    "expected": value.get("expected"),
                    "actual": value.get("actual"),
                })
        result["evidence"] = evidence
    return {key: value for key, value in result.items() if value not in ("", None)}


def _flatten_side_result(
    result: dict, selected: list[tuple[str, dict]], routes: dict,
    time_scope: str, year: int | None = None,
) -> tuple[list[dict], int]:
    assertions: list[dict] = []
    matched = 0
    for label, side in _side_pairs():
        for row in result.get(label, []):
            if not isinstance(row, dict):
                continue
            matched += 1
            topic = _topic_for_row(row, selected)
            if topic is None:
                continue
            source_rule = row.get("rule") or result.get("_rule", "")
            assertions.append(_normalize_row(
                row, side=side, topic=topic,
                method=routes["method_ids"].get(source_rule, source_rule),
                time_scope=time_scope, year=year, source_rule=source_rule,
            ))
    return assertions, matched


def analyze_natal(args: dict) -> dict:
    """本命盘 + 受控人生问题 → 结构化本命断语。"""
    pan = decode_chart_ref(args["chart_ref"])
    domain = args.get("domain")
    if domain is not None and domain not in ("bazi", "ziwei"):
        raise LikiToolError(
            f"domain 无效: {domain!r}，可选 'bazi'/'ziwei'/省略（省略=双术数全量）"
        )
    selected_pairs = _require_topics(args["topics"])
    routes = _load_routes()
    rule_order: list[str] = []
    for _, route in selected_pairs:
        for rule in route["natal_rules"]:
            if rule not in rule_order:
                rule_order.append(rule)
    all_assertions: list[dict] = []
    matched = 0
    seen: set[tuple[str, str]] = set()
    for rule in rule_order:
        result = query(rule, pan)
        result["_rule"] = rule
        part, count = _flatten_side_result(
            result, selected_pairs, routes, "natal"
        )
        matched += count
        for item in part:
            if domain is not None and item["side"] != domain:
                continue
            key = (item["assertion_id"], item["side"])
            if key not in seen:
                seen.add(key)
                all_assertions.append(item)
    return {
        "chart": {
            "id": chart_id(pan),
            "digest": pan.get("pan_digest"),
            "completeness": "full",
        },
        "query": {
            "scope": "natal",
            "topics": [topic for topic, _ in selected_pairs],
            "methods": [routes["method_ids"].get(rule, rule) for rule in rule_order],
        },
        "assertions": all_assertions,
        "counts": {"table_matches": matched, "returned": len(all_assertions)},
    }


def _resolve_time_scope(time_scope: dict) -> tuple[str, int, int]:
    if not isinstance(time_scope, dict):
        raise TopicRouteError("time_scope 必须是对象")
    allowed = {"type", "year", "start_year", "end_year"}
    unknown = set(time_scope) - allowed
    if unknown:
        raise TopicRouteError(f"time_scope 含无效字段: {sorted(unknown)}")
    scope_type = time_scope.get("type")
    if scope_type == "current_year":
        year, _ = resolve_current_year()
        return "year", year, year
    if scope_type == "year":
        year = time_scope.get("year")
        if not isinstance(year, int) or isinstance(year, bool) or year <= 0:
            raise TopicRouteError("time_scope.year 必须是正整数")
        return "year", year, year
    if scope_type == "year_range":
        start = time_scope.get("start_year")
        end = time_scope.get("end_year")
        if not all(isinstance(value, int) and not isinstance(value, bool)
                   and value > 0 for value in (start, end)):
            raise TopicRouteError("time_scope.start_year/end_year 必须是正整数")
        if start > end:
            raise TopicRouteError("time_scope.start_year 不能大于 end_year")
        return "year_range", start, end
    if scope_type == "current_decade":
        year, _ = resolve_current_year()
        return "decade", year, year
    if scope_type == "decade":
        year = time_scope.get("year")
        if not isinstance(year, int) or isinstance(year, bool) or year <= 0:
            raise TopicRouteError("time_scope.decade 需要 year 正整数")
        return "decade", year, year
    raise TopicRouteError(
        "time_scope.type 只支持 current_year/year/year_range/current_decade/decade"
    )


def _deduplicate(items: list[dict]) -> list[dict]:
    output, seen = [], set()
    for item in items:
        key = (item.get("year"), item.get("assertion_id"), item.get("side"))
        if key not in seen:
            seen.add(key)
            output.append(item)
    return output


def analyze_periods(args: dict) -> dict:
    """本命盘 + 时间范围 + 人生问题 → 大运/大限/流年断语。"""
    pan = decode_chart_ref(args["chart_ref"])
    return _analyze_periods(pan, args)


def _analyze_periods(pan: dict, args: dict, validate_pan: bool = True) -> dict:
    """analyze_periods 主体：pan 已就绪（完整盘或 analysis 组合盘）。"""
    selected_pairs = _require_topics(args["topics"])
    routes = _load_routes()
    scope_type, start, end = _resolve_time_scope(args["time_scope"])
    output_scopes: list[dict] = []

    if scope_type == "decade":
        rule_order: list[str] = []
        for _, route in selected_pairs:
            for rule in route["decade_rules"]:
                if rule not in rule_order:
                    rule_order.append(rule)
        assertions: list[dict] = []
        matched = 0
        for rule in rule_order:
            if rule not in CURRENT_LIMIT_RULES:
                raise TopicRouteError(f"decade route 含非限运规则: {rule}")
            result = query(rule, pan, year=start, validate_pan=validate_pan)
            result["_rule"] = rule
            part, count = _flatten_side_result(
                result, selected_pairs, routes, "decade", start
            )
            matched += count
            assertions.extend(part)
        unique = _deduplicate(assertions)
        output_scopes.append({
            "time_scope": {"type": "decade", "anchor_year": start},
            "assertions": unique,
            "counts": {"table_matches": matched, "returned": len(unique)},
        })
        year_basis = None
        current_year = start
        current_year_source = (
            "server"
            if args["time_scope"].get("type") == "current_decade"
            else "specified"
        )
    else:
        rule_order = []
        for _, route in selected_pairs:
            for rule in route["annual_rules"]:
                if rule not in rule_order:
                    rule_order.append(rule)
        if not rule_order:
            raise TopicRouteError("所选 topic 没有流年规则")
        raw = yearly_range(pan, start, end, rules=rule_order, detail=True, validate_pan=validate_pan)
        years = raw.get("years", {})
        for year in range(start, end + 1):
            result = years.get(str(year), {})
            if "error" in result:
                output_scopes.append({
                    "time_scope": {"type": "year", "year": year},
                    "error": {"code": "ENGINE_UNAVAILABLE",
                              "message": result["error"]},
                })
                continue
            assertions: list[dict] = []
            matched = 0
            for rule, rule_result in result.items():
                rule_result["_rule"] = rule
                part, count = _flatten_side_result(
                    rule_result, selected_pairs, routes, "annual", year
                )
                matched += count
                assertions.extend(part)
            unique = _deduplicate(assertions)
            output_scopes.append({
                "time_scope": {"type": "year", "year": year},
                "assertions": unique,
                "counts": {"table_matches": matched, "returned": len(unique)},
            })
        year_basis = raw.get("year_basis")
        current_year = raw.get("current_year")
        current_year_source = raw.get("current_year_source")

    return {
        "chart": {
            "id": chart_id(pan),
            "digest": pan.get("pan_digest"),
            "completeness": "full",
        },
        "query": {
            "scope": "periods",
            "topics": [topic for topic, _ in selected_pairs],
            "time_scope": args["time_scope"],
            "rules": [routes["method_ids"].get(rule, rule) for rule in rule_order],
        },
        "year_basis": year_basis,
        "current_year": current_year,
        "current_year_source": current_year_source,
        "periods": output_scopes,
    }


def compare_birth_charts(args: dict) -> dict:
    """两张不可变盘资源 → 合盘结构化资源引用结果。"""
    from paipan import bond

    pan_a = decode_chart_ref(args["chart_ref_a"])
    pan_b = decode_chart_ref(args["chart_ref_b"])
    result = bond(pan_a, pan_b)
    return {
        "charts": [
            {"id": chart_id(pan_a), "digest": pan_a.get("pan_digest")},
            {"id": chart_id(pan_b), "digest": pan_b.get("pan_digest")},
        ],
        "comparison": result,
    }


def calibrate_birth_time(args: dict) -> dict:
    """多候选出生输入 + 人生事件 → 结构化考时信号。"""
    from calibrate import calibrate

    candidates = args.get("candidates")
    events = args.get("events")
    if not isinstance(candidates, list) or not isinstance(events, list):
        raise BirthInputError("candidates 和 events 必须是数组")
    converted = [_calibration_candidate(item) for item in candidates]
    converted_events = []
    routes = _load_routes()
    configured = routes["topics"]
    for event in events:
        if not isinstance(event, dict):
            raise BirthInputError("events 每项必须是对象")
        topic = event.get("topic")
        if topic not in configured:
            raise TopicRouteError(
                f"events.topic 无效: {topic!r}。有效 topics: {sorted(configured)}"
            )
        year = event.get("year")
        label = event.get("label")
        if not isinstance(year, int) or isinstance(year, bool) or year <= 0:
            raise BirthInputError("events.year 必须是正整数")
        if not isinstance(label, str) or not label:
            raise BirthInputError("events.label 必须是非空字符串")
        annual_rules = list(configured[topic]["annual_rules"])
        if not annual_rules:
            raise TopicRouteError(
                f"topic '{topic}' 只支持本命分析，calibrate_birth_time events.topic 不支持它"
            )
        converted_events.append({
            "year": year,
            "label": label,
            "rule": annual_rules,
            "domains": list(configured[topic]["domains"]),
        })
    detail = args.get("detail", False)
    result = calibrate(converted, converted_events, detail=bool(detail))
    return {
        "candidates": [
            {
                "label": item["label"],
                "events": result.get(item["label"], []),
            }
            for item in converted
        ]
    }


def _calibration_candidate(candidate: dict) -> dict:
    if not isinstance(candidate, dict):
        raise BirthInputError("candidates 每项必须是对象")
    label = candidate.get("label")
    if not isinstance(label, str) or not label:
        raise BirthInputError("candidates 每项需要非空 label")
    gender = candidate.get("gender")
    if gender not in ("male", "female"):
        raise BirthInputError("candidates 每项 gender 只支持 male 或 female")
    timestamp, correct, longitude, _ = _resolve_birth_source(
        candidate.get("source")
    )
    return {
        "label": label,
        "gregorian": timestamp,
        "gender": gender,
        "longitude": longitude,
        "correct": correct,
    }
