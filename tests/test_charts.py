"""排盘层 RPC 重试契约：传输错误可重试，逻辑错误不重试。"""
import json
from unittest import mock
from urllib.error import HTTPError, URLError

import pytest
from app.natal.tools.charts import RPCError, call


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


def test_rpc_http_429_is_retried() -> None:
    success = json.dumps({"result": {"content": [{"type": "text", "text": json.dumps({"ok": True})}]}}).encode()
    ok_response = mock.MagicMock()
    ok_response.read.return_value = success
    ok_response.__enter__.return_value = ok_response
    ok_response.__exit__.return_value = False

    with mock.patch(
        "urllib.request.urlopen",
        side_effect=[HTTPError("liki", 429, "Too Many Requests", {}, None), ok_response],
    ) as urlopen:
        assert call("city.coords", {"city": "北京"}, retries=1)["data"] == {"ok": True}

    assert urlopen.call_count == 2
