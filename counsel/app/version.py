"""Counsel runtime version, sourced from the single liki VERSION.txt.

Resolution order:
1. ``LIKI_VERSION`` — injected by the image build.
2. ``app/VERSION.txt`` — legacy layout, if a deployment baked one.
3. ``skills/liki/VERSION.txt`` — repository layout used by tests.
"""

from __future__ import annotations

import os
from pathlib import Path

_HERE = Path(__file__).resolve()
_CANDIDATES = (
    _HERE.parent / "VERSION.txt",
    _HERE.parents[2] / "skills" / "liki" / "VERSION.txt",
)


def runtime_version() -> str:
    value = os.environ.get("LIKI_VERSION", "").strip()
    if value:
        return value
    for path in _CANDIDATES:
        if path.is_file():
            value = path.read_text(encoding="utf-8").strip()
            if value:
                return value
    raise RuntimeError("liki VERSION.txt is unavailable")
