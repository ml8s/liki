import asyncio
from contextlib import asynccontextmanager

from app.gateway import create_app
from httpx2 import ASGITransport, AsyncClient
from mcp.types import CallToolResult, ListToolsResult


def upstream_tools():
    engine = [
        "time_now",
        "subject_candidates",
        "city_info",
        "huangli_days",
        "bazi_chart",
        "ziwei_chart",
        "bazi_liupan",
        "ziwei_liupan",
        "bazi_bond",
        "ziwei_bond",
        "bazhai_chart",
        "bazhai_layout",
        "xuankong_chart",
        "xuankong_liunian",
        "qimen_chart",
        "liuyao_chart",
    ]
    counsel = [
        "bazi_yongshen",
        "natal_query",
        "natal_liunian_query",
        "naming_surname",
        "naming_pick",
        "naming_character",
        "naming_compose",
        "naming_check",
        "liuyao_ask",
        "qimen_ask",
    ]
    return {
        "engine": [
            {"name": name, "inputSchema": {"type": "object"}, "upstream": "engine"}
            for name in engine
        ],
        "counsel": [
            {"name": name, "inputSchema": {"type": "object"}, "upstream": "counsel"}
            for name in counsel
        ],
    }


def _list_tools_result(tools):
    return ListToolsResult(tools=tools)


def _call_tool_result(structured_content):
    return CallToolResult(content=[], structuredContent=structured_content)


def _call(payload):
    async def run():
        app = create_app()
        async with AsyncClient(
            transport=ASGITransport(app=app), base_url="http://test"
        ) as client:
            return await client.post("/mcp", json=payload)
    return asyncio.run(run())


def test_tools_list_filters_raw_divination(monkeypatch):
    requested = []

    class FakeSession:
        def __init__(self, upstream_name):
            self.upstream_name = upstream_name

        async def list_tools(self, params=None):
            requested.append(self.upstream_name)
            return _list_tools_result(upstream_tools()[self.upstream_name])

    @asynccontextmanager
    async def fake_upstream_session(config, **kwargs):
        yield FakeSession(config.name)

    monkeypatch.setattr("app.gateway.upstream_session", fake_upstream_session)
    response = _call({"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}})

    assert response.status_code == 200
    names = [tool["name"] for tool in response.json()["result"]["tools"]]
    assert names == [
        "time_now",
        "subject_candidates",
        "city_info",
        "huangli_days",
        "bazi_chart",
        "ziwei_chart",
        "bazi_liupan",
        "ziwei_liupan",
        "bazi_bond",
        "ziwei_bond",
        "bazhai_chart",
        "bazhai_layout",
        "xuankong_chart",
        "xuankong_liunian",
        "bazi_yongshen",
        "natal_query",
        "natal_liunian_query",
        "naming_surname",
        "naming_pick",
        "naming_character",
        "naming_compose",
        "naming_check",
        "liuyao_ask",
        "qimen_ask",
    ]


def test_tools_call_routes_without_retry_or_schema_change(monkeypatch):
    calls = []

    class FakeSession:
        async def call_tool(self, name, arguments=None):
            calls.append(("counsel", name, arguments))
            return _call_tool_result({"assertions": []})

    async def fake_upstream_session(config, **kwargs):
        yield FakeSession()

    fake_upstream_session = asynccontextmanager(fake_upstream_session)

    monkeypatch.setattr("app.gateway.upstream_session", fake_upstream_session)
    arguments = {"life_topics": ["career"]}
    response = _call({
        "jsonrpc": "2.0",
        "id": 7,
        "method": "tools/call",
        "params": {"name": "natal_query", "arguments": arguments},
    })

    assert response.status_code == 200
    assert calls == [("counsel", "natal_query", arguments)]
    assert response.json()["result"]["structuredContent"] == {"assertions": []}


def test_tools_call_denies_raw_divination(monkeypatch):
    async def fail_upstream_session(config, **kwargs):
        raise AssertionError("raw divination tools must not be routed")

    monkeypatch.setattr("app.gateway.upstream_session", fail_upstream_session)
    response = _call({
        "jsonrpc": "2.0",
        "id": 3,
        "method": "tools/call",
        "params": {"name": "qimen_chart", "arguments": {}},
    })
    assert response.status_code == 200
    assert response.json()["error"]["code"] == -32003


def test_initialize_passthrough_identity(monkeypatch):
    response = _call({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {"protocolVersion": "2025-06-18"},
    })
    assert response.status_code == 200
    result = response.json()["result"]
    assert result["protocolVersion"] == "2025-06-18"
    assert result["serverInfo"]["name"] == "hk.liki/liki"
    assert result["serverInfo"]["title"] == "Liki MCP"


def test_unsupported_modern_method_is_transport_not_found():
    response = _call({
        "jsonrpc": "2.0",
        "id": 9,
        "method": "subscriptions/listen",
        "params": {"notifications": {}},
    })

    assert response.status_code == 404
    assert response.json() == {
        "jsonrpc": "2.0",
        "id": 9,
        "error": {"code": -32601, "message": "method not found: subscriptions/listen"},
    }
