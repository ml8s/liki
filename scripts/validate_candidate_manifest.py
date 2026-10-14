#!/usr/bin/env python3
"""Validate and unpack a liki release candidate manifest."""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
from pathlib import Path
from typing import Any

SCRIPTS = Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPTS))

from build_candidate_manifest import IMAGE_NAMES  # noqa: E402

SHA256_RE = re.compile(r"^sha256:[0-9a-f]{64}$")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
TAG_RE = re.compile(r"^candidate-[0-9a-f]{7}-\d{14}$")


def unpack(document: Any, source_commit: str) -> dict[str, str]:
    if not isinstance(document, dict):
        raise ValueError("manifest must be a JSON object")
    required = {"registry", "repository", "source_commit", "candidate_tag", "schema_version", "images"}
    if set(document) != required:
        missing = sorted(required - set(document))
        unexpected = sorted(set(document) - required)
        raise ValueError(f"invalid manifest fields: missing={missing}, unexpected={unexpected}")
    if document["repository"] != "ml8s/liki":
        raise ValueError("manifest repository must be ml8s/liki")
    if document["source_commit"] != source_commit:
        raise ValueError("manifest source commit does not match the release tag")
    if document["schema_version"] != "liki.candidate-manifest.v1":
        raise ValueError("manifest contract must be liki.candidate-manifest.v1")
    if not isinstance(document["registry"], str) or not document["registry"]:
        raise ValueError("manifest registry must be non-empty")
    if not isinstance(document["candidate_tag"], str) or not TAG_RE.fullmatch(document["candidate_tag"]):
        raise ValueError("manifest candidate tag is invalid")
    images = document["images"]
    if not isinstance(images, dict) or set(images) != set(IMAGE_NAMES):
        expected = ",".join(IMAGE_NAMES)
        raise ValueError(f"manifest images must contain exactly: {expected}")
    for name in IMAGE_NAMES:
        if not isinstance(images[name], str) or not SHA256_RE.fullmatch(images[name]):
            raise ValueError(f"image {name} digest must be sha256 plus 64 lowercase hex characters")
    return {name: images[name] for name in IMAGE_NAMES}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-commit", required=True)
    parser.add_argument("--output", type=Path, help="append Actions outputs here")
    args = parser.parse_args()
    if not SHA_RE.fullmatch(args.source_commit):
        raise SystemExit("source_commit must be a lowercase 40-character Git SHA")
    try:
        document = json.loads(os.environ.get("DIGESTS_JSON", ""))
        images = unpack(document, args.source_commit)
    except (json.JSONDecodeError, TypeError, ValueError) as error:
        raise SystemExit(f"invalid candidate manifest: {error}") from error
    lines = [f"{name}_digest={digest}\n" for name, digest in images.items()]
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        with args.output.open("a", encoding="utf-8") as handle:
            handle.writelines(lines)
    else:
        sys.stdout.writelines(lines)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
