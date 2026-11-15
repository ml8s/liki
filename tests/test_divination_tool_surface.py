"""Divination tool-surface schema acceptance (schemas come from contracts)."""

from app.contracts import DIVINATION_TOOLS
from jsonschema import validate

_SCHEMAS = {contract.name: contract.input_schema for contract in DIVINATION_TOOLS}


def _liuyao_event(focus):
    return {
        "question": "这次面试能不能通过？",
        "origin": {"mode": "yaos", "yaos": [7, 7, 7, 7, 7, 7]},
        "casting": {
            "mode": "yaos", "order": "bottom_up",
            "yaos": [7, 7, 7, 7, 7, 7], "dong_yao": [],
        },
        "casting_context": {
            "standard_time": "2026-10-01T10:00:00+08:00",
            "longitude": 116.4,
            "focus": focus,
        },
    }


def test_ask_schemas_accept_primary_calls():
    focus_schema = (
        _SCHEMAS["liuyao_ask"]["properties"]["event"]["properties"]
        ["casting_context"]["properties"]["focus"]
    )
    validate({"kind": "matter", "matter": "career"}, focus_schema)
    validate({
        "kind": "matter", "matter": "relationship", "perspective": "male",
    }, focus_schema)
    validate({"kind": "yong_shen", "yong_shen": "官鬼"}, focus_schema)
    validate(
        {
            "event": _liuyao_event({"kind": "matter", "matter": "career"}),
        },
        _SCHEMAS["liuyao_ask"],
    )
    for invalid_focus in (
        {"kind": "matter", "matter": "relationship"},
        {"kind": "matter", "matter": "career", "perspective": "male"},
        {"kind": "matter", "matter": "career", "perspective": None},
        {"kind": "yong_shen", "yong_shen": "财神"},
    ):
        case = {"event": _liuyao_event(invalid_focus)}
        try:
            validate(case, _SCHEMAS["liuyao_ask"])
        except Exception:
            pass
        else:
            raise AssertionError(f"invalid focus accepted: {invalid_focus}")
    validate(
        {
            "question": "该往哪里推进？",
            "longitude": 121.47,
            "focus": {"kind": "matter", "matter": "wealth"},
            "standard_time": "2026-10-01T10:00:00+08:00",
            "method": {
                "scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu",
            },
        },
        _SCHEMAS["qimen_ask"],
    )


def test_ask_outputs_are_controlled_readings():
    for name in ("liuyao_ask", "qimen_ask"):
        properties = next(
            contract.output_schema["properties"]
            for contract in DIVINATION_TOOLS
            if contract.name == name
        )
        expected = {"scope", "evidence", "interpretation"}
        if name == "liuyao_ask":
            expected.update({"case_version", "event", "chart"})
        assert expected <= set(properties)
        assert "reading_digest" not in properties
        assert "case" not in properties
        assert "factors" not in properties
