from __future__ import annotations

import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_session  # noqa: E402


def _snapshot():
    return {
        "schema_version": "liuyao-snapshot-v2",
        "casting": {"fingerprint": "abc", "yaos": [7, 7, 7, 7, 7, 7], "dong_yao": []},
    }


def test_create_and_validate_session():
    session = liuyao_session.create_session(
        question="这次面试能不能通过？",
        snapshot=_snapshot(),
        headline="暂有阻碍，倾向待时",
        verdict="用神旬空，事未落实，出空或填实后再看。",
        confidence="medium",
    )
    assert session["schema_version"] == "liuyao-session-v1"
    assert session["$schema"] == "liki:liuyao-session-v1"
    assert set(session["integrity"]) == {"casting", "snapshot", "first_verdict", "report"}
    assert session["casting"]["fingerprint"] == "abc"
    assert liuyao_session.validate_session(session)["session_id"] == session["session_id"]


def test_append_reality_feedback_requires_outcome():
    session = liuyao_session.create_session(
        question="这次面试能不能通过？",
        snapshot=_snapshot(),
        headline="暂有阻碍",
        verdict="待条件落实。",
    )
    with pytest.raises(ValueError, match="outcome"):
        liuyao_session.append_followup(session, kind="reality_feedback", text="结果已出")
    updated = liuyao_session.append_followup(
        session, kind="reality_feedback", text="已收到通过通知", outcome="supported"
    )
    assert updated["followups"][-1]["outcome"] == "supported"


def test_validate_rejects_changed_casting():
    session = liuyao_session.create_session(
        question="测试", snapshot=_snapshot(), headline="首次结论", verdict="原卦结论。"
    )
    session["first_verdict"]["headline"] = "被篡改的结论"
    with pytest.raises(ValueError, match="session integrity mismatch"):
        liuyao_session.validate_session(session)
