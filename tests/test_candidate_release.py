from __future__ import annotations

import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPTS = ROOT / "scripts"
if str(SCRIPTS) not in sys.path:
    sys.path.insert(0, str(SCRIPTS))

import build_candidate_manifest as bcm


def _metadata(directory: Path) -> None:
    for name in bcm.IMAGE_NAMES:
        (directory / f"{name}.json").write_text(
            json.dumps({"containerimage.digest": "sha256:" + "0" * 64}),
            encoding="utf-8",
        )


def test_candidate_manifest_has_exact_required_shape(tmp_path):
    metadata = tmp_path / "metadata"
    metadata.mkdir()
    _metadata(metadata)
    result = bcm.manifest(
        "a" * 40, "v2026.10.16.0", metadata, "ghcr.io"
    )
    assert result == {
        "registry": "ghcr.io",
        "repository": "ml8s/liki",
        "source_commit": "a" * 40,
        "candidate_tag": "v2026.10.16.0",
        "schema_version": "liki.candidate-manifest.v1",
        "images": {
            "engine": "sha256:" + "0" * 64,
            "counsel": "sha256:" + "0" * 64,
            "experts": "sha256:" + "0" * 64,
        },
    }


def test_candidate_manifest_accepts_release_tag(tmp_path):
    metadata = tmp_path / "metadata"
    metadata.mkdir()
    _metadata(metadata)
    result = bcm.manifest(
        "a" * 40, "v2026.10.16.0", metadata, "ghcr.io"
    )
    assert result["candidate_tag"] == "v2026.10.16.0"
    assert result["repository"] == "ml8s/liki"


def test_candidate_manifest_rejects_malformed_identity(tmp_path):
    metadata = tmp_path / "metadata"
    metadata.mkdir()
    _metadata(metadata)
    try:
        bcm.manifest("short", "v2026.10.16.0", metadata, "ghcr.io")
    except SystemExit as error:
        assert "invalid source commit" in str(error)
    else:
        raise AssertionError("malformed source SHA accepted")
    try:
        bcm.manifest("a" * 40, "latest", metadata, "ghcr.io")
    except SystemExit as error:
        assert "invalid candidate tag" in str(error)
    else:
        raise AssertionError("malformed candidate tag accepted")
