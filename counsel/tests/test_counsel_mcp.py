"""counsel aggregate MCP endpoint tests."""

import asyncio
import json
import pathlib
from typing import Any

import pytest  # noqa: E402
from app import engine_mcp_client  # noqa: E402
from app.contracts import COUNSEL_TOOLS, NATAL_TOOLS  # noqa: E402
from app.contracts.schema import ToolContract  # noqa: E402
from app.counsel_mcp import (  # noqa: E402
    _make_app,
    app,
    create_counsel_root_mcp,
)


def test_docker_runtime_uses_configured_bind_host():
    dockerfile = (pathlib.Path(__file__).resolve().parents[1] / "Dockerfile").read_text(
        encoding="utf-8"
    )
    assert 'ENV LIKI_BIND_HOST=0.0.0.0' in dockerfile
    assert '--host \\"${LIKI_BIND_HOST:-0.0.0.0}\\"' in dockerfile
    assert '--host", "0.0.0.0' not in dockerfile
from app.engine_mcp_client import MCPError as EngineMCPError  # noqa: E402
from app.mcp_tool_adapter import add_contract_tool  # noqa: E402
from app.metrics import observe_tool_request, render_metrics, reset_metrics  # noqa: E402
from mcp import Client  # noqa: E402
from mcp.server import MCPServer  # noqa: E402
from mcp.types import ListToolsResult  # noqa: E402
from mcp_types.methods import serialize_server_result  # noqa: E402
from pydantic import ConfigDict, create_model  # noqa: E402
from starlette.applications import Starlette  # noqa: E402
from starlette.routing import Mount, Route  # noqa: E402


def _natal_subject():
    return {
        "standard_time": "1990-05-20T11:49:00+08:00",
        "longitude": 116.4,
        "gender": "male",
    }


def _call_tool_over_mcp(server, name, arguments):
    async def call():
        async with Client(server, mode="legacy") as client:
            return await client.call_tool(name, arguments)

    return asyncio.run(call())


async def _asgi_request(
    path,
    method="GET",
    payload=None,
    asgi_app=app,
    extra_headers=None,
):
    """Invoke the composed ASGI app directly without a TestClient portal."""
    if isinstance(payload, bytes):
        body = payload
    else:
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
    extra_by_name = {name.lower(): value for name, value in extra_headers or []}
    headers = [
        (name, value) for name, value in headers if name not in extra_by_name
    ] + list(extra_headers or [])

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


def _http_request(path, method="GET", payload=None, asgi_app=app):
    return asyncio.run(
        _asgi_request(
            path,
            method=method,
            payload=payload,
            asgi_app=asgi_app,
        )
    )


def _http_raw_request(path, body, asgi_app, extra_headers):
    return asyncio.run(
        _asgi_request(
            path,
            method="POST",
            payload=body,
            asgi_app=asgi_app,
            extra_headers=extra_headers,
        )
    )


@pytest.fixture(scope="module")
def subject():
    return _natal_subject()


def test_counsel_mcp_tools():
    srv = create_counsel_root_mcp()
    names = asyncio.run(srv.list_tools())
    assert [tool.name for tool in names] == [
        contract.name for contract in COUNSEL_TOOLS
    ]


@pytest.mark.parametrize(
    "protocol_version",
    ["2025-03-26", "2025-06-18", "2025-11-25", "2026-07-28"],
)
def test_counsel_root_tools_list_serializes_for_runtime(protocol_version):
    server = create_counsel_root_mcp()
    tools = asyncio.run(server.list_tools())
    assert len(tools) == len(COUNSEL_TOOLS)
    dumped = ListToolsResult(tools=tools).model_dump(
        by_alias=True, mode="json", exclude_none=True
    )
    serialized = serialize_server_result("tools/list", protocol_version, dumped)
    assert len(serialized["tools"]) == len(tools)


def test_bazi_yongshen_over_mcp(subject):
    srv = create_counsel_root_mcp()
    r = asyncio.run(srv.call_tool("bazi_yongshen", {"subject": subject}))
    out = r.structured_content
    assert out["selected_yongshen"]["yong"]
    assert out["status"] in ("resolved", "partial", "needs_review", "unresolved")
    assert "yongshen_digest" not in out
    assert "bazi_chart_digest" not in out
    assert "factors" not in out
    assert "alternatives" not in out
    assert out["ge_ju_context"]["pattern"]
    assert {item["source"] for item in out["evidence"]} == {
        "facts", "fu_yi", "tiao_hou", "ge_ju", "cong_ge",
    }
    assert r.structured_content == out


def test_natal_query_over_mcp(subject):
    srv = create_counsel_root_mcp()
    r = asyncio.run(srv.call_tool(
        "natal_query",
        {
            "subject": subject,
            "life_topics": ["chart_structure"],
        },
    ))
    out = r.structured_content
    assert out["bazi"]["assertions"]
    assert "ziwei" in out and isinstance(out["ziwei"]["assertions"], list)
    for side in ("bazi", "ziwei"):
        counts = out[side]["counts"]
        assert counts["returned"] == len(out[side]["assertions"])
        assert counts["filtered"] == counts["matched"] - counts["returned"]
        assert sum(counts["filtered_by"].values()) == counts["filtered"]
    assert all("technique" not in a for a in out["bazi"]["assertions"])
    assert out["natal_chart_digest"]


def test_natal_liunian_query_over_mcp(subject):
    srv = create_counsel_root_mcp()
    r = asyncio.run(srv.call_tool(
        "natal_liunian_query",
        {
            "subject": subject,
            "start_year": 2020,
            "year_count": 21,
            "life_topics": ["marriage"],
        },
    ))
    out = r.structured_content
    assert "hour_calibration" not in out
    years = out["years"]
    assert len(years) == 21  # 2020-2040 每年一段
    bazi_assertions = [a for p in years for a in p["bazi"]["assertions"]]
    ziwei_assertions = [a for p in years for a in p["ziwei"]["assertions"]]
    assert any(a["assertion_id"].startswith("ymar_") for a in bazi_assertions)
    assert bazi_assertions
    assert isinstance(ziwei_assertions, list)
    for year in years:
        for side in ("bazi", "ziwei"):
            if "counts" not in year[side]:
                continue
            counts = year[side]["counts"]
            assert counts["returned"] == len(year[side]["assertions"])
            assert counts["filtered"] == counts["matched"] - counts["returned"]
            assert sum(counts["filtered_by"].values()) == counts["filtered"]
    assert all("technique" not in a for a in bazi_assertions + ziwei_assertions)


def test_bazi_yongshen_rejects_partial_subject():
    result = _call_tool_over_mcp(
        create_counsel_root_mcp(),
        "bazi_yongshen",
        {"subject": {"standard_time": "1990-05-20T11:49:00+08:00"}},
    )
    assert result.is_error is True
    assert "longitude" in result.content[0].text


def test_counsel_tools_publish_the_contract_schemas_without_downgrading():
    expected = {contract.name: contract for contract in COUNSEL_TOOLS}
    tools = asyncio.run(create_counsel_root_mcp().list_tools())
    assert {tool.name: tool.input_schema for tool in tools} == {
        name: contract.input_schema for name, contract in expected.items()
    }
    assert {tool.name: tool.output_schema for tool in tools} == {
        name: contract.output_schema for name, contract in expected.items()
    }


def test_counsel_servers_advertise_the_distributed_version():
    from app.version import runtime_version

    version = runtime_version()
    assert create_counsel_root_mcp().version == version


def test_counsel_root_aggregate_exposes_expert_and_non_expert_tools(subject):
    srv = create_counsel_root_mcp()
    names = {tool.name for tool in asyncio.run(srv.list_tools())}
    assert {
        "bazi_yongshen", "natal_query", "natal_liunian_query",
        "naming_pick", "naming_compose", "liuyao_ask", "qimen_ask",
    } <= names


def test_counsel_root_aggregate_matches_tool_catalog():
    catalog_path = pathlib.Path(__file__).parents[2] / "tools" / "mcp-tool-catalog.json"
    catalog = json.loads(catalog_path.read_text(encoding="utf-8"))
    srv = create_counsel_root_mcp()
    names = {tool.name for tool in asyncio.run(srv.list_tools())}
    assert names == set(catalog["servers"]["counsel"]["tools"])


def test_counsel_root_aggregate_exposes_every_contract_schema():
    expected = {
        contract.name: (contract.input_schema, contract.output_schema)
        for contract in COUNSEL_TOOLS
    }

    srv = create_counsel_root_mcp()
    tools = asyncio.run(srv.list_tools())
    actual = {tool.name: (tool.input_schema, tool.output_schema) for tool in tools}
    assert actual == expected


def test_natal_query_accepts_full_topic_set_and_returns_topic_membership(subject):
    all_topics = [
        "marriage", "family", "children", "career", "wealth", "health",
        "mental", "study", "personality", "appearance", "origin",
        "relocation", "social", "property", "adversity", "chart_structure",
    ]
    result = asyncio.run(create_counsel_root_mcp().call_tool(
        "natal_query",
        {
            "subject": subject,
            "life_topics": all_topics,
        },
    ))
    out = result.structured_content
    assertions = out["bazi"]["assertions"] + out["ziwei"]["assertions"]
    assert assertions
    for assertion in assertions:
        assert isinstance(assertion.get("life_topics"), list)
        assert assertion["life_topics"]
        assert set(assertion["life_topics"]) <= set(all_topics)
    ids = [(side, a["assertion_id"]) for side in ("bazi", "ziwei") for a in out[side]["assertions"]]
    assert len(ids) == len(set(ids))


def test_natal_liunian_query_accepts_subject_only(subject):
    srv = create_counsel_root_mcp()
    result = asyncio.run(srv.call_tool(
        "natal_liunian_query",
        {
            "subject": subject,
            "start_year": 2026,
            "year_count": 1,
            "life_topics": ["career"],
        },
    ))
    out = result.structured_content
    assert "years" in out


def test_natal_liunian_query_contract_allows_full_topic_set():
    contract = next(c for c in NATAL_TOOLS if c.name == "natal_liunian_query")
    assert contract.input_schema["properties"]["life_topics"]["maxItems"] == 16


def test_counsel_root_aggregate_infers_natal_domain(subject):
    srv = create_counsel_root_mcp()
    result = asyncio.run(srv.call_tool(
        "natal_query",
        {
            "subject": subject,
            "life_topics": ["chart_structure"],
        },
    ))
    out = result.structured_content
    assert out["bazi"]["assertions"]
    assert "ziwei" in out and isinstance(out["ziwei"]["assertions"], list)
    assert result.structured_content == out


def test_counsel_tool_rejects_argument_outside_manifest_schema():
    result = _call_tool_over_mcp(
        create_counsel_root_mcp(),
        "natal_query",
        {"subject": {}, "life_topics": [], "unexpected": True},
    )
    assert result.is_error is True
    text = result.content[0].text
    assert "invalid_arguments" in text
    assert "unexpected" in text
    assert "Extra inputs are not permitted" in text


def _probe_contract(name: str, description: str = "test known failure mapping") -> ToolContract:
    """A throwaway no-argument contract for exercising the registration adapter."""
    empty = create_model(f"{name}_input", __config__=ConfigDict(extra="forbid"))
    return ToolContract(name=name, description=description, input=empty, output=dict[str, Any])


@pytest.mark.parametrize(
    ("exception", "code"),
    [
        (ValueError("candidate name is invalid"), "tool_failed"),
        (EngineMCPError("engine unavailable"), "engine_dependency_failed"),
    ],
)
def test_known_tool_failures_preserve_reason_over_mcp_protocol(exception, code):
    server = MCPServer(name="counsel-test", version="test")

    def invoke(**_arguments):
        raise exception

    add_contract_tool(server, _probe_contract("known_failure"), invoke)
    result = _call_tool_over_mcp(server, "known_failure", {})
    assert result.is_error is True
    text = result.content[0].text
    assert code in text
    assert str(exception) in text


def test_unexpected_tool_exception_is_not_leaked_over_mcp_protocol():
    server = MCPServer(name="counsel-test", version="test")

    def invoke(**_arguments):
        raise RuntimeError("internal secret")

    add_contract_tool(server, _probe_contract("boom", "test unknown exception handling"), invoke)
    result = _call_tool_over_mcp(server, "boom", {})
    assert result.is_error is True
    assert result.content[0].text == "Error executing tool boom"


def test_tool_adapter_records_terminal_status_and_duration():
    server = MCPServer(name="counsel-test", version="test")

    def invoke(**_arguments):
        return {"ok": True}

    add_contract_tool(server, _probe_contract("probe_metric"), invoke)
    reset_metrics()
    result = _call_tool_over_mcp(server, "probe_metric", {})
    assert not result.is_error
    text = render_metrics().decode("utf-8")
    assert 'tool="probe_metric",domain="aggregate",status="success"' in text
    assert "liki_counsel_tool_request_duration_seconds" in text


def test_aggregate_mcp_uses_standard_starlette_composition():
    """The official MCP apps must own protocol handling and ASGI lifespan."""
    assert isinstance(app, Starlette)
    routes = [route for route in app.routes if route.path.startswith("/mcp")]
    assert {route.path for route in routes} == {"/mcp"}
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
    assert results["/mcp"][0] == 200
    assert json.loads(results["/mcp"][1])["result"]["_meta"] is not None
    assert results["/mcp/bazi"][0] == 404
    assert results["/mcp/naming"][0] == 404

    # The public /counsel prefix is owned by the edge, not this process.
    status, _ = _http_request("/counsel/mcp/bazi", method="POST", payload=payload)
    assert status == 404


def test_readyz_reports_engine_dependency(monkeypatch):
    def unavailable(retries=0, timeout=None):
        raise EngineMCPError("engine unavailable")

    monkeypatch.setattr(engine_mcp_client, "server_info", unavailable)
    status, raw = _http_request("/readyz")
    assert status == 503
    document = json.loads(raw)
    assert document["status"] == "unavailable"
    assert document["reason"] == "engine_dependency_unreachable"
    assert "engine unavailable" not in raw.decode("utf-8")

    def available(retries=0, timeout=None):
        assert retries == 0
        assert 0 < timeout <= 1
        return {"name": "engine", "version": "test"}

    monkeypatch.setattr(engine_mcp_client, "server_info", available)
    status, raw = _http_request("/readyz")
    assert status == 200
    document = json.loads(raw)
    assert document["status"] == "ready"
    assert "engine_version" not in document


def test_healthz_remains_live_when_engine_dependency_is_down(monkeypatch):
    monkeypatch.setattr(
        engine_mcp_client,
        "server_info",
        lambda **_: (_ for _ in ()).throw(EngineMCPError("engine unavailable")),
    )
    status, raw = _http_request("/healthz")
    assert status == 200
    assert json.loads(raw)["status"] == "ok"


def test_readyz_rejects_invalid_timeout_configuration(monkeypatch):
    monkeypatch.setenv("LIKI_READYZ_TIMEOUT_SECONDS", "not-a-number")
    status, raw = _http_request("/readyz")
    assert status == 503
    assert json.loads(raw)["reason"] == "engine_dependency_unreachable"


def test_metrics_route_is_always_on():
    metrics_app = _make_app()
    status, _raw = _http_request("/metrics", asgi_app=metrics_app)
    assert status == 200


def test_metrics_route_renders_low_cardinality_counters(monkeypatch):
    metrics_app = _make_app()
    reset_metrics()
    observe_tool_request("bazi_yongshen", "bazi", "success", 0.01)

    status, raw = _http_request("/metrics", asgi_app=metrics_app)
    text = raw.decode("utf-8")
    print("DEBUG metrics text", text)
    assert status == 200
    assert 'tool="bazi_yongshen",domain="bazi",status="success"' in text
    assert "# TYPE liki_counsel_tool_requests_total counter" in text
    assert "# TYPE liki_counsel_tool_request_duration_seconds histogram" in text
    assert "# TYPE liki_counsel_engine_requests_total counter" in text


def test_counsel_mcp_rejects_oversized_request_body(monkeypatch):
    monkeypatch.setenv("LIKI_HTTP_MAX_BODY_BYTES", "16")
    limited_app = _make_app()
    status, raw = _http_request(
        "/mcp",
        method="POST",
        payload={"jsonrpc": "2.0", "id": 1, "method": "server/discover", "params": {}},
        asgi_app=limited_app,
    )
    assert status == 413
    assert json.loads(raw)["error"]["code"] == "payload_too_large"


def test_counsel_mcp_counts_actual_request_bytes_even_with_small_content_length(monkeypatch):
    monkeypatch.setenv("LIKI_HTTP_MAX_BODY_BYTES", "16")
    limited_app = _make_app()
    status, raw = _http_raw_request(
        "/mcp",
        b"x" * 17,
        limited_app,
        [(b"content-length", b"1")],
    )
    assert status == 413
    assert json.loads(raw)["error"]["code"] == "payload_too_large"


def test_engine_mcp_client_sends_protocol_headers_only(monkeypatch):
    requests = []

    class response:
        status = 200
        reason = "OK"

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc_value, traceback):
            return False

        @staticmethod
        def read():
            return b'{"result":{"_meta":{"io.modelcontextprotocol/serverInfo":{"name":"engine","version":"2026.09.26.0"}}}}'

    def urlopen(request, timeout):
        requests.append(request)
        return response()

    monkeypatch.setenv("LIKI_ENGINE_MCP_URL", "http://engine-mcp:8085/mcp/")
    monkeypatch.setattr(engine_mcp_client, "_HTTP_TRANSPORT", urlopen)

    engine_mcp_client.server_info(retries=0)

    assert len(requests) == 1
    assert {header.lower() for header, _ in requests[0].header_items()} == {
        "accept",
        "content-type",
        "mcp-method",
        "mcp-protocol-version",
    }

def test_engine_client_records_terminal_status_metrics(monkeypatch):
    timeouts = []

    class response:
        status = 200
        reason = "OK"

        def __enter__(self):
            return self

        def __exit__(self, exc_type, exc_value, traceback):
            return False

        @staticmethod
        def read():
            return json.dumps({
                "result": {
                    "_meta": {
                        "io.modelcontextprotocol/serverInfo": {
                            "name": "engine",
                            "version": "test",
                        }
                    }
                }
            }).encode()

    monkeypatch.setenv("LIKI_ENGINE_MCP_URL", "http://engine-mcp:8085/mcp")
    monkeypatch.setattr(
        engine_mcp_client,
        "_HTTP_TRANSPORT",
        lambda _request, timeout: (timeouts.append(timeout), response())[1],
    )
    reset_metrics()
    engine_mcp_client.server_info(retries=0, timeout=0.25)
    text = render_metrics().decode("utf-8")
    assert timeouts == [0.25]
    assert 'method="server/discover",status="success"' in text
    assert "liki_counsel_engine_request_duration_seconds" in text


def test_counsel_mcp_is_public_without_locking_health():
    public_app = _make_app()
    status, _raw = _http_request("/healthz", asgi_app=public_app)
    assert status == 200


def test_counsel_mcp_does_not_advertise_change_streams():
    """A static tool image must not invite clients to open listen streams."""
    public_app = _make_app()

    async def request_all():
        async with public_app.router.lifespan_context(public_app):
            discover = await _asgi_request(
                "/mcp",
                method="POST",
                payload={
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "server/discover",
                    "params": {
                        "_meta": {
                            "io.modelcontextprotocol/protocolVersion": "2026-07-28",
                            "io.modelcontextprotocol/clientInfo": {"name": "test", "version": "1.0"},
                            "io.modelcontextprotocol/clientCapabilities": {},
                        }
                    },
                },
                asgi_app=public_app,
            )
            listen_payload = {
                "jsonrpc": "2.0",
                "id": 2,
                "method": "subscriptions/listen",
                "params": {
                    "notifications": {},
                    "_meta": {
                        "io.modelcontextprotocol/protocolVersion": "2026-07-28",
                        "io.modelcontextprotocol/clientInfo": {"name": "test", "version": "1.0"},
                        "io.modelcontextprotocol/clientCapabilities": {},
                    },
                },
            }
            listen = await _asgi_request(
                "/mcp",
                method="POST",
                payload=listen_payload,
                asgi_app=public_app,
                extra_headers=[(b"mcp-method", b"subscriptions/listen")],
            )
            get = await _asgi_request("/mcp", method="GET", asgi_app=public_app)
            return discover, listen, get

    discover, listen, get = asyncio.run(request_all())
    assert discover[0] == 200
    capabilities = json.loads(discover[1])["result"]["capabilities"]
    assert not capabilities.get("tools", {}).get("listChanged", False)
    assert not capabilities.get("prompts", {}).get("listChanged", False)
    assert not capabilities.get("resources", {}).get("listChanged", False)
    assert not capabilities.get("resources", {}).get("subscribe", False)

    assert listen[0] == 404
    assert json.loads(listen[1])["error"]["code"] == -32601
    assert get[0] == 405


def test_engine_endpoint_requires_explicit_url(monkeypatch):
    monkeypatch.delenv("LIKI_ENGINE_MCP_URL", raising=False)
    with pytest.raises(EngineMCPError, match="LIKI_ENGINE_MCP_URL"):
        engine_mcp_client._endpoint()


def test_lifespan_fails_fast_without_engine_url(monkeypatch):
    """Startup must reject an empty LIKI_ENGINE_MCP_URL before touching any domain."""
    monkeypatch.delenv("LIKI_ENGINE_MCP_URL", raising=False)
    test_app = _make_app()

    async def enter_lifespan():
        async with test_app.router.lifespan_context(test_app):
            pass

    with pytest.raises(RuntimeError, match="LIKI_ENGINE_MCP_URL"):
        asyncio.run(enter_lifespan())
