"""排盘层 RPC 重试契约：传输错误可重试，逻辑错误不重试。"""
import json
from unittest import mock
from urllib.error import HTTPError, URLError

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


def test_rpc_http_429_is_retried() -> None:
    success = json.dumps({"result": {"data": {"ok": True}}}).encode()
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


def test_engine_compatibility_uses_rpc_discover(monkeypatch) -> None:
    calls = []

    def fake_call(method, params, retries=1):
        calls.append((method, params, retries))
        methods = [{"name": name} for name in paipan.REQUIRED_METHODS]
        return {"info": {"version": paipan.required_engine_version()}, "methods": methods}

    monkeypatch.setattr(paipan, "call", fake_call)
    paipan.ensure_engine_compatible()

    assert calls == [("rpc.discover", {"methods": ",".join(paipan.DISCOVER_SCOPES)}, 0)]


def test_engine_compatibility_rejects_old_engine(monkeypatch) -> None:
    methods = [{"name": name} for name in paipan.REQUIRED_METHODS]
    monkeypatch.setattr(paipan, "call", lambda *_a, **_k: {
        "info": {"version": "2026.09.10.9"}, "methods": methods
    })

    with pytest.raises(RPCError, match="incompatible"):
        paipan.ensure_engine_compatible()


def test_engine_compatibility_requires_installed_skill_version(monkeypatch) -> None:
    methods = [{"name": name} for name in paipan.REQUIRED_METHODS]
    required = paipan.required_engine_version()
    older = required.split(".")[:-1] + [str(int(required.split(".")[-1]) - 1)]
    monkeypatch.setattr(paipan, "call", lambda *_a, **_k: {
        "info": {"version": ".".join(older)}, "methods": methods
    })

    with pytest.raises(RPCError, match="skill VERSION requires engine"):
        paipan.ensure_engine_compatible()


def test_boundary_hint_names_current_and_alternate_shichen() -> None:
    """交界提示必须自带「当前时辰 + 跨后时辰」，调用方不做二次推断。"""
    hint = paipan._shichen_boundary_hint("1981-08-26T00:57:00+08:00")

    assert hint is not None
    assert hint["current_shichen"] == {
        "name": "子时", "branch": "子", "span": "23:00-01:00",
    }
    assert hint["alternate_shichen"]["name"] == "丑时"
    assert hint["direction"] == "later"
    assert hint["boundary_offset_minutes"] == -3
    assert hint["minutes_to_boundary"] == 3


def test_boundary_hint_mirrors_offset_sign_when_past_boundary() -> None:
    hint = paipan._shichen_boundary_hint("1981-08-26T01:03:00+08:00")

    assert hint is not None
    assert hint["current_shichen"]["name"] == "丑时"
    assert hint["alternate_shichen"]["name"] == "子时"
    assert hint["direction"] == "earlier"
    assert hint["boundary_offset_minutes"] == 3


def test_boundary_hint_wraps_to_previous_window_across_zi_shi() -> None:
    """子时是首个窗口——往回跨须环绕到亥时，不得越界或错位。"""
    hint = paipan._shichen_boundary_hint("1981-08-26T23:02:00+08:00")

    assert hint is not None
    assert hint["current_shichen"]["name"] == "子时"
    assert hint["alternate_shichen"] == {
        "name": "亥时", "branch": "亥", "span": "21:00-23:00",
    }
    assert hint["direction"] == "earlier"


def test_boundary_hint_maps_non_zi_shi_window() -> None:
    """非子时场景同样成立（防止只在子时正确）。"""
    hint = paipan._shichen_boundary_hint("1981-08-26T12:35:00+08:00")

    assert hint is not None
    assert hint["current_shichen"]["name"] == "午时"
    assert hint["current_shichen"]["span"] == "11:00-13:00"
    assert hint["alternate_shichen"]["name"] == "未时"
    assert hint["direction"] == "later"


def test_boundary_hint_carries_no_day_change_claim() -> None:
    """换日口径不属本层——提示只给窗口，不得夹带早/晚子时之类的换日立场。"""
    hint = paipan._shichen_boundary_hint("1981-08-26T23:02:00+08:00")

    assert hint is not None
    assert "day_part" not in hint["current_shichen"]
    assert "day_part" not in hint["alternate_shichen"]


def test_boundary_hint_is_silent_far_from_boundary_and_on_bad_input() -> None:
    assert paipan._shichen_boundary_hint("1981-08-26T00:15:00+08:00") is None
    assert paipan._shichen_boundary_hint("1981-08-26T12:28:00+08:00") is None
    assert paipan._shichen_boundary_hint("not-a-time") is None
