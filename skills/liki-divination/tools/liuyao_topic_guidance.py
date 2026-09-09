"""六爻专题断法库；返回规则引导，不产生吉凶结论。"""
from __future__ import annotations

import json
from pathlib import Path


PATH = Path(__file__).with_name("liuyao_topic_methods.json")
_DATA = None


def _data() -> dict:
    global _DATA
    if _DATA is None:
        _DATA = json.loads(PATH.read_text(encoding="utf-8"))
    return _DATA


def list_topics() -> list[str]:
    return sorted(_data()["topics"])


def get_topic(topic: str) -> dict:
    data = _data()
    if topic not in data["topics"]:
        raise ValueError(f"unknown topic: {topic}; valid: {', '.join(list_topics())}")
    return {
        "schema_version": data["schema_version"],
        "topic": topic,
        **data["topics"][topic],
        "conclusion_scope": data["conclusion_scope"],
    }


def project_topic_guidance(snapshot: dict, topic: str) -> dict:
    """把专题规则与 snapshot 主判层绑定；仍不输出吉凶。"""
    guidance = get_topic(topic)
    evidence = snapshot.get("evidence", {})
    primary_ids = [
        item.get("id")
        for item in evidence.get("primary", [])
        if isinstance(item, dict) and item.get("id")
    ]
    secondary_ids = [
        item.get("id")
        for item in evidence.get("secondary", [])
        if isinstance(item, dict) and item.get("id")
    ]
    return {
        "schema_version": "liuyao-topic-guidance-v1",
        "topic": topic,
        "name": guidance["name"],
        "yong_shen": guidance["yong_shen"],
        "force_roles": {
            "yuan_shen": guidance["yuan_shen"],
            "ji_shen": guidance["ji_shen"],
            "chou_shen": guidance["chou_shen"],
        },
        "available_primary_evidence": primary_ids,
        "available_secondary_evidence": secondary_ids,
        "primary_factors": guidance["primary_factors"],
        "reference_factors": guidance["reference_factors"],
        "timing_focus": guidance["timing_focus"],
        "common_misjudgments": guidance["common_misjudgments"],
        "boundary": guidance["boundary"],
        "conclusion_scope": guidance["conclusion_scope"],
    }
