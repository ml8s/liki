#!/usr/bin/env python3
"""Project an immutable Git commit into host-compatible semver."""

from __future__ import annotations

import os
import re
import subprocess
from datetime import datetime
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SEMVER_RE = re.compile(r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$")
COMMIT_RE = re.compile(r"^[0-9a-f]{40}$")


def _load_commit(repo_root: Path, spec: str) -> tuple[str, datetime]:
    if spec != "HEAD" and not COMMIT_RE.fullmatch(spec):
        raise ValueError(f"commit must be HEAD or a full 40-character SHA, got {spec!r}")
    raw = subprocess.check_output(
        ["git", "-C", str(repo_root), "show", "-s", "--format=%H%x09%cI", spec],
        text=True,
    ).strip()
    sha, committed_at_text = raw.split("\t", 1)
    if not COMMIT_RE.fullmatch(sha):
        raise ValueError(f"git returned an invalid commit SHA: {sha!r}")
    return sha, datetime.fromisoformat(committed_at_text)


def package_version(repo_root: Path = ROOT, commit: str | None = None) -> str:
    """Return a stable, host-compatible semver derived from one commit."""
    override = os.environ.get("LIKI_PACKAGE_VERSION", "").strip()
    if commit is None and override:
        if not SEMVER_RE.fullmatch(override):
            raise ValueError(f"LIKI_PACKAGE_VERSION must be semver, got {override!r}")
        return override

    sha, committed_at = _load_commit(repo_root, commit or "HEAD")
    return f"0.{int(committed_at.timestamp())}.{int(sha[:12], 16)}"


def main() -> int:
    print(package_version())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
