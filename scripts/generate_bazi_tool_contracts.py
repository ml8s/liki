#!/usr/bin/env python3
"""Generate Bazi Python-tool result contracts in skill-tools.json.

The generated manifest is deterministic so `--check` can prevent hand-edited
contracts from drifting away from tool_contracts.py.
"""
from __future__ import annotations

import argparse
import importlib.util
import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "skills/liki/bazi/tools/skill-tools.json"
CONTRACTS = ROOT / "skills/liki/bazi/tools/tool_contracts.py"
TOOLS = CONTRACTS.parent

if str(TOOLS) not in sys.path:
    sys.path.insert(0, str(TOOLS))


def _load_contracts():
    spec = importlib.util.spec_from_file_location("bazi_tool_contracts", CONTRACTS)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


def generate(manifest: dict) -> dict:
    contracts = _load_contracts()
    by_name = {tool["function"]["name"]: tool["function"] for tool in manifest["tools"]}
    expected = {
        "city_coords": contracts.city_result_schema(),
        "full_paipan": contracts.full_paipan_surface_result_schema(),
        "query": contracts.query_result_schema(),
        "yearly_range": contracts.yearly_result_schema(),
        "calibrate": contracts.calibrate_result_schema(),
        "bond": contracts.bond_result_schema(),
    }
    if set(by_name) != set(expected):
        raise ValueError(f"tool contract mismatch: {sorted(by_name)} != {sorted(expected)}")
    for name, schema in expected.items():
        by_name[name]["result_schema"] = schema
    return manifest


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true", help="verify generated manifest has no drift")
    args = parser.parse_args()

    original = MANIFEST.read_text(encoding="utf-8")
    generated = generate(json.loads(original))
    rendered = json.dumps(generated, ensure_ascii=False, indent=2) + "\n"
    if args.check:
        if rendered != original:
            raise SystemExit("Bazi tool contracts are stale; run scripts/generate_bazi_tool_contracts.py")
        print("  ✓ Bazi tool contracts up to date")
        return 0

    MANIFEST.write_text(rendered, encoding="utf-8")
    print("✅ Bazi tool contracts generated")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
