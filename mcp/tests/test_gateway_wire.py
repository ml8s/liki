"""契约回归：tools/list 线格式不得携带显式 null（对齐 SDK peer.py 序列化标准）。"""

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
    monkeypatch.setattr("app.tool_router.upstream_session", fake_upstream_session)
    response = _call({"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}})
    assert response.status_code == 200

    raw = response.text
    assert ": null" not in raw, "线格式携带显式 null，会被 Glama 等严格目录拒绝"

    tools = response.json()["result"]["tools"]
    assert tools, "应至少暴露一个工具"
