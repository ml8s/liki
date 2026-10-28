"""Divination tool-surface schema acceptance (schemas come from the contracts)."""

from app.contracts import DIVINATION_TOOLS
from jsonschema import validate

_SCHEMAS = {contract.name: contract.input_schema for contract in DIVINATION_TOOLS}


def _parameters(name: str) -> dict:
    return _SCHEMAS[name]


def test_snapshot_schemas_accept_primary_calls():
    validate(
        {"question": "这次面试能不能通过？", "matter": "career", "casting": {"mode": "auto"}, "longitude": 116.4, "standard_time": "2026-10-01T10:00:00+08:00"},
        _parameters("liuyao_snapshot"),
    )
    validate(
        {
            "question": "该往哪里推进？",
            "longitude": 121.47,
            "matter": "wealth",
            "standard_time": "2026-10-01T10:00:00+08:00",
            "method": {"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
        },
        _parameters("qimen_snapshot"),
    )


def test_ask_schemas_accept_primary_calls_and_bind_method():
    pairs = {"liuyao_query": "liuyao", "qimen_query": "qimen"}
    for name, method in pairs.items():
        snapshot = _parameters(name)["properties"]["snapshot"]
        assert snapshot["properties"]["method"]["const"] == method
        assert "snapshot_digest" in snapshot["required"]
