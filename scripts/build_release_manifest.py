#!/usr/bin/env python3
"""Build and verify the immutable liki release manifest.

The manifest deliberately contains only values that can be recomputed from the
release artifacts and repository metadata. Image digests are injected by the
release workflow after publication, using the LIKI_RELEASE_IMAGES JSON env var.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
DIST = ROOT / "dist"
REQUIRED_IMAGES = ("counsel_mcp", "engine", "engine_mcp", "experts")
IMAGE_DIGEST_RE = re.compile(r"^sha256:[0-9a-f]{64}$")


def digest(path: Path) -> str:
    return "sha256:" + hashlib.sha256(path.read_bytes()).hexdigest()


def parse_pin(path: Path) -> tuple[str, str]:
    values: dict[str, str] = {}
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or ":" not in line:
            continue
        key, value = line.split(":", 1)
        values[key.strip()] = value.strip()
    return values.get("version", ""), values.get("digest", "")


def runtime_version() -> str:
    return (ROOT / "skills" / "liki" / "VERSION.txt").read_text(encoding="utf-8").strip()


def agent_versions() -> dict[str, str]:
    versions: dict[str, str] = {}
    for path in sorted((ROOT / "agents").glob("*/agent.yaml")):
        match = re.search(r"^version:\s*(\S+)", path.read_text(encoding="utf-8"), re.MULTILINE)
        if not match:
            raise SystemExit(f"missing version: {path}")
        versions[path.parent.name] = match.group(1)
    return versions


def build(require_images: bool = False) -> dict[str, Any]:
    schema_version, schema_digest = parse_pin(ROOT / "contracts" / "agent-definition.version")
    actual_schema_digest = digest(ROOT / "contracts" / "agent-definition.schema.json")
    if schema_digest != actual_schema_digest:
        raise SystemExit(
            f"contract digest mismatch: pin={schema_digest} actual={actual_schema_digest}"
        )

    versions = agent_versions()
    expected_runtime = runtime_version()
    mismatched = {name: version for name, version in versions.items() if version != expected_runtime}
    if mismatched:
        raise SystemExit(f"agent runtime version drift: {mismatched} != {expected_runtime}")

    deployments: dict[str, Any] = {}
    for profile in ("experts", "single"):
        path = DIST / "agents" / profile / "deployment.json"
        if not path.is_file():
            raise SystemExit(f"missing generated deployment: {path}")
        document = json.loads(path.read_text(encoding="utf-8"))
        if document.get("metadata", {}).get("version") != expected_runtime:
            raise SystemExit(
                f"deployment version drift for {profile}: "
                f"{document.get('metadata', {}).get('version')} != {expected_runtime}"
            )
        deployments[profile] = {"sha256": digest(path)}

    skill_archive = DIST / "liki.tar.gz"
    skill_index = DIST / "index.json"
    if not skill_archive.is_file() or not skill_index.is_file():
        raise SystemExit("skill archive is missing; run make build-archive")

    skill_bundle: dict[str, str] = {
        "archive_sha256": digest(skill_archive),
        "index_sha256": digest(skill_index),
    }
    contract_dist = DIST / "contracts"
    contract_dist.mkdir(parents=True, exist_ok=True)
    shutil.copy2(
        ROOT / "contracts/agent-definition.schema.json",
        contract_dist / "agent-definition.schema.json",
    )
    shutil.copy2(
        ROOT / "contracts/agent-definition.version",
        contract_dist / "agent-definition.version",
    )
    web_bundle = DIST / "liki-web-skill-bundle.tar.gz"
    if web_bundle.is_file():
        skill_bundle["web_bundle_sha256"] = digest(web_bundle)

    images: dict[str, Any] = {}
    if os.environ.get("LIKI_RELEASE_IMAGES"):
        try:
            images = json.loads(os.environ["LIKI_RELEASE_IMAGES"])
        except json.JSONDecodeError as error:
            raise SystemExit(f"LIKI_RELEASE_IMAGES is not valid JSON: {error}") from error
        if not isinstance(images, dict):
            raise SystemExit("LIKI_RELEASE_IMAGES must be a JSON object")

    if require_images:
        missing = [name for name in REQUIRED_IMAGES if name not in images]
        if missing:
            raise SystemExit(
                "release manifest requires image digests for: " + ", ".join(missing)
            )
        invalid = [
            name for name, reference in images.items()
            if not isinstance(reference, str) or not IMAGE_DIGEST_RE.fullmatch(reference)
        ]
        if invalid:
            raise SystemExit(
                "image references must be digest-only (sha256:<64-hex>): "
                + ", ".join(sorted(invalid))
            )
        unexpected = sorted(set(images) - set(REQUIRED_IMAGES))
        if unexpected:
            raise SystemExit(
                "unexpected image manifest keys: " + ", ".join(unexpected)
            )
        if os.environ.get("LIKI_RELEASE_TAG", "unreleased") == "unreleased":
            raise SystemExit("--require-images requires LIKI_RELEASE_TAG")

    return {
        "schema_version": 1,
        "artifact": "liki-release-manifest.json",
        "release_tag": os.environ.get("LIKI_RELEASE_TAG", "unreleased"),
        "runtime_version": expected_runtime,
        "source_commit": os.environ.get(
            "LIKI_SOURCE_COMMIT",
            subprocess.check_output(["git", "-C", str(ROOT), "rev-parse", "HEAD"], text=True).strip(),
        ),
        "agent_contract": {
            "version": schema_version,
            "schema_sha256": schema_digest,
        },
        "agent_deployments": deployments,
        "skill_bundle": skill_bundle,
        "images": images,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="verify the existing manifest")
    parser.add_argument(
        "--require-images",
        action="store_true",
        help="require digest-pinned images and release metadata (release mode)",
    )
    args = parser.parse_args()

    manifest = build(require_images=args.require_images)
    output = DIST / "liki-release-manifest.json"
    canonical = json.dumps(manifest, ensure_ascii=False, sort_keys=True, indent=2) + "\n"

    if args.check:
        if not output.is_file():
            raise SystemExit(f"missing release manifest: {output}")
        expected = output.read_text(encoding="utf-8")
        if expected != canonical:
            raise SystemExit("release manifest drift detected")
        print(f"release manifest ok: {output}")
        return 0

    output.write_text(canonical, encoding="utf-8")
    checksum = hashlib.sha256(canonical.encode("utf-8")).hexdigest()
    (DIST / "liki-release-manifest.sha256").write_text(f"{checksum}  {output.name}\n", encoding="utf-8")
    print(f"release manifest: {output}")
    print(f"release manifest digest: sha256:{checksum}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
