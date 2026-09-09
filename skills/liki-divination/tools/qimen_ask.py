"""奇门 ask：校验 immutable snapshot 并生成结构化 answer。"""
from __future__ import annotations

import hashlib
import json

from divination_safety import assess, blocked_payload
from divination_snapshot import validate_snapshot
from divination_contracts import validate_document
from qimen_answer_core import build as build_core
from qimen_answer_core import validate_core
from qimen_snapshot import SCHEMA_VERSION


ANSWER_SCHEMA_VERSION = "qimen-answer-v1"


def _message_digest(message: str) -> str:
    raw = json.dumps({"message": message}, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def ask(snapshot: dict, *, message: str) -> dict:
    """基于同一 qimen snapshot 回答一次提问；不重排，不修改原 snapshot。"""
    if not isinstance(message, str) or not message.strip():
        raise ValueError("message must be non-empty text")
    message = message.strip()
    safety = assess(message)
    if safety["status"] != "allow":
        return blocked_payload(ANSWER_SCHEMA_VERSION, message, "qimen")
    validate_snapshot(snapshot, method="qimen", schema_version=SCHEMA_VERSION)

    core = build_core(snapshot)
    timing_count = len(core.get("timing_refs", []))
    core.update({
        "headline": "奇门 snapshot 条件性解读",
        "verdict": (
            f"当前局提供 {len(core.get('evidence_refs', []))} 项宫位因子"
            f"和 {timing_count} 个应期候选；只能表达态势倾向，不构成结果承诺。"
        ),
        "action": "先核查现实约束，再结合候选方向或时机推进。",
    })
    audit = validate_core(core, snapshot)
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
        "message_digest": _message_digest(message),
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
