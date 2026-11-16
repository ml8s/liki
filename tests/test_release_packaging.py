import re
import sys
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "scripts"))

from build_channel_package import load_market, skill_doc  # noqa: E402
from version_projection import package_version  # noqa: E402


def test_package_version_is_deterministic_host_semver() -> None:
    version = package_version(commit="2026-11-12T19:29:33+08:00")
    assert version == "2026.1112.192933"
    assert re.fullmatch(r"(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)", version)


def test_package_semver_omits_leading_zeroes() -> None:
    assert package_version(commit="2026-01-02T00:00:00+08:00") == "2026.102.0"


def test_current_repository_build_has_commit_identity() -> None:
    assert package_version()


def test_invalid_commit_is_rejected() -> None:
    with pytest.raises(ValueError):
        package_version(commit="invalid")


def test_package_version_accepts_semver_override(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LIKI_PACKAGE_VERSION", "2026.1116.200000")
    assert package_version() == "2026.1116.200000"


def test_package_version_rejects_invalid_override(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LIKI_PACKAGE_VERSION", "2026.11.16.1")
    with pytest.raises(ValueError):
        package_version()


def test_packaged_skill_projects_commit_time_semver() -> None:
    frontmatter = skill_doc(
        "free",
        "liki",
        load_market("skillhub", "free"),
        channel="skillhub",
    ).split("---\n")[1]
    assert re.search(r"(?m)^version: \d+\.\d+\.\d+$", frontmatter)
