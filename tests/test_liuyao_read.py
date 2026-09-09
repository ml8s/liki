from __future__ import annotations

import json
import sys
from pathlib import Path

from jsonschema import validate


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
FIXTURES = ROOT / "tests/fixtures/liuyao"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_read  # noqa: E402
import liuyao_report  # noqa: E402
import liuyao_session  # noqa: E402


def _fake_chart(**kwargs):
    snapshot = {
        "schema_version": "liuyao-snapshot-v2",
        "question": {"text": kwargs.get("question") or "", "domain": kwargs.get("matter") or "advanced"},
        "casting": {
            "mode": "coins", "fingerprint": "abc",
            "rounds": [], "yaos": [7] * 6, "dong_yao": [],
        },
        "board": {
            "name": "乾为天", "ben_gua": "乾", "bian_gua": None,
            "palace": "乾", "palace_wuxing": "金",
            "lines": [{
                "position": 1, "type": 7, "gan_zhi": "甲子", "wuxing": "水",
                "liu_qin": "子孙", "liu_shou": "螣蛇", "shi_ying": "",
                "chang_sheng_yue": "病",
                "flags": {"moving": False, "yue_po": False, "xun_kong": False,
                          "mu_ku": False, "mu_ku_branch": None, "mu_ku_element": None,
                          "dong_sheng": False, "dong_ke": False},
            }] * 6,
        },
        "focus": {
            "yong_shen": {"name": "官鬼", "position": 4, "wang_shuai": "相"},
            "yong_line": None, "world": None, "other": None,
        },
        "evidence": {"primary": [{"id": "yong-shen-state", "fact": {}}], "secondary": [], "reference": [], "conflicts": [], "ignored_scope": []},
        "facts": {
            "hidden_lines": [], "branch_relation_facts": [], "day_clash_facts": [],
            "moving_transformations": [], "san_he_candidates": [],
            "force_chain": {"yong_shen": "官鬼", "yong_element": "火", "position": 4, "entries": []},
            "yong_shen_candidates": [],
        },
        "timing_candidates": [],
        "policy": {"may_output": [], "must_cover_conflicts": True, "forbidden": []},
        "followup": {"locked": True, "casting_fingerprint": "abc", "yaos": [7] * 6, "dong_yao": [], "rules": []},
    }
    return {
        "matter": {"matter": "career"},
        "solar_time": "2026-09-08T12:00:00+08:00",
        "casting": {"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
        "chart": {"name": "乾为天"},
        "snapshot": snapshot,
    }


def test_read_binds_topic_timing_conditions_and_report(monkeypatch):
    monkeypatch.setattr(liuyao_read, "liuyao_chart", _fake_chart)
    result = liuyao_read.read(
        casting={"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
        matter="career",
        topic="career",
        question="这次面试能不能通过",
    )
    assert result["schema_version"] == "liuyao-reading-v1"
    assert result["topic"] == "career"
    assert result["topic_guidance"]["yong_shen"] == "官鬼"
    assert result["timing_plan"]["blocked"] is False
    assert result["condition_rules"]["policy"]["must_restore_hidden_conditions"] is True
    assert result["report_template"]["schema_version"] == "liuyao-report-v1"
    assert "liuyao_report validate" in result["next_actions"][1]


def test_snapshot_report_and_session_contracts_are_valid():
    from jsonschema import validate

    snapshot_schema = json.loads((TOOLS / "liuyao_snapshot_contract.json").read_text(encoding="utf-8"))
    report_schema = json.loads((TOOLS / "liuyao_report_contract.json").read_text(encoding="utf-8"))
    session_schema = json.loads((TOOLS / "liuyao_session_contract.json").read_text(encoding="utf-8"))

    monkeypatch_result = _fake_chart(casting={"mode": "coins"}, matter="career", question="测试")
    snapshot = monkeypatch_result["snapshot"]
    report = liuyao_report.template(snapshot)
    report.update({
        "headline": "事项当前稳定，倾向需条件确认",
        "verdict": "用神主线无发动阻碍，结果仍需核对现实条件。",
        "action": "先完成一次现实核查，再决定推进节奏。",
    })
    validate(snapshot, snapshot_schema)
    validate(report, report_schema)
    session = liuyao_session.create_session(
        question="测试", snapshot=snapshot,
        headline=report["headline"], verdict=report["verdict"],
        confidence=report["confidence"], report=report,
    )
    validate(session, session_schema)


def test_read_tool_schema_accepts_primary_call():
    schema = json.loads((TOOLS / "skill-tools.json").read_text(encoding="utf-8"))
    tool = next(item["function"] for item in schema["tools"] if item["function"]["name"] == "liuyao_read")
    validate(
        {
            "casting": {"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
            "matter": "career",
            "topic": "career",
            "question": "这次面试能不能通过",
        },
        tool["parameters"],
    )
