#!/usr/bin/env python3
"""Validate MCP tool schemas and generate their cross-repository catalog.

The counsel schemas are projected from the Pydantic contracts in
``counsel/app/contracts`` (the single source of truth); the engine names are read
from the typed Go registry. Nothing here validates a duplicated artifact --
``tools/mcp-tool-catalog.json`` is *generated* here, never hand-edited:

    python3 scripts/check_mcp_schema.py           # validate + drift check (read-only)
    python3 scripts/check_mcp_schema.py --emit    # regenerate the catalog
"""

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Any

import jsonschema
from mcp_engine_tools import engine_tool_names

ROOT = Path(__file__).resolve().parents[1]
CATALOG_PATH = ROOT / "tools/mcp-tool-catalog.json"
#: Context budget per counsel group, measured on the compact JSON a client pays for.
#: natal was raised from 24 KiB after 2026.10.22.1 strengthened contracts (closed
#: schemas, typed unions, self-describing fields) grew the group from 8 KiB at the
#: 24 KiB gate's introduction to ~51 KiB; 64 KiB keeps the guardrail functional.
COUNSEL_BUDGETS = {"natal": 64 * 1024, "naming": 24 * 1024, "divination": 64 * 1024}
COUNSEL_GROUPS = tuple(COUNSEL_BUDGETS)
SCHEMA_BUDGETS = {CATALOG_PATH: 16 * 1024}


def _counsel_groups(errors: list[str]) -> list[tuple[str, Any]]:
    """(group, contract) pairs from the Pydantic contract registry."""
    sys.path.insert(0, str(ROOT / "counsel"))
    try:
        from app.contracts import TOOL_GROUPS  # noqa: PLC0415 - deferred so the error is reported
    except Exception as exc:  # noqa: BLE001
        errors.append(
            f"counsel contracts: {exc} "
            "(install the hashed counsel dependencies: "
            "pip install --require-hashes -r counsel/requirements.lock, "
            "or run scripts/test-counsel.sh which bootstraps counsel/.venv)"
        )
        return []
    return [(group, tool) for group, tools in TOOL_GROUPS.items() for tool in tools]


def _catalog_document(engine_names: set[str], counsel_names: list[str]) -> dict[str, Any]:
    """The generated catalog: a sorted engine list and an ordered counsel list."""
    return {
        "servers": {
            "engine": {"tools": sorted(engine_names)},
            "counsel": {"tools": list(counsel_names)},
        },
    }


def _render(document: dict[str, Any]) -> str:
    return json.dumps(document, indent=2, ensure_ascii=False) + "\n"


def _check_schema(schema: dict, *, where: str, errors: list[str], is_output: bool = False) -> None:
    try:
        validator = jsonschema.validators.validator_for(schema)
        validator.check_schema(schema)
    except Exception as exc:
        errors.append(f"{where}: invalid JSON Schema: {exc}")
        return

    if "anyOf" in schema and "type" not in schema:
        if is_output:
            errors.append(
                f"{where}: output schema must be a top-level object "
                "(pre-2026 MCP wire models reject non-object outputSchema roots)"
            )
            return
        branches = schema.get("anyOf")
        if not isinstance(branches, list) or not branches:
            errors.append(f"{where}: anyOf must be a non-empty array")
            return
        if not is_output and not all(
            isinstance(branch, dict)
            and branch.get("type") == "object"
            and branch.get("additionalProperties") is False
            for branch in branches
        ):
            errors.append(f"{where}: input anyOf branches must be closed objects")
            return
        for index, branch in enumerate(branches):
            if isinstance(branch, dict) and branch.get("type") == "object":
                _check_schema(branch, where=f"{where}.anyOf[{index}]", errors=errors, is_output=is_output)
        return

    if schema.get("type") != "object":
        # MCP outputSchema must be type "object" at the root for every
        # supported protocol version; unions/arrays belong in fields.
        errors.append(f"{where}: schema type must be object")
        return

    properties = schema.get("properties", {})
    if not isinstance(properties, dict):
        errors.append(f"{where}: properties must be an object")
        return

    for name in schema.get("required", []):
        if name not in properties:
            errors.append(f"{where}: required field {name!r} has no property")

    if not is_output and schema.get("additionalProperties") is not False:
        errors.append(f"{where}: input schemas must set additionalProperties=false")

    if schema.get("type") == "object":
        for name in schema.get("required", []):
            prop = properties.get(name, {})
            if not prop.get("description"):
                errors.append(f"{where}: required output field {name!r} has no description")

    for name, prop in properties.items():
        if not isinstance(prop, dict):
            continue
        if name in {"schema_version", "contract_version"}:
            errors.append(f"{where}: wire contracts must not publish {name!r}")
        if "enum" in prop and not prop.get("description"):
            errors.append(f"{where}: enum field {name!r} has no description")
        nested = prop.get("items")
        if isinstance(nested, dict) and nested.get("type") == "object":
            _check_schema(nested, where=f"{where}.items", errors=errors, is_output=True)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Validate and generate the MCP tool catalog")
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument(
        "--check", action="store_true", help="read-only drift check (the default when no mode is given)"
    )
    mode.add_argument(
        "--emit", action="store_true", help="regenerate tools/mcp-tool-catalog.json"
    )
    args = parser.parse_args(argv)
    errors: list[str] = []
    catalog_errors: list[str] = []
    catalog_text: str | None = None

    try:
        catalog_text = CATALOG_PATH.read_text(encoding="utf-8")
        catalog = json.loads(catalog_text)
        catalog_tools = {
            server: set(definition.get("tools", []))
            for server, definition in catalog.get("servers", {}).items()
        }
        if set(catalog_tools) != {"engine", "counsel"}:
            catalog_errors.append("catalog: servers must be exactly engine and counsel")
        if "schema_version" in catalog:
            catalog_errors.append("catalog: generated tool names must not carry a contract schema_version")
    except Exception as exc:
        catalog_errors.append(f"catalog: {exc}")
        catalog_tools = {"engine": set(), "counsel": set()}

    counsel_tools: dict[str, Any] = {}
    counsel_bytes: dict[str, int] = {group: 0 for group in COUNSEL_GROUPS}
    for group, contract in _counsel_groups(errors):
        name = contract.name
        where = f"counsel:{name}"
        if name in counsel_tools:
            errors.append(f"{where}: duplicate tool")
        counsel_tools[name] = contract
        if not contract.description:
            errors.append(f"{where}: missing description")

        params = contract.input_schema
        if not isinstance(params, dict):
            errors.append(f"{where}: missing parameters schema")
            continue
        _check_schema(params, where=f"{where}:parameters", errors=errors)

        output = contract.output_schema
        if not isinstance(output, dict):
            errors.append(f"{where}: missing output_schema")
        else:
            _check_schema(
                output,
                where=f"{where}:output_schema",
                errors=errors,
                is_output=True,
            )
        counsel_bytes[group] += len(
            json.dumps(
                {
                    "name": name,
                    "description": contract.description,
                    "parameters": params,
                    "output_schema": output,
                },
                ensure_ascii=False,
            ).encode()
        )

    engine_names = engine_tool_names()
    if not engine_names:
        errors.append("engine: no tool declarations found")
    expected_catalog = _render(_catalog_document(engine_names, list(counsel_tools)))
    missing_engine = engine_names - catalog_tools.get("engine", set())
    extra_engine = catalog_tools.get("engine", set()) - engine_names
    if missing_engine:
        catalog_errors.append(f"catalog.engine missing: {sorted(missing_engine)}")
    if extra_engine:
        catalog_errors.append(f"catalog.engine unknown: {sorted(extra_engine)}")

    missing_counsel = set(counsel_tools) - catalog_tools.get("counsel", set())
    extra_counsel = catalog_tools.get("counsel", set()) - set(counsel_tools)
    if missing_counsel:
        catalog_errors.append(f"catalog.counsel missing: {sorted(missing_counsel)}")
    if extra_counsel:
        catalog_errors.append(f"catalog.counsel unknown: {sorted(extra_counsel)}")

    drift = missing_engine or extra_engine or missing_counsel or extra_counsel
    if not drift and catalog_text != expected_catalog:
        # Tool sets agree, so this is hand-edited ordering, formatting, or metadata.
        drift = True
    if drift:
        catalog_errors.append("catalog: regenerate with `python3 scripts/check_mcp_schema.py --emit`")

    for path, budget in SCHEMA_BUDGETS.items():
        size = path.stat().st_size
        if size > budget:
            errors.append(
                f"{path.relative_to(ROOT)}: schema context budget exceeded "
                f"({size} > {budget} bytes)"
            )

    for group, budget in COUNSEL_BUDGETS.items():
        size = counsel_bytes[group]
        if size > budget:
            errors.append(
                f"counsel/{group}: schema context budget exceeded "
                f"({size} > {budget} bytes)"
            )

    method_pattern = re.compile(
        r"`(?:"
        + "|".join(re.escape(name) for name in sorted(engine_names | set(counsel_tools)))
        + r")`"
    )
    for path in sorted((ROOT / "skills/liki").rglob("*.md")):
        for match in method_pattern.finditer(path.read_text(encoding="utf-8")):
            errors.append(
                f"{path.relative_to(ROOT)}: Skill documents must not hard-code MCP method "
                f"name {match.group()}; use semantic capability language and tools/list introspection"
            )

    if args.emit:
        if errors:
            # The inputs did not resolve; writing now would publish a truncated catalog.
            for error in errors:
                print(f"✗ {error}", file=sys.stderr)
            print("refusing to write the catalog while validation fails", file=sys.stderr)
            return 1
        CATALOG_PATH.write_text(expected_catalog, encoding="utf-8")
        note = " (replaced drifted catalog)" if catalog_errors else ""
        print(f"✓ wrote {CATALOG_PATH.relative_to(ROOT)}{note}")
        return 0

    if errors or catalog_errors:
        for error in [*errors, *catalog_errors]:
            print(f"✗ {error}", file=sys.stderr)
        return 1

    counts = ", ".join(f"{server}={len(tools)}" for server, tools in sorted(catalog_tools.items()))
    sizes = ", ".join(
        [f"{path.relative_to(ROOT)}={path.stat().st_size / 1024:.1f}KiB" for path in sorted(SCHEMA_BUDGETS)]
        + [f"counsel/{group}={counsel_bytes[group] / 1024:.1f}KiB" for group in COUNSEL_GROUPS]
    )
    print(f"✓ MCP tool schemas ({counts}; counsel output schemas validated)")
    print(f"✓ MCP schema context budget ({sizes})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
