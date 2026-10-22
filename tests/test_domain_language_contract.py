"""Contract tests for the canonical domain language registry and guard."""

from __future__ import annotations

import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

REGISTRY = ROOT / "tools/domain-language.json"


def _load() -> dict:
    return json.loads(REGISTRY.read_text(encoding="utf-8"))


def test_registry_has_contract_version_and_core_concepts():
    registry = _load()
    assert registry["schema_version"] == "liki.domain-language.v1"
    ids = {concept["id"] for concept in registry["concepts"]}
    assert {
        "instant", "zone", "apparent_solar_time", "lunar_date", "shichen",
        "solar_term_state", "chart", "factor_snapshot", "assertion",
        "snapshot", "answer", "yongshen",
    } <= ids


def test_instant_separates_domain_identity_from_utc_representation():
    concept = next(item for item in _load()["concepts"] if item["id"] == "instant")
    assert concept["surfaces"] == {
        "go": "time.Time",
        "python": "datetime.datetime",
        "json": "instant",
        "zh_cn": "精确时刻",
    }
    legacy_timestamp = "time" + "stamp"
    assert legacy_timestamp in concept["forbidden"]


def test_every_compatibility_alias_is_scoped_and_expires():
    for concept in _load()["concepts"]:
        for alias in concept["compatibility"]:
            assert alias["scope"], (concept["id"], alias["name"])
            assert alias["reason"], (concept["id"], alias["name"])
            assert alias["remove_in"], (concept["id"], alias["name"])


def test_generated_domain_document_is_fresh():
    result = subprocess.run(
        ["python3", str(ROOT / "scripts/sync_domain_language.py"), "--check"],
        cwd=ROOT,
        check=False,
        text=True,
        capture_output=True,
    )
    assert result.returncode == 0, result.stderr


def test_domain_language_guard_is_green():
    result = subprocess.run(
        ["python3", str(ROOT / "scripts/check_domain_language.py")],
        cwd=ROOT,
        check=False,
        text=True,
        capture_output=True,
    )
    assert result.returncode == 0, result.stderr
