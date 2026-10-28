"""Test-only conversions from reported instants to canonical civil/standard time."""

from datetime import datetime, timedelta, timezone
from zoneinfo import ZoneInfo


def _parse(value: str) -> datetime:
    return datetime.fromisoformat(value.replace("Z", "+00:00"))


def civil_time_from_instant(value: str, timezone_name: str) -> str:
    local = _parse(value).astimezone(ZoneInfo(timezone_name))
    return local.strftime("%Y-%m-%dT%H:%M:%S")


def standard_time_from_instant(value: str, timezone_name: str) -> str:
    instant = _parse(value).astimezone(timezone.utc)
    zone = ZoneInfo(timezone_name)
    local = instant.astimezone(zone)
    offset = local.utcoffset()
    if local.dst() not in (None, timedelta(0)):
        for days in (-180, 180, -90, 90):
            probe = (instant + timedelta(days=days)).astimezone(zone)
            if probe.dst() == timedelta(0):
                offset = probe.utcoffset()
                break
    return instant.astimezone(timezone(offset)).isoformat(timespec="seconds")
