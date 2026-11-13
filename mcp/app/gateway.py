"""LIKI-MCP HTTP endpoint and paid-call orchestration."""

from __future__ import annotations

import json
import logging
from typing import TYPE_CHECKING, Any

from app.config import load_config
from app.http_routes import create_app
from app.payment import PaywallResponse
from app.protocol import ProtocolError, handle_request, jsonrpc_error
from app.tool_router import handle_tools_call, handle_tools_list
from starlette.requests import Request
from starlette.responses import JSONResponse, Response

if TYPE_CHECKING:
    from app.config import AipayConfig, GatewayConfig
    from app.payment import Authorization, PaymentService

logger = logging.getLogger(__name__)

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
    messages = payload if isinstance(payload, list) else ([payload] if isinstance(payload, dict) else [])
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
        return JSONResponse(
            {
                "jsonrpc": "2.0",
                "id": None,
                "error": {"code": -32700, "message": "Parse error"},
            },
            status_code=400,
        )

    # AI 付费工具检查（application layer）。批量请求中出现付费工具时
    # fail-closed：单个 Payment-Proof 只绑定一个 resource_id，不能安全共享。
    paid_authorization: Authorization | None = None
    if config.aipay.enabled:
        paid_tools = [name for name in _request_paid_tools(payload) if name in config.aipay.paid_tools]
        if isinstance(payload, list) and paid_tools:
            return JSONResponse(
                {
                    "jsonrpc": "2.0",
                    "id": None,
                    "error": {
                        "code": -32600,
                        "message": "paid tools/call cannot be sent in a JSON-RPC batch",
                    },
                },
                status_code=400,
            )
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
            responses.append(jsonrpc_error(None, -32600, "Invalid Request"))
            continue

        method = message.get("method")
        request_id = message.get("id")
        params = message.get("params") or {}

        if not isinstance(method, str):
            if request_id is not None:
                responses.append(
                    {
                        "jsonrpc": "2.0",
                        "id": request_id,
                        "error": {"code": -32600, "message": "Invalid Request"},
                    }
                )
            continue

        if request_id is None:
            # Notifications have no response. LIKI exposes no server stream.
            continue

        has_request = True
        try:
            response = await handle_request(
                config,
                request_id,
                method,
                params,
                tools_handler=_handle_tools,
            )
        except ProtocolError as error:
            response = jsonrpc_error(request_id, error.code, error.message)
        except Exception:
            response = jsonrpc_error(request_id, -32603, "Internal error")
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


async def _handle_tools(config, request_id, params=None):
    if params is None:
        return await handle_tools_list(config, request_id)
    return await handle_tools_call(config, request_id, params)


app = create_app()
