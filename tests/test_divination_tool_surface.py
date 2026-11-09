"""Divination tool-surface schema acceptance (schemas come from contracts)."""

from app.contracts import DIVINATION_TOOLS
from jsonschema import validate

_SCHEMAS = {contract.name: contract.input_schema for contract in DIVINATION_TOOLS}


def test_ask_schemas_accept_primary_calls():
    validate(
        {
            "question": "这次面试能不能通过？",
            "focus": {"kind": "matter", "matter": "career"},
            "casting": {"mode": "auto"},
            "longitude": 116.4,
            "standard_time": "2026-10-01T10:00:00+08:00",
        },
        _SCHEMAS["liuyao_ask"],
    )
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
        assert {"scope", "evidence", "interpretation"} <= set(properties)
        assert "reading_digest" not in properties
        assert "case" not in properties
        assert "factors" not in properties
