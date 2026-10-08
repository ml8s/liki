from __future__ import annotations

import hashlib
import subprocess
import tarfile
from pathlib import Path

import yaml


ROOT = Path(__file__).resolve().parents[1]
DIST = ROOT / "dist"
REQUIRED_BUNDLE_FILES = {
    "agents/experts/deployment.json",
    "agents/single/deployment.json",
    "contracts/agent-definition.schema.json",
    "contracts/agent-definition.version",
    "index.json",
    "liki-release-manifest.json",
    "liki-release-manifest.sha256",
    "liki.tar.gz",
}


def run(command: list[str]) -> None:
    subprocess.run(command, cwd=ROOT, check=True)


def test_release_bundle_contains_all_verified_artifacts() -> None:
    run(["make", "build-release-manifest"])

    schema = DIST / "contracts/agent-definition.schema.json"
    pin = DIST / "contracts/agent-definition.version"
    assert schema.is_file(), "release bundle schema was not staged"
    assert pin.is_file(), "release bundle contract pin was not staged"

    pin_values = {}
    for line in pin.read_text(encoding="utf-8").splitlines():
        if ":" in line and not line.lstrip().startswith("#"):
            key, value = line.split(":", 1)
            pin_values[key.strip()] = value.strip()

    actual_schema_digest = hashlib.sha256(schema.read_bytes()).hexdigest()
    assert pin_values["digest"] == f"sha256:{actual_schema_digest}"

    bundle = DIST / "liki-release-artifacts.tar.gz"
    subprocess.run(
        [
            "tar",
            "-czf",
            str(bundle),
            "-C",
            str(DIST),
            *sorted(REQUIRED_BUNDLE_FILES),
        ],
        check=True,
    )
    with tarfile.open(bundle, "r:gz") as archive:
        members = {member.name for member in archive.getmembers() if member.isfile()}
    assert members == REQUIRED_BUNDLE_FILES

    manifest = (DIST / "liki-release-manifest.json").read_text(encoding="utf-8")
    assert actual_schema_digest in manifest


def test_release_profiles_match_runtime_and_catalog() -> None:
    runtime = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()
    catalog = (
        ROOT / "contracts/mcp-tool-catalog.json"
    ).read_text(encoding="utf-8")
    assert f'"runtime_version": "{runtime}"' in catalog

    for profile in (ROOT / "profiles").glob("*.json"):
        document = yaml.safe_load(profile.read_text(encoding="utf-8"))
        assert document["metadata"]["version"] == runtime
