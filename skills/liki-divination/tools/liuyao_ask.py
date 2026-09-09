"""六爻 ask：校验 immutable snapshot 并生成结构化 answer。"""
from __future__ import annotations

import hashlib
import json

from divination_safety import assess, blocked_payload
from divination_snapshot import validate_snapshot
from divination_contracts import validate_document
from liuyao_answer_core import build as build_core
from liuyao_answer_core import validate_core
from liuyao_snapshot import SCHEMA_VERSION


ANSWER_SCHEMA_VERSION = "liuyao-answer-v1"


def _message_digest(message: str) -> str:
    raw = json.dumps({"message": message}, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def ask(snapshot: dict, *, message: str) -> dict:
    """基于同一六爻 snapshot 回答一次提问；不重排，不修改原 snapshot。"""
    if not isinstance(message, str) or not message.strip():
        raise ValueError("message must be non-empty text")
    message = message.strip()
    safety = assess(message)
    if safety["status"] != "allow":
        return blocked_payload(ANSWER_SCHEMA_VERSION, message, "liuyao")
    validate_snapshot(snapshot, method="liuyao", schema_version=SCHEMA_VERSION)

    core = build_core(snapshot)
    primary_count = len(core.get("primary_evidence_refs", []))
    secondary_count = len(core.get("secondary_evidence_refs", []))
    timing_count = len(core.get("timing_refs", []))
    conflicts = snapshot.get("evidence", {}).get("conflicts", [])
    core.update({
        "headline": "六爻 snapshot 条件性解读",
        "verdict": (
            f"当前盘面提供 {primary_count} 项主要因子、{secondary_count} 项辅助因子"
            f"和 {timing_count} 个应期候选；只能表达条件性倾向，不构成结果承诺。"
        ),
        "action": "先核对主要因子对应的现实条件，再决定推进或调整节奏。",
        "confidence": "low" if conflicts else "medium",
    })
    audit = validate_core(core, snapshot)
    if not audit.get("accepted"):
        errors = "; ".join(
            str(item.get("reason") or item.get("field") or item)
            for item in audit.get("errors", [])
        )
        raise ValueError(f"liuyao answer failed validation: {errors}")

    answer = {
        "schema_version": ANSWER_SCHEMA_VERSION,
        "method": "liuyao",
        "snapshot_digest": snapshot["snapshot_digest"],
        "message_digest": _message_digest(message),
        "headline": core["headline"],
        "verdict": core["verdict"],
        "confidence": core["confidence"],
        "primary_evidence_refs": core["primary_evidence_refs"],
        "secondary_evidence_refs": core["secondary_evidence_refs"],
        "conflicts": core["conflicts"],
        "timing_refs": core["timing_refs"],
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
    validate_document("liuyao_answer", answer)
    return answer
