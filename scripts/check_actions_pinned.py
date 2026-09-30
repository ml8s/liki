#!/usr/bin/env python3
"""Fail CI when a reusable GitHub Action is referenced by a mutable ref."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
USES_RE = re.compile(r"^\s*(?:-\s+)?uses:\s*(\S+)")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")


def main() -> int:
    errors: list[str] = []
    patterns = [
        ROOT / ".github" / "workflows",
        ROOT / ".github" / "actions",
    ]
    files: list[Path] = []
    for pattern in patterns:
        if pattern.is_dir():
            files.extend(pattern.rglob("*.yml"))
            files.extend(pattern.rglob("*.yaml"))
    for path in sorted(set(files)):
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            match = USES_RE.match(line)
            if not match:
                continue
            reference = match.group(1)
            if reference.startswith("./") or reference.startswith("docker://"):
                continue
            ref = reference.rsplit("@", 1)[-1]
            if not SHA_RE.fullmatch(ref):
                errors.append(f"{path.relative_to(ROOT)}:{number}: mutable action reference: {reference}")
    if errors:
        print("Unpinned GitHub Actions found:", file=sys.stderr)
        print("\n".join(f"  {error}" for error in errors), file=sys.stderr)
        return 1
    print("GitHub Actions are SHA-pinned")
    return 0


if __name__ == "__main__":
    sys.exit(main())
