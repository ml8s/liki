#!/usr/bin/env python3
"""Project an immutable Git commit into host-compatible semver.

Liki has no release version and no calendar version. The only internal identity
is the Git commit. Some package hosts require three numeric semver identifiers,
so packaging projects one immutable commit into:

    0.<unix-commit-time>.<first-12-hex-digits-as-decimal>

The value is deterministic for a commit, unique across ordinary history, and is
an external packaging compatibility layer only.
"""

from __future__ import annotations

import os
import re
import subprocess
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SEMVER_RE = re.compile(r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$")
COMMIT_RE = re.compile(r"^[0-9a-f]{40}$")


@dataclass(frozen=True, slots=True)
class CommitIdentity:
    """The two immutable facts used by package version projection."""

    sha: str
    committed_at: datetime

    @property
    def numeric_identity(self) -> int:
        """Encode the leading commit hash as a numeric semver identifier."""
        return int(self.sha[:12], 16)


def _load_commit_identity(repo_root: Path, spec: str) -> CommitIdentity:
    if spec != "HEAD" and not COMMIT_RE.fullmatch(spec):
        raise ValueError(f"commit must be HEAD or a full 40-character SHA, got {spec!r}")
    raw = subprocess.check_output(
        ["git", "-C", str(repo_root), "show", "-s", "--format=%H%x09%cI", spec],
        text=True,
    ).strip()
    sha, committed_at_text = raw.split("\t", 1)
    if not COMMIT_RE.fullmatch(sha):
        raise ValueError(f"git returned an invalid commit SHA: {sha!r}")
    committed_at = datetime.fromisoformat(committed_at_text)
    return CommitIdentity(sha=sha, committed_at=committed_at)


def package_version(repo_root: Path = ROOT, commit: str | None = None) -> str:
    """Return host semver for an explicit commit, env override, or HEAD commit."""
    override = os.environ.get("LIKI_PACKAGE_VERSION", "").strip()
    if commit is None and override:
        if not SEMVER_RE.fullmatch(override):
            raise ValueError(f"LIKI_PACKAGE_VERSION must be semver, got {override!r}")
        return override
    identity = _load_commit_identity(repo_root, commit or "HEAD")
    return f"0.{int(identity.committed_at.timestamp())}.{identity.numeric_identity}"


def main() -> int:
    print(package_version())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
