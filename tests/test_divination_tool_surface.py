"""Divination tool-surface schema acceptance (schemas come from contracts)."""

from app.contracts import DIVINATION_TOOLS
from jsonschema import validate

_SCHEMAS = {contract.name: contract.input_schema for contract in DIVINATION_TOOLS}


def _liuyao_event():
    return {
        "casting": {
            "mode": "yaos", "order": "bottom_up",
            "yaos": [7, 7, 7, 7, 7, 7], "dong_yao": [],
        },
        "casting_context": {
            "standard_time": "2026-10-01T10:00:00+08:00",
            "longitude": 116.4,
            "yong_shen": "官鬼",
        },
    }


def test_ask_schemas_accept_primary_calls():
    validate({"event": _liuyao_event()}, _SCHEMAS["liuyao_ask"])
    # 旧事件仍可读：question/origin/focus 只作为历史兼容输入。
    validate({"event": {
        "casting": _liuyao_event()["casting"],
        "casting_context": {
            "standard_time": "2026-10-01T10:00:00+08:00",
            "longitude": 116.4,
            "focus": {"kind": "matter", "matter": "career"},
        },
    }}, _SCHEMAS["liuyao_ask"])
    for invalid_event in (
        {"casting": _liuyao_event()["casting"], "casting_context": {
            "standard_time": "2026-10-01T10:00:00+08:00",
            "longitude": 116.4,
        }},
        {"casting": _liuyao_event()["casting"], "casting_context": {
            "standard_time": "2026-10-01T10:00:00+08:00",
            "longitude": 116.4,
            "yong_shen": "官鬼",
            "focus": {"kind": "matter", "matter": "career"},
        }},
        {"casting": _liuyao_event()["casting"], "casting_context": {
            "standard_time": "2026-10-01T10:00:00+08:00",
            "longitude": 116.4,
            "yong_shen": "财神",
        }},
    ):
        try:
            validate({"event": invalid_event}, _SCHEMAS["liuyao_ask"])
        except Exception:
            pass
        else:
            raise AssertionError(f"invalid event accepted: {invalid_event}")
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
            expected = {
                "technique", "evidence", "timing_candidates",
                "condition_candidates", "interpretation", "safety_advisory",
            }
        assert expected <= set(properties)
        assert "reading_digest" not in properties
        assert "case" not in properties
        assert "factors" not in properties
        assert "event" not in properties
        assert "chart" not in properties
