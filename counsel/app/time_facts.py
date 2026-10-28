"""Shared civil/standard time helpers for counsel.

Civil time is the wall clock a user reports; standard time is the same zone
without DST. Engine calls always receive civil_time (no offset); engine returns
standard_time/civil_time and never instant.
"""

from datetime import datetime
from zoneinfo import ZoneInfo


def parse_instant(value: str) -> datetime:
    return datetime.fromisoformat(value.replace("Z", "+00:00"))


def civil_time_from_instant(value: str, timezone: str) -> str:
    local = parse_instant(value).astimezone(ZoneInfo(timezone))
    return local.strftime("%Y-%m-%dT%H:%M:%S")


def civil_wall(value: str) -> str:
    return parse_instant(value).strftime("%Y-%m-%dT%H:%M:%S")


def split_civil(value: str) -> tuple[str, str]:
    civil = civil_wall(value)
    date, clock = civil.split("T", 1)
    return date, clock
