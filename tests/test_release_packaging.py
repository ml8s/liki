from __future__ import annotations

import subprocess
import tarfile
from pathlib import Path

import yaml


ROOT = Path(__file__).resolve().parents[1]
DIST = ROOT / "dist"
REQUIRED_BUNDLE_FILES = {
    "agents/experts/deployment.json",
    "agents/single/deployment.json",
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


def run(command: list[str]) -> None:
    subprocess.run(command, cwd=ROOT, check=True)


def test_release_bundle_contains_all_verified_artifacts() -> None:
    run(["make", "build-web-skill-bundle"])
    run(["make", "build-deployment"])
    run(["python3", "scripts/build_release_manifest.py"])

    for profile in ("experts", "single"):
        assert (DIST / "agents" / profile / "deployment.json").is_file()

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
    assert '"runtime_version"' in manifest


def test_release_profiles_match_runtime() -> None:
    runtime = (ROOT / "skills/liki/VERSION.txt").read_text(encoding="utf-8").strip()

    for profile in (ROOT / "profiles").glob("*.json"):
        document = yaml.safe_load(profile.read_text(encoding="utf-8"))
        assert document["metadata"]["version"] == runtime
