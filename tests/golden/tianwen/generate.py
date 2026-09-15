#!/usr/bin/env python3
"""Generate two-oracle calendar consensus fixtures for the tianwen domain."""

from __future__ import annotations

import datetime as dt
import json
import os
import sys
from pathlib import Path

from lunar_python import Solar
import sxtwl


YEARS = (2024, 2025, 2026, 2027)
SOURCES = {
    "lunar-python": "1.4.8",
    "sxtwl": "2.0.7",
}


def lunar_python(date: dt.date) -> tuple[tuple[int, int, int, bool], str]:
    lunar = Solar.fromYmd(date.year, date.month, date.day).getLunar()
    month = lunar.getMonth()
    leap = month < 0
    return (lunar.getYear(), abs(month), lunar.getDay(), leap), lunar.getMonthZhiExact()


def sxtwl_lunar(date: dt.date) -> tuple[tuple[int, int, int, bool], str]:
    day = sxtwl.fromSolar(date.year, date.month, date.day)
    zhi = "子丑寅卯辰巳午未申酉戌亥"[day.getMonthGZ().dz]
    return (
        (day.getLunarYear(), day.getLunarMonth(), day.getLunarDay(), day.isLunarLeap()),
        zhi,
    )


def generate() -> dict:
    daily: dict[dt.date, tuple[tuple[int, int, int, bool], str | None]] = {}
    lunar_keys: dict[dt.date, tuple[int, int, bool]] = {}

    date = dt.date(YEARS[0], 1, 1)
    end = dt.date(YEARS[-1], 12, 31)
    while date <= end:
        lunar_python_value, month_zhi = lunar_python(date)
        sxtwl_value, sxtwl_month_zhi = sxtwl_lunar(date)
        if lunar_python_value != sxtwl_value:
            raise AssertionError(
                f"{date}: oracle mismatch lunar={lunar_python_value}/{month_zhi} "
                f"sxtwl={sxtwl_value}/{sxtwl_month_zhi}"
            )
        daily[date] = lunar_python_value, month_zhi if month_zhi == sxtwl_month_zhi else None
        lunar_keys[date] = (
            lunar_python_value[0],
            lunar_python_value[1],
            lunar_python_value[3],
        )
        date += dt.timedelta(days=1)

    selected: dict[dt.date, str] = {}
    ordered = sorted(daily)
    for index, current in enumerate(ordered):
        current_key = lunar_keys[current]
        previous = current - dt.timedelta(days=1) if index else None
        next_date = current + dt.timedelta(days=1) if index + 1 < len(ordered) else None
        previous_key = lunar_keys.get(previous)
        next_key = lunar_keys.get(next_date)
        if previous_key is None or previous_key != current_key:
            selected[current] = "month_start"
        if next_key is None or next_key != current_key:
            selected[current] = "month_end"

    cases: list[dict] = []
    for date, kind in sorted(selected.items()):
        (year, month, day, leap), month_zhi = daily[date]
        if month_zhi is None:
            continue
        cases.append({
            "name": f"{date.isoformat()}-{kind}",
            "kind": kind,
            "input": {"gregorian": date.isoformat()},
            "expected": {
                "lunar": {"year": year, "month": month, "day": day, "leap": leap},
                "jian_yue": month_zhi,
            },
            "consensus": list(SOURCES),
        })

    return {
        "version": 1,
        "policy": {
            "lunar_day_boundary": "local_midnight_utc8",
            "jian_yue": "solar_term",
            "consensus": "all_oracles_must_agree",
        },
        "scope": {
            "years": list(YEARS),
            "case_count": len(cases),
        },
        "sources": [
            {"id": source_id, "version": version, "fields": ["lunar", "jian_yue"]}
            for source_id, version in SOURCES.items()
        ],
        "known_limits": [
            "dates with date-level solar-term disagreement between oracles are not hard-consensus candidates",
            "true solar time is tested by the existing tianwen solar goldens",
            "day pillars are covered by bazi calendar goldens and domain oracle tables",
        ],
        "cases": cases,
    }


def main() -> int:
    output = Path(
        os.environ.get(
            "OUTPUT",
            "engine/internal/engine/tianwen/testdata/multi_oracle_golden.json",
        )
    )
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(generate(), ensure_ascii=False, indent=2), encoding="utf-8")
    return 0


if __name__ == "__main__":
    sys.exit(main())
