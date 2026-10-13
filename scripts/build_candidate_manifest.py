#!/usr/bin/env python3
"""Generate the digest manifest for locally published candidate images."""
from __future__ import annotations

import argparse
import json
import re
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SHA256_RE = re.compile(r"^sha256:[0-9a-f]{64}$")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
TAG_RE = re.compile(r"^candidate-[0-9a-f]{7}-\d{14}$")
IMAGE_NAMES = ("engine", "counsel", "experts", "single_expert")


def source_commit(explicit: str | None) -> str:
    value = explicit or subprocess.check_output(
        ["git", "-C", str(ROOT), "rev-parse", "HEAD"], text=True
    ).strip()
    if not SHA_RE.fullmatch(value):
        raise SystemExit(f"invalid source commit: {value!r}")
    return value


def digest_from_metadata(path: Path) -> str:
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise SystemExit(f"invalid buildx metadata {path}: {error}") from error
    value = payload.get("containerimage.digest")
    if not isinstance(value, str) or not SHA256_RE.fullmatch(value):
        raise SystemExit(f"missing containerimage.digest in buildx metadata: {path}")
    return value


def manifest(
    source_commit: str,
    candidate_tag: str,
    metadata_dir: Path,
    repository: str = "ml8s/liki",
) -> dict[str, object]:
    if not SHA_RE.fullmatch(source_commit):
        raise SystemExit(f"invalid source commit: {source_commit!r}")
    if not TAG_RE.fullmatch(candidate_tag):
        raise SystemExit(f"invalid candidate tag: {candidate_tag!r}")
    if not metadata_dir.is_dir():
        raise SystemExit(f"missing metadata directory: {metadata_dir}")
    images = {
        name: digest_from_metadata(metadata_dir / f"{name}.json")
        for name in IMAGE_NAMES
    }
    return {
        "repository": repository,
        "source_commit": source_commit,
        "candidate_tag": candidate_tag,
        "images": images,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source-commit")
    parser.add_argument("--candidate-tag", required=True)
    parser.add_argument("--metadata-dir", type=Path, required=True)
    parser.add_argument("--output", type=Path, default=ROOT / "build/candidate/candidate-digests.json")
    args = parser.parse_args()
    document = manifest(
        source_commit(args.source_commit), args.candidate_tag, args.metadata_dir
    )
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(
        json.dumps(document, ensure_ascii=False, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"candidate digest manifest: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
