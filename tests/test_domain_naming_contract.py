import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def test_tool_catalog_is_self_describing():
    catalog = json.loads((ROOT / "tools/mcp-tool-catalog.json").read_text())
    assert set(catalog) == {"servers"}


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


def test_only_production_expert_profiles_are_released():
    profiles = {path.stem for path in (ROOT / "profiles").glob("*.json")}
    assert profiles == {"natal", "divination", "fengshui", "naming"}
    assert not (ROOT / "agents/single_expert").exists()


def test_deployment_generator_emits_contract_version():
    import subprocess
    import tempfile

    # TemporaryDirectory：mkdtemp 不清理，历史上在 /tmp 泄漏过成百个目录。
    with tempfile.TemporaryDirectory(prefix="liki-deployment-") as tmp:
        output = Path(tmp) / "natal"
        subprocess.run(
            [
                "python3",
                str(ROOT / "scripts/generate_deployment.py"),
                "--profile",
                "natal",
                "--out",
                str(output.parent),
            ],
            check=True,
        )
        deployment = json.loads((output / "deployment.json").read_text())
    assert deployment["schema_version"] == "liki.agent-deployment.v1"


def test_candidate_manifest_is_self_describing():
    import sys
    import tempfile

    sys.path.insert(0, str(ROOT / "scripts"))
    import build_candidate_manifest as bcm

    # Test the pure dictionary shape with temporary metadata files.
    with tempfile.TemporaryDirectory() as directory:
        metadata_dir = Path(directory)
        for name in bcm.IMAGE_NAMES:
            (metadata_dir / f"{name}.json").write_text(
                json.dumps({"containerimage.digest": "sha256:" + "0" * 64})
            )
        result = bcm.manifest(
            "a" * 40,
            "v2026.10.16.0",
            metadata_dir,
            "ghcr.io",
        )
    assert result["schema_version"] == "liki.candidate-manifest.v1"
    assert set(result["images"]) == {"engine", "counsel", "experts"}


def test_release_manifest_contract_is_declared():
    source = (ROOT / "scripts/build_release_manifest.py").read_text()
    assert '"schema_version": "liki.release-manifest.v1"' in source
