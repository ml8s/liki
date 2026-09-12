from __future__ import annotations

import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_timing  # noqa: E402
import liuyao_topic_guidance  # noqa: E402


def test_topic_library_loads_all_topics():
    topics = sorted(liuyao_topic_guidance._data()["topics"])
    assert {
        "wealth", "career", "relationship", "study", "lost_item",
        "travel", "legal", "health_context"
    } <= set(topics)


def test_project_topic_guidance_returns_primary_layers():
    snapshot = {
        "evidence": {
            "primary": [{"id": "yong-shen-state"}],
            "secondary": [{"id": "pattern-1"}],
        }
    }
    result = liuyao_topic_guidance.project_topic_guidance(snapshot, "wealth")
    assert result["yong_shen"] == "妻财"
    assert result["available_primary_evidence"] == ["yong-shen-state"]
    assert result["available_secondary_evidence"] == ["pattern-1"]
    assert result["conclusion_scope"] == "rule_guidance_not_outcome"


def test_timing_ranks_candidates_without_dates():
    snapshot = {
        "timing_candidates": [
            {"id": "a", "mechanism": "静爻逢冲", "position": 2, "confidence": "candidate"},
            {"id": "b", "mechanism": "旬空填实", "position": 2, "confidence": "candidate"},
        ]
    }
    result = liuyao_timing.rank_timing_candidates(snapshot, topic="wealth")
    assert result["blocked"] is False
    assert result["candidates"][0]["id"] == "b"
    assert all(
        item["conclusion_scope"] == "conditional_timing_candidate_not_date"
        for item in result["candidates"]
    )


def test_timing_position_alone_is_not_target_evidence():
    snapshot = {
        "timing_candidates": [
            {"id": "with-position", "mechanism": "出月令", "position": 3, "confidence": "candidate"},
            {"id": "without-position", "mechanism": "出月令", "confidence": "candidate"},
        ]
    }
    result = liuyao_timing.rank_timing_candidates(snapshot, topic="wealth")
    scores = {item["id"]: item["rank_score"] for item in result["candidates"]}
    assert scores["with-position"] == scores["without-position"]


def test_timing_keeps_health_topic_with_unified_boundary():
    snapshot = {
        "timing_candidates": [
            {"id": "health-watch", "mechanism": "旬空填实", "confidence": "candidate"}
        ]
    }
    result = liuyao_timing.rank_timing_candidates(snapshot, topic="health_context")
    assert result["blocked"] is False
    assert [item["id"] for item in result["candidates"]] == ["health-watch"]
    assert "不构成医疗" in result["boundary"]
    assert result["policy"]["advice_scope"] == "traditional_reference_not_professional_advice"
