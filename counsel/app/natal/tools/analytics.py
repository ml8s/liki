"""面向用户问题的本命/应期分析编排。

TopicRouter 只做受控配置投影：topic → 命理规则 + 断语领域。这里不解释命理，
不修改断语真值表；输出统一使用英文契约字段，中文只保留在展示内容中。

counsel MCP 的 natal_query / natal_liunian_query 复用本模块的路由与展平逻辑
（`_require_life_topics` / `_load_routes` / `_flatten_side_result` / `_analyze_liunian`）。
"""

import json
import os
from functools import lru_cache

from app.domain.natal.duanyu import yearly_range
from app.natal.tools.errors import LikiToolError
from app.natal.tools.factor_constants import load_constants

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


def _require_life_topics(life_topics: list[str]) -> list[tuple[str, dict]]:
    routes = _load_routes()
    configured = routes["life_topics"]
    if not isinstance(life_topics, list) or not life_topics:
        raise TopicRouteError("life_topics 必须是非空数组")
    if len(life_topics) != len(set(life_topics)):
        raise TopicRouteError("life_topics 不能重复")
    if any(not isinstance(topic, str) for topic in life_topics):
        raise TopicRouteError("life_topics 每项必须是非空字符串")
    unknown = [topic for topic in life_topics if topic not in configured]
    if unknown:
        raise TopicRouteError(f"life_topics 含无效值: {unknown}。有效 life_topics: {sorted(configured)}")
    return [(topic, configured[topic]) for topic in life_topics]


def _life_topics_for_row(
    row: dict, selected: list[tuple[str, dict]], cross_cutting: frozenset[str] = frozenset()
) -> list[str]:
    """把断语行路由到所有命中的 topic；跨切域可同时归属多个已选 topic。"""
    domain = row.get("领域")
    is_cross = domain in cross_cutting
    return [topic_id for topic_id, route in selected if is_cross or domain in route["domains"]]


def _normalize_row(
    row: dict,
    *,
    life_topics: list[str],
    rule_id: str,
    year: int | None = None,
) -> dict:
    result = {
        "assertion_id": row.get("id"),
        "life_topics": life_topics,
        "rule_id": rule_id,
        "event_type": row.get("事件类型", ""),
        "event": row.get("事件", ""),
        "conclusion": row.get("结论", ""),
        "source": {
            "kind": "assertion_table",
            "rule_domain": row.get("领域", ""),
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
                item = {
                    "condition_group": group.get("condition_group"),
                    "factor": factor,
                    "expected": value.get("expected"),
                    "actual": value.get("actual"),
                }
                # 性别专属因子（如「女命官星得地格」，官=夫，据《女命赋》）：
                # 在异性命盘里仅作排除条件（恒 0），标注性别归属以免误读。
                if factor.startswith("女命"):
                    item["gender_scope"] = "女"
                elif factor.startswith("男命"):
                    item["gender_scope"] = "男"
                evidence.append(item)
        result["evidence"] = evidence
    return {key: value for key, value in result.items() if value not in ("", None)}


def _flatten_side_result(
    result: dict,
    selected: list[tuple[str, dict]],
    routes: dict,
    year: int | None = None,
) -> tuple[dict[str, list[dict]], dict[str, dict[str, int]]]:
    assertions: dict[str, list[dict]] = {}
    counts: dict[str, dict[str, int]] = {}
    cross_cutting = frozenset(routes.get("cross_cutting_domains") or ())
    for label, technique in _side_pairs():
        assertions[technique] = []
        counts[technique] = {"matched": 0, "topic_not_selected": 0}
        for row in result.get(label, []):
            if not isinstance(row, dict):
                continue
            counts[technique]["matched"] += 1
            life_topics = _life_topics_for_row(row, selected, cross_cutting)
            if not life_topics:
                counts[technique]["topic_not_selected"] += 1
                continue
            source_rule = row.get("rule") or result.get("_rule", "")
            assertions[technique].append(
                _normalize_row(
                    row,
                    life_topics=life_topics,
                    rule_id=routes["method_ids"].get(source_rule, source_rule),
                    year=year,
                )
            )
    return assertions, counts


def _empty_source_counts() -> dict[str, dict[str, int]]:
    return {technique: {"matched": 0, "topic_not_selected": 0} for technique in ("bazi", "ziwei")}


def _add_source_counts(total: dict[str, dict[str, int]], part: dict[str, dict[str, int]]) -> None:
    for technique in ("bazi", "ziwei"):
        for reason in ("matched", "topic_not_selected"):
            total[technique][reason] += part[technique][reason]


def _public_counts(
    matched: int,
    topic_not_selected: int,
    returned: int,
) -> dict:
    return {
        "matched": matched,
        "returned": returned,
        "filtered": matched - returned,
        "filtered_by": {
            "topic_not_selected": topic_not_selected,
            "deduplicated": matched - topic_not_selected - returned,
        },
    }


def _deduplicate(items: list[dict]) -> list[dict]:
    output, seen = [], set()
    for item in items:
        key = (item.get("year"), item.get("assertion_id"))
        if key not in seen:
            seen.add(key)
            output.append(item)
    return output


def _public_assertion(item: dict) -> dict:
    """Year belongs to the period envelope, not to each repeated assertion."""
    return {key: value for key, value in item.items() if key != "year"}


def _analyze_liunian(natal_chart: dict, args: dict, validate_pan: bool = True) -> dict:
    """natal_liunian_query 主体：输入是 engine series 的 start/count 口径。"""
    selected_pairs = _require_life_topics(args["life_topics"])
    routes = _load_routes()
    start, count = args["start_year"], args["year_count"]
    if not isinstance(start, int) or isinstance(start, bool) or start <= 0:
        raise TopicRouteError("start_year 必须是正整数")
    if not isinstance(count, int) or isinstance(count, bool) or not 1 <= count <= 50:
        raise TopicRouteError("year_count 必须在 1-50 之间")

    rule_order: list[str] = []
    for _, route in selected_pairs:
        for rule in route["annual_rules"]:
            if rule not in rule_order:
                rule_order.append(rule)
    if not rule_order:
        raise TopicRouteError("所选 topic 没有流年规则")

    raw = yearly_range(
        natal_chart,
        start,
        count,
        rules=rule_order,
        detail=True,
        validate_pan=validate_pan,
    )
    years = raw.get("years", {})
    output_scopes: list[dict] = []
    for year in range(start, start + count):
        result = years.get(str(year), {})
        if "error" in result:
            output_scopes.append(
                {
                    "year": year,
                    "error": {"code": "ENGINE_UNAVAILABLE", "message": result["error"]},
                }
            )
            continue
        grouped: dict[str, list[dict]] = {}
        source_counts = _empty_source_counts()
        for rule, rule_result in result.items():
            rule_result["_rule"] = rule
            part, matched_count = _flatten_side_result(rule_result, selected_pairs, routes, year)
            _add_source_counts(source_counts, matched_count)
            for technique in ("bazi", "ziwei"):
                grouped.setdefault(technique, []).extend(part[technique])
        side_results = {}
        for technique in ("bazi", "ziwei"):
            unique = [_public_assertion(item) for item in _deduplicate(grouped.get(technique, []))]
            side_results[technique] = {
                "assertions": unique,
                "counts": _public_counts(
                    source_counts[technique]["matched"],
                    source_counts[technique]["topic_not_selected"],
                    len(unique),
                ),
            }
        output_scopes.append(
            {
                "year": year,
                **side_results,
            }
        )

    return {
        "natal_chart_digest": natal_chart["natal_chart_digest"],
        "years": output_scopes,
    }
