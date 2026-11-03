"""Qimen answer-core reference and timing guardrails."""

import app.divination.tools.qimen_answer_core as core
import app.divination.tools.qimen_jinhan_answer_core as jinhan_core


def _case():
    return {
        "technique": "qimen",
        "question": "测试",
        "method": {"scope": "hour", "school": "zhuanpan"},
        "factors": {
            "patterns": [{"name": "test"}],
            "ying_qi": [],
        },
        "evidence": [{"id": "case:patterns"}],
        "special": {"assertions": [{"id": "assertion_only"}]},
        "judgments": {"results": []},
        "timing_candidates": [{"id": "timing-1", "related_to": "用神"}],
    }


def test_refs_preserve_case_and_assertion_namespaces():
    built = core.build(_case())
    assert "case:patterns" in built["refs"]
    assert "assertion:assertion_only" in built["refs"]

    audit = core.validate_core({**built, "refs": ["assertion:fake"]}, _case())
    assert audit["accepted"] is False
    assert any(error.get("field") == "refs" for error in audit["errors"])


def test_timing_ids_are_available_and_tampering_rejected():
    case = _case()
    built = core.build(case)
    assert "timing-1" in built["timing_refs"]
    audit = core.validate_core({**built, "timing_refs": ["timing-fake"]}, case)
    assert audit["accepted"] is False


def test_jinhan_rejects_assertion_refs():
    case = {
        "question": "测试",
        "method": {"scope": "day", "school": "jinhan_yujing"},
        "factors": {
            "kind": "jinhan", "scope": "day", "school": "jinhan_yujing",
            "palaces": [{"gong": str(index)} for index in range(9)],
            "day_spirits": [{"zhi": str(index)} for index in range(12)],
        },
    }
    built = jinhan_core.build(case)
    audit = jinhan_core.validate_core({**built, "refs": ["assertion:fake"]}, case)
    assert audit["accepted"] is False
