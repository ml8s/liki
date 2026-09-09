"""奇门 ask：校验 immutable snapshot 并生成结构化 answer。"""
from __future__ import annotations

import hashlib
import json

from divination_safety import assess, blocked_payload
from divination_snapshot import validate_snapshot
from qimen_report import template as report_template
from qimen_report import validate_report
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

    report = report_template(snapshot)
    timing_count = len(report.get("timing_refs", []))
    report.update({
        "headline": "奇门 snapshot 条件性解读",
        "verdict": (
            f"当前局提供 {len(report.get('evidence_refs', []))} 项宫位因子"
            f"和 {timing_count} 个应期候选；只能表达态势倾向，不构成结果承诺。"
        ),
        "action": "先核查现实约束，再结合候选方向或时机推进。",
    })
    audit = validate_report(report, snapshot)
    if not audit.get("accepted"):
        errors = "; ".join(
            str(item.get("reason") or item.get("field") or item)
            for item in audit.get("errors", [])
        )
        raise ValueError(f"qimen answer failed validation: {errors}")

    return {
        "schema_version": ANSWER_SCHEMA_VERSION,
        "method": "qimen",
        "snapshot_digest": snapshot["snapshot_digest"],
        "message_digest": _message_digest(message),
        "headline": report["headline"],
        "verdict": report["verdict"],
        "confidence": report["confidence"],
        "evidence_refs": report["evidence_refs"],
        "assertion_refs": report["assertion_refs"],
        "timing_refs": report["timing_refs"],
        "conflicts": report["conflicts"],
        "action": report["action"],
        "boundary": report["boundary"],
        "disclaimer": report["disclaimer"],
        "audit": {
            "accepted": audit["accepted"],
            "checked": audit["checked"],
            "unchecked": audit["unchecked"],
        },
    }
