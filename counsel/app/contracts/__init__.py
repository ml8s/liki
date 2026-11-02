"""Counsel MCP tool contracts.

Every counsel tool's published schema, argument validation, result validation and
serialization derive from the Pydantic v2 models in this package: the models are
the single source of truth, and no JSON manifest is read or written at runtime.
The schemas a client sees come from the MCP server's `tools/list` response.
"""

from app.contracts.divination import DIVINATION_TOOLS
from app.contracts.naming import NAMING_TOOLS
from app.contracts.natal import NATAL_TOOLS
from app.version import runtime_version

VERSION = runtime_version()

COUNSEL_TOOLS = NATAL_TOOLS + NAMING_TOOLS + DIVINATION_TOOLS

#: Counsel tool groups share their name with the skill package that documents them.
TOOL_GROUPS = {
    "natal": NATAL_TOOLS,
    "naming": NAMING_TOOLS,
    "divination": DIVINATION_TOOLS,
}




__all__ = [
    "COUNSEL_TOOLS",
    "DIVINATION_TOOLS",
    "NAMING_TOOLS",
    "NATAL_TOOLS",
    "TOOL_GROUPS",
    "VERSION",
]
