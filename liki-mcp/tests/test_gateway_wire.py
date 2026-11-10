"""契约回归：tools/list 线格式不得携带显式 null（Glama 严格 schema 拒绝）。

对齐官方 MCP Python SDK 的线序列化标准：
mcp/shared/peer.py —— model_dump(by_alias=True, mode="json", exclude_none=True)
"""

from contextlib import asynccontextmanager

from test_gateway import _call, _list_tools_result, upstream_tools


class FakeSession:
    def __init__(self, upstream_name):
        self.upstream_name = upstream_name

    async def list_tools(self, params=None):
        return _list_tools_result(upstream_tools()[self.upstream_name])


@asynccontextmanager
async def fake_upstream_session(config, **kwargs):
    yield FakeSession(config.name)


def test_tools_list_wire_has_no_null_fields(monkeypatch):
    monkeypatch.setattr("app.gateway.upstream_session", fake_upstream_session)
    response = _call({"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}})
    assert response.status_code == 200

    raw = response.text
    assert ": null" not in raw, "线格式携带显式 null，会被 Glama 等严格目录拒绝"

    tools = response.json()["result"]["tools"]
    assert tools, "应至少暴露一个工具"
    for tool in tools:
        assert "title" not in tool or isinstance(tool["title"], str), (
            f"{tool['name']}: title 必须省略或为字符串，不允许 null"
        )
