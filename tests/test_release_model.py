"""Release version model contract: CalVer runtime, SemVer product release."""
from pathlib import Path

from helpers import ROOT

DOC = ROOT / "docs" / "RELEASE_MODEL.md"


def test_release_model_documents_dual_versions() -> None:
    text = DOC.read_text(encoding="utf-8")
    assert "Runtime / engineering version: CalVer" in text
    assert "Product release: SemVer tag" in text
    assert "Do not replace it with SemVer." in text
    assert "v5.0.0" in text


def test_v5_baseline_is_recorded() -> None:
    text = DOC.read_text(encoding="utf-8")
    assert "Runtime CalVer: `2026.09.16.0`" in text
    assert "Baseline commit: `6f407df`" in text
    assert "unified `liki` skill" in text


def test_runtime_version_remains_calver() -> None:
    version = (ROOT / "skills" / "liki" / "VERSION").read_text(encoding="utf-8").strip()
    parts = version.split(".")
    assert len(parts) == 4
    assert all(part.isdigit() for part in parts)
    assert version.startswith("20")
