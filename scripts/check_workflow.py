#!/usr/bin/env python3
"""Validate GitHub workflow structure: YAML parses, needs references resolve.

A dangling `needs:` entry makes the whole workflow fail to start (zero jobs
reported), which is silent locally. Run as part of `make check`.
"""
from __future__ import annotations

import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent


def main() -> int:
    errors: list[str] = []
    for path in sorted(ROOT.glob(".github/workflows/*.yml")) + sorted(
        ROOT.glob(".github/workflows/*.yaml")
    ):
        try:
            document = yaml.safe_load(path.read_text(encoding="utf-8"))
        except yaml.YAMLError as error:
            errors.append(f"{path.relative_to(ROOT)}: invalid YAML: {error}")
            continue
        if not isinstance(document, dict):
            continue
        jobs = document.get("jobs") or {}
        if not isinstance(jobs, dict):
            errors.append(f"{path.relative_to(ROOT)}: jobs must be a mapping")
            continue
        for job_id, job in jobs.items():
            if not isinstance(job, dict):
                continue
            needs = job.get("needs")
            if needs is None:
                continue
            for ref in needs if isinstance(needs, list) else [needs]:
                if ref not in jobs:
                    errors.append(
                        f"{path.relative_to(ROOT)}: job {job_id!r} needs "
                        f"unknown job {ref!r}"
                    )
    if errors:
        for error in errors:
            print(f"❌ {error}", file=sys.stderr)
        return 1
    print("✓ workflow structure")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
