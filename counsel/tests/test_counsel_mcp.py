"""counsel 判断层 MCP 端点测试（分域 bazi/ziwei，3 工具）。"""
from __future__ import annotations

import asyncio
import json
import os
import pathlib

import pytest  # noqa: E402
from app import engine_client  # noqa: E402
from app.counsel_mcp import (  # noqa: E402
    _add_schema_tool,
    _make_app,
    app,
    create_counsel_mcp,
    create_counsel_root_mcp,
)
from app.engine_client import MCPError as EngineMCPError  # noqa: E402
from mcp import Client  # noqa: E402
from mcp.server import MCPServer  # noqa: E402
from starlette.applications import Starlette  # noqa: E402
from starlette.routing import Mount, Route  # noqa: E402


def _engine_bazi_chart():
    import urllib.request

    url = os.environ["LIKI_MCP_URL"] + "/bazi"
    body = json.dumps({
        "jsonrpc": "2.0", "id": 1, "method": "tools/call",
        "params": {"name": "bazi_chart", "arguments": {
            "solar_time": "1990-05-20T11:49:00+08:00", "gender": "male",
            "longitude": 116.4,
        }, "_meta": {
            "io.modelcontextprotocol/protocolVersion": "2026-07-28",
            "io.modelcontextprotocol/clientCapabilities": {},
        }},
    }).encode()
    req = urllib.request.Request(
        url, data=body,
        headers={
            "Content-Type": "application/json",
            "MCP-Protocol-Version": "2026-07-28",
            "Mcp-Method": "tools/call", "Mcp-Name": "bazi_chart",
            "Accept": "application/json, text/event-stream",
        },
    )
    with urllib.request.urlopen(req, timeout=8) as r:
        return json.loads(json.loads(r.read())["result"]["content"][0]["text"])


def _call_tool_over_mcp(server, name, arguments):
    async def call():
        async with Client(server, mode="legacy") as client:
            return await client.call_tool(name, arguments)

    return asyncio.run(call())


async def _asgi_request(path, method="GET", payload=None, asgi_app=app, authorization=None):
    """Invoke the composed ASGI app directly without a TestClient portal."""
    body = json.dumps(payload).encode("utf-8") if payload is not None else b""
    headers = []
    if payload is not None:
        headers.extend(
            [
                (b"content-type", b"application/json"),
                (b"accept", b"application/json, text/event-stream"),
                (b"mcp-protocol-version", b"2026-07-28"),
                (b"mcp-method", b"server/discover"),
            ]
        )
    if authorization is not None:
        headers.append((b"authorization", authorization.encode("latin-1")))

    messages = []

    async def receive():
        return {"type": "http.request", "body": body, "more_body": False}

    async def send(message):
        messages.append(message)

    scope = {
        "type": "http",
        "asgi": {"version": "3.0"},
        "http_version": "1.1",
        "method": method,
        "path": path,
        "raw_path": path.encode(),
        "headers": headers,
        "query_string": b"",
        "client": ("test-client", 1),
        "server": ("test-server", 80),
    }
    await asgi_app(scope, receive, send)
    status = next(item["status"] for item in messages if item["type"] == "http.response.start")
    raw = b"".join(
        item.get("body", b"") for item in messages if item["type"] == "http.response.body"
    )
    return status, raw


def _http_request(path, method="GET", payload=None, asgi_app=app, authorization=None):
    return asyncio.run(
        _asgi_request(
            path,
            method=method,
            payload=payload,
            asgi_app=asgi_app,
            authorization=authorization,
        )
    )


@pytest.fixture(scope="module")
def bazi_chart():
    return _engine_bazi_chart()


def test_counsel_mcp_tools():
    srv = create_counsel_mcp("bazi")
    names = asyncio.run(srv.list_tools())
    assert [t.name for t in names] == ["compute_factors", "natal_query", "period_query"]


def test_compute_factors_over_mcp(bazi_chart):
    srv = create_counsel_mcp("bazi")
    r = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    out = json.loads(r.content[0].text)
    assert set(out) == {"factors", "factors_digest", "context"}
    assert out["context"]["性别"] == "male"
    assert len(out["factors_digest"]) == 64  # sha256 hex


def test_natal_query_over_mcp(bazi_chart):
    srv = create_counsel_mcp("bazi")
    r = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    snapshot = json.loads(r.content[0].text)
    r = asyncio.run(srv.call_tool(
        "natal_query",
        {
            "factors": snapshot["factors"],
            "factors_digest": snapshot["factors_digest"],
            "topics": ["chart_structure"],
            "context": {"性别": "male"},
        },
    ))
    out = json.loads(r.content[0].text)
    assert out["assertions"] and all(a.get("side") == "bazi" for a in out["assertions"])


def test_period_query_over_mcp(bazi_chart):
    srv = create_counsel_mcp("bazi")
    r = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    snapshot = json.loads(r.content[0].text)
    r = asyncio.run(srv.call_tool(
        "period_query",
        {
            "factors": snapshot["factors"],
            "factors_digest": snapshot["factors_digest"],
            "time_scope": {"type": "year_range", "start_year": 2020, "end_year": 2040},
            "topics": ["marriage"],
            "chart": bazi_chart,
        },
    ))
    out = json.loads(r.content[0].text)
    periods = out["periods"]
    assert len(periods) == 21  # 2020-2040 每年一段
    assert sum(len(p["assertions"]) for p in periods) == 20
    assert all(a.get("side") == "bazi" for p in periods for a in p["assertions"])


def test_domain_rejects_wrong_chart():
    result = _call_tool_over_mcp(
        create_counsel_mcp("bazi"),
        "compute_factors",
        {"chart": {"ziwei": {}}},
    )
    assert result.is_error is True
    text = result.content[0].text
    assert "invalid_arguments" in text
    assert "counsel/bazi 只接收八字盘" in text


def test_counsel_tools_expose_manifest_schema_without_downgrading():
    srv = create_counsel_mcp("bazi")
    manifest = json.loads(
        (pathlib.Path(__file__).parents[1] / "app/natal/tools/counsel-tools.json")
        .read_text(encoding="utf-8")
    )
    expected = {
        fn["function"]["name"]: fn["function"]["parameters"]
        for fn in manifest["tools"]
    }
    tools = asyncio.run(srv.list_tools())
    assert {tool.name: tool.input_schema for tool in tools} == expected


def test_counsel_root_aggregate_exposes_non_expert_tools():
    srv = create_counsel_root_mcp()
    names = {tool.name for tool in asyncio.run(srv.list_tools())}
    assert {
        "qiming_pick", "qiming_compose", "liuyao_snapshot",
        "liuyao_query", "qimen_snapshot", "qimen_query",
    } <= names
    assert "compute_factors" not in names


def test_counsel_tool_rejects_argument_outside_manifest_schema():
    result = _call_tool_over_mcp(
        create_counsel_mcp("bazi"),
        "compute_factors",
        {"chart": {}, "unexpected": True},
    )
    assert result.is_error is True
    text = result.content[0].text
    assert "invalid_arguments" in text
    assert "unexpected" in text
    assert "Additional properties are not allowed" in text


@pytest.mark.parametrize(
    ("exception", "code"),
    [
        (ValueError("candidate name is invalid"), "tool_failed"),
        (EngineMCPError("engine unavailable"), "engine_dependency_failed"),
    ],
)
def test_known_tool_failures_preserve_reason_over_mcp_protocol(exception, code):
    server = MCPServer(name="counsel-test", version="test")

    def invoke(_arguments):
        raise exception

    _add_schema_tool(
        server,
        name="known_failure",
        description="test known failure mapping",
        schema={"type": "object", "properties": {}, "additionalProperties": False},
        invoke=invoke,
    )
    result = _call_tool_over_mcp(server, "known_failure", {})
    assert result.is_error is True
    text = result.content[0].text
    assert code in text
    assert str(exception) in text


def test_unexpected_tool_exception_is_not_leaked_over_mcp_protocol():
    server = MCPServer(name="counsel-test", version="test")

    def invoke(_arguments):
        raise RuntimeError("internal secret")

    _add_schema_tool(
        server,
        name="boom",
        description="test unknown exception handling",
        schema={"type": "object", "properties": {}, "additionalProperties": False},
        invoke=invoke,
    )
    result = _call_tool_over_mcp(server, "boom", {})
    assert result.is_error is True
    assert result.content[0].text == "Error executing tool boom"


def test_multi_domain_mcp_uses_standard_starlette_composition():
    """The official MCP apps must own protocol handling and ASGI lifespan."""
    assert isinstance(app, Starlette)
    routes = [route for route in app.routes if route.path.startswith("/mcp")]
    assert {route.path for route in routes} == {
        "/mcp",
        "/mcp/bazi",
        "/mcp/ziwei",
        "/mcp/liuyao",
        "/mcp/qimen",
        "/mcp/naming",
    }
    assert all(isinstance(route, Route) for route in routes)
    assert not any(isinstance(route, Mount) for route in app.routes)
    assert "/healthz" in {route.path for route in app.routes}


def test_counsel_mcp_http_surface():
    """The composed ASGI app serves health and MCP paths without edge rewrites."""
    status, raw = _http_request("/healthz")
    assert status == 200
    assert json.loads(raw)["status"] == "ok"

    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "server/discover",
        "params": {
            "_meta": {
                "io.modelcontextprotocol/protocolVersion": "2026-07-28",
                "io.modelcontextprotocol/clientCapabilities": {},
            }
        },
    }
    results = {}

    async def request_all():
        async with app.router.lifespan_context(app):
            for path in ("/mcp", "/mcp/bazi", "/mcp/naming"):
                results[path] = await _asgi_request(path, method="POST", payload=payload)

    asyncio.run(request_all())
    for path, (status, raw) in results.items():
        assert status == 200, path
        assert json.loads(raw)["result"]["_meta"] is not None

    # The public /counsel prefix is owned by the edge, not this process.
    status, _ = _http_request("/counsel/mcp/bazi", method="POST", payload=payload)
    assert status == 404


def test_counsel_health_is_not_consumed_by_mcp_rate_limit():
    """Liveness must remain available even when MCP clients exhaust quota."""
    for _ in range(260):
        status, _raw = _http_request("/healthz")
        assert status == 200


def test_readyz_reports_engine_dependency(monkeypatch):
    monkeypatch.setattr(engine_client, "engine_version", lambda: "2026.09.26.0")
    monkeypatch.setattr(engine_client, "required_engine_version", lambda: "2026.09.26.0")
    status, raw = _http_request("/readyz")
    assert status == 200
    assert json.loads(raw)["status"] == "ready"

    def unavailable():
        raise EngineMCPError("engine unavailable")

    monkeypatch.setattr(engine_client, "engine_version", unavailable)
    status, raw = _http_request("/readyz")
    assert status == 503
    assert json.loads(raw)["status"] == "unready"
    assert json.loads(raw)["reason"] == "engine dependency unavailable"


def test_counsel_mcp_rejects_oversized_request_body(monkeypatch):
    monkeypatch.setenv("LIKI_MCP_MAX_BODY_BYTES", "16")
    limited_app = _make_app()
    status, raw = _http_request(
        "/mcp",
        method="POST",
        payload={"jsonrpc": "2.0", "id": 1, "method": "server/discover", "params": {}},
        asgi_app=limited_app,
    )
    assert status == 413
    assert json.loads(raw)["error"]["code"] == "payload_too_large"


def test_authentication_runs_before_request_body_limit(monkeypatch):
    monkeypatch.setenv("LIKI_MCP_TOKEN", "test-secret")
    monkeypatch.setenv("LIKI_MCP_MAX_BODY_BYTES", "16")
    guarded_app = _make_app()
    status, raw = _http_request(
        "/mcp",
        method="POST",
        payload={"jsonrpc": "2.0", "id": 1, "method": "server/discover", "params": {}},
        asgi_app=guarded_app,
    )
    assert status == 401
    assert "unauthorized" in json.loads(raw)["error"]["code"]


def test_engine_client_forwards_private_engine_token(monkeypatch):
    requests = []

    class response:
        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc_value, traceback):
            return False

        @staticmethod
        def read():
            return b'{"result":{"_meta":{"io.modelcontextprotocol/serverInfo":{"name":"engine","version":"2026.09.26.0"}}}}'

    def urlopen(request, timeout):
        requests.append({"request": request, "timeout": timeout})
        return response()

    monkeypatch.setenv("LIKI_MCP_URL", "http://engine-mcp:8085/mcp/")
    monkeypatch.setenv("LIKI_ENGINE_MCP_TOKEN", "engine-secret")
    monkeypatch.setenv("LIKI_MCP_TOKEN", "counsel-secret")
    monkeypatch.setattr(engine_client.urllib.request, "urlopen", urlopen)

    engine_client.server_info(retries=0)

    assert len(requests) == 1
    request = requests[0]["request"]
    assert request.full_url == "http://engine-mcp:8085/mcp"
    assert request.get_header("Authorization") == "Bearer engine-secret"

    monkeypatch.setenv("LIKI_ENGINE_MCP_TOKEN", "")
    engine_client.server_info(retries=0)
    fallback = requests[1]["request"]
    assert fallback.get_header("Authorization") == "Bearer counsel-secret"


def test_optional_bearer_auth_protects_counsel_mcp_without_locking_health(monkeypatch):
    monkeypatch.setenv("LIKI_MCP_TOKEN", "test-secret")
    guarded_app = _make_app()
    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "server/discover",
        "params": {
            "_meta": {
                "io.modelcontextprotocol/protocolVersion": "2026-07-28",
                "io.modelcontextprotocol/clientCapabilities": {},
            }
        },
    }

    for authorization in (None, "Bearer wrong"):
        status, raw = _http_request(
            "/mcp",
            method="POST",
            payload=payload,
            asgi_app=guarded_app,
            authorization=authorization,
        )
        assert status == 401
        assert "unauthorized" in json.loads(raw)["error"]["code"]

    status, _raw = _http_request("/healthz", asgi_app=guarded_app)
    assert status == 200

    async def authorized_request():
        async with guarded_app.router.lifespan_context(guarded_app):
            return await _asgi_request(
                "/mcp",
                method="POST",
                payload=payload,
                asgi_app=guarded_app,
                authorization="Bearer test-secret",
            )

    status, _raw = asyncio.run(authorized_request())
    assert status == 200


def test_unauthorized_attempts_consume_rate_limit(monkeypatch):
    monkeypatch.setenv("LIKI_MCP_TOKEN", "test-secret")
    monkeypatch.setenv("LIKI_COUNSEL_RATE_LIMIT", "1")
    guarded_app = _make_app()
    payload = {"jsonrpc": "2.0", "id": 1, "method": "server/discover", "params": {}}

    first = _http_request(
        "/mcp", method="POST", payload=payload, asgi_app=guarded_app
    )
    second = _http_request(
        "/mcp", method="POST", payload=payload, asgi_app=guarded_app
    )

    assert first[0] == 401
    assert second[0] == 429
