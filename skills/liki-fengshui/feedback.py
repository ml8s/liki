#!/usr/bin/env python3
"""Standard autonomous feedback sender for a self-contained Liki skill.

Input: one JSON object on stdin (or via --payload-file).
Output: one JSON status object on stdout. The process never raises on network
failure; invalid payloads return submitted=false.
"""

from __future__ import annotations

import argparse
import json
import os
import pathlib
import re
import sys
import urllib.request

SCHEMA_VERSION = "feedback-v1"
DEFAULT_ENDPOINT = "https://liki.hk/api/feedback"
TIMEOUT_SECONDS = 2

SOURCE_VALUES = {"skill-agent", "user"}
ISSUE_TYPES = {"error", "gap", "conflict", "friction", "clarity"}
SEVERITIES = {"blocker", "high", "medium", "low"}
SKILL_NAMES = {"liki-bazi", "liki-divination", "liki-fengshui", "liki-naming"}
SESSION_HASH_PATTERN = r"^sha256:[a-f0-9]{64}$"
MAX_PAYLOAD_BYTES = 32 * 1024

TOP_LEVEL = {"schema_version", "meta", "agent", "llm", "problem"}
META_FIELDS = {"source", "skill", "skill_version", "engine_version", "session_hash"}
AGENT_FIELDS = {"name", "version"}
LLM_FIELDS = {"provider", "model", "model_version"}
PROBLEM_FIELDS = {"type", "severity", "tool", "summary", "expected", "observed"}


class FeedbackError(ValueError):
    pass


def _truthy(value: str | None) -> bool:
    return str(value).strip().lower() in {"1", "true", "yes", "on"}


def _load_context(payload: dict) -> dict:
    # Host context fills real gaps; explicit payload fields win. Only after
    # context has been applied do we fill unknown placeholders. Copy group
    # objects so malformed input cannot be mutated through aliases.
    payload = dict(payload)
    payload.setdefault("meta", {})

    context_path = os.environ.get("LIKI_FEEDBACK_CONTEXT")
    if context_path:
        try:
            with open(context_path, encoding="utf-8") as handle:
                context = json.load(handle)
        except (OSError, ValueError) as exc:
            raise FeedbackError(f"invalid LIKI_FEEDBACK_CONTEXT: {exc}") from exc
        if not isinstance(context, dict):
            raise FeedbackError("LIKI_FEEDBACK_CONTEXT must be a JSON object")
        for group in ("meta", "agent", "llm"):
            value = context.get(group)
            if value is None:
                continue
            if not isinstance(value, dict):
                raise FeedbackError(f"LIKI_FEEDBACK_CONTEXT.{group} must be an object")
            current = payload.get(group)
            if current is None:
                payload[group] = dict(value)
            elif not isinstance(current, dict):
                raise FeedbackError(f"{group} must be an object")
            else:
                payload[group] = {**value, **current}

    payload.setdefault("agent", {})
    payload.setdefault("llm", {})
    for key in ("name", "version"):
        payload["agent"].setdefault(key, "unknown")
    for key in ("provider", "model"):
        payload["llm"].setdefault(key, "unknown")

    skill_name = os.environ.get("LIKI_FEEDBACK_SKILL")
    if skill_name:
        payload["meta"]["skill"] = skill_name
        payload["meta"]["source"] = "skill-agent"
    skill_version = os.environ.get("LIKI_FEEDBACK_SKILL_VERSION")
    if skill_version:
        payload["meta"]["skill_version"] = skill_version
    engine_version = os.environ.get("LIKI_ENGINE_VERSION")
    if engine_version:
        payload["meta"]["engine_version"] = engine_version
    session_hash = os.environ.get("LIKI_FEEDBACK_SESSION_HASH")
    if session_hash:
        payload["meta"]["session_hash"] = session_hash
    return payload


def _required_string(group: dict, key: str, label: str, max_length: int = 120) -> str:
    value = group.get(key)
    if not isinstance(value, str) or not value.strip():
        raise FeedbackError(f"{label}.{key} is required and must be a non-empty string")
    if len(value) > max_length:
        raise FeedbackError(f"{label}.{key} exceeds {max_length} characters")
    return value


def _version_string(group: dict, key: str, label: str) -> str:
    value = _required_string(group, key, label, 32)
    if not re.fullmatch(r"\d{4}\.\d{2}\.\d{2}(?:\.\d+)?", value):
        raise FeedbackError(f"{label}.{key} must use CalVer x.y.z or x.y.z.w")
    return value


def validate(payload: dict) -> dict:
    if not isinstance(payload, dict):
        raise FeedbackError("payload must be a JSON object")
    unknown = set(payload) - TOP_LEVEL
    if unknown:
        raise FeedbackError(f"unknown payload fields: {', '.join(sorted(unknown))}")
    payload.setdefault("schema_version", SCHEMA_VERSION)
    if payload["schema_version"] != SCHEMA_VERSION:
        raise FeedbackError(f"schema_version must be {SCHEMA_VERSION}")

    meta = payload.get("meta")
    if meta is None:
        meta = {}
        payload["meta"] = meta
    if not isinstance(meta, dict):
        raise FeedbackError("meta must be an object")
    unknown = set(meta) - META_FIELDS
    if unknown:
        raise FeedbackError(f"unknown meta fields: {', '.join(sorted(unknown))}")
    meta.setdefault("source", "skill-agent")
    if meta.get("source") not in SOURCE_VALUES:
        raise FeedbackError("meta.source is invalid")
    skill = _required_string(meta, "skill", "meta", 64)
    if skill not in SKILL_NAMES:
        raise FeedbackError(f"meta.skill must be one of: {', '.join(sorted(SKILL_NAMES))}")
    _version_string(meta, "skill_version", "meta")
    _version_string(meta, "engine_version", "meta")
    session_hash = meta.get("session_hash")
    if session_hash is not None and not re.fullmatch(SESSION_HASH_PATTERN, session_hash):
        raise FeedbackError("meta.session_hash must match sha256:<64 lowercase hex characters>")

    agent = payload.get("agent")
    if agent is None:
        agent = {}
        payload["agent"] = agent
    if not isinstance(agent, dict):
        raise FeedbackError("agent must be an object")
    unknown = set(agent) - AGENT_FIELDS
    if unknown:
        raise FeedbackError(f"unknown agent fields: {', '.join(sorted(unknown))}")
    for key in ("name", "version"):
        _required_string(agent, key, "agent")

    llm = payload.get("llm")
    if llm is None:
        llm = {}
        payload["llm"] = llm
    if not isinstance(llm, dict):
        raise FeedbackError("llm must be an object")
    unknown = set(llm) - LLM_FIELDS
    if unknown:
        raise FeedbackError(f"unknown llm fields: {', '.join(sorted(unknown))}")
    for key in ("provider", "model"):
        _required_string(llm, key, "llm")
    if "model_version" in llm:
        _required_string(llm, "model_version", "llm")

    problem = payload.get("problem")
    if not isinstance(problem, dict):
        raise FeedbackError("problem is required and must be an object")
    unknown = set(problem) - PROBLEM_FIELDS
    if unknown:
        raise FeedbackError(f"unknown problem fields: {', '.join(sorted(unknown))}")
    if problem.get("type") not in ISSUE_TYPES:
        raise FeedbackError("problem.type is invalid")
    if problem.get("severity") not in SEVERITIES:
        raise FeedbackError("problem.severity is invalid")
    summary = problem.get("summary")
    if not isinstance(summary, str) or not 3 <= len(summary) <= 240:
        raise FeedbackError("problem.summary must contain 3-240 characters")
    tool = problem.get("tool")
    if tool is not None and (not isinstance(tool, str) or len(tool) > 120):
        raise FeedbackError("problem.tool must be a string of at most 120 characters")
    for key in ("expected", "observed"):
        value = problem.get(key)
        if value is None:
            continue
        if not isinstance(value, str) or len(value) > 500:
            raise FeedbackError(f"problem.{key} must be a string of at most 500 characters")
    return payload


def post(payload: dict) -> tuple[bool, str]:
    endpoint = os.environ.get("LIKI_FEEDBACK_URL", DEFAULT_ENDPOINT)
    encoded = json.dumps(payload, ensure_ascii=False, separators=(",", ":")).encode("utf-8")
    if len(encoded) > MAX_PAYLOAD_BYTES:
        return False, "payload too large"
    request = urllib.request.Request(
        endpoint,
        data=encoded,
        headers={"Content-Type": "application/json; charset=utf-8"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=TIMEOUT_SECONDS) as response:
            if 200 <= response.status < 300:
                return True, f"HTTP {response.status}"
            return False, f"HTTP {response.status}"
    except Exception as exc:  # noqa: BLE001 - telemetry must never block callers
        return False, type(exc).__name__


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--payload-file", help="read feedback JSON from this path instead of stdin")
    args = parser.parse_args(argv)

    status = {"ok": True, "submitted": False}
    try:
        if _truthy(os.environ.get("LIKI_FEEDBACK_DISABLED")):
            status["reason"] = "disabled"
        else:
            raw = pathlib.Path(args.payload_file).read_text(encoding="utf-8") if args.payload_file else sys.stdin.read()
            payload = json.loads(raw)
            context_path = pathlib.Path(__file__).with_name("feedback.context.json")
            if context_path.exists():
                os.environ.setdefault("LIKI_FEEDBACK_CONTEXT", str(context_path))
            payload = _load_context(payload)
            payload = validate(payload)
            sent, reason = post(payload)
            status["submitted"] = sent
            status["reason"] = reason
    except Exception as exc:  # noqa: BLE001 - telemetry must never block callers
        status = {"ok": True, "submitted": False, "reason": type(exc).__name__}
        if os.environ.get("LIKI_FEEDBACK_DEBUG"):
            status["detail"] = str(exc)

    print(json.dumps(status, ensure_ascii=True, separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
