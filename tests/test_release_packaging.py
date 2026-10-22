
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
PROFILES = ("natal", "divination", "fengshui", "naming")
REQUIRED_BUNDLE_FILES = {
    *(f"agents/{profile}/deployment.json" for profile in PROFILES),
    "index.json",
    "liki-release-manifest.json",
    "liki-release-manifest.sha256",
    "liki-web-skill-bundle.tar.gz",
    "liki.tar.gz",
}


def test_ci_release_bundle_list_matches_contract() -> None:
    """CI 的 release-integrity tar 清单必须与 REQUIRED_BUNDLE_FILES 一致。"""
    workflow = (ROOT / ".github/workflows/ci.yml").read_text(encoding="utf-8")
    document = yaml.safe_load(workflow)
    listed: set[str] = set()
    for job in document["jobs"].values():
        for step in job.get("steps", []):
            script = step.get("run") or ""
            if "liki-release-artifacts.tar.gz" not in script:
                continue
            for line in script.splitlines():
                token = line.strip().rstrip("\\").strip()
                if token in REQUIRED_BUNDLE_FILES:
                    listed.add(token)
    assert listed == REQUIRED_BUNDLE_FILES, f"CI tar 清单不一致: {listed ^ REQUIRED_BUNDLE_FILES}"


def test_release_integrity_validates_before_packaging_manifest() -> None:
    """Release artifacts cannot bypass the pinned liki-agent-runtime validator."""
    workflow = (ROOT / ".github/workflows/ci.yml").read_text(encoding="utf-8")
    document = yaml.safe_load(workflow)
    steps = document["jobs"]["release-integrity"]["steps"]
    commands = [step.get("run", "") for step in steps]
    validate_index = next(index for index, command in enumerate(commands) if "make runtime-validate" in command)
    manifest_index = next(
        index for index, command in enumerate(commands) if "make build-release-manifest" in command
    )
    assert validate_index < manifest_index


def test_release_profiles_match_runtime() -> None:
    runtime = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()

    for profile in (ROOT / "profiles").glob("*.json"):
        document = yaml.safe_load(profile.read_text(encoding="utf-8"))
        assert document["metadata"]["version"] == runtime
