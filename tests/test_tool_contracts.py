"""Natal tool parameter contract tests."""
import json
from pathlib import Path

import pytest
import _helpers  # noqa: F401 —— 注入 tools 路径
from agent_cli import _REQUIRED_ARGS

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = json.loads(
    (ROOT / "skills/liki/natal/tools/skill-tools.json").read_text(encoding="utf-8")
)
FUNCTIONS = {tool["function"]["name"]: tool["function"] for tool in MANIFEST["tools"]}


def test_all_tools_have_closed_args():
    required = {
        "create_birth_chart": ["gender", "source"],
        "analyze_natal": ["chart_ref", "topics"],
        "analyze_periods": ["chart_ref", "time_scope", "topics"],
        "compare_birth_charts": ["chart_ref_a", "chart_ref_b"],
        "calibrate_birth_time": ["candidates", "events"],
    }
    assert set(FUNCTIONS) == set(required)
    for name, fn in FUNCTIONS.items():
        assert fn["parameters"]["additionalProperties"] is False, name
        assert fn["parameters"]["required"] == required[name], name
        assert fn["parameters"]["type"] == "object", name


def test_required_args_match_cli_precheck():
    schema_required = {
        name: set(fn["parameters"]["required"])
        for name, fn in FUNCTIONS.items()
    }
    assert {name: set(args) for name, args in _REQUIRED_ARGS.items()} == schema_required


def test_topic_enum_is_route_configured():
    routes = json.loads(
        (ROOT / "skills/liki/natal/tools/topic_routes.json").read_text(encoding="utf-8")
    )
    enum = FUNCTIONS["analyze_natal"]["parameters"]["properties"]["topics"]["items"]["enum"]
    assert enum == list(routes["topics"])


def test_response_contract_covers_every_tool():
    response = json.loads(
        (ROOT / "skills/liki/natal/tools/response-contract.json").read_text(encoding="utf-8")
    )
    assert response["version"] == "natal-response-contract-v1"
    assert set(response["tools"]) == set(FUNCTIONS)
    for tool in MANIFEST["tools"]:
        assert "result_schema" not in tool["function"]


def test_topic_routes_are_valid_and_unambiguous():
    import sys

    sys.path.insert(0, str(ROOT / "skills/liki/natal/tools"))
    from factor_constants import load_constants

    routes_doc = json.loads(
        (ROOT / "skills/liki/natal/tools/topic_routes.json").read_text(encoding="utf-8")
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
            assert domain not in domains_by_topic, (domain, topic, domains_by_topic[domain])
            domains_by_topic[domain] = topic


def test_calibrate_rejects_natal_only_topic_before_engine_call():
    import sys

    sys.path.insert(0, str(ROOT / "skills/liki/natal/tools"))
    from analytics import calibrate_birth_time

    source = {
        "type": "hour",
        "timestamp": "1990-06-01T12:00:00+08:00",
        "precision": "hour",
        "solar_time_correction": "off",
    }
    args = {
        "candidates": [{"label": "A", "gender": "male", "source": source}],
        "events": [{"year": 2026, "topic": "appearance", "label": "外貌"}],
    }
    with pytest.raises(Exception, match="只支持本命分析"):
        calibrate_birth_time(args)
