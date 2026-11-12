"""Release version has one hand-maintained source; everything else is generated."""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def test_release_version_is_single_calver_source() -> None:
    path = ROOT / "skills/liki/VERSION.txt"
    version = path.read_text(encoding="utf-8").strip()
    parts = version.split(".")
    assert len(parts) == 4
    assert all(part.isdigit() for part in parts)
    assert parts[0].startswith("20")


def test_internal_sources_do_not_track_release_version() -> None:
    for relative in ("pyproject.toml", "counsel/pyproject.toml"):
        assert 'version = "0.0.0"' in (ROOT / relative).read_text(encoding="utf-8")
    profile = (ROOT / "distribution/liki-expert/profile.json").read_text(encoding="utf-8")
    assert '"version"' not in profile
