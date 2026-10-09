"""面向用户问题的本命/应期分析编排。

TopicRouter 只做受控配置投影：topic → 命理规则 + 断语领域。这里不解释命理，
不修改断语真值表；输出统一使用英文契约字段，中文只保留在展示内容中。

counsel MCP 的 natal_query / period_query 复用本模块的路由与展平逻辑
（`_require_topics` / `_load_routes` / `_flatten_side_result` / `_analyze_periods`）。
"""
from __future__ import annotations

import json
import os
from functools import lru_cache

from chart_id import chart_id
from duanyu import (
    CURRENT_LIMIT_RULES,
    query,
    resolve_current_year,
    yearly_range,
)
from errors import LikiToolError
from factor_constants import load_constants

TOOLS_DIR = os.path.dirname(os.path.abspath(__file__))
ROUTE_PATH = os.path.join(TOOLS_DIR, "topic_routes.json")


def _side_pairs() -> list[tuple[str, str]]:
    config = load_constants()["命理侧"]
    return [(label, code) for code, label in config["标签"].items()]


class TopicRouteError(LikiToolError, ValueError):
    """topic 或分析参数不在受控契约内。"""


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


def _analyze_periods(pan: dict, args: dict, validate_pan: bool = True) -> dict:
    """period_query 主体：pan 已就绪（完整盘或分侧盘）。"""
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
