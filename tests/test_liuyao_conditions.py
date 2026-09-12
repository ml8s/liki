from __future__ import annotations

import sys
from pathlib import Path
import json

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
            ],
            "secondary": [
                {"id": f"pattern-{index}", "fact": pattern}
                for index, pattern in enumerate(patterns or [], 1)
            ],
        },
        "facts": {},
    }


def test_rules_all_have_restoration():
    rules = liuyao_conditions.load_rules()["rules"]
    assert len(rules) >= 5
    assert all(rule["restored"].strip() for rule in rules)


def test_rule_paths_stay_inside_snapshot_contract():
    contract = json.loads(
        (ROOT / "skills/liki-divination/tools/liuyao_snapshot_contract.json")
        .read_text(encoding="utf-8")
    )
    allowed_roots = set(contract["properties"]) | {"topic"}
    for rule in liuyao_conditions.load_rules()["rules"]:
        for path in rule.get("applies_if", {}):
            root = path.split(".", 1)[0]
            assert root in allowed_roots, f"{rule['id']} uses stale path: {path}"


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
    assert item["state"] == "weak"
    assert "耗损" in item["conclusion"]

    result = liuyao_conditions.evaluate(
        _snapshot(wang="旺", relations=["克用"]), topic="wealth"
    )
    item = next(rule for rule in result["results"] if rule["id"] == "brother-moving-not-certain-loss")
    assert item["applies"] is True
    assert item["state"] == "strong"
    assert "不直接断破财" in item["conclusion"]

    result = liuyao_conditions.evaluate(
        _snapshot(relations=["克用"]), topic="career"
    )
    item = next(rule for rule in result["results"] if rule["id"] == "brother-moving-not-certain-loss")
    assert item["applies"] is False


def test_chonghe_rules_read_projected_secondary_evidence():
    for sub_type, rule_id in (
        ("六合", "liuhe-not-always-success"),
        ("六冲", "six-chong-not-always-failure"),
    ):
        snapshot = _snapshot(patterns=[{"type": "冲合", "sub_type": sub_type}])
        result = liuyao_conditions.evaluate(snapshot, topic="relationship")
        item = next(
            rule for rule in result["results"] if rule["id"] == rule_id
        )
        assert item["applies"] is True


def test_all_outputs_are_conditional_not_absolute():
    result = liuyao_conditions.evaluate(_snapshot(), topic="wealth")
    assert result["policy"]["forbid_absolute_old_phrases"] is True
    assert all(rule["conclusion_scope"] == "conditional_rule_restoration_not_absolute_verdict" for rule in result["results"])


def test_array_conditions_support_hidden_and_transformed_facts():
    snapshot = _snapshot()
    snapshot["focus"]["yong_shen"]["is_hidden"] = True
    result = liuyao_conditions.evaluate(snapshot, topic="wealth")
    item = next(rule for rule in result["results"] if rule["id"] == "hidden-yong-shen-not-always-unavailable")
    assert item["applies"] is True

    snapshot = _snapshot()
    snapshot["facts"]["moving_transformations"] = [
        {"position": 3, "return_relation": "回头克"}
    ]
    result = liuyao_conditions.evaluate(snapshot, topic="career")
    item = next(rule for rule in result["results"] if rule["id"] == "returning-control-not-absolute-damage")
    assert item["applies"] is True


def test_complete_sanhe_structure_is_conditioned():
    snapshot = _snapshot()
    snapshot["facts"]["san_he_candidates"] = [
        {"complete": True, "positions": [1, 3, 5], "branches": ["申", "子", "辰"]}
    ]
    result = liuyao_conditions.evaluate(snapshot, topic="wealth")
    item = next(rule for rule in result["results"] if rule["id"] == "complete-sanhe-is-candidate-not-verdict")
    assert item["applies"] is True
    assert item["conclusion_scope"] == "conditional_rule_restoration_not_absolute_verdict"
