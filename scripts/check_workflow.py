#!/usr/bin/env python3
"""Validate GitHub workflow structure: YAML parses, needs references resolve.

A dangling `needs:` entry makes the whole workflow fail to start (zero jobs
reported), which is silent locally. Run as part of `make check`.
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
REQUIRED_IMAGE_DIGEST_ARTIFACTS = {
    "image-digest-counsel_mcp",
    "image-digest-engine",
    "image-digest-engine_mcp",
    "image-digest-experts",
}


def main() -> int:
    errors: list[str] = []
    for path in sorted(ROOT.glob(".github/workflows/*.yml")) + sorted(
        ROOT.glob(".github/workflows/*.yaml")
    ):
        try:
            document = yaml.safe_load(path.read_text(encoding="utf-8"))
        except yaml.YAMLError as error:
            errors.append(f"{path.relative_to(ROOT)}: invalid YAML: {error}")
            continue
        if not isinstance(document, dict):
            continue
        jobs = document.get("jobs") or {}
        if not isinstance(jobs, dict):
            errors.append(f"{path.relative_to(ROOT)}: jobs must be a mapping")
            continue
        concurrency = document.get("concurrency")
        if not isinstance(concurrency, dict):
            errors.append(f"{path.relative_to(ROOT)}: missing concurrency group")
        else:
            group = concurrency.get("group")
            cancel = concurrency.get("cancel-in-progress")
            if not isinstance(group, str) or "github.workflow" not in group:
                errors.append(f"{path.relative_to(ROOT)}: concurrency.group must include github.workflow")
            if cancel != "${{ github.event_name == 'pull_request' }}":
                errors.append(
                    f"{path.relative_to(ROOT)}: concurrency.cancel-in-progress must be "
                    "true only for pull_request"
                )
        pins: dict[str, str] = {}
        for job_id, job in jobs.items():
            if not isinstance(job, dict):
                continue
            timeout = job.get("timeout-minutes")
            if isinstance(timeout, bool) or not isinstance(timeout, int) or not 1 <= timeout <= 60:
                errors.append(
                    f"{path.relative_to(ROOT)}: job {job_id!r} needs timeout-minutes in 1..60"
                )
            for step in job.get("steps") or []:
                if not isinstance(step, dict):
                    continue
                run_script = step.get("run")
                if isinstance(run_script, str):
                    if re.search(r"--identity-token\s+[\"']?\$RUNNER_TEMP/", run_script):
                        errors.append(
                            f"{path.relative_to(ROOT)}: job {job_id!r} passes an OIDC token "
                            "file path to cosign; consume the token string in the same step"
                        )
                    if '--identity-token "$oidc_token"' in run_script and not re.search(
                        r"oidc_token=\"\$\(.*\)\"", run_script, flags=re.DOTALL
                    ):
                        errors.append(
                            f"{path.relative_to(ROOT)}: job {job_id!r} uses oidc_token "
                            "without fetching it in the same shell step"
                        )
                uses = step.get("uses")
                if not isinstance(uses, str) or "@" not in uses:
                    continue
                if uses.rpartition("@")[0] == "docker/build-push-action":
                    if "id" not in step:
                        errors.append(
                            f"{path.relative_to(ROOT)}: job {job_id!r} uses "
                            "docker/build-push-action without a step id"
                        )
                    elif "id" in (step.get("with") or {}):
                        errors.append(
                            f"{path.relative_to(ROOT)}: job {job_id!r} puts "
                            "docker/build-push-action step id in `with`; outputs.digest "
                            "will be empty"
                        )
                if uses.rpartition("@")[0] == "sigstore/cosign-installer":
                    with_block = step.get("with") or {}
                    if with_block.get("cosign-release") != "v2.4.1":
                        errors.append(
                            f"{path.relative_to(ROOT)}: job {job_id!r} must pin "
                            "cosign-installer to cosign-release v2.4.1"
                        )
                action, _, ref = uses.rpartition("@")
                if not all(c in "0123456789abcdef" for c in ref) or len(ref) != 40:
                    continue
                if action in pins and pins[action] != ref:
                    errors.append(
                        f"{path.relative_to(ROOT)}: job {job_id!r} pins {action} "
                        f"to {ref} but another job uses {pins[action]}"
                    )
                pins.setdefault(action, ref)
        for job_id, job in jobs.items():
            if not isinstance(job, dict):
                continue
            needs = job.get("needs")
            if needs is None:
                continue
            for ref in needs if isinstance(needs, list) else [needs]:
                if ref not in jobs:
                    errors.append(
                        f"{path.relative_to(ROOT)}: job {job_id!r} needs "
                        f"unknown job {ref!r}"
                    )
    if errors:
        for error in errors:
            print(f"❌ {error}", file=sys.stderr)
            return 1
    for workflow in sorted(ROOT.glob(".github/workflows/*.yml")):
        try:
            document = yaml.safe_load(workflow.read_text(encoding="utf-8"))
        except yaml.YAMLError:
            continue
        for job in (document.get("jobs") or {}).values():
            if not isinstance(job, dict):
                continue
            for step in job.get("steps") or []:
                if not isinstance(step, dict):
                    continue
                uses = step.get("uses")
                if not isinstance(uses, str) or uses.rpartition("@")[0] != "actions/upload-artifact":
                    continue
                with_block = step.get("with") or {}
                if isinstance(with_block.get("name"), str):
                    REQUIRED_IMAGE_DIGEST_ARTIFACTS.discard(with_block["name"])
    if REQUIRED_IMAGE_DIGEST_ARTIFACTS:
        errors.append(
            "missing release image-digest artifacts: "
            + ", ".join(sorted(REQUIRED_IMAGE_DIGEST_ARTIFACTS))
        )
    if errors:
        for error in errors:
            print(f"❌ {error}", file=sys.stderr)
        return 1
    print("✓ workflow structure")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
