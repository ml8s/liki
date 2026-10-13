from __future__ import annotations

from pathlib import Path

import pytest


ROOT = Path(__file__).resolve().parents[1]
MAKEFILE = (ROOT / "Makefile").read_text(encoding="utf-8")


def test_python_coverage_target_runs_domain_suite() -> None:
    assert "test-python-coverage:" in MAKEFILE
    assert "coverage erase" in MAKEFILE
    assert "coverage run --source=counsel/app" in MAKEFILE
    assert "-m pytest tests" in MAKEFILE


def test_python_coverage_fails_closed_and_emits_reports() -> None:
    assert "coverage report --fail-under=65" in MAKEFILE
    assert "coverage xml" in MAKEFILE
    assert "coverage html" in MAKEFILE


def test_coverage_artifacts_are_not_tracked(tmp_path: Path) -> None:
    gitignore = (ROOT / ".gitignore").read_text(encoding="utf-8")
    for artifact in (".coverage", ".coverage.*", "coverage.xml", "htmlcov/"):
        assert artifact in gitignore, artifact


def test_ci_runs_python_coverage_and_uploads_reports() -> None:
    workflow = (ROOT / ".github/workflows/ci.yml").read_text(encoding="utf-8")
    assert "make test-python-coverage" in workflow
    assert "name: python-coverage" in workflow
    assert "coverage.xml" in workflow
    assert "htmlcov" in workflow
