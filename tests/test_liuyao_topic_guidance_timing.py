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
    topics = liuyao_topic_guidance.list_topics()
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


def test_timing_blocks_health_topic():
    result = liuyao_timing.rank_timing_candidates({"timing_candidates": []}, topic="health_context")
    assert result["blocked"] is True
