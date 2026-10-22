#!/usr/bin/env python3
"""Generate the canonical domain-language document from its registry."""


import argparse
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REGISTRY = ROOT / "tools/domain-language.json"
DOCUMENT = ROOT / "docs/DOMAIN_LANGUAGE.md"

HEADER = """# Liki canonical domain language

This document is generated from [`tools/domain-language.json`](../tools/domain-language.json).
Do not edit the tables by hand. One domain concept has one ID and one canonical
projection per surface. A legacy name is accepted only while it is declared in
that concept's compatibility list.
"""


def _aliases(concept: dict) -> str:
    forbidden = ", ".join(f"`{name}`" for name in concept["forbidden"])
    compatible = ", ".join(
        f"`{item['name']}`（{item['remove_in']}）"
        for item in concept["compatibility"]
    )
    parts = []
    if forbidden:
        parts.append(f"禁止：{forbidden}")
    if compatible:
        parts.append(f"兼容：{compatible}")
    return "<br>".join(parts).replace("|", "\\|") if parts else "—"


def _document(registry: dict) -> str:
    lines = [HEADER.rstrip(), "", "## Concepts", ""]
    lines.extend(
        [
            "| Concept ID | Status | Definition | Go | Python | JSON | 中文 | Aliases |",
            "| --- | --- | --- | --- | --- | --- | --- | --- |",
        ]
    )
    for concept in registry["concepts"]:
        surfaces = concept["surfaces"]
        lines.append(
            "| `{id}` | {status} | {definition} | `{go}` | `{python}` | `{json}` | {zh_cn} | {aliases} |".format(
                id=concept["id"],
                status=concept["status"],
                definition=concept["definition"].replace("|", "\\|"),
                go=surfaces["go"].replace("|", "\\|"),
                python=surfaces["python"].replace("|", "\\|"),
                json=surfaces["json"].replace("|", "\\|"),
                zh_cn=surfaces["zh_cn"].replace("|", "\\|"),
                aliases=_aliases(concept),
            )
        )
    lines.extend(
        [
            "",
            "## Rules",
            "",
            "1. The `id` is immutable and is never reused for a different concept.",
            "2. Code, contracts, comments, tests, and documentation must use the canonical projection for their surface.",
            "3. A compatibility name must be scoped, must state its reason, and must declare a removal version.",
            "4. New code may not introduce a forbidden alias or an unregistered synonym.",
            "5. Runtime and deployment names remain governed by [DOMAIN_NAMING.md](./DOMAIN_NAMING.md).",
            "",
        ]
    )
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="fail when the document differs")
    args = parser.parse_args()
    registry = json.loads(REGISTRY.read_text(encoding="utf-8"))
    expected = _document(registry)
    if args.check:
        actual = DOCUMENT.read_text(encoding="utf-8") if DOCUMENT.exists() else ""
        if actual != expected:
            print("domain language document differs from tools/domain-language.json", file=sys.stderr)
            return 1
        print(f"[sync-domain-language] ✓ {len(registry['concepts'])} concepts")
        return 0
    DOCUMENT.write_text(expected, encoding="utf-8")
    print(f"[sync-domain-language] wrote {DOCUMENT.relative_to(ROOT)} ({len(registry['concepts'])} concepts)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
