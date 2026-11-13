"""JSON-RPC and MCP protocol envelope rules for LIKI MCP."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any, Awaitable, Callable, TypeAlias

from app.config import LATEST_PROTOCOL_VERSION, SUPPORTED_PROTOCOL_VERSIONS
from starlette.responses import JSONResponse

if TYPE_CHECKING:
    from app.config import GatewayConfig


__all__ = [
    "GatewayRequestHandler",
    "ProtocolError",
    "handle_request",
    "jsonrpc_error",
    "jsonrpc_result",
    "method_not_found",
]


class ProtocolError(Exception):
    def __init__(self, code: int, message: str, *, tool: str | None = None):
        super().__init__(message)
        self.code = code
        self.message = message
        self.tool = tool


def jsonrpc_result(request_id: Any, result: Any) -> JSONResponse:
    return JSONResponse({"jsonrpc": "2.0", "id": request_id, "result": result})


def jsonrpc_error(request_id: Any, code: int, message: str) -> JSONResponse:
    return JSONResponse(
        {
            "jsonrpc": "2.0",
            "id": request_id if request_id is not None else None,
            "error": {"code": code, "message": message},
        }
    )


def method_not_found(request_id: Any, method: str) -> JSONResponse:
    """Answer an unsupported 2026 RPC with the transport's required status."""
    response = jsonrpc_error(request_id, -32601, f"method not found: {method}")
    response.status_code = 404
    return response


async def handle_request(
    config: GatewayConfig,
    request_id: Any,
    method: str,
    params: dict[str, Any],
    *,
    tools_handler: GatewayRequestHandler,
) -> JSONResponse:
    if method == "initialize":
        requested = params.get("protocolVersion")
        negotiated = requested if requested in SUPPORTED_PROTOCOL_VERSIONS else LATEST_PROTOCOL_VERSION
        return jsonrpc_result(
            request_id,
            {
                "protocolVersion": negotiated,
                "capabilities": {"tools": {"listChanged": False}},
                "serverInfo": {
                    "name": config.server_name,
                    "title": config.server_title,
                    "version": config.server_version,
                },
            },
        )
    if method == "ping":
        return jsonrpc_result(request_id, {})
    if method == "tools/list":
        return await tools_handler(config, request_id)
    if method == "tools/call":
        return await tools_handler(config, request_id, params)
    return method_not_found(request_id, method)


GatewayRequestHandler: TypeAlias = Callable[..., Awaitable[JSONResponse]]
