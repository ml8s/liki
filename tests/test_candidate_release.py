from __future__ import annotations

import hashlib
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPTS = ROOT / "scripts"
if str(SCRIPTS) not in sys.path:
    sys.path.insert(0, str(SCRIPTS))

import build_candidate_manifest as bcm
import validate_candidate_manifest as vcm


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
        "a" * 40, "candidate-abcdef0-20260101010101", metadata, "127.0.0.1:5000"
    )
    assert result == {
        "registry": "127.0.0.1:5000",
        "repository": "ml8s/liki",
        "source_commit": "a" * 40,
        "candidate_tag": "candidate-abcdef0-20260101010101",
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
        bcm.manifest("short", "candidate-abcdef0-20260101010101", metadata, "127.0.0.1:5000")
    except SystemExit as error:
        assert "invalid source commit" in str(error)
    else:
        raise AssertionError("malformed source SHA accepted")
    try:
        bcm.manifest("a" * 40, "latest", metadata, "127.0.0.1:5000")
    except SystemExit as error:
        assert "invalid candidate tag" in str(error)
    else:
        raise AssertionError("malformed candidate tag accepted")


def test_publish_candidate_rejects_wrong_evidence_contract(tmp_path):
    manifest = tmp_path / "candidate-digests.json"
    manifest.write_text(
        json.dumps({"registry": "127.0.0.1:5000"}),
        encoding="utf-8",
    )
    manifest_sha256 = hashlib.sha256(manifest.read_bytes()).hexdigest()
    manifest.with_name("assembly-evidence.json").write_text(
        json.dumps(
            {
                "status": "pass",
                "gate": "liki-deploy/candidate-assembly",
                "manifest_sha256": manifest_sha256,
                "schema_version": "legacy",
            }
        ),
        encoding="utf-8",
    )
    result = subprocess.run(
        [str(ROOT / "scripts/publish-candidate.sh")],
        env={"MANIFEST": str(manifest), "PATH": "/usr/bin:/bin"},
        text=True,
        capture_output=True,
        check=False,
    )
    assert result.returncode != 0
    assert "liki.assembly-evidence.v1" in result.stderr


def test_candidate_manifest_validator_unpacks_exact_digest_set():
    digest = "sha256:" + "0" * 64
    manifest = {
        "registry": "127.0.0.1:5000",
        "repository": "ml8s/liki",
        "source_commit": "a" * 40,
        "candidate_tag": "candidate-abcdef0-20260101010101",
        "schema_version": "liki.candidate-manifest.v1",
        "images": {name: digest for name in bcm.IMAGE_NAMES},
    }
    assert vcm.unpack(manifest, "a" * 40) == {
        name: digest for name in bcm.IMAGE_NAMES
    }


def test_candidate_manifest_validator_rejects_contract_drift():
    digest = "sha256:" + "0" * 64
    manifest = {
        "registry": "127.0.0.1:5000",
        "repository": "ml8s/liki",
        "source_commit": "a" * 40,
        "candidate_tag": "candidate-abcdef0-20260101010101",
        "schema_version": "legacy",
        "images": {name: digest for name in bcm.IMAGE_NAMES},
    }
    try:
        vcm.unpack(manifest, "a" * 40)
    except ValueError as error:
        assert "liki.candidate-manifest.v1" in str(error)
    else:
        raise AssertionError("legacy candidate contract accepted")
