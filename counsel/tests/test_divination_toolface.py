"""Divination tool-face contract tests."""

import pytest
from app.contracts.divination import LiuyaoAskInput, QimenAskInput
from pydantic import BaseModel, ValidationError


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


def _rejects(model: type[BaseModel], instance: dict) -> None:
    with pytest.raises(ValidationError):
        model.model_validate(instance)


def _accepts(model: type[BaseModel], instance: dict) -> None:
    model.model_validate(instance)


def test_liuyao_ask_input_rejects_ambiguous_sources_and_modes():
    base = {"event": _liuyao_event({"kind": "matter", "matter": "career"})}
    auto_event = dict(base["event"])
    auto_event["casting"] = {"mode": "auto"}
    _rejects(LiuyaoAskInput, {**base, "casting": base["event"]["casting"]})
    _rejects(LiuyaoAskInput, {"event": auto_event})
    _rejects(LiuyaoAskInput, {**base, "perspective": "male"})
    _accepts(LiuyaoAskInput, base)


def test_liuyao_ask_input_rejects_origin_receipt_mismatch():
    event = _liuyao_event({"kind": "matter", "matter": "career"})
    event["origin"] = {"mode": "auto"}
    _rejects(LiuyaoAskInput, {"event": event})


def test_liuyao_ask_input_accepts_explicit_null_topic():
    base = {"event": _liuyao_event({"kind": "matter", "matter": "career"})}
    _accepts(LiuyaoAskInput, {**base, "topic": None})
    schema = LiuyaoAskInput.model_json_schema()["properties"]["topic"]
    assert any(branch.get("type") == "null" for branch in schema.get("anyOf", []))
    _rejects(LiuyaoAskInput, {**base, "topic": "not-a-topic"})


def test_liuyao_relationship_focus_requires_perspective():
    def instance(perspective):
        focus = {"kind": "matter", "matter": "relationship"}
        if perspective:
            focus["perspective"] = perspective
        return {"event": _liuyao_event(focus)}
    _rejects(LiuyaoAskInput, instance(None))
    _accepts(LiuyaoAskInput, {
        **instance("male"),
    })


def test_qimen_ask_input_rejects_invalid_event_and_focus():
    hour = {"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"}
    _rejects(QimenAskInput, {
        "question": "该往哪里推进？",
        "longitude": 121.47,
        "standard_time": "2026-10-01T10:00:00+08:00",
        "method": hour,
    })
    _rejects(QimenAskInput, {
        "question": "该往哪里推进？",
        "longitude": 121.47,
        "standard_time": "2026-10-01T10:00:00+08:00",
        "method": hour,
        "focus": {"kind": "matter", "matter": "wealth", "yong_shen": ["生门"]},
    })
    _rejects(QimenAskInput, {
        "question": "该往哪里推进？",
        "longitude": 121.47,
        "standard_time": "2026-10-01T10:00:00+08:00",
        "method": {
            "scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu",
        },
        "focus": {"kind": "matter", "matter": "wealth"},
        "quarter_rule": "ten_minute_sanyuan",
    })
    _accepts(QimenAskInput, {
        "question": "该往哪里推进？",
        "longitude": 121.47,
        "standard_time": "2026-10-01T10:00:00+08:00",
        "method": hour,
        "focus": {"kind": "matter", "matter": "wealth"},
    })


def test_qimen_ask_method_schema_rejects_incompatible_forms():
    event = {
        "question": "该往哪里推进？",
        "longitude": 121.47,
        "standard_time": "2026-10-01T10:00:00+08:00",
    }

    _rejects(QimenAskInput, {
        **event,
        "method": {"scope": "hour", "school": "jinhan_yujing"},
        "rule": "lost_property",
    })
    _rejects(QimenAskInput, {
        **event,
        "method": {"scope": "day", "school": "mingfa_feipan", "dingju_method": "chaibu"},
    })
    _rejects(QimenAskInput, {
        **event,
        "method": {
            "scope": "quarter", "school": "mingfa_feipan",
            "quarter_rule": "ten_minute_sanyuan",
        },
    })
    _rejects(QimenAskInput, {
        **event,
        "method": {
            "scope": "quarter", "school": "zhuanpan",
            "quarter_rule": "ten_minute_sanyuan", "base_dingju_method": "chaibu",
        },
    })
    _rejects(QimenAskInput, {
        **event,
        "method": {"scope": "month", "school": "zhuanpan", "dingju_method": "chaibu"},
    })
    _rejects(QimenAskInput, {
        **event,
        "method": {"scope": "hour", "school": "zhuanpan", "dingju_method": "bad"},
    })

    _accepts(QimenAskInput, {
        **event,
        "method": {
            "scope": "quarter", "school": "zhuanpan",
            "quarter_rule": "twelve_minute_ten_division",
            "base_dingju_method": "chaibu",
            "dun_source": "hour_branch",
            "hour_boundary": "late_zi",
        },
    })
