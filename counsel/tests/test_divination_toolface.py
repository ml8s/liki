"""Divination tool-face contract tests."""

import pytest
from app.contracts.divination import LiuyaoSnapshotInput, QimenSnapshotInput
from pydantic import BaseModel, ValidationError


def _rejects(model: type[BaseModel], instance: dict) -> None:
    with pytest.raises(ValidationError):
        model.model_validate(instance)


def _accepts(model: type[BaseModel], instance: dict) -> None:
    model.model_validate(instance)


def test_liuyao_snapshot_input_rejects_ambiguous_sources_and_modes():
    base = {
        "question": "这次面试能不能通过？",
        "focus": {"kind": "matter", "matter": "career"},
        "qigua_longitude": 116.4,
        "qigua_time": "2026-10-01T10:00:00+08:00",
    }
    _rejects(LiuyaoSnapshotInput, {**base, "yong_shen": "官鬼"})
    _rejects(LiuyaoSnapshotInput, {**base, "casting": {"mode": "coins"}})
    _rejects(LiuyaoSnapshotInput, {**base, "casting": {"mode": "yaos"}})
    _rejects(LiuyaoSnapshotInput, {
        **base, "casting": {"mode": "auto", "rounds": [["正", "反", "反"]] * 6},
    })
    _rejects(LiuyaoSnapshotInput, {**base, "perspective": "male"})
    _accepts(LiuyaoSnapshotInput, {
        **base, "casting": {"mode": "coins", "rounds": [["正", "反", "反"]] * 6},
    })
    _accepts(LiuyaoSnapshotInput, base)


def test_liuyao_relationship_focus_requires_perspective():
    _rejects(LiuyaoSnapshotInput, {
        "question": "感情如何？",
        "focus": {"kind": "matter", "matter": "relationship"},
        "qigua_longitude": 116.4,
        "qigua_time": "2026-10-01T10:00:00+08:00",
    })
    _accepts(LiuyaoSnapshotInput, {
        "question": "感情如何？",
        "focus": {"kind": "matter", "matter": "relationship", "perspective": "male"},
        "qigua_longitude": 116.4,
        "qigua_time": "2026-10-01T10:00:00+08:00",
    })


def test_qimen_snapshot_input_rejects_ambiguous_location_and_focus():
    hour = {"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"}
    _rejects(QimenSnapshotInput, {
        "question": "该往哪里推进？",
        "location": {"kind": "city", "city": "上海"},
        "qiju_longitude": 121.47,
        "qiju_time": "2026-10-01T10:00:00+08:00",
        "method": hour,
    })
    _rejects(QimenSnapshotInput, {
        "question": "该往哪里推进？",
        "location": {"kind": "longitude", "qiju_longitude": 121.47},
        "qiju_time": "2026-10-01T10:00:00+08:00",
        "method": hour,
        "focus": {"kind": "matter", "matter": "wealth", "yong_shen": ["生门"]},
    })
    _rejects(QimenSnapshotInput, {
        "question": "该往哪里推进？",
        "location": {"kind": "longitude", "qiju_longitude": 121.47},
        "qiju_time": "2026-10-01T10:00:00+08:00",
        "method": {
            "scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu",
        },
        "focus": {"kind": "matter", "matter": "wealth"},
        "quarter_rule": "ten_minute_sanyuan",
    })
    _accepts(QimenSnapshotInput, {
        "question": "该往哪里推进？",
        "location": {"kind": "longitude", "qiju_longitude": 121.47},
        "qiju_time": "2026-10-01T10:00:00+08:00",
        "method": hour,
        "focus": {"kind": "matter", "matter": "wealth"},
    })


def test_qimen_snapshot_method_schema_rejects_incompatible_forms():
    location = {
        "question": "该往哪里推进？",
        "location": {"kind": "longitude", "qiju_longitude": 121.47},
        "qiju_time": "2026-10-01T10:00:00+08:00",
    }

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
            "base_dingju_method": "chaibu",
            "dun_source": "hour_branch",
            "hour_boundary": "late_zi",
        },
    })
