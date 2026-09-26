"""Counsel natal tool 参数契约测试（skill-tools → counsel-tools）。"""
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = json.loads(
    (ROOT / "counsel/app/natal/tools/counsel-tools.json").read_text(encoding="utf-8")
)
FUNCTIONS = {tool["function"]["name"]: tool["function"] for tool in MANIFEST["tools"]}


def test_all_tools_have_closed_args():
    required = {
        "compute_factors": ["chart"],
        "natal_query": ["factors", "factors_digest", "topics", "context"],
        "period_query": ["factors", "factors_digest", "time_scope", "topics", "chart"],
    }
    assert set(FUNCTIONS) == set(required)
    for name, fn in FUNCTIONS.items():
        assert fn["parameters"]["additionalProperties"] is False, name
        assert fn["parameters"]["required"] == required[name], name
        assert fn["parameters"]["type"] == "object", name


def test_topic_enum_is_generated_from_topic_routes():
    routes = json.loads(
        (ROOT / "counsel/app/natal/tools/topic_routes.json").read_text(encoding="utf-8")
    )
    expected = list(routes["topics"])
    for name in ("natal_query", "period_query"):
        schema = FUNCTIONS[name]["parameters"]["properties"]["topics"]
        assert schema["items"]["enum"] == expected, name


def test_topic_routes_are_valid_and_unambiguous():
    sys.path.insert(0, str(ROOT / "counsel/app/natal/tools"))
    from factor_constants import load_constants

    routes_doc = json.loads(
        (ROOT / "counsel/app/natal/tools/topic_routes.json").read_text(encoding="utf-8")
    )
    routes = routes_doc["topics"]
    domain_config = load_constants()["命理域"]
    natal = set(domain_config["本命规则"])
    annual = set(domain_config["流年规则"]) | set(domain_config["场景别名"])
    decades = {"大运", "大限"}

    domains_by_topic: dict[str, str] = {}
    for topic, route in routes.items():
        assert route["label"]
        assert route["domains"]
        assert all(rule in natal for rule in route["natal_rules"]), topic
        assert all(rule in annual for rule in route["annual_rules"]), topic
        assert all(rule in decades for rule in route["decade_rules"]), topic
        for domain in route["domains"]:
            if topic == "chart_structure":
                continue  # chart_structure 跨域（含旺衰/用神断语领域），不参与唯一性检查
            assert domain not in domains_by_topic, (domain, topic, domains_by_topic[domain])
            domains_by_topic[domain] = topic
