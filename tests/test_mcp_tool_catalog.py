"""Keep the MCP tool list aligned with producer sources."""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
CATALOG = ROOT / "tools" / "mcp-tool-catalog.json"


def test_catalog_matches_registered_engine_rpc_methods():
    source = "\n".join(
        path.read_text(encoding="utf-8")
        for path in (ROOT / "engine" / "internal" / "agent").glob("tools_*.go")
    )
    methods = sorted(
        name.replace(".", "_")
        for name in re.findall(r'Name:\s*"([a-z]+(?:\.[a-z_]+)+)"', source)
    )
    catalog = json.loads(CATALOG.read_text(encoding="utf-8"))
    assert set(catalog) == {"schema_version", "servers"}
    assert set(catalog["servers"]) == {"engine", "counsel"}
    assert sorted(catalog["servers"]["engine"]["tools"]) == methods


def test_catalog_uses_unique_tool_names():
    catalog = json.loads(CATALOG.read_text(encoding="utf-8"))
    for server, definition in catalog["servers"].items():
        tools = definition["tools"]
        assert tools and all(re.fullmatch(r"[a-z][a-z0-9_]{0,127}", tool) for tool in tools)
        assert len(tools) == len(set(tools)), server
