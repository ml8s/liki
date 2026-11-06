"""LIKI-MCP: a thin profile-free MCP passthrough gateway.

The gateway owns only one product decision: which upstream tool names are
exposed. It does not rewrite schemas, arguments, results, errors, or names.
"""

from __future__ import annotations

from dataclasses import dataclass
import json
import logging
from typing import TYPE_CHECKING, Any

import os

from app.config import (
    LATEST_PROTOCOL_VERSION,
    SUPPORTED_PROTOCOL_VERSIONS,
    GatewayConfig,
    load_config,
)
# alipay-sdk-python 和 order_store 在 _get_payment_service 中懒加载：
# paywall disabled 时不需要 SDK。
from app.upstream import UpstreamError, upstream_session

logger = logging.getLogger(__name__)

if TYPE_CHECKING:
    from app.config import AipayConfig
    from app.payment import Authorization, PaymentService

from app.payment import PaywallResponse

from starlette.applications import Starlette
from starlette.requests import Request
from starlette.responses import JSONResponse, Response
from starlette.routing import Route

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


def _method_not_found(request_id: Any, method: str) -> JSONResponse:
    """Answer an unsupported 2026 RPC with the transport's required status."""
    response = _jsonrpc_error(request_id, -32601, f"method not found: {method}")
    response.status_code = 404
    return response


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
        logger.exception("LIKI tools/call upstream failed")
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

    return _method_not_found(request_id, method)



_payment_service: "PaymentService | None" = None
_payment_config: "AipayConfig | None" = None


def _get_payment_service(config: GatewayConfig) -> "PaymentService":
    """懒初始化支付基础设施。"""
    global _payment_service, _payment_config
    if _payment_service is None or _payment_config != config.aipay:
        previous_service = _payment_service
        from app.order_store import OrderStore
        from app.payment import PaymentService

        _payment_service = PaymentService(
            config.aipay,
            OrderStore(config.aipay.db_path),
        )
        _payment_config = config.aipay
        if previous_service is not None:
            previous_service.close()
    return _payment_service


def _json_response_body(response: JSONResponse) -> Any:
    return json.loads(bytes(response.body))


def _request_paid_tools(payload: Any) -> list[str]:
    """Return paid tool names in JSON-RPC requests only (notifications excluded)."""
    messages = payload if isinstance(payload, list) else (
        [payload] if isinstance(payload, dict) else []
    )
    names: list[str] = []
    for message in messages:
        if not isinstance(message, dict) or message.get("id") is None:
            continue
        if message.get("method") != "tools/call":
            continue
        params = message.get("params")
        if isinstance(params, dict) and isinstance(params.get("name"), str):
            names.append(params["name"])
    return names


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

    # AI 付费工具检查（application layer）。批量请求中出现付费工具时
    # fail-closed：单个 Payment-Proof 只绑定一个 resource_id，不能安全共享。
    paid_authorization: Authorization | None = None
    if config.aipay.enabled:
        paid_tools = [
            name for name in _request_paid_tools(payload)
            if name in config.aipay.paid_tools
        ]
        if isinstance(payload, list) and paid_tools:
            return JSONResponse({
                "jsonrpc": "2.0",
                "id": None,
                "error": {
                    "code": -32600,
                    "message": "paid tools/call cannot be sent in a JSON-RPC batch",
                },
            }, status_code=400)
        if paid_tools:
            tool_name = paid_tools[0]
            proof = request.headers.get("Payment-Proof")
            payment_service = _get_payment_service(config)
            authorization = payment_service.authorize_paid_call(tool_name, proof)

            if isinstance(authorization, PaywallResponse):
                return JSONResponse(
                    authorization.body,
                    status_code=authorization.status_code,
                    headers=authorization.headers,
                )
            if authorization is None:
                authorization = payment_service.build_402_response(tool_name)
            reservation = payment_service.reserve_fulfillment(authorization)
            if reservation is not None:
                return JSONResponse(
                    reservation.body,
                    status_code=reservation.status_code,
                    headers=reservation.headers,
                )
            paid_authorization = authorization

    messages = payload if isinstance(payload, list) else [payload]
    responses: list[JSONResponse] = []
    has_request = False

    for message in messages:
        if not isinstance(message, dict):
            responses.append(_jsonrpc_error(None, -32600, "Invalid Request"))
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

    # After upstream delivery succeeds, send the mandatory A2M fulfillment
    # confirmation and attach Payment-Validation to the successful response.
    if paid_authorization is not None and has_request:
        payment = _get_payment_service(config)
        # A paid call never reaches this point inside a JSON-RPC batch.
        upstream_response = responses[0]
        upstream_body = _json_response_body(upstream_response)
        if "error" in upstream_body or upstream_body.get("result", {}).get("isError"):
            return upstream_response
        result = upstream_body.get("result")
        if not isinstance(result, dict):
            return upstream_response

        delivery_evidence = {
            "resource_id": paid_authorization.resource_id,
            "content": result.get("structuredContent") or result,
        }
        # A2M validators inspect the top-level HTTP JSON object; MCP clients
        # continue reading the unchanged standard jsonrpc result.
        upstream_body["resource_id"] = paid_authorization.resource_id
        upstream_body["content"] = delivery_evidence
        upstream_response = JSONResponse(
            upstream_body,
            status_code=upstream_response.status_code,
        )
        responses[0] = upstream_response

        validation = payment.complete_fulfillment(paid_authorization)
        if isinstance(validation, PaywallResponse):
            return JSONResponse(
                validation.body,
                status_code=validation.status_code,
                headers=validation.headers,
            )
        upstream_response.headers.append("Payment-Validation", validation.header())

    if not has_request:
        return Response(status_code=202)
    if len(responses) == 1:
        return responses[0]
    return JSONResponse([_json_response_body(response) for response in responses])


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
