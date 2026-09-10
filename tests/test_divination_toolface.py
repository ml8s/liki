from __future__ import annotations

import json
import sys
from pathlib import Path

import pytest
from jsonschema import validate
from jsonschema.exceptions import ValidationError


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))


def _schema(name: str) -> dict:
    doc = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    return next(item["function"]["parameters"] for item in doc["tools"] if item["function"]["name"] == name)


def _assert_rejects(schema: dict, instance: dict) -> None:
    with pytest.raises(ValidationError):
        validate(instance, schema)


def test_liuyao_snapshot_schema_rejects_ambiguous_sources_and_modes():
    schema = _schema("liuyao_snapshot")
    base = {"question": "这次面试能不能通过？", "matter": "career"}
    _assert_rejects(schema, {**base, "yong_shen": "官鬼"})
    _assert_rejects(schema, {"question": "这次面试能不能通过？", "mode": "coins"})
    _assert_rejects(schema, {"question": "这次面试能不能通过？", "mode": "yaos"})
    _assert_rejects(schema, {
        "question": "这次面试能不能通过？", "mode": "auto",
        "rounds": [["正", "反", "反"]] * 6,
    })
    _assert_rejects(schema, {
        "question": "这次面试能不能通过？", "matter": "career", "perspective": "male",
    })
    validate({
        "question": "这次面试能不能通过？", "mode": "coins",
        "rounds": [["正", "反", "反"]] * 6, "matter": "career",
    }, schema)


def test_qimen_snapshot_schema_rejects_ambiguous_location_and_focus():
    schema = _schema("qimen_snapshot")
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "city": "上海", "longitude": 121.47,
    })
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "matter": "wealth", "yong_shen": ["生门"],
    })
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "longitude": 121.47, "quarter_rule": "ten_minute_sanyuan",
    })
    validate({
        "question": "该往哪里推进？", "longitude": 121.47, "matter": "wealth",
    }, schema)


def test_huangli_schema_rejects_ambiguous_range():
    schema = _schema("huangli_days")
    _assert_rejects(schema, {
        "question": "哪天适合签约？", "event": "sign",
        "end_date": "2026-10-31", "days": 7,
    })
    validate({
        "question": "哪天适合签约？", "event": "sign", "days": 7,
    }, schema)


def test_qimen_snapshot_schema_rejects_incompatible_methods():
    schema = _schema("qimen_snapshot")
    base = {"question": "该往哪里推进？", "longitude": 121.47, "matter": "wealth"}

    _assert_rejects(schema, {
        **base,
        "school": "jinhan_yujing",
        "scope": "hour",
        "rule": "lost_property",
    })
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "longitude": 121.47,
        "school": "mingfa_feipan", "scope": "day", "dingju_method": "chaibu",
    })
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "scope": "quarter", "school": "mingfa_feipan",
        "quarter_rule": "ten_minute_sanyuan",
    })
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "scope": "quarter", "school": "zhuanpan",
        "quarter_rule": "ten_minute_sanyuan", "base_dingju_method": "chaibu",
    })
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "scope": "month", "dingju_method": "chaibu",
    })
    _assert_rejects(schema, {
        "question": "该往哪里推进？", "scope": "hour", "school": "zhuanpan",
        "dingju_method": "maoshan",
    })

    validate({
        "question": "该往哪里推进？", "longitude": 121.47,
        "scope": "quarter", "school": "zhuanpan",
        "quarter_rule": "twelve_minute_ten_division",
        "base_dingju_method": "zhirun", "dun_source": "solar_term",
        "hour_boundary": "zi_zheng",
    }, schema)
    validate({
        "question": "今天整体态势如何？", "longitude": 121.47,
        "scope": "day", "school": "jinhan_yujing",
    }, schema)
    validate({
        "question": "钥匙丢了还能找到吗？", "longitude": 121.47,
        "scope": "hour", "school": "zhuanpan", "rule": "lost_property",
    }, schema)


def test_qimen_snapshot_schema_rejects_invalid_special_focus_policy():
    schema = _schema("qimen_snapshot")
    location = {"longitude": 121.47}

    _assert_rejects(schema, {
        "question": "钥匙丢了还能找到吗？", "scope": "hour", "school": "zhuanpan",
        "rule": "lost_property", "matter": "lost_item", **location,
    })
    _assert_rejects(schema, {
        "question": "钥匙丢了还能找到吗？", "scope": "hour", "school": "zhuanpan",
        "rule": "lost_property", "yong_shen": ["生门"], **location,
    })
    _assert_rejects(schema, {
        "question": "家人走失了", "scope": "hour", "school": "zhuanpan",
        "rule": "missing_person", **location,
    })
    _assert_rejects(schema, {
        "question": "家人走失了", "scope": "hour", "school": "zhuanpan",
        "rule": "missing_person", "matter": "career", **location,
    })

    validate({
        "question": "家人走失了", "scope": "hour", "school": "zhuanpan",
        "rule": "missing_person", "matter": "missing_person", **location,
    }, schema)
