"""LIKI-MCP: a thin profile-free MCP passthrough gateway.

The gateway owns only one product decision: which upstream tool names are
exposed. It does not rewrite schemas, arguments, results, errors, or names.
"""

from __future__ import annotations

from dataclasses import dataclass
import json
import logging
from typing import Any

import os

from app.config import GatewayConfig, load_config
# alipay-sdk-python 和 order_store 在 _get_payment_service 中懒加载：
# paywall disabled 时不需要 SDK。
from app.upstream import UpstreamError, upstream_session

logger = logging.getLogger(__name__)
from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import JSONResponse, Response
from starlette.routing import Route

SUPPORTED_PROTOCOL_VERSIONS = {
    "2025-03-26",
    "2025-06-18",
    "2025-11-25",
    "2026-07-28",
}
LATEST_PROTOCOL_VERSION = "2026-07-28"


class ProtocolError(Exception):
    def __init__(self, code: int, message: str, *, tool: str | None = None):
        super().__init__(message)
        self.code = code
        self.message = message
        self.tool = tool


@dataclass
class UpstreamRoutes:
    engine: str
    counsel: str


def _jsonrpc_result(request_id: Any, result: Any) -> JSONResponse:
    return JSONResponse({"jsonrpc": "2.0", "id": request_id, "result": result})


def _jsonrpc_error(request_id: Any, code: int, message: str) -> JSONResponse:
    return JSONResponse({
        "jsonrpc": "2.0",
        "id": request_id if request_id is not None else None,
        "error": {"code": code, "message": message},
    })


async def _upstream_tools_list(
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
            result = await session.list_tools(
                params={"cursor": cursor} if cursor else None
            )
            for tool in result.tools:
                tools.append(tool.model_dump(by_alias=True, mode="json"))

            cursor = result.next_cursor
            if cursor is None:
                break

    return tools


async def _load_visible_tools(config: GatewayConfig) -> list[dict[str, Any]]:
    engine_tools = await _upstream_tools_list(config, "engine")
    counsel_tools = await _upstream_tools_list(config, "counsel")
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


async def _handle_tools_list(config: GatewayConfig, request_id: Any) -> JSONResponse:
    try:
        tools = await _load_visible_tools(config)
    except UpstreamError as error:
        return _jsonrpc_error(request_id, -32000, str(error))
    except Exception as error:
        logger.exception("LIKI tools/list upstream failed")
        return _jsonrpc_error(request_id, -32000, f"LIKI upstream unavailable: {error}")
    return _jsonrpc_result(request_id, {"tools": tools})


def _route_tool(config: GatewayConfig, tool_name: str) -> str:
    if tool_name in config.forbidden_tools:
        raise ProtocolError(-32003, f"tool is forbidden on LIKI: {tool_name}", tool=tool_name)
    upstream_name = config.exposed_tools.get(tool_name)
    if upstream_name is None:
        raise ProtocolError(-32602, f"unknown tool: {tool_name}", tool=tool_name)
    return upstream_name


async def _handle_tools_call(
    config: GatewayConfig,
    request_id: Any,
    params: dict[str, Any],
) -> JSONResponse:
    tool_name = params.get("name")
    if not isinstance(tool_name, str) or not tool_name:
        return _jsonrpc_error(request_id, -32602, "tools/call requires a tool name")

    try:
        upstream_name = _route_tool(config, tool_name)
    except ProtocolError as error:
        return _jsonrpc_error(request_id, error.code, error.message)

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
        return _jsonrpc_error(request_id, -32000, str(error))
    except Exception:
        import traceback
        traceback.print_exc()
        raise

    return _jsonrpc_result(request_id, result.model_dump(by_alias=True, mode="json"))


async def _handle_request(
    config: GatewayConfig,
    request_id: Any,
    method: str,
    params: dict[str, Any],
) -> JSONResponse:
    if method == "initialize":
        requested = params.get("protocolVersion")
        negotiated = requested if requested in SUPPORTED_PROTOCOL_VERSIONS else LATEST_PROTOCOL_VERSION
        return _jsonrpc_result(request_id, {
            "protocolVersion": negotiated,
            "capabilities": {"tools": {"listChanged": False}},
            "serverInfo": {
                "name": config.server_name,
                "title": config.server_title,
                "version": config.server_version,
            },
        })

    if method == "ping":
        return _jsonrpc_result(request_id, {})

    if method == "tools/list":
        return await _handle_tools_list(config, request_id)

    if method == "tools/call":
        return await _handle_tools_call(config, request_id, params)

    raise ProtocolError(-32601, f"method not found: {method}")



def _extract_tool_name(payload: dict) -> str | None:
    """从 MCP JSON-RPC tools/call 请求中提取工具名。"""
    params = payload.get("params")
    if isinstance(params, dict):
        name = params.get("name")
        return str(name) if name else None
    return None


_payment_service: PaymentService | None = None
_order_store: OrderStore | None = None


def _get_payment_service(config: GatewayConfig) -> PaymentService:
    """懒初始化支付基础设施。"""
    global _payment_service, _order_store
    if _payment_service is None:
        from app.order_store import OrderStore
        from app.payment import PaymentService

        db_path = os.environ.get("LIKI_AIPAY_DB_PATH", "/data/aipay.db")
        _order_store = OrderStore(db_path)
        _payment_service = PaymentService(config.aipay, _order_store)
    return _payment_service


async def mcp_endpoint(request: Request) -> Response:
    config = load_config()

    if request.method != "POST":
        return Response(status_code=405)

    try:
        payload = await request.json()
    except Exception:
        return JSONResponse({
            "jsonrpc": "2.0",
            "id": None,
            "error": {"code": -32700, "message": "Parse error"},
        }, status_code=400)

    # SkillPay 付费工具检查（application layer）
    if config.aipay.enabled:
        tool_name = _extract_tool_name(payload)
        if tool_name:
            proof = request.headers.get("Payment-Proof")
            payment_service = _get_payment_service(config)
            payment_result = payment_service.enforce_paywall(tool_name, proof)
            if payment_result is not None:
                return JSONResponse(
                    payment_result.body,
                    status_code=payment_result.status_code,
                    headers=payment_result.headers,
                )

    messages = payload if isinstance(payload, list) else [payload]
    responses: list[dict[str, Any]] = []
    has_request = False

    for message in messages:
        if not isinstance(message, dict):
            responses.append({
                "jsonrpc": "2.0",
                "id": None,
                "error": {"code": -32600, "message": "Invalid Request"},
            })
            continue

        method = message.get("method")
        request_id = message.get("id")
        params = message.get("params") or {}

        if not isinstance(method, str):
            if request_id is not None:
                responses.append({
                    "jsonrpc": "2.0",
                    "id": request_id,
                    "error": {"code": -32600, "message": "Invalid Request"},
                })
            continue

        if request_id is None:
            # Notifications have no response. LIKI exposes no server stream.
            continue

        has_request = True
        try:
            response = await _handle_request(config, request_id, method, params)
        except ProtocolError as error:
            response = _jsonrpc_error(request_id, error.code, error.message)
        except Exception:
            response = _jsonrpc_error(request_id, -32603, "Internal error")
        responses.append(response)

    if not has_request:
        return Response(status_code=202)
    if len(responses) == 1:
        return responses[0]
    return JSONResponse(responses)


async def healthz(_: Request) -> JSONResponse:
    return JSONResponse({"status": "ok"})


async def readyz(_: Request) -> JSONResponse:
    config = load_config()
    try:
        async with upstream_session(
            config.engine,
            client_name=config.server_name,
            client_version=config.server_version,
        ) as session:
            await session.ping()
        async with upstream_session(
            config.counsel,
            client_name=config.server_name,
            client_version=config.server_version,
        ) as session:
            await session.ping()
    except Exception as error:
        return JSONResponse({"status": "unavailable", "reason": str(error)}, status_code=503)
    return JSONResponse({"status": "ready"})


async def metrics(_: Request) -> Response:
    return Response(
        "# HELP liki_upstream_requests LIKI upstream requests.\n"
        "# TYPE liki_upstream_requests counter\n",
        media_type="text/plain; version=0.0.4",
    )


def create_app() -> Starlette:
    return Starlette(routes=[
        Route("/mcp", mcp_endpoint, methods=["GET", "POST", "DELETE"]),
        Route("/healthz", healthz, methods=["GET"]),
        Route("/readyz", readyz, methods=["GET"]),
        Route("/metrics", metrics, methods=["GET"]),
    ])


app = create_app()
