"""counsel 断语引擎 MCP 端点测试（bazi 4 工具，ziwei 3 工具）。"""

import asyncio
import json
import os
import pathlib
from typing import Any

import pytest  # noqa: E402
from app import engine_mcp_client  # noqa: E402
from app.contracts import COUNSEL_TOOLS, NATAL_TOOLS  # noqa: E402
from app.contracts.schema import ToolContract  # noqa: E402
from app.counsel_mcp import (  # noqa: E402
    _make_app,
    app,
    create_counsel_mcp,
    create_counsel_root_mcp,
)
from app.engine_mcp_client import MCPError as EngineMCPError  # noqa: E402
from app.mcp_tool_adapter import add_contract_tool  # noqa: E402
from mcp import Client  # noqa: E402
from mcp.server import MCPServer  # noqa: E402
from mcp.types import ListToolsResult  # noqa: E402
from mcp_types.methods import serialize_server_result  # noqa: E402
from pydantic import ConfigDict, create_model  # noqa: E402
from starlette.applications import Starlette  # noqa: E402
from starlette.routing import Mount, Route  # noqa: E402


def _engine_bazi_chart():
    import urllib.request

    url = os.environ["LIKI_MCP_URL"] + "/bazi"
    body = json.dumps({
        "jsonrpc": "2.0", "id": 1, "method": "tools/call",
        "params": {"name": "bazi_chart", "arguments": {
            "instant": "1990-05-20T03:49:00Z", "timezone": "Asia/Shanghai",
            "longitude": 116.4, "gender": "male",
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


async def _asgi_request(
    path,
    method="GET",
    payload=None,
    asgi_app=app,
    authorization=None,
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
    if authorization is not None:
        headers.append((b"authorization", authorization.encode("latin-1")))
    headers.extend(extra_headers or [])

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
def bazi_chart():
    return _engine_bazi_chart()


def test_counsel_mcp_tools():
    srv = create_counsel_mcp("bazi")
    names = asyncio.run(srv.list_tools())
    assert [t.name for t in names] == [
        "bazi_yongshen", "compute_factors", "natal_query", "period_query",
    ]


@pytest.mark.parametrize(
    "domain",
    ["bazi", "ziwei", "naming", "liuyao", "qimen"],
)
def test_counsel_tools_list_serializes_for_runtime(domain):
    """Runtime readiness consumes tools/list through the MCP wire serializer."""
    server = create_counsel_mcp(domain)
    tools = asyncio.run(server.list_tools())
    assert tools
    dumped = ListToolsResult(tools=tools).model_dump(
        by_alias=True, mode="json", exclude_none=True
    )
    serialized = serialize_server_result("tools/list", "2026-07-28", dumped)
    assert [tool["name"] for tool in serialized["tools"]] == [
        tool.name for tool in tools
    ]


def test_counsel_root_tools_list_serializes_for_runtime():
    server = create_counsel_root_mcp()
    tools = asyncio.run(server.list_tools())
    assert len(tools) == len(COUNSEL_TOOLS)
    dumped = ListToolsResult(tools=tools).model_dump(
        by_alias=True, mode="json", exclude_none=True
    )
    serialized = serialize_server_result("tools/list", "2026-07-28", dumped)
    assert len(serialized["tools"]) == len(tools)


def test_ziwei_mcp_does_not_expose_bazi_yongshen():
    srv = create_counsel_mcp("ziwei")
    names = asyncio.run(srv.list_tools())
    assert [t.name for t in names] == ["compute_factors", "natal_query", "period_query"]


def test_bazi_yongshen_over_mcp(bazi_chart):
    srv = create_counsel_mcp("bazi")
    r = asyncio.run(srv.call_tool("bazi_yongshen", {"chart": bazi_chart}))
    out = json.loads(r.content[0].text)
    assert out["selected"]["yong"]
    assert out["status"] in ("resolved", "partial", "needs_review", "unresolved")
    assert len(out["yongshen_digest"]) == 64
    assert "factors" not in out
    assert "alternatives" not in out
    assert out["ge_ju_context"]["ge_ju"]
    assert {item["source"] for item in out["evidence"]} == {
        "facts", "fu_yi", "tiao_hou", "ge_ju", "cong_ge",
    }
    assert r.structured_content == out


def test_compute_factors_over_mcp(bazi_chart):
    srv = create_counsel_mcp("bazi")
    r = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    out = json.loads(r.content[0].text)
    assert set(out) == {"factors", "factors_digest"}
    assert "selected_yongshen" in out["factors"]
    assert out["factors"]["_provenance"]["context"]["性别"] == "male"
    assert len(out["factors_digest"]) == 64  # sha256 hex


def test_compute_factors_has_one_yongshen_source(bazi_chart):
    srv = create_counsel_mcp("bazi")
    out = json.loads(asyncio.run(srv.call_tool(
        "compute_factors", {"chart": bazi_chart}
    )).content[0].text)
    selected = out["factors"]["selected_yongshen"]
    provenance = out["factors"]["_provenance"]
    direct = json.loads(asyncio.run(srv.call_tool(
        "bazi_yongshen", {"chart": bazi_chart}
    )).content[0].text)
    assert selected == direct
    assert selected["policy_id"] == "liki.bazi.yongshen"
    assert selected["yongshen_digest"]
    assert "yongshen" not in provenance
    assert provenance["yongshen_digest"] == selected["yongshen_digest"]
    assert provenance["policy_id"] == selected["policy_id"]


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
            "time_scope": {"type": "year_range", "start_year": 2020, "end_year": 2040, "timezone": "Asia/Shanghai"},
            "topics": ["marriage"],
            "chart": bazi_chart,
        },
    ))
    out = json.loads(r.content[0].text)
    assert out["calibration_facts"]["hour_pillar"] == {
        "gan": bazi_chart["shi"]["gan"],
        "zhi": bazi_chart["shi"]["zhi"],
    }
    assert out["calibration_facts"]["hour_stem_ten_god"]
    periods = out["periods"]
    assert len(periods) == 21  # 2020-2040 每年一段
    assertions = [a for p in periods for a in p["assertions"]]
    # marriage 的 annual_rules 含 yingqi（应期触发）；跨切域 应期 随 topic 呈现。
    assert any(a["assertion_id"].startswith(("ying_", "ysz_")) for a in assertions)
    assert sum(len(p["assertions"]) for p in periods) == 31
    assert all(a.get("side") == "bazi" for a in assertions)


def test_domain_rejects_wrong_chart():
    """结构上既不是八字盘也不是紫微盘：在 schema 边界即拒。"""
    result = _call_tool_over_mcp(
        create_counsel_mcp("bazi"),
        "compute_factors",
        {"chart": {"ziwei": {}}},
    )
    assert result.is_error is True
    text = result.content[0].text
    assert "invalid_arguments" in text
    # 联合两支的必填字段缺失都会出现在校验错误里（八字缺四柱，紫微缺 school/gong_wei）。
    assert "nian" in text and "school" in text


def test_bazi_yongshen_rejects_ziwei_chart():
    """bazi_yongshen 的 chart 是强类型 BaziChart：紫微盘缺四柱必填字段，边界即拒。"""
    result = _call_tool_over_mcp(
        create_counsel_mcp("bazi"),
        "bazi_yongshen",
        {"chart": {"ziwei": {}}},
    )
    assert result.is_error is True
    assert "invalid_arguments" in result.content[0].text
    assert "nian" in result.content[0].text


def test_counsel_tools_publish_the_contract_schemas_without_downgrading():
    expected = {contract.name: contract for contract in NATAL_TOOLS}
    tools = asyncio.run(create_counsel_mcp("bazi").list_tools())
    assert {tool.name: tool.input_schema for tool in tools} == {
        name: contract.input_schema for name, contract in expected.items()
    }
    assert {tool.name: tool.output_schema for tool in tools} == {
        name: contract.output_schema for name, contract in expected.items()
    }


def test_counsel_servers_advertise_the_distributed_version():
    from app.version import runtime_version

    version = runtime_version()
    for domain in ("bazi", "ziwei", "liuyao", "qimen", "naming"):
        assert create_counsel_mcp(domain).version == version
    assert create_counsel_root_mcp().version == version


def test_counsel_root_aggregate_exposes_expert_and_non_expert_tools(bazi_chart):
    srv = create_counsel_root_mcp()
    names = {tool.name for tool in asyncio.run(srv.list_tools())}
    assert {
        "bazi_yongshen", "compute_factors", "natal_query", "period_query",
        "naming_pick", "naming_compose", "liuyao_snapshot",
        "liuyao_query", "qimen_snapshot", "qimen_query",
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


def test_natal_query_accepts_full_topic_set_and_returns_topic_membership(bazi_chart):
    snapshot = json.loads(asyncio.run(create_counsel_root_mcp().call_tool(
        "compute_factors", {"chart": bazi_chart}
    )).content[0].text)
    all_topics = [
        "marriage", "family", "children", "career", "wealth", "health",
        "mental", "study", "personality", "appearance", "origin",
        "relocation", "social", "property", "adversity", "chart_structure",
    ]
    result = asyncio.run(create_counsel_root_mcp().call_tool(
        "natal_query",
        {
            "factors": snapshot["factors"],
            "factors_digest": snapshot["factors_digest"],
            "topics": all_topics,
        },
    ))
    out = json.loads(result.content[0].text)
    assert out["assertions"]
    for assertion in out["assertions"]:
        assert isinstance(assertion.get("topics"), list)
        assert assertion["topics"]
        assert set(assertion["topics"]) <= set(all_topics)
    ids = [assertion["assertion_id"] for assertion in out["assertions"]]
    assert len(ids) == len(set(ids))


def test_period_query_accepts_compute_factors_context(bazi_chart):
    srv = create_counsel_mcp("bazi")
    snapshot = json.loads(asyncio.run(srv.call_tool(
        "compute_factors", {"chart": bazi_chart}
    )).content[0].text)
    result = asyncio.run(srv.call_tool(
        "period_query",
        {
            "factors": snapshot["factors"],
            "factors_digest": snapshot["factors_digest"],
            "time_scope": {"type": "current_year", "timezone": "Asia/Shanghai"},
            "topics": ["career"],
            "chart": bazi_chart,
        },
    ))
    out = json.loads(result.content[0].text)
    assert "periods" in out


def test_period_query_contract_allows_full_topic_set():
    contract = next(c for c in NATAL_TOOLS if c.name == "period_query")
    assert contract.input_schema["properties"]["topics"]["maxItems"] == 16


def test_counsel_root_aggregate_infers_natal_domain(bazi_chart):
    srv = create_counsel_root_mcp()
    result = asyncio.run(srv.call_tool("compute_factors", {"chart": bazi_chart}))
    snapshot = json.loads(result.content[0].text)
    assert snapshot["factors"]["_provenance"]["side"] == "bazi"
    result = asyncio.run(srv.call_tool(
        "natal_query",
        {
            "factors": snapshot["factors"],
            "factors_digest": snapshot["factors_digest"],
            "topics": ["chart_structure"],
        },
    ))
    out = json.loads(result.content[0].text)
    assert out["assertions"] and all(a.get("side") == "bazi" for a in out["assertions"])
    assert result.structured_content == out


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


def test_readyz_does_not_depend_on_engine_version(monkeypatch):
    def unavailable():
        raise EngineMCPError("engine unavailable")

    monkeypatch.setattr(engine_mcp_client, "server_info", unavailable)
    status, raw = _http_request("/readyz")
    assert status == 200
    document = json.loads(raw)
    assert document["status"] == "ready"
    assert "engine_version" not in document


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


def test_counsel_mcp_counts_actual_request_bytes_even_with_small_content_length(monkeypatch):
    monkeypatch.setenv("LIKI_MCP_MAX_BODY_BYTES", "16")
    limited_app = _make_app()
    status, raw = _http_raw_request(
        "/mcp",
        b"x" * 17,
        limited_app,
        [(b"content-length", b"1")],
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


def test_engine_mcp_client_forwards_private_engine_token(monkeypatch):
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
    monkeypatch.setattr(engine_mcp_client.urllib.request, "urlopen", urlopen)

    engine_mcp_client.server_info(retries=0)

    assert len(requests) == 1
    request = requests[0]["request"]
    assert request.full_url == "http://engine-mcp:8085/mcp"
    assert request.get_header("Authorization") == "Bearer engine-secret"

    monkeypatch.setenv("LIKI_ENGINE_MCP_TOKEN", "")
    engine_mcp_client.server_info(retries=0)
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


def test_unauthorized_attempts_are_not_rate_limited(monkeypatch):
    """Rate limiting is owned by the edge; the app must never answer 429."""
    monkeypatch.setenv("LIKI_MCP_TOKEN", "test-secret")
    guarded_app = _make_app()
    payload = {"jsonrpc": "2.0", "id": 1, "method": "server/discover", "params": {}}

    for _ in range(5):
        status, _raw = _http_request(
            "/mcp", method="POST", payload=payload, asgi_app=guarded_app
        )
        assert status == 401


def test_engine_endpoint_requires_explicit_url(monkeypatch):
    monkeypatch.delenv("LIKI_MCP_URL", raising=False)
    with pytest.raises(EngineMCPError, match="LIKI_MCP_URL"):
        engine_mcp_client._endpoint()


def test_lifespan_fails_fast_without_engine_url(monkeypatch):
    """Startup must reject an empty LIKI_MCP_URL before touching any domain."""
    monkeypatch.delenv("LIKI_MCP_URL", raising=False)
    test_app = _make_app()

    async def enter_lifespan():
        async with test_app.router.lifespan_context(test_app):
            pass

    with pytest.raises(RuntimeError, match="LIKI_MCP_URL"):
        asyncio.run(enter_lifespan())
