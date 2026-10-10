#!/usr/bin/env python3
"""Read and validate the liki-agents runtime base pin.

`contracts/liki-agents.base` 锁定 liki 依赖的 liki-agents 发布版本
（version + immutable digest）。测试与部署共用该 pin，不做与 liki 版本的
同名对齐。CI 装配与本地 `make agents-validate` 都从这里取版本。

用法:
    python3 scripts/liki_agents_pin.py --version   # 打印 pin 版本
    python3 scripts/liki_agents_pin.py --digest    # 打印 pin digest
    python3 scripts/liki_agents_pin.py --ref       # 打印 image:version@digest
    python3 scripts/liki_agents_pin.py --check      # 校验（失败返回非零）
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
PIN_PATH = ROOT / "contracts" / "liki-agents.base"
DEFAULT_REGISTRY = "ghcr.io/ml8s"

_CALVER_RE = re.compile(r"^\d{4}\.\d{2}\.\d{2}\.\d+$")
_DIGEST_RE = re.compile(r"^sha256:[0-9a-f]{64}$")


def read_pin() -> dict[str, str]:
    values: dict[str, str] = {}
    for raw in PIN_PATH.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or ":" not in line:
            continue
        key, value = line.split(":", 1)
        values[key.strip()] = value.strip()
    return values


def validate(values: dict[str, str]) -> list[str]:
    errors: list[str] = []
    version = values.get("version", "")
    digest = values.get("digest", "")
    if not _CALVER_RE.fullmatch(version):
        errors.append(f"liki-agents.base version 非法: {version!r}")
    if not _DIGEST_RE.fullmatch(digest):
        errors.append(f"liki-agents.base digest 非法: {digest!r}")
    return errors


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--registry", default=DEFAULT_REGISTRY)
    group = parser.add_mutually_exclusive_group()
    group.add_argument("--version", action="store_true")
    group.add_argument("--digest", action="store_true")
    group.add_argument("--ref", action="store_true")
    group.add_argument("--check", action="store_true")
    args = parser.parse_args()

    if not PIN_PATH.is_file():
        print(f"missing pin: {PIN_PATH.relative_to(ROOT)}", file=sys.stderr)
        return 1
    values = read_pin()

    if args.check:
        errors = validate(values)
        for line in errors:
            print(f"[liki-agents-pin] ✗ {line}", file=sys.stderr)
        if errors:
            return 1
        print(
            f"[liki-agents-pin] ✓ {values['version']} @ {values['digest']}"
        )
        return 0

    errors = validate(values)
    if errors:
        for line in errors:
            print(f"[liki-agents-pin] ✗ {line}", file=sys.stderr)
        return 1

    if args.digest:
        print(values["digest"])
    elif args.ref:
        print(f"{args.registry}/liki-agents:{values['version']}@{values['digest']}")
    else:  # --version（默认）
        print(values["version"])
    return 0


if __name__ == "__main__":
    sys.exit(main())
