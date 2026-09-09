from __future__ import annotations

import sys
from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_audit  # noqa: E402
import liuyao_read  # noqa: E402
import liuyao_report  # noqa: E402
import liuyao_session  # noqa: E402


def _fake_chart(**kwargs):
    snapshot = {
        "schema_version": "liuyao-snapshot-v2",
        "question": {"text": kwargs.get("question", ""), "domain": kwargs.get("matter") or "advanced"},
        "casting": {
            "mode": "coins", "fingerprint": "abc", "rounds": [],
            "yaos": [7, 7, 7, 7, 7, 7], "dong_yao": [],
        },
        "board": {
            "name": "乾为天", "ben_gua": "乾", "bian_gua": None,
            "palace": "乾", "palace_wuxing": "金",
            "lines": [],
        },
        "focus": {
            "yong_shen": {"name": "官鬼", "position": 4, "wang_shuai": "相"},
            "yong_line": None, "world": None, "other": None,
        },
        "evidence": {
            "primary": [{"id": "yong-shen-state", "fact": {}}],
            "secondary": [], "reference": [], "conflicts": [], "ignored_scope": [],
        },
        "facts": {
            "hidden_lines": [], "branch_relation_facts": [], "day_clash_facts": [],
            "moving_transformations": [], "san_he_candidates": [],
            "force_chain": {"yong_shen": "官鬼", "yong_element": "火", "position": 4, "entries": []},
            "yong_shen_candidates": [],
        },
        "timing_candidates": [],
        "policy": {"may_output": [], "must_cover_conflicts": True, "forbidden": []},
        "followup": {
            "locked": True, "casting_fingerprint": "abc", "yaos": [7] * 6,
            "dong_yao": [], "rules": [],
        },
    }
    return {
        "matter": {"matter": "career"},
        "solar_time": "2026-09-08T12:00:00+08:00",
        "casting": {"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
        "chart": {"name": "乾为天", "ben_gua": "乾", "lines": [
            {"position": i + 1, "liu_qin": liu_qin, "shi_ying": shi_ying}
            for i, (liu_qin, shi_ying) in enumerate([
                ("子孙", ""), ("妻财", ""), ("父母", "应"),
                ("官鬼", ""), ("兄弟", ""), ("父母", "世"),
            ])
        ]},
        "snapshot": snapshot,
    }


def test_read_report_audit_session_flow(monkeypatch):
    monkeypatch.setattr(liuyao_read, "liuyao_chart", _fake_chart)
    reading = liuyao_read.read(
        casting={"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
        matter="career",
        topic="career",
        question="这次面试能不能通过",
    )
    report = liuyao_report.template(reading["snapshot"])
    report.update({
        "headline": "面试当前有支持条件，结果仍需现实核查",
        "verdict": "用神主线未受直接冲击，倾向偏可推进，但应期须等条件确认。",
        "action": "先确认面试时间与岗位要求，再按当前准备节奏推进。",
    })
    report_audit = liuyao_audit.audit_report(report, reading)
    assert report_audit["accepted"] is True, report_audit
    session = liuyao_session.create_session(
        question="这次面试能不能通过",
        snapshot=reading["snapshot"],
        headline=report["headline"],
        verdict=report["verdict"],
        confidence=report["confidence"],
        report=report,
    )
    assert session["casting"]["fingerprint"] == "abc"
    assert session["first_verdict"]["report"]["schema_version"] == "liuyao-report-v1"
    assert session["snapshot_digest"] == liuyao_session._digest(reading["snapshot"])


def test_session_rejects_invalid_report_before_locking(monkeypatch):
    monkeypatch.setattr(liuyao_read, "liuyao_chart", _fake_chart)
    reading = liuyao_read.read(
        casting={"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
        matter="career", question="测试",
    )
    report = liuyao_report.template(reading["snapshot"])
    report.update({
        "headline": "必然成功",
        "verdict": "此事百分百会成。",
        "action": "直接推进。",
        "primary_evidence_refs": ["not-real"],
    })
    with pytest.raises(ValueError, match="report must pass validation"):
        liuyao_session.create_session(
            question="测试", snapshot=reading["snapshot"],
            headline=report["headline"], verdict=report["verdict"],
            report=report,
        )
