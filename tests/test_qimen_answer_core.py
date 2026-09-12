from __future__ import annotations

import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TOOLS = ROOT / "skills/liki-divination/tools"
if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))

import qimen_answer_core as core  # noqa: E402
import qimen_jinhan_answer_core as jinhan_core  # noqa: E402


def _snapshot() -> dict:
    return {
        "schema_version": "qimen-snapshot-v4",
        "question": "测试",
        "method_context": {"scope": "hour", "school": "zhuanpan"},
        "matter": {"matter": "career"},
        "factors": {
            "scope": "hour",
            "school": "zhuanpan",
            "patterns": [{"name": "test"}],
            "ying_qi": [],
        },
        "special": {"assertions": [{"id": "assertion_only"}]},
    }


def test_evidence_and_assertion_refs_use_disjoint_namespaces() -> None:
    snapshot = _snapshot()
    built = core.build(snapshot)
    assert built["assertion_refs"] == ["assertion:assertion_only"]

    tampered = {**built, "evidence_refs": ["assertion:assertion_only"]}
    audit = core.validate_core(tampered, snapshot)

    assert audit["accepted"] is False
    assert any(
        error.get("field") == "evidence_refs"
        for error in audit["errors"]
    )


def test_empty_factor_values_are_not_available_evidence() -> None:
    snapshot = _snapshot()
    snapshot["factors"]["yong_shen"] = []
    snapshot["factors"]["nian_gan_gong"] = None
    snapshot["factors"]["nian_gan"] = ""

    all_ids, _ = core.available_ids(snapshot)

    assert "snapshot:patterns" in all_ids
    assert "snapshot:ying_qi" not in all_ids
    assert "snapshot:yong_shen" not in all_ids
    assert "snapshot:nian_gan_gong" not in all_ids
    assert "snapshot:nian_gan" not in all_ids


def test_jinhan_rejects_assertion_and_timing_refs() -> None:
    snapshot = {
        "schema_version": "qimen-snapshot-v4",
        "question": "测试",
        "method_context": {"scope": "day", "school": "jinhan_yujing"},
        "factors": {
            "kind": "jinhan",
            "scope": "day",
            "school": "jinhan_yujing",
            "palaces": [{"gong": str(index)} for index in range(9)],
            "day_spirits": [{"zhi": str(index)} for index in range(12)],
        },
    }
    built = jinhan_core.build(snapshot)
    tampered = {**built, "assertion_refs": ["assertion:fake"]}
    audit = jinhan_core.validate_core(tampered, snapshot)

    assert audit["accepted"] is False
    assert any(
        error.get("field") == "assertion_refs"
        for error in audit["errors"]
    )
