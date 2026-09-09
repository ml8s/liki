from __future__ import annotations

import json
import sys
from pathlib import Path

from jsonschema import validate


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

from divination_router import route_question  # noqa: E402


def test_event_outcome_routes_to_liuyao():
    result = route_question("这次面试能不能通过？")
    assert result["route"] == "liuyao"
    assert result["category"] == "event_outcome"
    assert result["dual_divination"] is False


def test_strategy_routes_to_qimen():
    result = route_question("我该主动联系还是继续等待？")
    assert result["route"] == "qimen"
    assert result["category"] in {"strategy_decision", "timing_action"}


def test_date_selection_routes_to_huangli():
    result = route_question("哪天适合搬家？")
    assert result["route"] == "huangli"
    assert result["category"] == "date_selection"


def test_mixed_goal_asks_user_to_choose_one():
    result = route_question("这笔生意能不能赚钱？我该往哪个方向推进？")
    assert result["route"] == "clarify"
    assert result["category"] == "mixed_outcome_strategy"
    assert result["dual_divination"] is False
    assert result["clarify"]["question"].startswith("本次优先看哪一个")


def test_user_specified_method_overrides_default():
    result = route_question("这次面试能不能通过？", specified_method="qimen")
    assert result["route"] == "qimen"
    assert result["confidence"] == "specified"


def test_safety_redirect_blocks_divination():
    result = route_question("医生说我病情很重，还能活多久？")
    assert result["route"] == "blocked"
    assert result["safety_category"] in {"serious_medical", "emergency_medical"}


def test_ambiguous_goal_does_not_cast():
    result = route_question("帮我看看最近怎么样")
    assert result["route"] == "clarify"
    assert result["category"] == "ambiguous"


def test_router_tool_schema_accepts_examples():
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    tool = next(
        item["function"]
        for item in schema["tools"]
        if item["function"]["name"] == "divination_route"
    )
    for args in (
        {"question": "这次面试能不能通过？"},
        {"question": "我该主动还是等待？", "category": "strategy_decision"},
        {"question": "哪天适合签约？", "specified_method": "huangli"},
    ):
        validate(args, tool["parameters"])
