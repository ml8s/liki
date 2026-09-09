from __future__ import annotations

import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_conditions  # noqa: E402


def _snapshot(wang="休", yuepo=True, xunkong=False, moving=False, patterns=None, relations=None):
    return {
        "focus": {
            "yong_shen": {"wang_shuai": wang, "yue_po": yuepo, "xun_kong": xunkong},
            "yong_line": {"flags": {"moving": moving}},
        },
        "evidence": {
            "primary": [
                {"id": "moving-relation", "fact": {"relation": relation}}
                for relation in (relations or [])
            ]
        },
        "facts": {},
    }


def test_list_rules_all_have_restoration():
    rules = liuyao_conditions.list_rules()
    assert len(rules) >= 5
    assert all(rule["restored"].strip() for rule in rules)


def test_month_break_weak_is_true_break_candidate():
    result = liuyao_conditions.evaluate(_snapshot(wang="死", yuepo=True), topic="wealth")
    item = next(rule for rule in result["results"] if rule["id"] == "month-break-not-always-failed")
    assert item["applies"] is True
    assert item["state"] == "weak"
    assert "当下无力" in item["conclusion"]


def test_month_break_strong_is_false_break_candidate():
    result = liuyao_conditions.evaluate(_snapshot(wang="旺", yuepo=True), topic="wealth")
    item = next(rule for rule in result["results"] if rule["id"] == "month-break-not-always-failed")
    assert item["state"] == "strong"
    assert "假破" in item["conclusion"]


def test_void_moving_is_false_void_candidate():
    result = liuyao_conditions.evaluate(
        _snapshot(wang="相", xunkong=True, moving=True), topic="wealth"
    )
    item = next(rule for rule in result["results"] if rule["id"] == "void-not-always-useless")
    assert item["state"] == "strong"
    assert "假空" in item["conclusion"]


def test_brother_rule_requires_wealth_and_relation():
    result = liuyao_conditions.evaluate(
        _snapshot(relations=["克用"]), topic="wealth"
    )
    item = next(rule for rule in result["results"] if rule["id"] == "brother-moving-not-certain-loss")
    assert item["applies"] is True

    result = liuyao_conditions.evaluate(
        _snapshot(relations=["克用"]), topic="career"
    )
    item = next(rule for rule in result["results"] if rule["id"] == "brother-moving-not-certain-loss")
    assert item["applies"] is False


def test_all_outputs_are_conditional_not_absolute():
    result = liuyao_conditions.evaluate(_snapshot(), topic="wealth")
    assert result["policy"]["forbid_absolute_old_phrases"] is True
    assert all(rule["conclusion_scope"] == "conditional_rule_restoration_not_absolute_verdict" for rule in result["results"])
