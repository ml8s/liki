#!/usr/bin/env python3
"""Validate and smoke-test the assembled Liki expert image."""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
import tarfile
import tempfile
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
DEPLOYMENT_NAME = "liki-expert"
AGENT_NAME = "liki"


def git_revision() -> str:
    return subprocess.check_output(
        ["git", "-C", str(ROOT), "rev-parse", "HEAD"], text=True
    ).strip()


def docker(*args: str, capture: bool = True, check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["docker", *args],
        check=check,
        capture_output=capture,
        text=True,
    )


def docker_json(args: list[str]) -> Any:
    result = docker(*args)
    return json.loads(result.stdout)


def validate_deployment(image: str, dist: Path) -> None:
    print("1/3 Validating the generated AgentDeployment with AgentGo")
    dist_root = dist.resolve()
    docker(
        "run",
        "--pull=never",
        "--rm",
        "--entrypoint=/agentgo",
        f"--volume={dist_root}:/deployment:ro",
        image,
        "validate",
        "-deployment",
        f"/deployment/{DEPLOYMENT_NAME}/deployment.json",
    )


def image_labels(image: str) -> dict[str, str]:
    result = docker("image", "inspect", "--format={{json .Config.Labels}}", image)
    labels = json.loads(result.stdout)
    return {key: value for key, value in labels.items() if value is not None}


def normalize_tar_name(name: str) -> str:
    return name.removeprefix("./").lstrip("/")


def assert_image_content(image: str, dist: Path, revision: str) -> None:
    print("2/3 Checking the expert image content")
    labels = image_labels(image)
    actual_revision = labels.get("org.opencontainers.image.revision")
    if actual_revision != revision:
        raise SystemExit(
            f"image revision mismatch: got {actual_revision!r}, want {revision!r}"
        )

    deployment_paths = (
        f"deployment/{DEPLOYMENT_NAME}/deployment.json",
        f"deployment/{DEPLOYMENT_NAME}/agents/{AGENT_NAME}/instruction.md",
    )
    skill_root = ROOT / "skills" / "liki"
    references = sorted(
        path.name
        for path in (skill_root / "references").glob("*.md")
    )

    with tempfile.TemporaryDirectory(prefix="liki-image-export-") as temp:
        export_path = Path(temp) / "image.tar"
        container = f"liki-image-export-{uuid.uuid4().hex}"
        docker("create", "--name", container, image)
        try:
            docker("export", "--output", str(export_path), container)
        finally:
            docker("rm", container, check=False)

        with tarfile.open(export_path, "r:*") as archive:
            members = {
                normalize_tar_name(member.name): member
                for member in archive.getmembers()
                if member.isfile()
            }
            extracted = {
                name: archive.extractfile(member).read()
                for name, member in members.items()
            }

        for relative in deployment_paths:
            member = members.get(relative)
            if member is None:
                raise SystemExit(f"missing image file: /{relative}")
            source = dist / relative.removeprefix("deployment/")
            if extracted[relative] != source.read_bytes():
                raise SystemExit(f"image file differs from source: /{relative}")

        if members.get(f"skills/{skill_root.name}/SKILL.md") is None:
            raise SystemExit("missing image file: /skills/liki/SKILL.md")

        image_references = {
            name.removeprefix(f"skills/{skill_root.name}/references/")
            for name in members
            if name.startswith(f"skills/{skill_root.name}/references/") and name.endswith(".md")
        }
        if image_references != set(references):
            raise SystemExit(
                "image methodology files differ from the canonical skill source: "
                f"{sorted(image_references ^ set(references))}"
            )


def mapped_port(container: str, port: int) -> int:
    result = docker("port", container, str(port))
    mappings = [line.strip() for line in result.stdout.splitlines() if line.strip()]
    if not mappings:
        raise SystemExit(f"container port {port} is not published")
    preferred = next((line for line in mappings if line.startswith("127.0.0.1:")), mappings[-1])
    return int(preferred.rsplit(":", 1)[1])


def wait_for_livez(url: str, container: str, timeout: float) -> None:
    deadline = time.monotonic() + timeout
    last_error: Exception | None = None
    while time.monotonic() < deadline:
        state = docker_json(
            ["container", "inspect", "--format={{json .State}}", container]
        )
        if state.get("Status") == "exited":
            logs = docker("logs", container).stdout + docker("logs", container).stderr
            raise SystemExit(f"expert image exited before becoming live:\n{logs}")

        try:
            with urllib.request.urlopen(url, timeout=2) as response:
                if response.status == 200:
                    return
                last_error = RuntimeError(f"/livez returned HTTP {response.status}")
        except (OSError, urllib.error.URLError) as exc:
            last_error = exc
        except urllib.error.HTTPError as exc:
            last_error = RuntimeError(f"/livez returned HTTP {exc.code}")
        time.sleep(0.5)
    raise SystemExit(f"expert image did not become live within {timeout:g}s: {last_error}")


def smoke_runtime(image: str, dist: Path, timeout: float) -> None:
    print("3/3 Running the expert image runtime smoke")
    deployment = json.loads(
        (dist / DEPLOYMENT_NAME / "deployment.json").read_text(encoding="utf-8")
    )
    dist_root = dist.resolve()
    container = f"liki-expert-smoke-{uuid.uuid4().hex}"
    try:
        args = [
            "run",
            "--pull=never",
            "--detach",
            "--name",
            container,
            "-v",
            f"{dist_root}:/deployment:ro",
            "-p",
            "127.0.0.1::8083",
            "-p",
            "127.0.0.1::8085",
            "--env",
            "AGENTGO_ENV=development",
            "--env",
            f"AGENTGO_DEPLOYMENT_FILE=/deployment/{DEPLOYMENT_NAME}/deployment.json",
            "--env",
            "AGENTGO_AUDIT_DB_PATH=/data/liki-test-audit.db",
            "--env",
            "AGENTGO_PROTOCOL_STATE_DB_PATH=/data/liki-test-protocol-state.db",
            "--env",
            "AGENTGO_ADK_SESSION_DB_PATH=/data/liki-test-adk-session.db",
            "--env",
            "AGENTGO_LLM_BASE_URL=http://127.0.0.1:9/v1",
            "--env",
            "AGENTGO_LLM_API_KEY=local-image-test",
            "--env",
            "AGENTGO_LLM_MODEL=local-image-test",
            "--env",
            "AGENTGO_LLM_PROVIDER=openai",
        ]
        for server in deployment["spec"]["mcpServers"]:
            args.extend(["--env", f"{server['endpointEnv']}=http://127.0.0.1:9/mcp"])
        args.append(image)

        docker(*args)
        ops_port = mapped_port(container, 8085)
        wait_for_livez(
            f"http://127.0.0.1:{ops_port}/livez", container, timeout=timeout
        )
    finally:
        logs = docker("logs", container, check=False)
        if logs.stdout:
            print(logs.stdout.rstrip())
        docker("rm", "--force", container, check=False)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--image", required=True, help="assembled Liki expert image")
    parser.add_argument("--dist", type=Path, default=ROOT / "dist")
    parser.add_argument("--revision", default=git_revision())
    parser.add_argument("--runtime-timeout", type=float, default=30)
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    dist = args.dist.resolve()
    deployment = dist / DEPLOYMENT_NAME / "deployment.json"
    if not deployment.is_file():
        raise SystemExit(f"deployment artifact missing: {deployment}; run make build-expert-deployment")

    validate_deployment(args.image, dist)
    assert_image_content(args.image, dist, args.revision)
    smoke_runtime(args.image, dist, timeout=args.runtime_timeout)
    print(f"✓ expert image smoke passed: {args.image}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
