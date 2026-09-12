"""排盘层 RPC 重试契约：传输错误可重试，逻辑错误不重试。"""
import json
from unittest import mock
from urllib.error import URLError

import pytest

import _helpers  # noqa: F401 —— 注入 tools 路径
import paipan
from paipan import RPCError, call


def test_rpc_logical_error_is_not_retried() -> None:
    payload = json.dumps({"error": {"message": "bad params"}}).encode()
    response = mock.MagicMock()
    response.read.return_value = payload
    response.__enter__.return_value = response
    response.__exit__.return_value = None

    with mock.patch("urllib.request.urlopen", return_value=response) as urlopen:
        with pytest.raises(RPCError, match="bad params"):
            call("bazi.chart", {}, retries=2)

    urlopen.assert_called_once()


def test_rpc_transport_error_is_retried() -> None:
    with mock.patch(
        "urllib.request.urlopen",
        side_effect=URLError("connection refused"),
    ) as urlopen:
        with pytest.raises(RPCError, match="connection refused"):
            call("bazi.chart", {}, retries=1)

    assert urlopen.call_count == 2


def test_engine_compatibility_uses_rpc_discover(monkeypatch) -> None:
    calls = []

    def fake_call(method, params, retries=1):
        calls.append((method, params, retries))
        return {"info": {"version": "2026.09.12.2"}}

    monkeypatch.setattr(paipan, "call", fake_call)
    paipan.ensure_engine_compatible()

    assert calls == [("rpc.discover", {"methods": "bazi.fullchart"}, 0)]


def test_engine_compatibility_rejects_old_engine(monkeypatch) -> None:
    monkeypatch.setattr(
        paipan, "call", lambda *_a, **_k: {"info": {"version": "2026.09.10.9"}}
    )

    with pytest.raises(RPCError, match="incompatible"):
        paipan.ensure_engine_compatible()
