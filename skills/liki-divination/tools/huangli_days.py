"""黄历择日编排：安全分流、日期范围与 engine 事件适配结果投影。"""
from __future__ import annotations

from datetime import date, timedelta

from divination_contracts import validate_document
from divination_rpc import engine_items, server_time
from divination_safety import assess, blocked_payload


SCHEMA_VERSION = "huangli-days-v1"
MAX_DAYS = 30

def _server_date() -> date:
    return date.fromisoformat(server_time()[:10])


def _date(value: str, field: str) -> date:
    try:
        return date.fromisoformat(value)
    except (TypeError, ValueError) as error:
        raise ValueError(f"{field} must be YYYY-MM-DD") from error


def _normalize_event(event: str | None) -> str | None:
    if event is None:
        return None
    if event not in {
        "wedding", "engage", "opening", "sign", "move",
        "travel", "build", "exam", "medical", "sacrifice", "cleaning",
        "renovation", "bed_install", "income", "funeral",
    }:
        raise ValueError(f"unknown event: {event}")
    return event


def _project_day(day: dict) -> dict:
    """Project engine event classification into the stable tool contract."""
    warnings = [day.get(field) for field in ("gan_ji", "zhi_ji") if day.get(field)]
    return {
        "date": day.get("date"),
        "jian_chu": day.get("jian_chu"),
        "suitability": day.get("suitability") or "unspecified",
        "reason": day.get("reason") or "未指定事项，仅列黄历事实。",
        "warnings": warnings,
        "day": day,
    }


def days(
    *,
    question: str,
    event: str | None = None,
    start_date: str | None = None,
    end_date: str | None = None,
    days: int | None = None,
) -> dict:
    """查询黄历日期范围并投影 engine 的建除事项适配结果。"""
    if not isinstance(question, str) or not question.strip():
        raise ValueError("question must be non-empty text")
    question = question.strip()
    safety = assess(question)
    if safety["status"] != "allow":
        return blocked_payload(question=question, method="huangli", safety=safety)
    if len(question) < 2:
        raise ValueError("question must be at least 2 characters")

    event_key = _normalize_event(event)
    if end_date is not None and days is not None:
        raise ValueError("provide only one of end_date or days")
    if days is not None:
        if isinstance(days, bool) or not isinstance(days, int):
            raise ValueError("days must be an integer")
        if not 1 <= days <= MAX_DAYS:
            raise ValueError(f"days must contain 1-{MAX_DAYS} days")
    if start_date is not None:
        start = _date(start_date, "start_date")
    else:
        today = _server_date()
        start = today + timedelta(days=1)
    if end_date is not None:
        end = _date(end_date, "end_date")
        if end < start:
            raise ValueError("end_date must not be before start_date")
        count = (end - start).days + 1
    else:
        count = days if days is not None else 30
    if isinstance(count, bool) or not isinstance(count, int) or not 1 <= count <= MAX_DAYS:
        raise ValueError(f"date range must contain 1-{MAX_DAYS} days")

    raw_days = engine_items("huangli.days", {"start_date": start.isoformat(), "count": count})
    if any(not isinstance(day, dict) for day in raw_days):
        raise ValueError("huangli.days returned a non-object day")

    candidates = [_project_day(day) for day in raw_days]
    recommended = [item for item in candidates if item["suitability"] == "recommended"]
    unsuitable = [item for item in candidates if item["suitability"] == "unsuitable"]
    result = {
        "schema_version": SCHEMA_VERSION,
        "question": question,
        "event": event_key,
        "event_label": (candidates[0]["day"].get("event_label") if candidates and event_key else None),
        "range": {
            "start_date": start.isoformat(),
            "end_date": (start + timedelta(days=count - 1)).isoformat(),
            "days": count,
        },
        "safety": safety,
        "recommended": recommended,
        "unsuitable": unsuitable,
        "candidates": candidates,
        "policy": {
            "candidate_scope": "traditional_huangli_labels_only",
            "no_bazi_composite_without_birth_data": True,
            "no_guarantee": True,
        },
    }
    validate_document("huangli_days", result)
    return result
