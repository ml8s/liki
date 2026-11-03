#!/usr/bin/env python3
"""Smoke-test engine-mcp's aggregate MCP endpoint."""

import argparse
import json
import os
import sys
import urllib.error
import urllib.request

PROTOCOL_VERSION = "2026-07-28"


def post(endpoint: str, method: str) -> dict:
    params = {
        "_meta": {
            "io.modelcontextprotocol/protocolVersion": PROTOCOL_VERSION,
            "io.modelcontextprotocol/clientCapabilities": {},
        }
    }
    body = json.dumps(
        {"jsonrpc": "2.0", "id": 1, "method": method, "params": params}
    ).encode("utf-8")
    request = urllib.request.Request(
        endpoint,
        data=body,
        headers={
            "Content-Type": "application/json",
            "MCP-Protocol-Version": PROTOCOL_VERSION,
            "Mcp-Method": method,
            "Accept": "application/json, text/event-stream",
        },
        method="POST",
    )
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


def call(endpoint: str, tool_name: str, arguments: dict) -> dict:
    params = {
        "name": tool_name,
        "arguments": arguments,
        "_meta": {
            "io.modelcontextprotocol/protocolVersion": PROTOCOL_VERSION,
            "io.modelcontextprotocol/clientCapabilities": {},
        },
    }
    body = json.dumps(
        {"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": params}
    ).encode("utf-8")
    request = urllib.request.Request(
        endpoint,
        data=body,
        headers={
            "Content-Type": "application/json",
            "MCP-Protocol-Version": PROTOCOL_VERSION,
            "Mcp-Method": "tools/call",
            "Mcp-Name": tool_name,
            "Accept": "application/json, text/event-stream",
        },
        method="POST",
    )
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
    if document.get("error") or document.get("result", {}).get("isError"):
        raise RuntimeError(f"{tool_name} failed: {document}")
    return document.get("result", {})


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "base_url",
        nargs="?",
        default=os.environ.get("LIKI_ENGINE_MCP_URL", ""),
        help="engine MCP endpoint, for example http://127.0.0.1:8081/mcp",
    )
    args = parser.parse_args()
    if not args.base_url:
        parser.error("base_url required or LIKI_ENGINE_MCP_URL must be set")
    endpoint = args.base_url.rstrip("/")

    info = post(endpoint, "server/discover")
    server_info = info.get("_meta", {}).get(
        "io.modelcontextprotocol/serverInfo", {}
    )
    if not server_info.get("version"):
        raise RuntimeError("serverInfo.version is missing")

    listed = post(endpoint, "tools/list").get("tools", [])
    names = {tool["name"]: tool for tool in listed}
    required = {
        "bazhai_chart", "bazhai_layout", "bazi_bond", "bazi_chart",
        "bazi_liupan", "city_info", "huangli_days", "liuyao_chart",
        "liuyao_qigua", "qimen_chart", "subject_candidates",
        "time_now", "xuankong_chart", "xuankong_liunian", "ziwei_bond",
        "ziwei_chart", "ziwei_liupan",
    }
    missing = required - names.keys()
    if missing:
        raise RuntimeError(f"aggregate endpoint missing tools: {sorted(missing)}")
    if len(names) != 17:
        raise RuntimeError(f"aggregate endpoint exposed {len(names)} tools, want 17")

    qigua = names["liuyao_qigua"].get("annotations") or {}
    if qigua.get("idempotentHint") is not False:
        raise RuntimeError("liuyao_qigua must advertise idempotentHint=false")
    if names["liuyao_chart"].get("annotations", {}).get("idempotentHint") is not True:
        raise RuntimeError("liuyao_chart must advertise idempotentHint=true")

    aggregate_root = endpoint.removesuffix("/mcp")
    for path in ("/mcp/bazi", "/mcp/qimen", "/mcp/aux"):
        try:
            post(aggregate_root + path, "server/discover")
        except urllib.error.HTTPError as error:
            if error.code != 404:
                raise
        else:
            raise RuntimeError(f"retired domain endpoint still exists: {path}")

    result = call(endpoint, "time_now", {"timezone": "Asia/Shanghai"})
    payload = result.get("structuredContent")
    if not payload or not payload.get("local_standard_time"):
        raise RuntimeError("time_now returned no canonical standard_time")
    print(f"✓ /mcp: {len(names)} tools, aggregate-only, event hints valid")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as error:
        print(f"MCP smoke failed: {error}", file=sys.stderr)
        raise SystemExit(1)
