#!/usr/bin/env python3
"""Bump the single release version source; all other versions are generated."""

from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo

ROOT = Path(__file__).resolve().parents[1]
VERSION_FILE = ROOT / "skills/liki/VERSION.txt"


def next_version() -> str:
    today = datetime.now(ZoneInfo("Asia/Shanghai")).strftime("%Y.%m.%d")
    current = VERSION_FILE.read_text(encoding="utf-8").strip()
    serial = -1 if current.startswith(today + ".") else 0
    if serial == -1:
        serial = int(current.rsplit(".", 1)[1])
    return f"{today}.{serial + 1}"


def main() -> None:
    version = next_version()
    VERSION_FILE.write_text(version + "\n", encoding="utf-8")
    print(f"release version → {version}")


if __name__ == "__main__":
    main()
