"""排盘层 RPC 重试契约：传输错误可重试，逻辑错误不重试。"""
import json
from unittest import mock
from urllib.error import HTTPError, URLError

import app.natal.tools.charts as charts
import pytest
from app.natal.tools.charts import RPCError, call


@pytest.fixture(autouse=True)
def skip_shared_engine_compatibility(monkeypatch):
    """Pure transport tests exercise call(); the shared gate has its own tests."""
    import app.engine_mcp_client as engine_mcp_client
    monkeypatch.setattr(engine_mcp_client, "_COMPATIBILITY_CHECKED", True)


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


def test_engine_compatibility_reads_engine_version(monkeypatch) -> None:
    calls = []

    def fake_engine_version():
        calls.append(1)
        return charts.required_engine_version()

    monkeypatch.setattr(charts, "engine_version", fake_engine_version)
    charts.ensure_engine_compatible()

    assert calls == [1]


def test_engine_compatibility_rejects_old_engine(monkeypatch) -> None:
    monkeypatch.setattr(charts, "engine_version", lambda: "2026.09.10.9")

    with pytest.raises(RPCError, match="incompatible"):
        charts.ensure_engine_compatible()


def test_engine_compatibility_requires_installed_skill_version(monkeypatch) -> None:
    required = charts.required_engine_version()
    older = required.split(".")[:-1] + [str(int(required.split(".")[-1]) - 1)]
    monkeypatch.setattr(charts, "engine_version", lambda: ".".join(older))

    with pytest.raises(RPCError, match="VERSION.txt requires engine"):
        charts.ensure_engine_compatible()
