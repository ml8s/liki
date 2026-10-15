"""Keep the MCP tool list aligned with producer sources."""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
CATALOG = ROOT / "tools" / "mcp-tool-catalog.json"
sys.path.insert(0, str(ROOT / "scripts"))

from mcp_engine_tools import engine_tool_names  # noqa: E402 - needs the path above


def test_catalog_matches_registered_engine_tools():
    # The typed mcpcontract registry is the only engine tool source: the
    # legacy raw-JSON registry has been retired.
    catalog = json.loads(CATALOG.read_text(encoding="utf-8"))
    assert set(catalog) == {"schema_version", "servers"}
    assert set(catalog["servers"]) == {"engine", "counsel"}
    assert sorted(catalog["servers"]["engine"]["tools"]) == sorted(engine_tool_names())


def test_retired_engine_raw_json_registry_stays_gone():
    """The typed registry replaced the raw-JSON sources; they must not return."""
    removed = (
        "engine/internal/agent/rpc_registry.go",
        "engine/internal/agent/tools_bazi.go",
        "engine/internal/agent/tools_ziwei.go",
        "engine/internal/agent/tools_other.go",
        "engine/internal/agent/types.go",
        "engine/internal/agent/qimen_schema.json",
        "engine/internal/agent/ziwei_fullchart_handler.go",
        "engine/cmd/mcpdtogen",
    )
    for path in removed:
        assert not (ROOT / path).exists(), f"{path} still exists"
    for name in ("bazi_wire.go", "ziwei_wire.go", "qimen_wire.go", "wire_enums.go"):
        assert (ROOT / "engine/internal/mcpcontract" / name).is_file(), name


def test_catalog_uses_unique_tool_names():
    catalog = json.loads(CATALOG.read_text(encoding="utf-8"))
    for server, definition in catalog["servers"].items():
        tools = definition["tools"]
        assert tools and all(re.fullmatch(r"[a-z][a-z0-9_]{0,127}", tool) for tool in tools)
        assert len(tools) == len(set(tools)), server
