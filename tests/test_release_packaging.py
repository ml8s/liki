"""Packaging version projection contracts."""

import re
import sys
from datetime import datetime

import pytest

from _helpers import ROOT

sys.path.insert(0, str(ROOT / "scripts"))

from build_channel_package import load_market, skill_doc  # noqa: E402
from version_projection import package_version  # noqa: E402

COMMIT = "0123456789abcdef0123456789abcdef01234567"
COMMITTED_AT = "2026-11-12T11:29:33+00:00"


def _patch_git(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setattr(
        "version_projection.subprocess.check_output",
        lambda *_args, **_kwargs: f"{COMMIT}\t{COMMITTED_AT}",
    )


def test_commit_projects_to_deterministic_semver(monkeypatch: pytest.MonkeyPatch) -> None:
    _patch_git(monkeypatch)
    timestamp = int(datetime.fromisoformat(COMMITTED_AT).timestamp())
    version = package_version(commit=COMMIT)
    assert version == f"0.{timestamp}.{int(COMMIT[:12], 16)}"


def test_repository_and_package_versions_are_semver() -> None:
    assert re.fullmatch(r"0\.\d+\.\d+", package_version())
    frontmatter = skill_doc(
        "free", "liki", load_market("skillhub", "free"), channel="skillhub"
    ).split("---\n")[1]
    assert re.search(r"(?m)^version: 0\.\d+\.\d+$", frontmatter)


def test_semver_override_is_accepted_but_calver_is_rejected(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LIKI_PACKAGE_VERSION", "0.1.0")
    assert package_version() == "0.1.0"
    monkeypatch.setenv("LIKI_PACKAGE_VERSION", "2026.11.16.1")
    with pytest.raises(ValueError):
        package_version()
