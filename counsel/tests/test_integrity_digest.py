"""Cross-runtime canonical digest contract."""

import re

import pytest

from app.integrity.digest import require_sha256, sha256_excluding, sha256_json


def test_canonical_json_matches_cross_runtime_reference_vector():
    value = {
        "b": 1,
        "a": ["甲", {"z": False, "y": None}],
        "n": 12345678901234567890,
        "s": '<a>&"猴"\u2028',
    }
    assert sha256_json(value) == (
        "bf15cbf7215513b63017a1a28b7f2a7590fffc4a23414a8a26e14fd119154d6d"
    )


def test_sha256_excluding_ignores_digest_metadata():
    value = {"chart": {"gan": "甲"}, "chart_digest": "ignored"}
    digest = sha256_excluding(value, "chart_digest")
    assert digest == sha256_json({"chart": {"gan": "甲"}})
    assert re.fullmatch(r"[0-9a-f]{64}", digest)


def test_require_sha256_rejects_non_canonical_form():
    with pytest.raises(ValueError):
        require_sha256("short", "digest")
    with pytest.raises(ValueError):
        require_sha256("A" * 64, "digest")
