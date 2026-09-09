from __future__ import annotations

import json
from pathlib import Path

from jsonschema import validate
from jsonschema.exceptions import ValidationError


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"


def _tool(name: str) -> dict:
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    return next(item["function"] for item in schema["tools"] if item["function"]["name"] == name)


def test_snapshot_schemas_accept_primary_calls():
    validate(
        {"question": "这次面试能不能通过？", "matter": "career", "mode": "auto"},
        _tool("liuyao_snapshot")["parameters"],
    )
    validate(
        {"question": "该往哪里推进？", "longitude": 121.47, "matter": "wealth"},
        _tool("qimen_snapshot")["parameters"],
    )


def test_ask_schemas_accept_primary_calls_and_bind_method():
    liuyao = _tool("liuyao_ask")["parameters"]["properties"]["snapshot"]
    qimen = _tool("qimen_ask")["parameters"]["properties"]["snapshot"]
    validate(
        {
            "method": "liuyao",
            "schema_version": "liuyao-snapshot-v3",
            "snapshot_digest": "a" * 16,
        },
        liuyao,
    )
    validate(
        {
            "method": "qimen",
            "schema_version": "qimen-snapshot-v3",
            "snapshot_digest": "a" * 16,
        },
        qimen,
    )
    try:
        validate(
            {
                "method": "qimen",
                "schema_version": "qimen-snapshot-v3",
                "snapshot_digest": "a" * 16,
            },
            liuyao,
        )
    except ValidationError:
        pass
    else:
        raise AssertionError("liuyao_ask accepted qimen snapshot")
