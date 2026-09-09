from __future__ import annotations

import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import liuyao_matters  # noqa: E402
import liuyao_read  # noqa: E402


def _fake_chart(**kwargs):
    snapshot = {
        "schema_version": "liuyao-snapshot-v2",
        "casting": {"mode": "coins", "fingerprint": "abc", "yaos": [7]*6, "dong_yao": []},
    }
    return {
        "matter": {"matter": "career"},
        "solar_time": "1984-02-15T08:00:00+08:00",
        "casting": {"mode": "coins", "fingerprint": "abc", "yaos": [7]*6, "dong_yao": []},
        "chart": {"name": "乾为天"},
        "snapshot": snapshot,
    }
import liuyao_topic_guidance  # noqa: E402


def test_canonical_matter_aliases_are_normalized():
    assert liuyao_matters.resolve_matter("study")["name"] == "学业文书"
    assert liuyao_matters.resolve_matter("academic")["name"] == "学业文书"
    assert liuyao_matters.resolve_matter("legal")["name"] == "争议纠纷"
    assert liuyao_matters.resolve_matter("legal_risk")["name"] == "争议纠纷"
    assert liuyao_matters.resolve_matter("marriage", perspective="female")["yong_shen"] == "官鬼"


def test_topic_guidance_uses_canonical_topic_and_primary_roles():
    guidance = liuyao_topic_guidance.get_topic("relationship")
    assert guidance["topic"] == "relationship"
    assert "xi_shen" not in guidance
    projected = liuyao_topic_guidance.project_topic_guidance({"evidence": {"primary": [], "secondary": []}}, "relationship")
    assert set(projected["force_roles"]) == {"yuan_shen", "ji_shen", "chou_shen"}


def test_liuyao_read_is_reading_aggregate(monkeypatch):
    monkeypatch.setattr(liuyao_read, "liuyao_chart", _fake_chart)
    result = liuyao_read.read(
        casting={"mode": "coins", "fingerprint": "abc", "yaos": [7] * 6, "dong_yao": []},
        solar_time="1984-02-15T08:00:00+08:00",
        matter="career",
        question="这次面试能不能通过",
    )
    assert result["$schema"] == "liki:liuyao-reading-v1"
    assert result["schema_version"] == "liuyao-reading-v1"
    assert len(result["reading_id"]) == 64
    assert len(result["reading_digest"]) == 64
    assert result["topic"] == "career"
    assert result["snapshot"]["schema_version"] == "liuyao-snapshot-v2"
