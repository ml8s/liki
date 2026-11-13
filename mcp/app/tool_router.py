"""Upstream tool discovery and exposure policy."""

from __future__ import annotations

import logging
from typing import TYPE_CHECKING, Any

from app.protocol import ProtocolError, jsonrpc_error, jsonrpc_result
from app.upstream import UpstreamError, upstream_session

if TYPE_CHECKING:
    from app.config import GatewayConfig

logger = logging.getLogger(__name__)

# Retired names intentionally have no compatibility aliases. A structured
# migration hint keeps old clients from interpreting a generic unknown-tool
# error as a runtime outage.
RETIRED_TOOLS = {
    "compute_factors": "removed; factors are internal; use natal_query",
    "liuyao_snapshot": "renamed; use liuyao_ask",
    "period_query": "renamed; use natal_liunian_query",
    "qimen_snapshot": "renamed; use qimen_ask",
    "tianwen_resolve": "renamed; use subject_candidates",
}


async def upstream_tools_list(
    config: GatewayConfig,
    upstream_name: str,
) -> list[dict[str, Any]]:
    """Return every paginated tool descriptor exposed by one upstream."""
    upstream_config = getattr(config, upstream_name)
    cursor = None
    tools: list[dict[str, Any]] = []

    async with upstream_session(
        upstream_config,
        client_name=config.server_name,
        client_version=config.server_version,
    ) as session:
        while True:
            result = await session.list_tools(params={"cursor": cursor} if cursor else None)
            for tool in result.tools:
                tools.append(tool.model_dump(by_alias=True, mode="json", exclude_none=True))

            cursor = result.next_cursor
            if cursor is None:
                break

    return tools


async def load_visible_tools(config: GatewayConfig) -> list[dict[str, Any]]:
    engine_tools = await upstream_tools_list(config, "engine")
    counsel_tools = await upstream_tools_list(config, "counsel")
    tools_by_name: dict[str, dict[str, Any]] = {}

    for upstream_name, tools in (
        ("engine", engine_tools),
        ("counsel", counsel_tools),
    ):
        for tool in tools:
            name = tool["name"]
            previous = tools_by_name.get(name)
            if previous is not None and previous != tool:
                raise UpstreamError(f"tool name conflict across upstreams: {name}")
            tools_by_name[name] = tool

    visible: list[dict[str, Any]] = []
    for name in config.exposed_tools:
        if name in config.forbidden_tools:
            raise UpstreamError(f"forbidden tool configured for LIKI exposure: {name}")
        if name not in tools_by_name:
            raise UpstreamError(f"exposed upstream tool is unavailable: {name}")
        visible.append(tools_by_name[name])
    return visible


async def handle_tools_list(config: GatewayConfig, request_id: Any) -> Any:
    try:
        tools = await load_visible_tools(config)
    except UpstreamError as error:
        return jsonrpc_error(request_id, -32000, str(error))
    except Exception as error:
        logger.exception("LIKI tools/list upstream failed")
        return jsonrpc_error(request_id, -32000, f"LIKI upstream unavailable: {error}")
    return jsonrpc_result(request_id, {"tools": tools})


def route_tool(config: GatewayConfig, tool_name: str) -> str:
    if tool_name in config.forbidden_tools:
        raise ProtocolError(-32003, f"tool is forbidden on LIKI: {tool_name}", tool=tool_name)
    upstream_name = config.exposed_tools.get(tool_name)
    if upstream_name is None:
        migration = RETIRED_TOOLS.get(tool_name)
        detail = f"; {migration}" if migration else ""
        raise ProtocolError(-32602, f"unknown tool: {tool_name}{detail}", tool=tool_name)
    return upstream_name


async def handle_tools_call(config: GatewayConfig, request_id: Any, params: dict[str, Any]) -> Any:
    tool_name = params.get("name")
    if not isinstance(tool_name, str) or not tool_name:
        return jsonrpc_error(request_id, -32602, "tools/call requires a tool name")

    try:
        upstream_name = route_tool(config, tool_name)
    except ProtocolError as error:
        return jsonrpc_error(request_id, error.code, error.message)

    upstream_config = getattr(config, upstream_name)
    try:
        async with upstream_session(
            upstream_config,
            client_name=config.server_name,
            client_version=config.server_version,
        ) as session:
            result = await session.call_tool(
                tool_name,
                arguments=params.get("arguments") or {},
            )
    except UpstreamError as error:
        # Fail closed. In particular, casting=auto must never be retried.
        return jsonrpc_error(request_id, -32000, str(error))
    except Exception:
        logger.exception("LIKI tools/call upstream failed")
        raise

    return jsonrpc_result(request_id, result.model_dump(by_alias=True, mode="json"))
