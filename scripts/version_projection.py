#!/usr/bin/env python3
"""Project an immutable Git commit into the semver required by package hosts.

WorkBuddy rejects four-part CalVer and requires three numeric identifiers.
We do not maintain a release version. The package version is derived only at
package-build time from the commit's committer instant in Asia/Shanghai:

    YYYY.MMDD.HHMMSS

All three identifiers are numeric and contain no leading zeroes. The value is
deterministic for one commit and does not create an internal release contract.
"""

from __future__ import annotations

import os
import re
import subprocess
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo

ROOT = Path(__file__).resolve().parents[1]
TIMEZONE = ZoneInfo("Asia/Shanghai")
SEMVER_RE = re.compile(r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$")


def package_version(repo_root: Path = ROOT, commit: str | None = None) -> str:
    """Return semver for an explicit commit, or the env override / HEAD commit."""
    if commit is None:
        override = os.environ.get("LIKI_PACKAGE_VERSION", "").strip()
        if override:
            if not SEMVER_RE.fullmatch(override):
                raise ValueError(f"LIKI_PACKAGE_VERSION must be semver, got {override!r}")
            return override
        raw = subprocess.check_output(
            ["git", "-C", str(repo_root), "show", "-s", "--format=%cI", "HEAD"],
            text=True,
        ).strip()
    else:
        raw = commit
    committed_at = datetime.fromisoformat(raw).astimezone(TIMEZONE)
    minor = int(f"{dt_month(committed_at)}")
    patch = int(committed_at.strftime("%H%M%S"))
    return f"{committed_at.year}.{minor}.{patch}"


def dt_month(value: datetime) -> str:
    """Return MMDD without a leading zero in the resulting numeric identifier."""
    return f"{value.month}{value.day:02d}"


def main() -> int:
    print(package_version())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
