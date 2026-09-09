from __future__ import annotations

import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_answer_core  # noqa: E402


def _snapshot():
    return {
        "evidence": {
            "primary": [{"id": "yong-shen-state", "fact": {}}],
            "secondary": [{"id": "pattern-1", "fact": {}}],
            "reference": [{"id": "hexagram-name", "fact": {}}],
            "conflicts": [{"id": "strong-but-void", "reason": "须辨真假空"}],
        },
        "timing_candidates": [{"id": "void-fill-1-子", "mechanism": "旬空填实"}],
    }


def test_build_uses_snapshot_refs():
    core = liuyao_answer_core.build(_snapshot())
    assert core["primary_evidence_refs"] == ["yong-shen-state"]
    assert core["secondary_evidence_refs"] == ["pattern-1"]
    assert core["timing_refs"] == ["void-fill-1-子"]
    assert core["conflicts"][0]["id"] == "strong-but-void"


def test_validate_accepts_conditioned_core():
    core = liuyao_answer_core.build(_snapshot())
    core.update({
        "headline": "事项暂受旬空影响，倾向迟成",
        "verdict": "用神旺相但旬空，事有基础而当下未实，宜待条件补足。",
        "action": "先完成现实核查，再择机推进。",
    })
    result = liuyao_answer_core.validate_core(core, _snapshot())
    assert result["accepted"] is True
    assert result["errors"] == []


def test_validate_rejects_unknown_refs_and_absolutism():
    core = liuyao_answer_core.build(_snapshot())
    core.update({
        "headline": "必然成功",
        "verdict": "此事百分百会成。",
        "action": "直接推进。",
        "primary_evidence_refs": ["not-real"],
        "timing_refs": ["not-timing"],
        "conflicts": [],
        "unknown_field": True,
    })
    result = liuyao_answer_core.validate_core(core, _snapshot())
    assert result["accepted"] is False
    codes = {error.get("reason", error.get("field")) for error in result["errors"]}
    assert "unknown evidence id" in codes
    assert "unknown timing candidate id" in codes
    assert "contains 必然" in codes
    assert "unknown field" in codes
