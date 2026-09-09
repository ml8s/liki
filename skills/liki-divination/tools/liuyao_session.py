"""六爻追问会话；固化原卦与首次结论，不重排、不改写。"""
from __future__ import annotations

import hashlib
import json
import uuid
from typing import Any

from liuyao_report import validate_report


SCHEMA_VERSION = "liuyao-session-v1"
FOLLOWUP_KINDS = {"clarification", "strategy", "reality_feedback", "term_explanation"}


def _digest(value: Any) -> str:
    raw = json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(raw.encode("utf-8")).hexdigest()


def _casting_from_snapshot(snapshot: dict) -> dict:
    casting = snapshot.get("casting")
    if not isinstance(casting, dict):
        raise ValueError("snapshot lacks casting")
    yaos = casting.get("yaos")
    dong_yao = casting.get("dong_yao")
    if not isinstance(yaos, list) or len(yaos) != 6:
        raise ValueError("casting.yaos must contain exactly 6 values")
    if not isinstance(dong_yao, list):
        raise ValueError("casting.dong_yao must be an array")
    return {
        "fingerprint": casting.get("fingerprint"),
        "yaos": yaos,
        "dong_yao": dong_yao,
    }


def create_session(
    *,
    question: str,
    snapshot: dict,
    headline: str,
    verdict: str,
    confidence: str = "medium",
    report: dict | None = None,
) -> dict:
    if not isinstance(question, str) or len(question.strip()) < 2:
        raise ValueError("question must be non-empty")
    if not isinstance(snapshot, dict) or snapshot.get("schema_version") != "liuyao-snapshot-v2":
        raise ValueError("snapshot must be liuyao-snapshot-v2")
    if not headline.strip() or not verdict.strip():
        raise ValueError("headline and verdict are required")
    if confidence not in {"low", "medium", "high"}:
        raise ValueError("confidence must be low/medium/high")
    if report is not None:
        report_audit = validate_report(report, snapshot)
        if not report_audit.get("accepted"):
            errors = "; ".join(
                str(item.get("reason") or item.get("field") or item)
                for item in report_audit.get("errors", [])
            )
            raise ValueError(f"report must pass validation before session create: {errors}")
        if report.get("headline") != headline.strip() or report.get("verdict") != verdict.strip():
            raise ValueError("headline/verdict must match the structured report")
    casting = _casting_from_snapshot(snapshot)
    first_verdict = {
        "headline": headline.strip(),
        "verdict": verdict.strip(),
        "confidence": confidence,
        "report": report,
    }
    return {
        "$schema": "liki:liuyao-session-v1",
        "schema_version": SCHEMA_VERSION,
        "session_id": uuid.uuid4().hex,
        "question": question.strip(),
        "casting": casting,
        "snapshot_digest": _digest(snapshot),
        "snapshot": snapshot,
        "first_verdict": first_verdict,
        "integrity": {
            "casting": _digest(casting),
            "snapshot": _digest(snapshot),
            "first_verdict": _digest(first_verdict),
            "report": _digest(report) if report is not None else None,
        },
        "followups": [],
        "policy": {
            "immutable_casting": True,
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
    item = {
        "kind": kind,
        "text": text.strip(),
        "answer": answer.strip(),
    }
    if kind == "reality_feedback":
        item["outcome"] = outcome
    session["followups"].append(item)
    return session


def validate_session(session: dict) -> dict:
    if not isinstance(session, dict):
        raise ValueError("session must be an object")
    if session.get("schema_version") != SCHEMA_VERSION:
        raise ValueError("session schema_version must be liuyao-session-v1")
    for key in ("session_id", "question", "casting", "snapshot_digest", "snapshot", "first_verdict", "followups"):
        if key not in session:
            raise ValueError(f"session lacks {key}")
    if _casting_from_snapshot(session.get("snapshot", {})) != session["casting"]:
        raise ValueError("casting has changed; follow-up must use the original casting")
    if _digest(session.get("snapshot", {})) != session["snapshot_digest"]:
        raise ValueError("snapshot has changed; follow-up must use the original snapshot")
    integrity = session.get("integrity", {})
    expected_integrity = {
        "casting": _digest(session["casting"]),
        "snapshot": _digest(session.get("snapshot", {})),
        "first_verdict": _digest(session.get("first_verdict", {})),
        "report": _digest(session["first_verdict"].get("report")) if session.get("first_verdict", {}).get("report") is not None else None,
    }
    if integrity != expected_integrity:
        raise ValueError("session integrity mismatch; casting, first verdict or report has changed")
    policy = session.get("policy", {})
    if not all([
        policy.get("immutable_casting") is True,
        policy.get("immutable_first_verdict") is True,
        policy.get("no_recast_without_new_event") is True,
    ]):
        raise ValueError("session policy must lock original casting and verdict")
    return session
