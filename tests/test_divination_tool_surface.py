"""Divination tool-surface schema acceptance (schemas come from the contracts)."""
from __future__ import annotations

from app.contracts import DIVINATION_TOOLS
from jsonschema import validate
from jsonschema.exceptions import ValidationError

_SCHEMAS = {contract.name: contract.input_schema for contract in DIVINATION_TOOLS}


def _parameters(name: str) -> dict:
    return _SCHEMAS[name]


def test_snapshot_schemas_accept_primary_calls():
    validate(
        {"question": "这次面试能不能通过？", "matter": "career", "mode": "auto"},
        _parameters("liuyao_snapshot"),
    )
    validate(
        {"question": "该往哪里推进？", "longitude": 121.47, "matter": "wealth"},
        _parameters("qimen_snapshot"),
    )


def test_ask_schemas_accept_primary_calls_and_bind_method():
    liuyao = _parameters("liuyao_query")["properties"]["snapshot"]
    qimen = _parameters("qimen_query")["properties"]["snapshot"]
    validate(
        {
            "method": "liuyao",
            "schema_version": "liuyao-snapshot-v7",
            "snapshot_digest": "a" * 16,
        },
        liuyao,
    )
    validate(
        {
            "method": "qimen",
            "schema_version": "qimen-snapshot-v5",
            "snapshot_digest": "a" * 16,
        },
        qimen,
    )
    try:
        validate(
            {
                "method": "qimen",
                "schema_version": "qimen-snapshot-v5",
                "snapshot_digest": "a" * 16,
            },
            liuyao,
        )
    except ValidationError:
        pass
    else:
        raise AssertionError("liuyao_query accepted qimen snapshot")
