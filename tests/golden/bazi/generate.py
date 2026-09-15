#!/usr/bin/env python3
"""Generate checked-in, three-oracle consensus fixtures for BaZi pillars.

This is a development-time tool. It intentionally has no repository runtime
dependency: lunar-python and sxtwl must be installed in the development Python
environment, and bazi-calculator is loaded through a separate Node process.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import subprocess
import sys
from pathlib import Path

from lunar_python import Solar
import sxtwl


GAN = "甲乙丙丁戊己庚辛壬癸"
ZHI = "子丑寅卯辰巳午未申酉戌亥"
TERMS = [
    "小寒", "大寒", "立春", "雨水", "惊蛰", "春分", "清明", "谷雨",
    "立夏", "小满", "芒种", "夏至", "小暑", "大暑", "立秋", "处暑",
    "白露", "秋分", "寒露", "霜降", "立冬", "小雪", "大雪", "冬至",
]
YEARS = (2024, 2025, 2026, 2027)
ORACLE_VERSIONS = {
    "lunar-python": "1.4.8",
    "sxtwl": "2.0.7",
    "bazi-calculator": "1.0.0",
}


def sxtwl_pillars(moment: dt.datetime) -> list[str]:
    day = sxtwl.fromSolar(moment.year, moment.month, moment.day)
    # False keeps 23:00 on the current day, matching liki / lunar-python.
    hour = day.getHourGZ(moment.hour, False)
    gz = lambda item: GAN[item.tg] + ZHI[item.dz]
    return [
        gz(day.getYearGZ()),
        gz(day.getMonthGZ()),
        gz(day.getDayGZ()),
        gz(hour),
    ]


def bazi_calculator_pillars(
    calculator_dir: Path, moments: list[dt.datetime]
) -> list[list[str]]:
    payload = "\n".join(
        json.dumps({
            "year": item.year,
            "month": item.month,
            "day": item.day,
            "hour": item.hour,
        })
        for item in moments
    )
    runner = Path(__file__).with_name("bazi_calculator_runner.mjs")
    process = subprocess.run(
        ["node", str(runner)],
        input=payload,
        text=True,
        cwd=calculator_dir,
        capture_output=True,
        check=False,
    )
    if process.returncode != 0:
        raise RuntimeError(process.stderr.strip() or "bazi-calculator runner failed")
    return [json.loads(line) for line in process.stdout.splitlines()]


def make_moment(term_solar: str, side: str) -> dt.datetime:
    day, clock = term_solar.split(" ")
    year, month, day_num = map(int, day.split("-"))
    hour, minute = map(int, clock.split(":")[:2])
    # One day around each exact boundary. If the anchor lands on 23:00, move one
    # more hour so the hard consensus fixture does not encode a late-zi policy.
    minutes = -25 * 60 if side == "before" else 25 * 60
    moment = dt.datetime(year, month, day_num, hour, minute) + dt.timedelta(minutes=minutes)
    if moment.hour == 23:
        moment += dt.timedelta(hours=-1 if side == "before" else 1)
    return moment


def generate() -> dict:
    inputs: list[dt.datetime] = []
    lunar_expectations: list[list[str]] = []

    for year in YEARS:
        table = Solar.fromYmd(year, 6, 1).getLunar().getJieQiTable()
        for term in TERMS:
            for side in ("before", "after"):
                moment = make_moment(table[term].toYmdHms(), side)
                eight = Solar.fromYmdHms(
                    moment.year, moment.month, moment.day, moment.hour, moment.minute, 0
                ).getLunar().getEightChar()
                inputs.append(moment)
                lunar_expectations.append([
                    eight.getYear(), eight.getMonth(), eight.getDay(), eight.getTime()
                ])

    calculator_dir = Path(os.environ.get("BAZI_CALCULATOR_DIR", "/tmp/bazi-calculator"))
    calculator_expectations = bazi_calculator_pillars(calculator_dir, inputs)

    cases = []
    seen_ids: set[str] = set()
    for index, (moment, lunar, calculator) in enumerate(
        zip(inputs, lunar_expectations, calculator_expectations)
    ):
        shoushu = sxtwl_pillars(moment)
        year = YEARS[index // (len(TERMS) * 2)]
        term_index = (index // 2) % len(TERMS)
        term = TERMS[term_index]
        side = ("before", "after")[index % 2]
        name = f"{year}-{term}-{side}"
        if name in seen_ids:
            raise AssertionError(f"duplicate case id: {name}")
        seen_ids.add(name)

        if lunar != shoushu or lunar != calculator:
            raise AssertionError(
                f"{name}: oracle mismatch\n"
                f"  lunar-python:   {lunar}\n"
                f"  sxtwl:          {shoushu}\n"
                f"  bazi-calculator: {calculator}"
            )

        cases.append({
            "name": name,
            "year": year,
            "term_index": term_index,
            "jie_qi": term,
            "side": side,
            "input": {
                "solar": moment.strftime("%Y-%m-%dT%H:%M:%S"),
                "timezone_offset": 8,
            },
            "expected": {
                "pillars": {
                    "nian": lunar[0],
                    "yue": lunar[1],
                    "ri": lunar[2],
                    "shi": lunar[3],
                },
            },
            "consensus": list(ORACLE_VERSIONS),
        })

    return {
        "version": 1,
        "policy": {
            "year_boundary": "lichun",
            "month_boundary": "jie",
            "late_zi": "same_day",
            "consensus": "all_oracles_must_agree",
        },
        "scope": {
            "years": list(YEARS),
            "terms_per_year": 24,
            "anchors_per_term": 2,
            "case_count": len(cases),
        },
        "sources": [
            {
                "id": source_id,
                "version": ORACLE_VERSIONS[source_id],
                "fields": ["pillars"],
            }
            for source_id in ORACLE_VERSIONS
        ],
        "known_limits": [
            "bazi-calculator uses date-level solar terms; anchors are ±25h from lunar-python term time",
            "late-zi cases are moved off 23:00; exact late-zi policy remains covered by lunar-typescript goldens",
            "bazi-calculator is authoritative only for its declared 2000-2099 range",
        ],
        "cases": cases,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--output",
        default="engine/internal/engine/bazi/testdata/multi_oracle_golden.json",
    )
    args = parser.parse_args()
    document = generate()
    output = Path(args.output)
    output.write_text(
        json.dumps(document, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(f"wrote {len(document['cases'])} cases to {output}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
