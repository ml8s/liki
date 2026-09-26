#!/usr/bin/env python3
"""Smoke-test engine-mcp's internal root and domain endpoints."""
from __future__ import annotations

import argparse
import json
import sys
import urllib.request

PROTOCOL_VERSION = "2026-07-28"


def post(endpoint: str, method: str, tool_name: str | None = None, params: dict | None = None) -> dict:
    params = dict(params or {})
    params["_meta"] = {
        "io.modelcontextprotocol/protocolVersion": PROTOCOL_VERSION,
        "io.modelcontextprotocol/clientCapabilities": {},
    }
    body = json.dumps(
        {"jsonrpc": "2.0", "id": 1, "method": method, "params": params}
    ).encode("utf-8")
    headers = {
        "Content-Type": "application/json",
        "MCP-Protocol-Version": PROTOCOL_VERSION,
        "Mcp-Method": method,
        "Accept": "application/json, text/event-stream",
    }
    if tool_name:
        headers["Mcp-Name"] = tool_name
    request = urllib.request.Request(endpoint, data=body, headers=headers, method="POST")
    with urllib.request.urlopen(request, timeout=10) as response:
        raw = response.read().decode("utf-8")
    document = json.loads(raw) if raw.lstrip().startswith("{") else None
    if document is None:
        for line in raw.splitlines():
            if line.startswith("data: "):
                document = json.loads(line[6:])
                break
    if document is None:
        raise RuntimeError(f"unparseable MCP response from {endpoint}: {raw[:200]!r}")
    if "error" in document:
        raise RuntimeError(f"{method} failed on {endpoint}: {document['error']}")
    return document.get("result", {})


def tool_names(endpoint: str) -> set[str]:
    result = post(endpoint, "tools/list")
    return {item["name"] for item in result.get("tools", [])}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("base_url", help="engine origin, for example http://127.0.0.1:8081")
    args = parser.parse_args()
    base = args.base_url.rstrip("/")

    expectations = {
        "/mcp": {"bazi_chart", "ziwei_chart", "huangli_days", "bazhai_chart"},
        "/mcp/bazi": {"bazi_chart", "bazi_fullchart"},
        "/mcp/huangli": {"huangli_days"},
        "/mcp/fengshui": {"bazhai_chart", "xuankong_chart", "time_now"},
        "/mcp/aux": {"time_now", "tianwen_time", "city_coords"},
    }
    for path, required in expectations.items():
        endpoint = base + path
        info = post(endpoint, "server/discover")
        meta = info.get("_meta", {})
        server_info = meta.get("io.modelcontextprotocol/serverInfo", {})
        if not server_info.get("version"):
            raise RuntimeError(f"{endpoint}: serverInfo.version is missing")
        names = tool_names(endpoint)
        missing = required - names
        if missing:
            raise RuntimeError(f"{endpoint}: missing tools: {sorted(missing)}")
        print(f"✓ {path}: {len(names)} tools")

    root = base + "/mcp"
    result = post(root, "tools/call", "time_now", {"name": "time_now", "arguments": {}})
    content = result.get("content") or []
    text = content[0].get("text", "") if content else ""
    payload = json.loads(text)
    if not payload.get("cst"):
        raise RuntimeError("time_now returned no CST timestamp")
    print("✓ tools/call time_now")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as error:
        print(f"MCP smoke failed: {error}", file=sys.stderr)
        raise SystemExit(1)
