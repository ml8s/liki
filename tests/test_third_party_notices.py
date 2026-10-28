"""Third-party data notices must stay connected to the generated city dataset."""

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def test_city_attribution_is_generated_from_manifest():
    result = subprocess.run(
        [sys.executable, str(ROOT / "scripts" / "generate_city_attribution.py"), "--check"],
        cwd=ROOT,
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, result.stderr


def test_notice_declares_non_mit_geographic_data():
    notice = (ROOT / "NOTICE.md").read_text(encoding="utf-8")
    attribution = (
        ROOT / "engine" / "internal" / "engine" / "city" / "ATTRIBUTION.txt"
    ).read_text(encoding="utf-8")

    assert "**not** relicensed" in notice
    for term in ("DataV.GeoAtlas", "GeoNames", "GeoBoundaries"):
        assert term in notice
        assert term in attribution
    assert "data/sources/manifest.json" in attribution


def test_engine_image_contains_attribution_notice():
    dockerfile = (ROOT / "engine" / "dev" / "Dockerfile").read_text(encoding="utf-8")
    assert "ATTRIBUTION.txt" in dockerfile
    assert "/usr/share/licenses/liki-engine/ATTRIBUTION.txt" in dockerfile
