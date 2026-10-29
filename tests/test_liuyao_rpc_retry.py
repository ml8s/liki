"""Non-idempotent auto qigua must not be hidden behind generic MCP retries."""

import json
from urllib.error import HTTPError

import pytest


def _sse_result(payload: dict) -> bytes:
    message = json.dumps(
        {
            "jsonrpc": "2.0",
            "id": 1,
            "result": {
                "content": [
                    {
                        "type": "text",
                        "text": json.dumps(payload, ensure_ascii=False),
                    }
                ]
            },
        },
        ensure_ascii=False,
    )
    return f"event: message\ndata: {message}\n\n".encode()


class _Response:
    def __init__(self, payload: bytes):
        self._payload = payload
        self.status = 200
        self.reason = "OK"

    def read(self):
        return self._payload

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return False


def test_auto_qigua_is_not_retried(monkeypatch):
    import app.engine_mcp_client as client
    from app.engine_mcp_client import MCPError

    monkeypatch.setenv("LIKI_ENGINE_MCP_URL", "http://engine-mcp:8081/mcp")
    requests = []

    def transport(request, timeout):
        requests.append(request)
        raise HTTPError(request.full_url, 503, "Service Unavailable", {}, None)

    monkeypatch.setattr(client, "_HTTP_TRANSPORT", transport)

    with pytest.raises(MCPError, match="503"):
        client.call("liuyao.qigua", {"mode": "auto"}, retries=2)

    assert len(requests) == 1


def test_explicit_coin_qigua_can_retry(monkeypatch):
    import app.engine_mcp_client as client

    monkeypatch.setenv("LIKI_ENGINE_MCP_URL", "http://engine-mcp:8081/mcp")
    requests = []

    def transport(request, timeout):
        requests.append(request)
        if len(requests) == 1:
            raise HTTPError(request.full_url, 503, "Service Unavailable", {}, None)
        return _Response(_sse_result({"casting": {"mode": "coins"}}))

    monkeypatch.setattr(client, "_HTTP_TRANSPORT", transport)

    result = client.call(
        "liuyao.qigua",
        {"mode": "coins", "rounds": [["正", "反", "反"]] * 6},
        retries=1,
    )

    assert result == {"data": {"casting": {"mode": "coins"}}}
    assert len(requests) == 2
