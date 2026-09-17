"""LLM-facing Bazi tool contract tests."""
import json
import subprocess
import sys
from pathlib import Path

import pytest
from jsonschema import Draft202012Validator

import _helpers  # noqa: F401 —— 注入 tools 路径
from agent_cli import _REQUIRED_ARGS
from factor_tokens import FACTOR_WILDCARD
from tool_contracts import (
    FACTOR_VALUE_SCHEMA,
    TRACE_FACTOR_SCHEMA,
    brief_assertion_hit_schema,
    detailed_assertion_hit_schema,
)

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = json.loads(
    (ROOT / "skills/liki/bazi/tools/skill-tools.json").read_text(encoding="utf-8")
)
FUNCTIONS = {tool["function"]["name"]: tool["function"] for tool in MANIFEST["tools"]}


def _errors(schema, payload):
    return list(Draft202012Validator(schema).iter_errors(payload))


def test_factor_value_is_closed_scalar_contract():
    assert not _errors(FACTOR_VALUE_SCHEMA, 0)
    assert not _errors(FACTOR_VALUE_SCHEMA, 1)
    assert not _errors(FACTOR_VALUE_SCHEMA, "木")
    assert not _errors(FACTOR_VALUE_SCHEMA, "身弱")
    assert not _errors(FACTOR_VALUE_SCHEMA, "")
    assert _errors(FACTOR_VALUE_SCHEMA, FACTOR_WILDCARD)

    for value in (None, 2, True, False, [], {}, {"value": 1}):
        assert _errors(FACTOR_VALUE_SCHEMA, value), value


def test_trace_factor_is_expected_actual_pair():
    assert not _errors(TRACE_FACTOR_SCHEMA, {"expected": 1, "actual": 0})
    assert not _errors(TRACE_FACTOR_SCHEMA, {"expected": "木", "actual": "火"})
    assert _errors(TRACE_FACTOR_SCHEMA, {"expected": 1})
    assert _errors(TRACE_FACTOR_SCHEMA, {"expected": 1, "actual": None})
    assert _errors(TRACE_FACTOR_SCHEMA, {"expected": 1, "actual": {"value": 1}})
    assert _errors(
        TRACE_FACTOR_SCHEMA,
        {"expected": FACTOR_WILDCARD, "actual": FACTOR_WILDCARD},
    )


def test_brief_and_detailed_hits_have_disjoint_shapes():
    brief = {
        "id": "ying_h09",
        "领域": "应期",
        "事件类型": "引动",
        "时间层": "流年",
        "事件": "三刑成立",
        "结论": "断事年",
    }
    detailed = {
        **brief,
        "约束组": [{"流年地支相刑寅巳申": 1}],
        "依据": "三刑主刑伤断事",
        "经典依据": "《三命通会》论三刑",
        "trace": [{"condition_group": 1, "factors": {
            "流年地支相刑寅巳申": {"expected": 1, "actual": 1},
        }}],
    }

    assert not _errors(brief_assertion_hit_schema(), brief)
    assert _errors(brief_assertion_hit_schema(), detailed)
    assert not _errors(detailed_assertion_hit_schema(), detailed)
    assert _errors(detailed_assertion_hit_schema(), brief)


def test_query_result_has_stable_sides_and_optional_context():
    schema = FUNCTIONS["query"]["result_schema"]
    payload = {
        "八字": [],
        "紫微": [],
        "合参": [],
        "yong_shen_context": {
            "yong_shen": {},
            "element_states": [],
            "ten_god_states": [],
        },
        "current_year": 2026,
        "current_year_source": "server",
    }
    assert not _errors(schema, payload)
    assert _errors(schema, {**payload, "evidence": {}})


def test_yearly_result_separates_rule_success_and_year_error():
    schema = FUNCTIONS["yearly_range"]["result_schema"]
    payload = {
        "current_year": 2026,
        "current_year_source": "server",
        "year_basis": {"八字": "a", "紫微": "b", "usage": "c"},
        "years": {
            "2026": {"年合会": {"八字": [], "紫微": [], "合参": []}},
            "2027": {"error": "RPCError: timeout"},
        },
    }
    assert not _errors(schema, payload)
    assert _errors(schema, {**payload, "years": {"x": "not-object"}})
    # error-only year must not also validate as a rule-result map.
    assert not _errors(schema, {
        **payload,
        "years": {"2027": {"error": "RPCError: timeout"}},
    })


def test_all_tools_have_valid_result_and_closed_args():
    required = {
        "city_coords": ["city"],
        "full_paipan": ["gregorian", "gender"],
        "query": ["rule", "pan"],
        "yearly_range": ["pan", "start", "end", "rules"],
        "calibrate": ["candidates", "events"],
        "bond": ["pan_a", "pan_b"],
    }
    for name, fn in FUNCTIONS.items():
        assert fn["parameters"]["additionalProperties"] is False, name
        assert fn["parameters"]["required"] == required[name], name
        result = fn.get("result_schema")
        assert result, name
        Draft202012Validator.check_schema(result)


def test_generated_tool_contracts_have_no_drift():
    result = subprocess.run(
        [sys.executable, str(ROOT / "scripts/generate_bazi_tool_contracts.py"), "--check"],
        cwd=ROOT,
        text=True,
        capture_output=True,
    )
    assert result.returncode == 0, result.stdout + result.stderr
