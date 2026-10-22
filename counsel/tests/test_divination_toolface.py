"""Divination tool-face contract tests (moved from the root pydantic-free suite).

These cases used to validate instances against the hand-written
``skill-tools.json`` schema. They now validate against the Pydantic models that
own the contract, so the cross-field clauses the legacy schema declared as
``allOf`` are exercised at runtime instead of statically.
"""

from __future__ import annotations

import pytest
from app.contracts.divination import LiuyaoSnapshotInput, QimenSnapshotInput
from pydantic import BaseModel, ValidationError


def _rejects(model: type[BaseModel], instance: dict) -> None:
    with pytest.raises(ValidationError):
        model.model_validate(instance)


def _accepts(model: type[BaseModel], instance: dict) -> None:
    model.model_validate(instance)


def test_liuyao_snapshot_input_rejects_ambiguous_sources_and_modes():
    base = {"question": "这次面试能不能通过？", "matter": "career", "longitude": 116.4, "timezone": "Asia/Shanghai"}
    _rejects(LiuyaoSnapshotInput, {**base, "yong_shen": "官鬼"})
    _rejects(LiuyaoSnapshotInput, {**base, "mode": "coins"})
    _rejects(LiuyaoSnapshotInput, {**base, "mode": "yaos"})
    _rejects(LiuyaoSnapshotInput, {
        **base, "mode": "auto",
        "rounds": [["正", "反", "反"]] * 6,
    })
    _rejects(LiuyaoSnapshotInput, {**base, "perspective": "male"})
    _accepts(LiuyaoSnapshotInput, {
        **base, "mode": "coins",
        "rounds": [["正", "反", "反"]] * 6,
    })


def test_qimen_snapshot_input_rejects_ambiguous_location_and_focus():
    hour = {"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"}
    _rejects(QimenSnapshotInput, {
        "question": "该往哪里推进？", "city": "上海", "longitude": 121.47,
        "timezone": "Asia/Shanghai", "method": hour,
    })
    _rejects(QimenSnapshotInput, {
        "question": "该往哪里推进？", "longitude": 121.47,
        "timezone": "Asia/Shanghai", "method": hour,
        "matter": "wealth", "yong_shen": ["生门"],
    })
    _rejects(QimenSnapshotInput, {
        "question": "该往哪里推进？", "longitude": 121.47,
        "timezone": "Asia/Shanghai", "matter": "wealth",
        "quarter_rule": "ten_minute_sanyuan",
    })
    _accepts(QimenSnapshotInput, {
        "question": "该往哪里推进？", "longitude": 121.47,
        "timezone": "Asia/Shanghai", "method": hour, "matter": "wealth",
    })


def test_qimen_snapshot_method_schema_rejects_incompatible_forms():
    location = {"question": "该往哪里推进？", "longitude": 121.47, "timezone": "Asia/Shanghai"}

    _rejects(QimenSnapshotInput, {
        **location,
        "method": {"scope": "hour", "school": "jinhan_yujing"},
        "rule": "lost_property",
    })
    _rejects(QimenSnapshotInput, {
        **location,
        "method": {"scope": "day", "school": "mingfa_feipan", "dingju_method": "chaibu"},
    })
    _rejects(QimenSnapshotInput, {
        **location,
        "method": {
            "scope": "quarter", "school": "mingfa_feipan",
            "quarter_rule": "ten_minute_sanyuan",
        },
    })
    _rejects(QimenSnapshotInput, {
        **location,
        "method": {
            "scope": "quarter", "school": "zhuanpan",
            "quarter_rule": "ten_minute_sanyuan", "base_dingju_method": "chaibu",
        },
    })
    _rejects(QimenSnapshotInput, {
        **location,
        "method": {"scope": "month", "school": "zhuanpan", "dingju_method": "chaibu"},
    })
    _rejects(QimenSnapshotInput, {
        **location,
        "method": {"scope": "hour", "school": "zhuanpan", "dingju_method": "bad"},
    })

    _accepts(QimenSnapshotInput, {
        **location,
        "method": {
            "scope": "quarter", "school": "zhuanpan",
            "quarter_rule": "twelve_minute_ten_division",
            "base_dingju_method": "zhirun", "dun_source": "solar_term",
            "hour_boundary": "zi_zheng",
        },
    })
    _accepts(QimenSnapshotInput, {
        "question": "今天整体态势如何？", "longitude": 121.47,
        "timezone": "Asia/Shanghai", "method": {"scope": "day", "school": "jinhan_yujing"},
    })
    _accepts(QimenSnapshotInput, {
        "question": "钥匙丢了还能找到吗？", "longitude": 121.47,
        "timezone": "Asia/Shanghai",
        "method": {"scope": "hour", "school": "zhuanpan", "dingju_method": "maoshan"},
        "rule": "lost_property",
    })


def test_qimen_snapshot_input_rejects_invalid_special_focus_policy():
    hour = {"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"}
    location = {"longitude": 121.47, "timezone": "Asia/Shanghai"}

    _rejects(QimenSnapshotInput, {
        "question": "钥匙丢了还能找到吗？", "method": hour,
        "rule": "lost_property", "matter": "lost_item", **location,
    })
    _rejects(QimenSnapshotInput, {
        "question": "钥匙丢了还能找到吗？", "method": hour,
        "rule": "lost_property", "yong_shen": ["生门"], **location,
    })
    _rejects(QimenSnapshotInput, {
        "question": "家人走失了", "method": hour,
        "rule": "missing_person", **location,
    })
    _rejects(QimenSnapshotInput, {
        "question": "家人走失了", "method": hour,
        "rule": "missing_person", "matter": "career", **location,
    })

    _accepts(QimenSnapshotInput, {
        "question": "家人走失了", "method": hour,
        "rule": "missing_person", "matter": "missing_person", **location,
    })
