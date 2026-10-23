"""SKILL metadata platform-format contract, without pinning marketing copy."""

from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
SKILL_MD = ROOT / "skills/liki/SKILL.md"


def _metadata() -> dict:
    text = SKILL_MD.read_text(encoding="utf-8")
    assert text.startswith("---\n")
    metadata = yaml.safe_load(text.split("---\n")[1])
    assert isinstance(metadata, dict)
    return metadata


def test_metadata_values_are_adk_compatible_scalars():
    metadata = _metadata()["metadata"]
    assert isinstance(metadata, dict)
    assert metadata
    for key, value in metadata.items():
        assert isinstance(key, str) and key.strip(), key
        assert isinstance(value, str), f"metadata.{key} must be a string, got {type(value).__name__}"


def test_description_is_within_adk_limit():
    description = _metadata().get("description") or ""
    assert description.strip()
    assert len(description.encode("utf-8")) <= 1024


def test_optional_keywords_are_flat_and_bounded():
    raw = _metadata()["metadata"].get("keywords")
    if raw is None:
        return
    assert isinstance(raw, str)
    keywords = [keyword.strip() for keyword in raw.split(",") if keyword.strip()]
    assert keywords == list(dict.fromkeys(keywords))
    assert len(keywords) <= 32
    for keyword in keywords:
        assert 1 <= len(keyword) <= 64
        assert keyword == keyword.strip()
        assert not any(char.isspace() and char != " " for char in keyword)
