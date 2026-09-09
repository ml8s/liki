from __future__ import annotations

import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import qimen_report  # noqa: E402
import qimen_session  # noqa: E402


def _read_result():
    snapshot = {
        "method": {"scope": "hour", "school": "zhuanpan"},
        "ri_gan_gong": "震",
        "shi_gan_gong": "兑",
        "ying_qi": [
            {"type": "ma_xing", "branch": "寅", "gong": "艮", "related_to": [{"symbol": "日干"}]}
        ],
    }
    special = {
        "rule": "lost_property",
        "assertions": [{
            "id": "qimen_lost_property_direction",
            "name": "失物方向",
            "conclusion": "候选方向",
            "basis": "表",
            "evidence": [],
        }],
    }
    return {
        "schema_version": "qimen-read-v1",
        "question": "钥匙还能找到吗",
        "input": {
            "city": "上海", "longitude": 121.47,
            "local_time": "2026-09-08T12:00:00+08:00",
            "solar_time": "2026-09-08T11:57:00+08:00",
        },
        "snapshot": snapshot,
        "special": special,
    }


def test_report_template_contains_snapshot_and_assertion_refs():
    read_result = _read_result()
    report = qimen_report.template(read_result)
    assert "snapshot:method" in report["evidence_refs"]
    assert "snapshot:ri_gan_gong" in report["evidence_refs"]
    assert "assertion:qimen_lost_property_direction" in report["assertion_refs"]
    assert report["timing_refs"]


def test_validate_accepts_conditioned_report():
    report = qimen_report.template(_read_result())
    report.update({
        "headline": "失物方向有候选，暂不宜断定寻回",
        "verdict": "马星与用神方向给出线索，须结合寻找范围核实。",
        "action": "先按候选方向和近期路径查找。",
    })
    audit = qimen_report.validate_report(report, _read_result())
    assert audit["accepted"] is True
    assert audit["errors"] == []


def test_validate_rejects_unknown_refs_and_absolutism():
    report = qimen_report.template(_read_result())
    report.update({
        "headline": "必然找到",
        "verdict": "百分百在东北。",
        "action": "直接去。",
        "evidence_refs": ["snapshot:not-real"],
        "assertion_refs": ["assertion:not-real"],
        "timing_refs": ["timing:not-real"],
    })
    report["unknown_field"] = True
    audit = qimen_report.validate_report(report, _read_result())
    assert audit["accepted"] is False
    reasons = {item.get("reason") for item in audit["errors"]}
    assert {"unknown ref", "contains 必然", "contains 百分百", "unknown field"} <= reasons


def test_session_create_append_validate():
    read_result = _read_result()
    report = qimen_report.template(read_result)
    report.update({
        "headline": "有寻找方向候选",
        "verdict": "当前线索支持方向排查，不承诺寻回。",
        "action": "沿候选方向和近期路径查找。",
    })
    session = qimen_session.create_session(
        read_result=read_result, headline=report["headline"],
        verdict=report["verdict"], confidence="medium", report=report,
    )
    assert session["method"] == {"scope": "hour", "school": "zhuanpan"}
    assert session["$schema"] == "liki:qimen-session-v1"
    assert set(session["integrity"]) == {"input", "method", "snapshot", "first_verdict", "report"}
    updated = qimen_session.append_followup(
        session, kind="reality_feedback", text="已在该方向找到",
        outcome="supported",
    )
    assert updated["followups"][-1]["outcome"] == "supported"
    qimen_session.validate_session(updated)


def test_session_rejects_invalid_report():
    report = qimen_report.template(_read_result())
    report.update({
        "headline": "必然找到",
        "verdict": "百分百在东北。",
        "action": "直接去。",
    })
    with pytest.raises(ValueError, match="report must pass validation"):
        qimen_session.create_session(
            read_result=_read_result(), headline=report["headline"],
            verdict=report["verdict"], report=report,
        )


def test_session_rejects_tampered_first_verdict():
    read_result = _read_result()
    report = qimen_report.template(read_result)
    report.update({
        "headline": "有寻找方向候选",
        "verdict": "当前线索支持方向排查，不承诺寻回。",
        "action": "沿候选方向和近期路径查找。",
    })
    session = qimen_session.create_session(
        read_result=read_result,
        headline=report["headline"],
        verdict=report["verdict"],
        report=report,
    )
    session["first_verdict"]["headline"] = "被篡改"
    import pytest
    with pytest.raises(ValueError, match="session integrity mismatch"):
        qimen_session.validate_session(session)
