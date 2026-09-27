"""校验生成的 AgentDeployment 工件对照契约 schema。

契约 pin 在 contracts/agent-definition.version（version + digest）。
校验流程：
  1. 从 pin 文件读取期望的 schema version 与 digest。
  2. 解析 contracts 下缓存的 schema（或从参数指定的 schema 文件）。
  3. 用 jsonschema 库校验 dist/agents/<profile>/deployment.json。
  4. 核对 schema digest 与 pin 一致（schema 本身未被篡改/漂移）。

用法:
    python3 scripts/check_deployment_schema.py --profile experts
    python3 scripts/check_deployment_schema.py --schema <path> --profile experts
"""
from __future__ import annotations

import argparse
import hashlib
import json
import sys
from pathlib import Path

import jsonschema

_ROOT = Path(__file__).resolve().parent.parent
PIN_FILE = _ROOT / "contracts" / "agent-definition.version"
SCHEMA_CACHE = _ROOT / "contracts" / "agent-definition.schema.json"


def load_pin() -> dict:
    if not PIN_FILE.exists():
        sys.exit(f"contract pin not found: {PIN_FILE}")
    pin = {}
    for line in PIN_FILE.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if ":" not in line:
            continue
        key, _, value = line.partition(":")
        pin[key.strip()] = value.strip()
    if "version" not in pin or "digest" not in pin:
        sys.exit(f"contract pin is incomplete: {PIN_FILE}")
    return pin


def resolve_schema(schema_path: str | None) -> tuple[dict, bytes]:
    path = Path(schema_path) if schema_path else SCHEMA_CACHE
    if not path.exists():
        sys.exit(
            f"schema not found: {path}. 请从 GHCR OCI artifact 拉取 "
            f"ghcr.io/ml8s/liki-contracts 或用 --schema 指定"
        )
    raw = path.read_bytes()
    return json.loads(raw), raw


def digest_of(raw: bytes) -> str:
    return "sha256:" + hashlib.sha256(raw).hexdigest()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--profile", required=True)
    parser.add_argument("--schema", default=None, help="schema JSON 路径（默认 contracts/）")
    parser.add_argument(
        "--out", default="dist/agents", help="生成产物目录（默认 dist/agents）"
    )
    args = parser.parse_args()

    pin = load_pin()
    schema, schema_raw = resolve_schema(args.schema)

    actual_digest = digest_of(schema_raw)
    if actual_digest != pin["digest"]:
        sys.exit(
            f"schema digest mismatch: pin={pin['digest']} actual={actual_digest} "
            f"(schema 与 contracts/agent-definition.version 不一致)"
        )
    print(f"schema ok version={pin['version']} digest={pin['digest']}")

    manifest = Path(args.out) / args.profile / "deployment.json"
    if not manifest.exists():
        sys.exit(f"deployment not found: {manifest}（先运行 generate_deployment.py）")
    document = json.loads(manifest.read_text(encoding="utf-8"))

    resolver = jsonschema.RefResolver.from_schema(schema)
    jsonschema.validate(document, schema, resolver=resolver)
    print(f"deployment ok {manifest}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
