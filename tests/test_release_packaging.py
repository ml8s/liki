import re
import sys
from datetime import datetime
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[1]
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



def test_package_version_is_deterministic_commit_semver(monkeypatch: pytest.MonkeyPatch) -> None:
    _patch_git(monkeypatch)
    timestamp = int(datetime.fromisoformat(COMMITTED_AT).timestamp())
    identity = int(COMMIT[:12], 16)
    version = package_version(commit=COMMIT)
    assert version == f"0.{timestamp}.{identity}"
    assert re.fullmatch(r"0\.\d+\.\d+", version)
    assert package_version(commit=COMMIT) == version


def test_current_repository_build_has_commit_identity() -> None:
    version = package_version()
    assert re.fullmatch(r"0\.\d+\.\d+", version)


def test_invalid_commit_is_rejected() -> None:
    with pytest.raises(ValueError):
        package_version(commit="invalid")


def test_package_version_accepts_semver_override(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LIKI_PACKAGE_VERSION", "0.1.0")
    assert package_version() == "0.1.0"


def test_package_version_rejects_invalid_override(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LIKI_PACKAGE_VERSION", "2026.11.16.1")
    with pytest.raises(ValueError):
        package_version()


def test_packaged_skill_projects_commit_semver() -> None:
    frontmatter = skill_doc(
        "free",
        "liki",
        load_market("skillhub", "free"),
        channel="skillhub",
    ).split("---\n")[1]
    assert re.search(r"(?m)^version: 0\.\d+\.\d+$", frontmatter)
