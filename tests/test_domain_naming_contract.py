import json
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parents[1]


def test_tool_catalog_is_self_describing():
    catalog = json.loads((ROOT / "tools/mcp-tool-catalog.json").read_text())
    assert catalog["schema_version"] == "liki.mcp-tool-catalog.v1"


def test_naming_tools_use_domain_prefix():
    catalog = json.loads((ROOT / "tools/mcp-tool-catalog.json").read_text())
    tools = set(catalog["servers"]["counsel"]["tools"])
    assert tools >= {
        "naming_surname",
        "naming_pick",
        "naming_char",
        "naming_compose",
        "naming_check",
    }
    assert not any(name.startswith("qiming_") for name in tools)


def test_only_production_expert_profile_is_released():
    profiles = {path.stem for path in (ROOT / "profiles").glob("*.json")}
    assert profiles == {"experts"}
    assert not (ROOT / "agents/single_expert").exists()


def test_deployment_generator_emits_contract_version():
    import subprocess
    import tempfile

    output = Path(tempfile.mkdtemp(prefix="liki-deployment-")) / "experts"
    subprocess.run(
        [
            "python3",
            str(ROOT / "scripts/generate_deployment.py"),
            "--profile",
            "experts",
            "--out",
            str(output.parent),
        ],
        check=True,
    )
    deployment = json.loads((output / "deployment.json").read_text())
    assert deployment["schema_version"] == "liki.agent-deployment.v1"


def test_candidate_manifest_is_self_describing():
    import sys

    sys.path.insert(0, str(ROOT / "scripts"))
    import build_candidate_manifest as bcm

    metadata = ROOT / "build/nonexistent"
    manifest = bcm.manifest.__wrapped__ if hasattr(bcm.manifest, "__wrapped__") else bcm.manifest
    # Test the pure dictionary shape with temporary metadata files.
    import tempfile

    with tempfile.TemporaryDirectory() as directory:
        metadata_dir = Path(directory)
        for name in bcm.IMAGE_NAMES:
            (metadata_dir / f"{name}.json").write_text(
                json.dumps({"containerimage.digest": "sha256:" + "0" * 64})
            )
        result = bcm.manifest(
            "a" * 40,
            "candidate-abcdef0-20260101010101",
            metadata_dir,
            "127.0.0.1:5000",
        )
    assert result["schema_version"] == "liki.candidate-manifest.v1"
    assert set(result["images"]) == {"engine", "counsel", "experts"}


def test_release_manifest_contract_is_declared():
    source = (ROOT / "scripts/build_release_manifest.py").read_text()
    assert '"schema_version": "liki.release-manifest.v1"' in source
