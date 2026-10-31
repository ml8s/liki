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
    profiles = [path.stem for path in (ROOT / "profiles").glob("*.json")]
    assert profiles == ["liki"]
    assert not (ROOT / "agents/single_expert").exists()


def test_deployment_generator_emits_contract_version():
    import subprocess
    import tempfile

    # TemporaryDirectory：mkdtemp 不清理，历史上在 /tmp 泄漏过成百个目录。
    with tempfile.TemporaryDirectory(prefix="liki-deployment-") as tmp:
        output = Path(tmp) / "liki"
        subprocess.run(
            [
                "python3",
                str(ROOT / "scripts/generate_deployment.py"),
                "--profile",
                "liki",
                "--out",
                str(output.parent),
            ],
            check=True,
        )
        deployment = json.loads((output / "deployment.json").read_text())
    assert deployment["schema_version"] == "agent.deployment.v1"
