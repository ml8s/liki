"""Counsel MCP tool contracts.

Every counsel tool's published schema, argument validation, result validation and
serialization derive from the Pydantic v2 models in this package: the models are
the single source of truth, and no JSON manifest is read or written at runtime.
The schemas a client sees come from the MCP server's `tools/list` response.
"""


from pathlib import Path
from typing import Any

from app.contracts.divination import DIVINATION_TOOLS
from app.contracts.naming import NAMING_TOOLS
from app.contracts.natal import NATAL_TOOLS

VERSION_PATH = Path(__file__).resolve().parents[1] / "VERSION.txt"
VERSION = VERSION_PATH.read_text(encoding="utf-8").strip()

COUNSEL_TOOLS = NATAL_TOOLS + NAMING_TOOLS + DIVINATION_TOOLS

#: Counsel tool groups share their name with the skill package that documents them.
TOOL_GROUPS = {
    "natal": NATAL_TOOLS,
    "naming": NAMING_TOOLS,
    "divination": DIVINATION_TOOLS,
}


def _collect_schema_names(node: Any, names: set[str]) -> None:
    """Collect every property name and enum value a projected schema exposes.

    The Go document check walks the engine schemas the same way
    (``collectToolVocabulary`` in skill_docs_sync_test.go); keep both walks in
    step so a field resolves identically on either side of the boundary.
    """
    if isinstance(node, dict):
        properties = node.get("properties")
        if isinstance(properties, dict):
            names.update(properties)
            for child in properties.values():
                _collect_schema_names(child, names)
        const = node.get("const")
        if isinstance(const, str):
            names.add(const)
        enum = node.get("enum")
        if isinstance(enum, list):
            names.update(value for value in enum if isinstance(value, str))
        for key, value in node.items():
            if key != "properties":
                _collect_schema_names(value, names)
    elif isinstance(node, list):
        for value in node:
            _collect_schema_names(value, names)


def tool_vocabulary() -> dict[str, list[str]]:
    """Per-group vocabulary a skill documentation check can resolve against.

    Each group lists its tool names plus every property name and enum value the
    tools publish in their input and output schemas. Skill documents reference
    fields by backticked name; a checker resolves those references against this
    vocabulary, so the documentation needs no separately maintained index.
    """
    vocabulary: dict[str, list[str]] = {}
    for group, tools in TOOL_GROUPS.items():
        names: set[str] = set()
        for tool in tools:
            names.add(tool.name)
            _collect_schema_names(tool.input_schema, names)
            output_schema = tool.output_schema
            if output_schema is not None:
                _collect_schema_names(output_schema, names)
        vocabulary[group] = sorted(names)
    return vocabulary


__all__ = [
    "COUNSEL_TOOLS",
    "DIVINATION_TOOLS",
    "NAMING_TOOLS",
    "NATAL_TOOLS",
    "TOOL_GROUPS",
    "VERSION",
    "VERSION_PATH",
    "tool_vocabulary",
]
