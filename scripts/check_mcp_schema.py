#!/usr/bin/env python3
"""Validate static MCP tool schemas and their cross-repository catalog."""
from __future__ import annotations

import json
import re
import sys
from pathlib import Path

import jsonschema

ROOT = Path(__file__).resolve().parents[1]
CATALOG_PATH = ROOT / "tools/mcp-tool-catalog.json"
COUNSEL_MANIFESTS = (
    ROOT / "counsel/app/natal/tools/counsel-tools.json",
    ROOT / "counsel/app/naming/tools/skill-tools.json",
    ROOT / "counsel/app/divination/tools/skill-tools.json",
)
SCHEMA_BUDGETS = {
    ROOT / "counsel/app/natal/tools/counsel-tools.json": 32 * 1024,
    ROOT / "counsel/app/naming/tools/skill-tools.json": 32 * 1024,
    ROOT / "counsel/app/divination/tools/skill-tools.json": 128 * 1024,
    CATALOG_PATH: 16 * 1024,
}
ENGINE_TOOL_FILES = (
    ROOT / "engine/internal/agent/tools_bazi.go",
    ROOT / "engine/internal/agent/tools_ziwei.go",
    ROOT / "engine/internal/agent/tools_other.go",
    ROOT / "engine/internal/mcpcontract/time.go",
    ROOT / "engine/internal/mcpcontract/huangli.go",
    ROOT / "engine/internal/mcpcontract/tianwen.go",
    ROOT / "engine/internal/mcpcontract/city.go",
    ROOT / "engine/internal/mcpcontract/bazhai.go",
    ROOT / "engine/internal/mcpcontract/xuankong.go",
    ROOT / "engine/internal/mcpcontract/liuyao.go",
)


def _tool_functions(path: Path) -> list[dict]:
    doc = json.loads(path.read_text(encoding="utf-8"))
    return [item["function"] for item in doc["tools"]]


def _check_schema(schema: dict, *, where: str, errors: list[str], is_output: bool = False) -> None:
    try:
        validator = jsonschema.validators.validator_for(schema)
        validator.check_schema(schema)
    except Exception as exc:
        errors.append(f"{where}: invalid JSON Schema: {exc}")
        return

    if is_output and schema.get("type") == "array":
        pass
    elif schema.get("type") != "object":
        # Nullable scalar/object unions and top-level arrays are allowed only for outputs.
        if not (is_output and ("anyOf" in schema or "items" in schema)):
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
        if "enum" in prop and not prop.get("description"):
            errors.append(f"{where}: enum field {name!r} has no description")
        nested = prop.get("items")
        if isinstance(nested, dict) and nested.get("type") == "object":
            _check_schema(nested, where=f"{where}.items", errors=errors, is_output=True)


def main() -> int:
    errors: list[str] = []

    try:
        catalog = json.loads(CATALOG_PATH.read_text(encoding="utf-8"))
        catalog_tools = {
            server: set(definition.get("tools", []))
            for server, definition in catalog.get("servers", {}).items()
        }
        if set(catalog_tools) != {"engine", "counsel"}:
            errors.append("catalog: servers must be exactly engine and counsel")
    except Exception as exc:
        errors.append(f"catalog: {exc}")
        catalog_tools = {"engine": set(), "counsel": set()}

    counsel_tools: dict[str, dict] = {}
    for path in COUNSEL_MANIFESTS:
        try:
            functions = _tool_functions(path)
        except Exception as exc:
            errors.append(f"{path.relative_to(ROOT)}: {exc}")
            continue
        for fn in functions:
            name = fn.get("name")
            where = f"{path.relative_to(ROOT)}:{name}"
            if not name:
                errors.append(f"{where}: missing tool name")
                continue
            if name in counsel_tools:
                errors.append(f"{where}: duplicate tool")
            counsel_tools[name] = fn
            if not fn.get("description"):
                errors.append(f"{where}: missing description")

            params = fn.get("parameters")
            if not isinstance(params, dict):
                errors.append(f"{where}: missing parameters schema")
                continue
            _check_schema(params, where=f"{where}:parameters", errors=errors)

            output = fn.get("output_schema")
            if not isinstance(output, dict):
                errors.append(f"{where}: missing output_schema")
            else:
                _check_schema(
                    output,
                    where=f"{where}:output_schema",
                    errors=errors,
                    is_output=True,
                )

    engine_names: set[str] = set()
    for path in ENGINE_TOOL_FILES:
        text = path.read_text(encoding="utf-8")
        for match in re.finditer(r'Name:\s*"([a-z]+(?:\.[a-z_]+)+)"', text):
            engine_names.add(match.group(1).replace(".", "_"))
    if not engine_names:
        errors.append("engine: no tool declarations found")
    missing_engine = engine_names - catalog_tools.get("engine", set())
    extra_engine = catalog_tools.get("engine", set()) - engine_names
    if missing_engine:
        errors.append(f"catalog.engine missing: {sorted(missing_engine)}")
    if extra_engine:
        errors.append(f"catalog.engine unknown: {sorted(extra_engine)}")

    missing_counsel = set(counsel_tools) - catalog_tools.get("counsel", set())
    extra_counsel = catalog_tools.get("counsel", set()) - set(counsel_tools)
    if missing_counsel:
        errors.append(f"catalog.counsel missing: {sorted(missing_counsel)}")
    if extra_counsel:
        errors.append(f"catalog.counsel unknown: {sorted(extra_counsel)}")

    for path, budget in SCHEMA_BUDGETS.items():
        size = path.stat().st_size
        if size > budget:
            errors.append(
                f"{path.relative_to(ROOT)}: schema context budget exceeded "
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

    if errors:
        for error in errors:
            print(f"✗ {error}", file=sys.stderr)
        return 1

    counts = ", ".join(f"{server}={len(tools)}" for server, tools in sorted(catalog_tools.items()))
    sizes = ", ".join(
        f"{path.relative_to(ROOT)}={path.stat().st_size / 1024:.1f}KiB"
        for path in sorted(SCHEMA_BUDGETS)
    )
    print(f"✓ MCP tool schemas ({counts}; counsel output schemas validated)")
    print(f"✓ MCP schema context budget ({sizes})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
