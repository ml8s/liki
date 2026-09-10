"""奇门 ask：校验 immutable snapshot 并生成结构化 answer。"""
from __future__ import annotations

from divination_contracts import validate_document
from divination_hashing import message_digest
from divination_safety import assess, blocked_payload
from divination_snapshot import validate_snapshot
from qimen_answer_core import build as build_standard_core, validate_core as validate_standard_core
from qimen_projection import validate as validate_standard_factors
from qimen_jinhan_answer_core import build as build_jinhan_core, validate_core as validate_jinhan_core
from qimen_jinhan_factors import validate as validate_jinhan_factors
from qimen_snapshot import SCHEMA_VERSION


ANSWER_SCHEMA_VERSION = "qimen-answer-v1"


def ask(snapshot: dict, *, message: str) -> dict:
    """基于同一 qimen snapshot 回答一次提问；不重排，不修改原 snapshot。"""
    if not isinstance(message, str):
        raise ValueError("message must be at least 2 characters")
    message = message.strip()
    if len(message) < 2:
        raise ValueError("message must be at least 2 characters")
    safety = assess(message)
    if safety["status"] != "allow":
        return blocked_payload(question=message, method="qimen", safety=safety)

    validate_snapshot(snapshot, method="qimen", schema_version=SCHEMA_VERSION)
    validate_document("qimen_snapshot", snapshot)

    factors = snapshot["factors"]
    if factors.get("school") == "jinhan_yujing":
        validate_jinhan_factors(factors)
        core = build_jinhan_core(snapshot)
        audit = validate_jinhan_core(core, snapshot)
    else:
        validate_standard_factors(factors)
        core = build_standard_core(snapshot)
        audit = validate_standard_core(core, snapshot)

    timing_count = len(core.get("timing_refs", []))
    core.update({
        "headline": "奇门 snapshot 条件性解读",
        "verdict": (
            f"当前局提供 {len(core.get('evidence_refs', []))} 项宫位因子"
            f"和 {timing_count} 个应期候选；只能表达态势倾向，不构成结果承诺。"
        ),
        "action": "先核查现实约束，再结合候选方向或时机推进。",
    })
    audit = (
        validate_jinhan_core(core, snapshot)
        if factors.get("school") == "jinhan_yujing"
        else validate_standard_core(core, snapshot)
    )
    if not audit.get("accepted"):
        errors = "; ".join(
            str(item.get("reason") or item.get("field") or item)
            for item in audit.get("errors", [])
        )
        raise ValueError(f"qimen answer failed validation: {errors}")

    answer = {
        "schema_version": ANSWER_SCHEMA_VERSION,
        "method": "qimen",
        "snapshot_digest": snapshot["snapshot_digest"],
        "message_digest": message_digest(message),
        "headline": core["headline"],
        "verdict": core["verdict"],
        "confidence": core["confidence"],
        "evidence_refs": core["evidence_refs"],
        "assertion_refs": core["assertion_refs"],
        "timing_refs": core["timing_refs"],
        "conflicts": core["conflicts"],
        "action": core["action"],
        "boundary": core["boundary"],
        "disclaimer": core["disclaimer"],
        "audit": {
            "accepted": audit["accepted"],
            "errors": audit["errors"],
            "checked": audit["checked"],
            "unchecked": audit["unchecked"],
        },
    }
    validate_document("qimen_answer", answer)
    return answer
