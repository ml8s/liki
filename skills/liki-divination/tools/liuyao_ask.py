"""六爻 ask：校验 immutable snapshot 并生成结构化 answer。"""
from __future__ import annotations

import hashlib
import json

from divination_safety import assess, blocked_payload
from divination_snapshot import validate_snapshot
from liuyao_report import template as report_template
from liuyao_report import validate_report
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

    report = report_template(snapshot)
    primary_count = len(report.get("primary_evidence_refs", []))
    secondary_count = len(report.get("secondary_evidence_refs", []))
    timing_count = len(report.get("timing_refs", []))
    conflicts = snapshot.get("evidence", {}).get("conflicts", [])
    report.update({
        "headline": "六爻 snapshot 条件性解读",
        "verdict": (
            f"当前盘面提供 {primary_count} 项主要因子、{secondary_count} 项辅助因子"
            f"和 {timing_count} 个应期候选；只能表达条件性倾向，不构成结果承诺。"
        ),
        "action": "先核对主要因子对应的现实条件，再决定推进或调整节奏。",
        "confidence": "low" if conflicts else "medium",
    })
    audit = validate_report(report, snapshot)
    if not audit.get("accepted"):
        errors = "; ".join(
            str(item.get("reason") or item.get("field") or item)
            for item in audit.get("errors", [])
        )
        raise ValueError(f"liuyao answer failed validation: {errors}")

    return {
        "schema_version": ANSWER_SCHEMA_VERSION,
        "method": "liuyao",
        "snapshot_digest": snapshot["snapshot_digest"],
        "message_digest": _message_digest(message),
        "headline": report["headline"],
        "verdict": report["verdict"],
        "confidence": report["confidence"],
        "primary_evidence_refs": report["primary_evidence_refs"],
        "secondary_evidence_refs": report["secondary_evidence_refs"],
        "conflicts": report["conflicts"],
        "timing_refs": report["timing_refs"],
        "action": report["action"],
        "boundary": report["boundary"],
        "disclaimer": report["disclaimer"],
        "audit": {
            "accepted": audit["accepted"],
            "checked": audit["checked"],
            "unchecked": audit["unchecked"],
        },
    }
