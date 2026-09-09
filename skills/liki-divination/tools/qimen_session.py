"""奇门追问会话；固化原局时间、方法、snapshot 和首次结论。"""
from __future__ import annotations

import hashlib
import json
import uuid

from qimen_report import validate_report


SCHEMA_VERSION = "qimen-session-v1"
FOLLOWUP_KINDS = {"clarification", "strategy", "reality_feedback", "term_explanation"}


def _digest(value) -> str:
    raw = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _read_core(read_result: dict) -> dict:
    if read_result.get("schema_version") != "qimen-read-v1":
        raise ValueError("read_result must be qimen-read-v1")
    for key in ("question", "input", "snapshot"):
        if key not in read_result:
            raise ValueError(f"read_result lacks {key}")
    return {
        "question": read_result["question"],
        "input": read_result["input"],
        "snapshot": read_result["snapshot"],
    }


def create_session(
    *,
    read_result: dict,
    headline: str,
    verdict: str,
    confidence: str = "medium",
    report: dict | None = None,
) -> dict:
    if not headline.strip() or not verdict.strip():
        raise ValueError("headline and verdict are required")
    if confidence not in {"low", "medium", "high"}:
        raise ValueError("confidence must be low/medium/high")
    core = _read_core(read_result)
    if report is not None:
        audit = validate_report(report, read_result)
        if not audit.get("accepted"):
            errors = "; ".join(
                str(item.get("reason") or item.get("field"))
                for item in audit.get("errors", [])
            )
            raise ValueError(f"report must pass validation before session create: {errors}")
        if report.get("headline") != headline.strip() or report.get("verdict") != verdict.strip():
            raise ValueError("headline/verdict must match the structured report")
    snapshot = core["snapshot"]
    method = snapshot.get("method", {})
    first_verdict = {
        "headline": headline.strip(),
        "verdict": verdict.strip(),
        "confidence": confidence,
        "report": report,
    }
    return {
        "$schema": "liki:qimen-session-v1",
        "schema_version": SCHEMA_VERSION,
        "session_id": uuid.uuid4().hex,
        "question": core["question"],
        "input": core["input"],
        "method": {"scope": method.get("scope"), "school": method.get("school")},
        "snapshot_digest": _digest(snapshot),
        "snapshot": snapshot,
        "integrity": {
            "input": _digest(core["input"]),
            "method": _digest(method),
            "snapshot": _digest(snapshot),
            "first_verdict": _digest(first_verdict),
            "report": _digest(report) if report is not None else None,
        },
        "first_verdict": first_verdict,
        "followups": [],
        "policy": {
            "immutable_input": True,
            "immutable_method": True,
            "immutable_first_verdict": True,
            "no_recast_without_new_event": True,
        },
    }


def append_followup(session: dict, *, kind: str, text: str, answer: str = "", outcome: str | None = None) -> dict:
    validate_session(session)
    if kind not in FOLLOWUP_KINDS:
        raise ValueError(f"kind must be one of: {', '.join(sorted(FOLLOWUP_KINDS))}")
    if not text.strip():
        raise ValueError("followup text is required")
    if kind == "reality_feedback" and outcome not in {"supported", "contradicted", "unresolved"}:
        raise ValueError("reality_feedback requires outcome supported/contradicted/unresolved")
    item = {"kind": kind, "text": text.strip(), "answer": answer.strip()}
    if kind == "reality_feedback":
        item["outcome"] = outcome
    session["followups"].append(item)
    return session


def validate_session(session: dict) -> dict:
    if session.get("schema_version") != SCHEMA_VERSION:
        raise ValueError("session schema_version must be qimen-session-v1")
    for key in ("session_id", "question", "input", "method", "snapshot_digest", "snapshot", "first_verdict", "followups"):
        if key not in session:
            raise ValueError(f"session lacks {key}")
    if _digest(session.get("snapshot", {})) != session.get("snapshot_digest"):
        raise ValueError("snapshot has changed; follow-up must use the original chart")
    integrity = session.get("integrity", {})
    first_verdict = session.get("first_verdict", {})
    expected_integrity = {
        "input": _digest(session.get("input", {})),
        "method": _digest(session.get("method", {})),
        "snapshot": _digest(session.get("snapshot", {})),
        "first_verdict": _digest(first_verdict),
        "report": _digest(first_verdict.get("report")) if first_verdict.get("report") is not None else None,
    }
    if integrity != expected_integrity:
        raise ValueError("session integrity mismatch; input, method, first verdict or report has changed")
    policy = session.get("policy", {})
    if not all(policy.get(key) is True for key in (
        "immutable_input", "immutable_method", "immutable_first_verdict", "no_recast_without_new_event"
    )):
        raise ValueError("session policy must lock original chart and verdict")
    return session
