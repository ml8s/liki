"""排盘层 RPC 重试契约：传输错误可重试，逻辑错误不重试。"""
import json
from datetime import datetime
from zoneinfo import ZoneInfo
from unittest import mock
from urllib.error import HTTPError, URLError

import pytest

import _helpers  # noqa: F401 —— 注入 tools 路径
from factor_constants import load_constants
import paipan
from paipan import RPCError, call


@pytest.fixture(autouse=True)
def skip_shared_engine_compatibility(monkeypatch):
    """Pure transport tests exercise call(); the shared gate has its own tests."""
    import engine_client

    monkeypatch.setattr(engine_client, "_COMPATIBILITY_CHECKED", True)


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
        return paipan.required_engine_version()

    monkeypatch.setattr(paipan, "engine_version", fake_engine_version)
    paipan.ensure_engine_compatible()

    assert calls == [1]


def test_engine_compatibility_rejects_old_engine(monkeypatch) -> None:
    monkeypatch.setattr(paipan, "engine_version", lambda: "2026.09.10.9")

    with pytest.raises(RPCError, match="incompatible"):
        paipan.ensure_engine_compatible()


def test_engine_compatibility_requires_installed_skill_version(monkeypatch) -> None:
    required = paipan.required_engine_version()
    older = required.split(".")[:-1] + [str(int(required.split(".")[-1]) - 1)]
    monkeypatch.setattr(paipan, "engine_version", lambda: ".".join(older))

    with pytest.raises(RPCError, match="VERSION.txt requires engine"):
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
    hint = paipan._shichen_boundary_hint("1981-08-26T12:57:00+08:00")

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


@pytest.mark.parametrize("index", range(12))
def test_boundary_hint_covers_every_shichen_boundary(index: int) -> None:
    """12 个交界都要验证前/后侧窗口；不得只覆盖子时或午时等代表点。"""
    zhi = load_constants()["地支"]
    boundaries = paipan.SHICHEN_BOUNDARY_START_HOURS
    start_hour = boundaries[index]
    previous = (index - 1) % len(boundaries)

    before = datetime(
        1981, 8, 25, (start_hour - 1) % 24, 59, tzinfo=ZoneInfo("Asia/Shanghai")
    )

    hint = paipan._shichen_boundary_hint(before.isoformat())
    assert hint is not None
    assert hint["current_shichen"]["branch"] == zhi[previous]
    assert hint["alternate_shichen"]["branch"] == zhi[index]
    assert hint["direction"] == "later"
    assert hint["boundary_offset_minutes"] == -1

    after = datetime(
        1981, 8, 25, start_hour, 1, tzinfo=ZoneInfo("Asia/Shanghai")
    )
    hint = paipan._shichen_boundary_hint(after.isoformat())
    assert hint is not None
    assert hint["current_shichen"]["branch"] == zhi[index]
    assert hint["alternate_shichen"]["branch"] == zhi[previous]
    assert hint["direction"] == "earlier"
    assert hint["boundary_offset_minutes"] == 1


@pytest.mark.parametrize(
    ("time", "expect_hint"),
    [
        ("1981-08-26T00:53:00+08:00", True),   # 距 01:00 边界 7 分钟 < 8 → 提示
        ("1981-08-26T00:52:00+08:00", True),   # 距边界恰 8 分钟（>8 才沉默）→ 提示
        ("1981-08-26T00:51:00+08:00", False),  # 距边界 9 分钟 > 8 → 沉默
    ],
)
def test_boundary_hint_threshold_is_eight_minutes(time: str, expect_hint: bool) -> None:
    """临界阈值 8 分钟：<=8 触发提示，>8 沉默（锁边界防回归）。"""
    hint = paipan._shichen_boundary_hint(time)
    if expect_hint:
        assert hint is not None, f"{time} 应在临界窗口内触发提示"
    else:
        assert hint is None, f"{time} 不应触发提示"
